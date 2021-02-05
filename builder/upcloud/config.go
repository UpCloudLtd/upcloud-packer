package upcloud

import (
	"errors"
	"time"

	"github.com/hashicorp/packer-plugin-sdk/common"
	"github.com/hashicorp/packer-plugin-sdk/communicator"
	"github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/template/config"
	"github.com/hashicorp/packer-plugin-sdk/template/interpolate"
)

type Config struct {
	common.PackerConfig `mapstructure:",squash"`
	Comm                communicator.Config `mapstructure:",squash"`

	// Required configuration values
	Username       string `mapstructure:"username"`
	Password       string `mapstructure:"password"`
	Zone           string `mapstructure:"zone"`
	StorageUUID    string `mapstructure:"storage_uuid"`
	TemplatePrefix string `mapstructure:"template_prefix"`

	// Optional configuration values
	StorageSize int `mapstructure:"storage_size"`

	Timeout time.Duration
	ctx     interpolate.Context
}

func (c *Config) Prepare(raws ...interface{}) ([]string, error) {
	err := config.Decode(c, &config.DecodeOpts{
		Interpolate:        true,
		InterpolateContext: &c.ctx,
	}, raws...)

	if err != nil {
		return nil, err
	}

	// validate
	var errs *packer.MultiError
	if es := c.Comm.Prepare(&c.ctx); len(es) > 0 {
		errs = packer.MultiErrorAppend(errs, es...)
	}

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

	c.Timeout = 5 * time.Minute
	return nil, nil
}
