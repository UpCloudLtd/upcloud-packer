package upcloud

import (
	"context"
	"fmt"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/packer"
)

// StepCreateImage represents the step that creates a storage template from the newly created server
type StepCreateImage struct{}

// Run runs the actual step
func (s *StepCreateImage) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	serverUuid := state.Get("server_uuid").(string)
	serverTitle := state.Get("server_title").(string)

	ui := state.Get("ui").(packer.Ui)
	driver := state.Get("driver").(Driver)

	ui.Say(fmt.Sprintf("Creating storage image for server %q...", serverTitle))

	err := driver.CreateTemplate(serverUuid)
	if err != nil {
		return StepHaltWithError(state, err)
	}
	ui.Say(fmt.Sprintf("Image for server %q created", serverTitle))
	return multistep.ActionContinue
}

// Cleanup cleans up after the step
func (s *StepCreateImage) Cleanup(state multistep.StateBag) {}
