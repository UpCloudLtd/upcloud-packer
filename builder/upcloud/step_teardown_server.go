package upcloud

import (
	"context"
	"fmt"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/packer"
)

// StepTeardownServer represents the step that stops the server before creating the image
type StepTeardownServer struct{}

// Run runs the actual step
func (s *StepTeardownServer) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	// Extract server details
	serverUuid := state.Get("server_uuid").(string)
	serverTitle := state.Get("server_title").(string)

	ui := state.Get("ui").(packer.Ui)
	driver := state.Get("driver").(Driver)

	ui.Say(fmt.Sprintf("Stopping server %q...", serverTitle))

	err := driver.StopServer(serverUuid)
	if err != nil {
		return StepHaltWithError(state, err)
	}

	ui.Say(fmt.Sprintf("Server %q is now in 'stopped' state", serverTitle))

	return multistep.ActionContinue
}

func (s *StepTeardownServer) Cleanup(state multistep.StateBag) {}
