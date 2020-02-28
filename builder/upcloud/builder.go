package upcloud

import (
	"context"
	"fmt"
	"github.com/UpCloudLtd/upcloud-go-api/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/upcloud/request"
	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/hashicorp/packer/common"
	"github.com/hashicorp/packer/helper/communicator"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"log"
)

// The unique ID for this builder.
const BuilderId = "upcloudltd.upcloud"

// Builder represents a Packer Builder.
type Builder struct {
	config *Config
	runner multistep.Runner
}

func (b *Builder) ConfigSpec() hcldec.ObjectSpec {
	return b.config.FlatMapstructure().HCL2Spec()
}

// Prepare processes the build configuration parameters and validates the configuration
func (b *Builder) Prepare(raws ...interface{}) ([]string, []string, error) {
	var err error
	// Parse and create the configuration
	b.config, err = NewConfig(raws...)

	if err != nil {
		return nil, nil, err
	}

	// Check that the client/service is usable
	service := b.config.GetService()

	if _, err := service.GetAccount(); err != nil {
		return nil, nil, err
	}

	// Check that the specified storage device is a template
	storageDetails, err := service.GetStorageDetails(&request.GetStorageDetailsRequest{
		UUID: b.config.StorageUUID,
	})
	if err != nil {
		return nil, nil, err
	}

	if storageDetails.Type != upcloud.StorageTypeTemplate {
		return nil, nil, fmt.Errorf("The specified storage UUID is of invalid type \"%s\"", storageDetails.Type)
	}

	return nil, nil, nil
}

// Run executes the actual build steps
func (b *Builder) Run(ctx context.Context, ui packer.Ui, hook packer.Hook) (packer.Artifact, error) {
	// Create the service
	service := b.config.GetService()

	// Set up the state which is used to share state between the steps
	state := new(multistep.BasicStateBag)
	state.Put("config", *b.config)
	state.Put("service", *service)
	state.Put("hook", hook)
	state.Put("ui", ui)

	// Build the steps
	steps := []multistep.Step{
		&StepCreateSSHKey{
			Debug:        b.config.PackerDebug,
			DebugKeyPath: fmt.Sprintf("packer-builder-upcloud-%s.pem", b.config.PackerBuildName),
		},
		new(StepCreateServer),
		&communicator.StepConnect{
			Config:    &b.config.Comm,
			Host:      sshHostCallback,
			SSHConfig: sshConfigCallback,
		},
		new(common.StepProvision),
		new(StepTemplatizeStorage),
	}

	// Create the runner which will run the steps we just build
	b.runner = &multistep.BasicRunner{Steps: steps}
	b.runner.Run(ctx, state)

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
