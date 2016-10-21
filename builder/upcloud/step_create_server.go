package upcloud

import (
	"fmt"
	"github.com/Jalle19/upcloud-go-sdk/upcloud"
	"github.com/Jalle19/upcloud-go-sdk/upcloud/request"
	"github.com/Jalle19/upcloud-go-sdk/upcloud/service"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"time"
)

// StepCreateServer represents the Packer step that creates a new server instance
type StepCreateServer struct {
}

// Run performs the actual step
func (s *StepCreateServer) Run(state multistep.StateBag) multistep.StepAction {
	// Extract state
	ui := state.Get("ui").(packer.Ui)
	service := state.Get("service").(service.Service)
	config := state.Get("config").(Config)

	// Create the request
	title := fmt.Sprintf("packer-builder-upcloud-%d", time.Now().Unix())
	hostname := title

	createServerRequest := request.CreateServerRequest{
		Title:            title,
		Hostname:         hostname,
		Zone:             config.Zone,
		PasswordDelivery: request.PasswordDeliveryNone,
		CoreNumber:       2,
		MemoryAmount:     2048,
		StorageDevices: []upcloud.CreateServerStorageDevice{
			{
				Action:  upcloud.CreateServerStorageDeviceActionClone,
				Storage: config.StorageUUID,
				Title:   fmt.Sprintf("%s-disk1", title),
				Size:    config.StorageSize,
				Tier:    upcloud.StorageTierMaxIOPS,
			},
		},
		IPAddresses: []request.CreateServerIPAddress{
			{
				Access: upcloud.IPAddressAccessPrivate,
				Family: upcloud.IPAddressFamilyIPv4,
			},
			{
				Access: upcloud.IPAddressAccessPublic,
				Family: upcloud.IPAddressFamilyIPv4,
			},
			{
				Access: upcloud.IPAddressAccessPublic,
				Family: upcloud.IPAddressFamilyIPv6,
			},
		},
		LoginUser: &request.LoginUser{
			CreatePassword: "no",
			Username:       config.Comm.SSHUsername,
			SSHKeys: []string{
				state.Get("ssh_public_key").(string),
			},
		},
	}

	// Create the server
	ui.Say(fmt.Sprintf("Creating server \"%s\" ...", createServerRequest.Title))

	serverDetails, err := service.CreateServer(&createServerRequest)
	if err != nil {
		return handleError(fmt.Errorf("Error creating server instance: %s", err), state)
	}

	// Store the server details in the state immediately
	state.Put("server_details", serverDetails)

	ui.Say(fmt.Sprintf("Waiting for server \"%s\" to enter the \"started\" state ...", serverDetails.Title))
	serverDetails, err = service.WaitForServerState(&request.WaitForServerStateRequest{
		UUID:         serverDetails.UUID,
		DesiredState: upcloud.ServerStateStarted,
		Timeout:      config.StateTimeoutDuration,
	})

	if err != nil {
		return handleError(fmt.Errorf("Error while waiting for server \"%s\" to enter the \"started\" state: %s", serverDetails.Title, err), state)
	}

	// Update the state
	state.Put("server_details", serverDetails)

	ui.Say(fmt.Sprintf("Server \"%s\" is now in \"started\" state", serverDetails.Title))
	return multistep.ActionContinue
}

// Cleanup stops and destroys the server if server details are found in the state
func (s *StepCreateServer) Cleanup(state multistep.StateBag) {
	// Extract state, return if no state has been stored
	rawDetails, ok := state.GetOk("server_details")

	if !ok {
		return
	}

	serverDetails := rawDetails.(*upcloud.ServerDetails)

	ui := state.Get("ui").(packer.Ui)
	config := state.Get("config").(Config)
	service := state.Get("service").(service.Service)

	// Ensure the instance is not in maintenance state
	ui.Say(fmt.Sprintf("Waiting for server \"%s\" to exit the \"maintenance\" state ...", serverDetails.Title))
	_, err := service.WaitForServerState(&request.WaitForServerStateRequest{
		UUID:           serverDetails.UUID,
		UndesiredState: upcloud.ServerStateMaintenance,
		Timeout:        config.StateTimeoutDuration,
	})

	if err != nil {
		ui.Error(fmt.Sprintf("Error while waiting for server \"%s\" to exit the \"maintenance\" state: %s", serverDetails.Title, err))
		return
	}

	// Stop the server if it hasn't been stopped yet
	newServerDetails, err := service.GetServerDetails(&request.GetServerDetailsRequest{
		UUID: serverDetails.UUID,
	})

	if err != nil {
		ui.Error(fmt.Sprintf("Failed to get details for server \"%s\": %s", serverDetails.Title, err))
		return
	}

	if newServerDetails.State != upcloud.ServerStateStopped {
		ui.Say(fmt.Sprintf("Stopping server \"%s\" ...", serverDetails.Title))
		_, err = service.StopServer(&request.StopServerRequest{
			UUID: serverDetails.UUID,
		})

		if err != nil {
			ui.Error(fmt.Sprintf("Failed to stop server \"%s\": %s", serverDetails.Title, err))
			return
		}

		// Wait for the server to stop
		ui.Say(fmt.Sprintf("Waiting for server \"%s\" to enter the \"stopped\" state ...", serverDetails.Title))
		_, err = service.WaitForServerState(&request.WaitForServerStateRequest{
			UUID:         serverDetails.UUID,
			DesiredState: upcloud.ServerStateStopped,
			Timeout:      config.StateTimeoutDuration,
		})

		if err != nil {
			ui.Error(fmt.Sprintf("Error while waiting for server \"%s\" to enter the \"stopped\" state: %s", serverDetails.Title, err))
			return
		}
	}

	// Store the disk UUID so we can delete it once the server is deleted
	storageUUID := ""
	storageTitle := ""

	for _, storage := range newServerDetails.StorageDevices {
		if storage.Type == upcloud.StorageTypeDisk {
			storageUUID = storage.UUID
			storageTitle = storage.Title
			break
		}
	}

	// Delete the server
	ui.Say(fmt.Sprintf("Deleting server \"%s\" ...", serverDetails.Title))
	err = service.DeleteServer(&request.DeleteServerRequest{
		UUID: serverDetails.UUID,
	})

	if err != nil {
		ui.Error(fmt.Sprintf("Failed to delete server \"%s\": %s", serverDetails.Title, err))
	}

	// Delete the disk
	if storageUUID != "" {
		ui.Say(fmt.Sprintf("Deleting disk \"%s\" ...", storageTitle))
		err = service.DeleteStorage(&request.DeleteStorageRequest{
			UUID: storageUUID,
		})

		if err != nil {
			ui.Error(fmt.Sprintf("Failed to delete disk \"%s\": %s", storageTitle, err))
		}
	}
}
