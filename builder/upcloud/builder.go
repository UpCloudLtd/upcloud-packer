package upcloud

import (
	"context"
	"fmt"

	"github.com/UpCloudLtd/upcloud-go-api/upcloud"
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
	driver Driver
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
	b.driver = NewDriver(&b.config)

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
			GeneratedData: generatedData,
		},
		&communicator.StepConnect{
			Config:    &b.config.Comm,
			Host:      sshHostCallback,
			SSHConfig: b.config.Comm.SSHConfigFunc(),
		},
		&commonsteps.StepProvision{},
		&commonsteps.StepCleanupTempKeys{
			Comm: &b.config.Comm,
		},
		&StepTeardownServer{},
		&StepCreateTemplate{
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

	template, ok := state.GetOk("template")
	if !ok {
		return nil, fmt.Errorf("Failed to find 'template' in state")
	}

	artifact := &Artifact{
		Template:  template.(*upcloud.Storage),
		config:    &b.config,
		driver:    b.driver,
		StateData: map[string]interface{}{"generated_data": state.Get("generated_data")},
	}
	return artifact, nil
}
