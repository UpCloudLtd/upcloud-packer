package upcloud

import (
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
)

func StepHaltWithError(state multistep.StateBag, err error) multistep.StepAction {
	ui := state.Get("ui").(packersdk.Ui)
	state.Put("error", err)
	ui.Error(err.Error())
	return multistep.ActionHalt
}
