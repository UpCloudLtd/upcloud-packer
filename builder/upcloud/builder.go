package upcloud

import (
	"errors"
	"fmt"
	"github.com/jalle19/upcloud-go-sdk/upcloud"
	"github.com/jalle19/upcloud-go-sdk/upcloud/client"
	"github.com/jalle19/upcloud-go-sdk/upcloud/service"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/common"
	"github.com/mitchellh/packer/helper/communicator"
	"github.com/mitchellh/packer/helper/config"
	"github.com/mitchellh/packer/packer"
	"github.com/mitchellh/packer/template/interpolate"
	"log"
	"time"
)

// The unique ID for this builder.
const BuilderId = "jalle19.upcloud"

// Config represents the configuration for this builder
type Config struct {
	common.PackerConfig `mapstructure:",squash"`
	Comm                communicator.Config `mapstructure:",squash"`

	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`

	Plan         string `mapstructure:"plan"`
	CoreNumber   int    `mapstructure:"core_number"`
	MemoryAmount int    `mapstructure:"memory_amount"`
	Title        string `mapstructure:"title"`
	Hostname     string `mapstructure:"hostname"`
	Zone         string `mapstructure:"zone"`
	StorageUUID  string `mapstructure:"storage_uuid"`
	StorageTitle string `mapstructure:"storage_title"`
	StorageSize  int    `mapstructure:"storage_size"`
	StorageTier  string `mapstructure:"storage_tier"`

	RawStateTimeoutDuration string `mapstructure:"state_timeout_duration"`
	StateTimeoutDuration    time.Duration

	ctx interpolate.Context
}

// Builder represents a Packer Builder.
type Builder struct {
	config Config
	runner multistep.Runner
}

// Prepare processes the build configuration parameters and validates the configuration
func (self *Builder) Prepare(raws ...interface{}) (parms []string, retErr error) {
	// Parse configuration
	err := config.Decode(&self.config, &config.DecodeOpts{
		Interpolate:        true,
		InterpolateContext: &self.config.ctx,
	}, raws...)

	if err != nil {
		return nil, err
	}

	// Assign default values if possible
	if self.config.Title == "" {
		self.config.Title = fmt.Sprintf("packer-builder-upcloud-%d", time.Now().Unix())
	}

	if self.config.Hostname == "" {
		self.config.Hostname = fmt.Sprintf("%s.example.com", self.config.Title)
	}

	if self.config.StorageSize == 0 {
		self.config.StorageSize = 30
	}

	if self.config.StorageTitle == "" {
		self.config.StorageTitle = fmt.Sprintf("%s-disk1", self.config.Title)
	}

	if self.config.StorageTier == "" {
		self.config.StorageTier = upcloud.StorageTierMaxIOPS
	}

	if self.config.Comm.SSHUsername == "" {
		self.config.Comm.SSHUsername = "root"
	}

	if self.config.RawStateTimeoutDuration == "" {
		self.config.RawStateTimeoutDuration = "5m"
	}

	// Validation
	var errs *packer.MultiError
	errs = packer.MultiErrorAppend(errs, self.config.Comm.Prepare(&self.config.ctx)...)

	// Check for required configurations that will display errors if not set
	if self.config.Username == "" {
		errs = packer.MultiErrorAppend(
			errs, errors.New("\"username\" must be specified"))
	}

	if self.config.Password == "" {
		errs = packer.MultiErrorAppend(
			errs, errors.New("\"password\" must be specified"))
	}

	if self.config.Plan == "" && (self.config.CoreNumber == 0 || self.config.MemoryAmount == 0) {
		errs = packer.MultiErrorAppend(
			errs, errors.New("\"core_number\" and \"memory_amount\" must be specified if \"plan\" is not specified"))
	}

	if self.config.Plan != "" && (self.config.CoreNumber > 0 || self.config.MemoryAmount > 0) {
		errs = packer.MultiErrorAppend(
			errs, errors.New("\"core_number\" and \"memory_amount\" must not be specified when \"plan\" is specified"))
	}

	if self.config.Zone == "" {
		errs = packer.MultiErrorAppend(
			errs, errors.New("\"zone\" must be specified"))
	}

	if self.config.StorageUUID == "" {
		errs = packer.MultiErrorAppend(
			errs, errors.New("\"storage_uuid\" must be specified"))
	}

	stateTimeout, err := time.ParseDuration(self.config.RawStateTimeoutDuration)
	if err != nil {
		errs = packer.MultiErrorAppend(
			errs, fmt.Errorf("Failed to parse state_timeout_duration: %s", err))
	}
	self.config.StateTimeoutDuration = stateTimeout

	log.Println(common.ScrubConfig(self.config, self.config.Password, self.config.Username))

	if len(errs.Errors) > 0 {
		retErr = errors.New(errs.Error())
	}

	return nil, retErr
}

// Run executes the actual build steps
func (self *Builder) Run(ui packer.Ui, hook packer.Hook, cache packer.Cache) (packer.Artifact, error) {
	// Create the service
	client := client.New(self.config.Username, self.config.Password)
	service := service.New(client)

	// Set up the state which is used to share state between the steps
	state := new(multistep.BasicStateBag)
	state.Put("config", self.config)
	state.Put("service", *service)
	state.Put("hook", hook)
	state.Put("ui", ui)

	// Build the steps
	steps := []multistep.Step{
		&StepCreateSSHKey{
			Debug:        self.config.PackerDebug,
			DebugKeyPath: fmt.Sprintf("packer-builder-upcloud-%s.pem", self.config.PackerBuildName),
		},
		new(StepCreateServer),
		&communicator.StepConnect{
			Config:    &self.config.Comm,
			Host:      sshHostCallback,
			SSHConfig: sshConfigCallback,
		},
		new(common.StepProvision),
		new(StepTemplatizeStorage),
	}

	// Create the runner which will run the steps we just build
	self.runner = &multistep.BasicRunner{Steps: steps}
	self.runner.Run(state)

	if rawErr, ok := state.GetOk("error"); ok {
		return nil, rawErr.(error)
	}

	// Extract the final storage details from the state
	rawDetails, ok := state.GetOk("storage_details")

	if !ok {
		log.Println("No storage details found in state, the build was probably cancelled")
		return nil, nil
	}

	storageDetails := rawDetails.(*upcloud.StorageDetails)

	// Create an artifact and return it
	artifact := &Artifact{
		UUID:    storageDetails.UUID,
		Zone:    storageDetails.Zone,
		Title:   storageDetails.Title,
		service: service,
	}

	return artifact, nil
}

// Cancel is called when the build is cancelled
func (self *Builder) Cancel() {
	if self.runner != nil {
		log.Println("Cancelling the step runner ...")
		self.runner.Cancel()
	}

	fmt.Println("Cancelling the builder ...")
}
