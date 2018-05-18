package upcloud

import (
	"errors"
	"fmt"
	"github.com/UpCloudLtd/upcloud-go-sdk/upcloud/client"
	"github.com/UpCloudLtd/upcloud-go-sdk/upcloud/service"
	"github.com/mitchellh/packer/common"
	"github.com/mitchellh/packer/helper/communicator"
	"github.com/mitchellh/packer/helper/config"
	"github.com/mitchellh/packer/packer"
	"github.com/mitchellh/packer/template/interpolate"
	"time"
	"github.com/UpCloudLtd/upcloud-go-sdk/upcloud/request"
	"github.com/UpCloudLtd/upcloud-go-sdk/upcloud"
)

// Config represents the configuration for this builder
type Config struct {
	common.PackerConfig      `mapstructure:",squash"`
	Comm communicator.Config `mapstructure:",squash"`

	// Required configuration values
	Username       string `mapstructure:"username"`
	Password       string `mapstructure:"password"`
	Zone           string `mapstructure:"zone"`
	StorageUUID    string `mapstructure:"storage_uuid"`
	TemplatePrefix string `mapstructure:"template_prefix"`

	// Optional configuration values
	StorageSize             int    `mapstructure:"storage_size"`
	RawStateTimeoutDuration string `mapstructure:"state_timeout_duration"`
	IPs                     []IP   `mapstructure:"ips"`

	StateTimeoutDuration time.Duration
	ctx                  interpolate.Context
}

type IP struct {
	Access string `mapstructure:"access"`
	Family string `mapstructure:"family"`
}

func (c *Config) GetIPAddresses() (ips []request.CreateServerIPAddress) {

	if len(c.IPs) == 0 {
		// default to standard setup for backward compatibility
		return []request.CreateServerIPAddress{
			{Access: upcloud.IPAddressAccessPrivate, Family: upcloud.IPAddressFamilyIPv4,},
			{Access: upcloud.IPAddressAccessPublic, Family: upcloud.IPAddressFamilyIPv4,},
			{Access: upcloud.IPAddressAccessPublic, Family: upcloud.IPAddressFamilyIPv6,},
		}
	}

	for _, ip := range c.IPs {
		ips = append(ips, request.CreateServerIPAddress{
			Access: ip.Access,
			Family: ip.Family,
		})
	}

	return ips
}

// GetService returns a service object using the credentials specified in the configuration
func (c *Config) GetService() *service.Service {
	client := client.New(c.Username, c.Password)
	return service.New(client)
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
	c.Comm.SSHUsername = "root"

	if c.StorageSize == 0 {
		c.StorageSize = 30
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
