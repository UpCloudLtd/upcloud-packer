package upcloud

import (
	"errors"
	"fmt"
	"github.com/jalle19/upcloud-go-sdk/upcloud"
	"github.com/mitchellh/packer/common"
	"github.com/mitchellh/packer/helper/communicator"
	"github.com/mitchellh/packer/helper/config"
	"github.com/mitchellh/packer/packer"
	"github.com/mitchellh/packer/template/interpolate"
	"time"
)

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

// NewConfig creates a new configuration, setting default values and validating it along the way
func NewConfig(raws ...interface{}) (*Config, error) {
	c := new(Config)

	err := config.Decode(c, &config.DecodeOpts{
		Interpolate:        true,
		InterpolateContext: &c.ctx,
	}, raws...)

	if err != nil {
		return nil, err
	}

	// Assign default values if possible
	if c.Title == "" {
		c.Title = fmt.Sprintf("packer-builder-upcloud-%d", time.Now().Unix())
	}

	if c.Hostname == "" {
		c.Hostname = fmt.Sprintf("%s.example.com", c.Title)
	}

	if c.StorageSize == 0 {
		c.StorageSize = 30
	}

	if c.StorageTitle == "" {
		c.StorageTitle = fmt.Sprintf("%s-disk1", c.Title)
	}

	if c.StorageTier == "" {
		c.StorageTier = upcloud.StorageTierMaxIOPS
	}

	if c.Comm.SSHUsername == "" {
		c.Comm.SSHUsername = "root"
	}

	if c.RawStateTimeoutDuration == "" {
		c.RawStateTimeoutDuration = "5m"
	}

	// Validation
	var errs *packer.MultiError
	errs = packer.MultiErrorAppend(errs, c.Comm.Prepare(&c.ctx)...)

	// Check for required configurations that will display errors if not set
	if c.Username == "" {
		errs = packer.MultiErrorAppend(
			errs, errors.New("\"username\" must be specified"))
	}

	if c.Password == "" {
		errs = packer.MultiErrorAppend(
			errs, errors.New("\"password\" must be specified"))
	}

	if c.Plan == "" && (c.CoreNumber == 0 || c.MemoryAmount == 0) {
		errs = packer.MultiErrorAppend(
			errs, errors.New("\"core_number\" and \"memory_amount\" must be specified if \"plan\" is not specified"))
	}

	if c.Plan != "" && (c.CoreNumber > 0 || c.MemoryAmount > 0) {
		errs = packer.MultiErrorAppend(
			errs, errors.New("\"core_number\" and \"memory_amount\" must not be specified when \"plan\" is specified"))
	}

	if c.Zone == "" {
		errs = packer.MultiErrorAppend(
			errs, errors.New("\"zone\" must be specified"))
	}

	if c.StorageUUID == "" {
		errs = packer.MultiErrorAppend(
			errs, errors.New("\"storage_uuid\" must be specified"))
	}

	stateTimeout, err := time.ParseDuration(c.RawStateTimeoutDuration)
	if err != nil {
		errs = packer.MultiErrorAppend(
			errs, fmt.Errorf("Failed to parse state_timeout_duration: %s", err))
	}
	c.StateTimeoutDuration = stateTimeout

	if len(errs.Errors) > 0 {
		return nil, errors.New(errs.Error())
	}

	return c, nil
}
