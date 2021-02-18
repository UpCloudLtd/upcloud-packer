package upcloud

import (
	"context"
	"fmt"

	"github.com/UpCloudLtd/upcloud-go-api/upcloud"
	internal "github.com/UpCloudLtd/upcloud-packer/internal"
	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/hashicorp/packer-plugin-sdk/communicator"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/multistep/commonsteps"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/packerbuilderdata"
)

const BuilderId = "upcloud.builder"

type Builder struct {
	config Config
	runner multistep.Runner
	driver internal.Driver
}

func (b *Builder) ConfigSpec() hcldec.ObjectSpec { return b.config.FlatMapstructure().HCL2Spec() }

func (b *Builder) Prepare(raws ...interface{}) (generatedVars []string, warnings []string, err error) {
	warnings, errs := b.config.Prepare(raws...)
	if errs != nil {
		return nil, warnings, errs
	}

	buildGeneratedData := []string{
		"ServerUUID",
		"ServerTitle",
		"ServerSize",
		"TemplateUUID",
		"TemplateTitle",
		"TemplateSize",
	}
	return buildGeneratedData, nil, nil
}

func (b *Builder) Run(ctx context.Context, ui packersdk.Ui, hook packersdk.Hook) (packersdk.Artifact, error) {
	// Setup the state bag and initial state for the steps
	b.driver = internal.NewDriver(&internal.DriverConfig{
		Username:    b.config.Username,
		Password:    b.config.Password,
		Timeout:     b.config.Timeout,
		SSHUsername: b.config.Comm.SSHUsername,
	})

	state := new(multistep.BasicStateBag)
	state.Put("config", &b.config)
	state.Put("hook", hook)
	state.Put("ui", ui)
	state.Put("driver", b.driver)

	generatedData := &packerbuilderdata.GeneratedData{State: state}

	// Build the steps
	steps := []multistep.Step{
		&StepCreateSSHKey{
			Debug:        b.config.PackerDebug,
			DebugKeyPath: fmt.Sprintf("ssh_key-%s.pem", b.config.PackerBuildName),
		},
		&StepCreateServer{
			Config:        &b.config,
			GeneratedData: generatedData,
		},
		&communicator.StepConnect{
			Config:    &b.config.Comm,
			Host:      internal.SshHostCallback,
			SSHConfig: b.config.Comm.SSHConfigFunc(),
		},
		&commonsteps.StepProvision{},
		&commonsteps.StepCleanupTempKeys{
			Comm: &b.config.Comm,
		},
		&StepTeardownServer{},
		&StepCreateTemplate{
			Config:        &b.config,
			GeneratedData: generatedData,
		},
	}

	// Run
	b.runner = commonsteps.NewRunner(steps, b.config.PackerConfig, ui)
	b.runner.Run(ctx, state)

	// If there was an error, return that
	if err, ok := state.GetOk("error"); ok {
		return nil, err.(error)
	}

	templates, ok := state.GetOk("templates")
	if !ok {
		return nil, fmt.Errorf("No template found in state, the build was probably cancelled")
	}

	artifact := &Artifact{
		Templates: templates.([]*upcloud.Storage),
		config:    &b.config,
		driver:    b.driver,
		StateData: map[string]interface{}{
			"generated_data":  state.Get("generated_data"),
			"template_prefix": b.config.TemplatePrefix,
		},
	}
	return artifact, nil
}
