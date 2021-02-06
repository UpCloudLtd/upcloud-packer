package upcloud

import (
	"errors"
	"os"
	"time"

	"github.com/hashicorp/packer-plugin-sdk/common"
	"github.com/hashicorp/packer-plugin-sdk/communicator"
	"github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/template/config"
	"github.com/hashicorp/packer-plugin-sdk/template/interpolate"
)

const (
	DefaultImageName   = "custom-image"
	DefaultStorageSize = 30
	DefaultTimeout     = 5 * time.Minute
)

type Config struct {
	common.PackerConfig `mapstructure:",squash"`
	Comm                communicator.Config `mapstructure:",squash"`

	// Required configuration values
	Username     string `mapstructure:"username"`
	Password     string `mapstructure:"password"`
	Zone         string `mapstructure:"zone"`
	TemplateUUID string `mapstructure:"template_uuid"`
	TemplateName string `mapstructure:"template_name"`

	// Optional configuration values
	ImageName   string        `mapstructure:"image_name"`
	StorageSize int           `mapstructure:"storage_size"`
	Timeout     time.Duration `mapstructure:"timeout"`

	// deprecated, left for backward compatibility
	StorageUUID    string `mapstructure:"storage_uuid"`
	TemplatePrefix string `mapstructure:"template_prefix"`

	ctx interpolate.Context
}

func (c *Config) Prepare(raws ...interface{}) ([]string, error) {
	err := config.Decode(c, &config.DecodeOpts{
		Interpolate:        true,
		InterpolateContext: &c.ctx,
	}, raws...)

	if err != nil {
		return nil, err
	}

	c.setEnv()
	c.setDeprecated()

	// validate
	var errs *packer.MultiError
	if es := c.Comm.Prepare(&c.ctx); len(es) > 0 {
		errs = packer.MultiErrorAppend(errs, es...)
	}

	if c.Username == "" {
		errs = packer.MultiErrorAppend(
			errs, errors.New("'username' must be specified"),
		)
	}

	if c.Password == "" {
		errs = packer.MultiErrorAppend(
			errs, errors.New("'password' must be specified"),
		)
	}

	if c.Zone == "" {
		errs = packer.MultiErrorAppend(
			errs, errors.New("'zone' must be specified"),
		)
	}

	if c.TemplateUUID == "" && c.TemplateName == "" {
		errs = packer.MultiErrorAppend(
			errs, errors.New("'template_uuid' or `template_name` must be specified"),
		)
	}

	if errs != nil && len(errs.Errors) > 0 {
		return nil, errs
	}

	// defaults
	if c.ImageName == "" {
		c.ImageName = DefaultImageName
	}

	if c.StorageSize == 0 {
		c.StorageSize = DefaultStorageSize
	}

	if c.Timeout == 0 {
		c.Timeout = DefaultTimeout
	}
	return nil, nil
}

// get params from environment
func (c *Config) setEnv() {
	username := os.Getenv("UPCLOUD_API_USER")
	if username != "" && c.Username == "" {
		c.Username = username
	}

	password := os.Getenv("UPCLOUD_API_PASSWORD")
	if password != "" && c.Password == "" {
		c.Password = password
	}
}

// set deprecated params if exists
func (c *Config) setDeprecated() {
	if c.StorageUUID != "" {
		c.TemplateUUID = c.StorageUUID
	}
	if c.TemplatePrefix != "" {
		c.ImageName = c.TemplatePrefix
	}
}
