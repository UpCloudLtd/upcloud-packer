package upcloud

import (
	"fmt"
	"github.com/jalle19/upcloud-go-sdk/upcloud"
	"github.com/jalle19/upcloud-go-sdk/upcloud/client"
	"github.com/jalle19/upcloud-go-sdk/upcloud/service"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/common"
	"github.com/mitchellh/packer/helper/communicator"
	"github.com/mitchellh/packer/packer"
	"log"
)

// The unique ID for this builder.
const BuilderId = "jalle19.upcloud"

// Builder represents a Packer Builder.
type Builder struct {
	config *Config
	runner multistep.Runner
}

// Prepare processes the build configuration parameters and validates the configuration
func (self *Builder) Prepare(raws ...interface{}) (parms []string, err error) {
	self.config, err = NewConfig(raws...)

	if err != nil {
		return nil, err
	}

	log.Println(common.ScrubConfig(self.config, self.config.Password, self.config.Username))

	return nil, nil
}

// Run executes the actual build steps
func (self *Builder) Run(ui packer.Ui, hook packer.Hook, cache packer.Cache) (packer.Artifact, error) {
	// Create the service
	client := client.New(self.config.Username, self.config.Password)
	service := service.New(client)

	// Set up the state which is used to share state between the steps
	state := new(multistep.BasicStateBag)
	state.Put("config", *self.config)
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
