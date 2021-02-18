package upcloud

import (
	"context"
	"fmt"

	"github.com/UpCloudLtd/upcloud-go-api/upcloud"
	internal "github.com/UpCloudLtd/upcloud-packer/internal"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/packerbuilderdata"
)

// StepCreateTemplate represents the step that creates a storage template from the newly created server
type StepCreateTemplate struct {
	Config        *Config
	GeneratedData *packerbuilderdata.GeneratedData
}

// Run runs the actual step
func (s *StepCreateTemplate) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	serverUuid := state.Get("server_uuid").(string)

	ui := state.Get("ui").(packer.Ui)
	driver := state.Get("driver").(internal.Driver)

	// get storage details
	storage, err := driver.GetServerStorage(serverUuid)
	if err != nil {
		return internal.StepHaltWithError(state, err)
	}

	// clonning to zones
	cleanupStorageUuid := []string{}
	storageUuids := []string{}
	storageUuids = append(storageUuids, storage.UUID)

	for _, zone := range s.Config.CloneZones {
		ui.Say(fmt.Sprintf("Cloning storage %q to zone %q...", storage.UUID, zone))
		title := fmt.Sprintf("packer-%s-%s-cloned-disk1", s.Config.TemplatePrefix, internal.GetNowString())
		clonedStorage, err := driver.CloneStorage(storage.UUID, zone, title)
		if err != nil {
			return internal.StepHaltWithError(state, err)
		}
		storageUuids = append(storageUuids, clonedStorage.UUID)
		cleanupStorageUuid = append(cleanupStorageUuid, clonedStorage.UUID)
	}
	ui.Say("Clonning completed...")

	// creating template
	templates := []*upcloud.Storage{}

	for _, uuid := range storageUuids {
		ui.Say(fmt.Sprintf("Creating template for storage %q...", uuid))

		t, err := driver.CreateTemplate(uuid, s.Config.TemplatePrefix)
		if err != nil {
			return internal.StepHaltWithError(state, err)
		}
		templates = append(templates, t)
		ui.Say(fmt.Sprintf("Template for storage %q created...", uuid))
	}

	state.Put("cleanup_storage_uuids", cleanupStorageUuid)
	state.Put("templates", templates)

	return multistep.ActionContinue
}

// Cleanup cleans up after the step
func (s *StepCreateTemplate) Cleanup(state multistep.StateBag) {
	rawStorageUuids, ok := state.GetOk("cleanup_storage_uuids")

	if !ok {
		return
	}

	storageUuids := rawStorageUuids.([]string)

	ui := state.Get("ui").(packer.Ui)
	driver := state.Get("driver").(internal.Driver)

	for _, uuid := range storageUuids {
		ui.Say(fmt.Sprintf("Delete storage %q...", uuid))

		err := driver.DeleteTemplate(uuid)
		if err != nil {
			ui.Error(err.Error())
		}
	}
}
