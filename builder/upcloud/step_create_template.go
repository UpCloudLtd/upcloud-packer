package upcloud

import (
	"context"
	"fmt"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/packerbuilderdata"
)

// StepCreateTemplate represents the step that creates a storage template from the newly created server
type StepCreateTemplate struct {
	GeneratedData *packerbuilderdata.GeneratedData
}

// Run runs the actual step
func (s *StepCreateTemplate) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	serverUuid := state.Get("server_uuid").(string)
	serverTitle := state.Get("server_title").(string)

	ui := state.Get("ui").(packer.Ui)
	driver := state.Get("driver").(Driver)

	ui.Say(fmt.Sprintf("Creating storage template for server %q...", serverTitle))

	template, err := driver.CreateTemplate(serverUuid)
	if err != nil {
		return StepHaltWithError(state, err)
	}
	ui.Say(fmt.Sprintf("Storage template for server %q created", serverTitle))

	state.Put("template", template)
	s.GeneratedData.Put("TemplateUUID", template.UUID)
	s.GeneratedData.Put("TemplateTitle", template.Title)
	s.GeneratedData.Put("TemplateSize", template.Size)

	return multistep.ActionContinue
}

// Cleanup cleans up after the step
func (s *StepCreateTemplate) Cleanup(state multistep.StateBag) {}
