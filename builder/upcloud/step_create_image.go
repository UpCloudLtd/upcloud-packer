package upcloud

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/UpCloudLtd/upcloud-go-api/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/upcloud/request"
	"github.com/UpCloudLtd/upcloud-go-api/upcloud/service"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/packer"
)

// StepCreateImage represents the step that creates a storage template from the newly created server
type StepCreateImage struct{}

// Run runs the actual step
func (s *StepCreateImage) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	// Store a success indicator in the state
	state.Put("step_templatize_storage_success", false)

	// Extract state
	ui := state.Get("ui").(packer.Ui)
	api := state.Get("api").(*service.Service)
	config := state.Get("config").(*Config)
	serverDetails := state.Get("server_details").(*upcloud.ServerDetails)

	// Templatize the first disk device in the server
	for _, storage := range serverDetails.StorageDevices {
		if storage.Type == upcloud.StorageTypeDisk {
			ui.Say(fmt.Sprintf("Creating storage image %q ...", storage.Title))

			// Determine the prefix to use for the template title
			prefix := storage.Title
			if config.TemplatePrefix != "" {
				prefix = config.TemplatePrefix
			}

			storageDetails, err := api.TemplatizeStorage(&request.TemplatizeStorageRequest{
				UUID:  storage.UUID,
				Title: fmt.Sprintf("%s-template-%d", prefix, time.Now().Unix()),
			})

			if err != nil {
				return StepHaltWithError(state, fmt.Errorf("Error creating image: %s", err))
			}

			// Wait for the newly templatized storage to enter the "online" state
			ui.Say(fmt.Sprintf("Waiting for storage %q to enter the 'online' state", storageDetails.Title))
			storageDetails, err = api.WaitForStorageState(&request.WaitForStorageStateRequest{
				UUID:         storageDetails.UUID,
				DesiredState: upcloud.StorageStateOnline,
				Timeout:      config.Timeout,
			})

			if err != nil {
				return StepHaltWithError(state, fmt.Errorf("Error creating image: %s", err))
			}

			// Storage the details about the templatized storage in the state. Also update our success
			// boolean
			state.Put("storage_details", storageDetails)
			state.Put("step_templatize_storage_success", true)

			return multistep.ActionContinue
		}
	}

	// No storage found, we'll have to abort
	return StepHaltWithError(state, errors.New("Unable to find the storage device to templatize"))
}

// Cleanup cleans up after the step
func (s *StepCreateImage) Cleanup(state multistep.StateBag) {
	// Don't perform any cleanup if the step executed successfully
	if state.Get("step_templatize_storage_success").(bool) {
		return
	}

	// Extract state, return if no state has been stored
	if rawDetails, ok := state.GetOk("storage_details"); ok {
		storageDetails := rawDetails.(*upcloud.StorageDetails)

		api := state.Get("api").(*service.Service)
		ui := state.Get("ui").(packer.Ui)

		// Delete the storage device
		err := api.DeleteStorage(&request.DeleteStorageRequest{
			UUID: storageDetails.UUID,
		})

		if err != nil {
			ui.Error(fmt.Sprintf("%s", err))
			return
		}
	}
}
