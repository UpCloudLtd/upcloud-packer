package upcloud

import (
	"fmt"

	"github.com/UpCloudLtd/upcloud-go-api/upcloud"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
)

// wraps error logic
func StepHaltWithError(state multistep.StateBag, err error) multistep.StepAction {
	ui := state.Get("ui").(packersdk.Ui)
	state.Put("error", err)
	ui.Error(err.Error())
	return multistep.ActionHalt
}

// parse public ip from server details
func GetServerIp(details *upcloud.ServerDetails) (string, error) {
	for _, ipAddress := range details.IPAddresses {
		if ipAddress.Access == upcloud.IPAddressAccessPublic && ipAddress.Family == upcloud.IPAddressFamilyIPv4 {
			return ipAddress.Address, nil
		}
	}
	return "", fmt.Errorf("Unable to find the public IPv4 address of the server")
}
