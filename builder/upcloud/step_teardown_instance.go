package upcloud

import (
	"context"
	"fmt"

	"github.com/UpCloudLtd/upcloud-go-api/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/upcloud/request"
	"github.com/UpCloudLtd/upcloud-go-api/upcloud/service"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/packer"
)

type StepTeardownInstance struct{}

func (s *StepTeardownInstance) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	// Store a success indicator in the state
	state.Put("step_templatize_storage_success", false)

	// Extract state
	ui := state.Get("ui").(packer.Ui)
	api := state.Get("api").(*service.Service)
	config := state.Get("config").(*Config)
	serverDetails := state.Get("server_details").(*upcloud.ServerDetails)

	// Stop the server and wait until it has stopped
	ui.Say(fmt.Sprintf("Stopping server \"%s\" ...", serverDetails.Title))
	serverDetails, err := api.StopServer(&request.StopServerRequest{
		UUID: serverDetails.UUID,
	})

	if err != nil {
		return StepHaltWithError(state, fmt.Errorf("Error teardown server: %s", err))
	}

	ui.Say(fmt.Sprintf("Waiting for server %q to enter the 'stopped' state...", serverDetails.Title))
	serverDetails, err = api.WaitForServerState(&request.WaitForServerStateRequest{
		UUID:         serverDetails.UUID,
		DesiredState: upcloud.ServerStateStopped,
		Timeout:      config.Timeout,
	})

	if err != nil {
		return StepHaltWithError(state, fmt.Errorf("Error waiting for server: %s", err))
	}

	ui.Say(fmt.Sprintf("Server %q is now in 'stopped' state", serverDetails.Title))

	return multistep.ActionContinue
}

func (s *StepTeardownInstance) Cleanup(state multistep.StateBag) {
	// no cleanup
}
