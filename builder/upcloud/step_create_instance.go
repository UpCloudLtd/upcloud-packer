package upcloud

import (
	"context"
	"fmt"
	"time"

	"github.com/UpCloudLtd/upcloud-go-api/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/upcloud/request"
	"github.com/UpCloudLtd/upcloud-go-api/upcloud/service"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/packer"
)

type StepCreateInstance struct {
	Debug bool

	// GeneratedData *packerbuilderdata.GeneratedData
}

func (s *StepCreateInstance) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	config := state.Get("config").(Config)
	api := state.Get("api").(service.Service)

	title := fmt.Sprintf("custom-image-%d", time.Now().Unix())
	hostname := title

	createServerRequest := request.CreateServerRequest{
		Title:            title,
		Hostname:         hostname,
		Zone:             config.Zone,
		PasswordDelivery: request.PasswordDeliveryNone,
		CoreNumber:       2,
		MemoryAmount:     2048,
		StorageDevices: []request.CreateServerStorageDevice{
			{
				Action:  request.CreateServerStorageDeviceActionClone,
				Storage: config.StorageUUID,
				Title:   fmt.Sprintf("%s-disk1", title),
				Size:    config.StorageSize,
				Tier:    upcloud.StorageTierMaxIOPS,
			},
		},
		Networking: &request.CreateServerNetworking{
			Interfaces: []request.CreateServerInterface{
				{
					IPAddresses: []request.CreateServerIPAddress{
						{
							Family: upcloud.IPAddressFamilyIPv4,
						},
					},
					Type: upcloud.IPAddressAccessPublic,
				},
				{
					IPAddresses: []request.CreateServerIPAddress{
						{
							Family: upcloud.IPAddressFamilyIPv4,
						},
					},
					Type: upcloud.IPAddressAccessPrivate,
				},
				{
					IPAddresses: []request.CreateServerIPAddress{
						{
							Family: upcloud.IPAddressFamilyIPv6,
						},
					},
					Type: upcloud.IPAddressAccessPublic,
				},
			},
		},
		LoginUser: &request.LoginUser{
			CreatePassword: "no",
			Username:       config.Communicator.SSHUsername,
			SSHKeys: []string{
				state.Get("ssh_key_public").(string),
			},
		},
	}
	// Create the server
	ui.Say(fmt.Sprintf("Creating server %q...", title))

	serverDetails, err := api.CreateServer(&createServerRequest)
	if err != nil {
		return StepHaltWithError(state, fmt.Errorf("Error creating server: %s", err))
	}

	// Store the server details in the state immediately
	state.Put("server_details", serverDetails)

	ui.Say(fmt.Sprintf("Waiting for server %q to enter the 'started' state ...", title))
	serverDetails, err = api.WaitForServerState(&request.WaitForServerStateRequest{
		UUID:         serverDetails.UUID,
		DesiredState: upcloud.ServerStateStarted,
		Timeout:      config.Timeout,
	})

	if err != nil {
		return StepHaltWithError(state, fmt.Errorf("Error while waiting for server: %s", err))
	}

	// Update the state
	state.Put("server_details", serverDetails)

	ui.Say(fmt.Sprintf("Server %q is now in 'started' state", title))

	return multistep.ActionContinue
}

// Cleanup stops and destroys the server if server details are found in the state
func (s *StepCreateInstance) Cleanup(state multistep.StateBag) {
	// Extract state, return if no state has been stored
	rawDetails, ok := state.GetOk("server_details")

	if !ok {
		return
	}

	serverDetails := rawDetails.(*upcloud.ServerDetails)
	uuid := serverDetails.UUID
	title := serverDetails.Title

	ui := state.Get("ui").(packer.Ui)
	config := state.Get("config").(Config)
	api := state.Get("api").(service.Service)

	// Ensure the instance is not in maintenance state
	ui.Say(fmt.Sprintf("Waiting for server %q to exit the 'maintenance' state ...", title))
	_, err := api.WaitForServerState(&request.WaitForServerStateRequest{
		UUID:           uuid,
		UndesiredState: upcloud.ServerStateMaintenance,
		Timeout:        config.Timeout,
	})

	if err != nil {
		ui.Error(fmt.Sprintf("Error while waiting for server %q to exit the 'maintenance' state: %s", title, err))
		return
	}

	// Stop the server if it hasn't been stopped yet
	newServerDetails, err := api.GetServerDetails(&request.GetServerDetailsRequest{
		UUID: serverDetails.UUID,
	})

	if err != nil {
		ui.Error(fmt.Sprintf("Failed to get details for server %q: %s", title, err))
		return
	}

	if newServerDetails.State != upcloud.ServerStateStopped {
		ui.Say(fmt.Sprintf("Stopping server %q...", title))
		_, err = api.StopServer(&request.StopServerRequest{
			UUID: serverDetails.UUID,
		})

		if err != nil {
			ui.Error(fmt.Sprintf("Failed to stop server %q: %s", title, err))
			return
		}

		// Wait for the server to stop
		ui.Say(fmt.Sprintf("Waiting for server %q to enter the 'stopped' state ...", title))
		_, err = api.WaitForServerState(&request.WaitForServerStateRequest{
			UUID:         serverDetails.UUID,
			DesiredState: upcloud.ServerStateStopped,
			Timeout:      config.Timeout,
		})

		if err != nil {
			ui.Error(fmt.Sprintf("Error while waiting for server %q to enter the 'stopped' state: %s", title, err))
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
	ui.Say(fmt.Sprintf("Deleting server %q...", title))
	err = api.DeleteServer(&request.DeleteServerRequest{
		UUID: serverDetails.UUID,
	})

	if err != nil {
		ui.Error(fmt.Sprintf("Failed to delete server %q: %s", title, err))
	}

	// Delete the disk
	if storageUUID != "" {
		ui.Say(fmt.Sprintf("Deleting disk \"%s\" ...", storageTitle))
		err = api.DeleteStorage(&request.DeleteStorageRequest{
			UUID: storageUUID,
		})

		if err != nil {
			ui.Error(fmt.Sprintf("Failed to delete disk %q: %s", storageTitle, err))
		}
	}
}
