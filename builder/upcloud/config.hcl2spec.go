package upcloud

import (
	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/zclconf/go-cty/cty"
)

type FlatConfig struct {
	Username       string `mapstructure:"username"`
	Password       string `mapstructure:"password"`
	Zone           string `mapstructure:"zone"`
	StorageUUID    string `mapstructure:"storage_uuid"`
	TemplatePrefix string `mapstructure:"template_prefix"`

	// Optional configuration values
	StorageSize             int    `mapstructure:"storage_size"`
	RawStateTimeoutDuration string `mapstructure:"state_timeout_duration"`
}

// FlatMapstructure returns a new FlatConfig.
// FlatConfig is an auto-generated flat version of Config.
// Where the contents a fields with a `mapstructure:,squash` tag are bubbled up.
func (*Config) FlatMapstructure() interface{ HCL2Spec() map[string]hcldec.Spec } {
	return new(FlatConfig)
}

// HCL2Spec returns the hcl spec of a Config.
// This spec is used by HCL to read the fields of Config.
// The decoded values from this spec will then be applied to a FlatConfig.
func (*FlatConfig) HCL2Spec() map[string]hcldec.Spec {
	s := map[string]hcldec.Spec{
		"username": &hcldec.AttrSpec{Name: "username", Type: cty.String, Required: false},
		"password": &hcldec.AttrSpec{Name: "password", Type: cty.String, Required: false},
		"zone": &hcldec.AttrSpec{Name: "zone", Type: cty.String, Required: true},
		"storage_uuid": &hcldec.AttrSpec{Name: "storage_uuid", Type: cty.String, Required: true},
		"template_prefix": &hcldec.AttrSpec{Name: "template_prefix", Type: cty.String, Required: false},
		"storage_size": &hcldec.AttrSpec{Name: "storage_size", Type: cty.Number, Required: false},
		"state_timeout_duration": &hcldec.AttrSpec{Name: "state_timeout_duration", Type: cty.String, Required: false},
	}
	return s
}
