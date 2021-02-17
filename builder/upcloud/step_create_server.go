package upcloud

import (
	"context"
	"fmt"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/packerbuilderdata"
)

// StepCreateServer represents the step that creates a server
type StepCreateServer struct {
	Config        *Config
	GeneratedData *packerbuilderdata.GeneratedData
}

// Run runs the actual step
func (s *StepCreateServer) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	driver := state.Get("driver").(Driver)

	rawSshKeyPublic, ok := state.GetOk("ssh_key_public")
	if !ok {
		return StepHaltWithError(state, fmt.Errorf("SSH public key is missing"))
	}
	sshKeyPublic := rawSshKeyPublic.(string)

	ui.Say("Getting storage...")

	storage, err := driver.GetStorage(s.Config.StorageUUID, s.Config.StorageName)
	if err != nil {
		return StepHaltWithError(state, err)
	}

	ui.Say(fmt.Sprintf("Creating server based on storage %q...", storage.Title))

	response, err := driver.CreateServer(
		storage.UUID,
		s.Config.Zone,
		s.Config.TemplatePrefix,
		sshKeyPublic,
		s.Config.StorageSize,
	)
	if err != nil {
		return StepHaltWithError(state, err)
	}

	serverUuid := response.UUID
	serverTitle := response.Title
	serverIp, err := GetServerIp(response)
	if err != nil {
		return StepHaltWithError(state, err)
	}

	ui.Say(fmt.Sprintf("Server %q created and in 'started' state", serverTitle))

	state.Put("server_uuid", serverUuid)
	state.Put("server_title", serverTitle)
	state.Put("server_ip", serverIp)

	s.GeneratedData.Put("ServerUUID", serverUuid)
	s.GeneratedData.Put("ServerTitle", serverTitle)
	s.GeneratedData.Put("ServerSize", serverIp)

	return multistep.ActionContinue
}

// Cleanup stops and destroys the server if server details are found in the state
func (s *StepCreateServer) Cleanup(state multistep.StateBag) {
	// Extract server uuid, return if no uuid has been stored
	rawServerUuid, ok := state.GetOk("server_uuid")

	if !ok {
		return
	}

	serverUuid := rawServerUuid.(string)
	serverTitle := state.Get("server_title").(string)

	ui := state.Get("ui").(packer.Ui)
	driver := state.Get("driver").(Driver)

	// stop server
	ui.Say(fmt.Sprintf("Stopping server %q...", serverTitle))

	err := driver.StopServer(serverUuid)
	if err != nil {
		ui.Error(err.Error())
		return
	}

	// delete server
	ui.Say(fmt.Sprintf("Deleting server %q...", serverTitle))

	err = driver.DeleteServer(serverUuid)
	if err != nil {
		ui.Error(err.Error())
		return
	}
}
