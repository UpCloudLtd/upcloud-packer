package upcloud

import (
	"fmt"

	"github.com/UpCloudLtd/upcloud-go-api/upcloud"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
)

// sshHostCallback retrieves the server details from the state and returns the public IPv4 address of the server
func sshHostCallback(state multistep.StateBag) (string, error) {
	serverDetails := state.Get("server_details").(*upcloud.ServerDetails)

	for _, ipAddress := range serverDetails.IPAddresses {
		if ipAddress.Access == upcloud.IPAddressAccessPublic && ipAddress.Family == upcloud.IPAddressFamilyIPv4 {
			return ipAddress.Address, nil
		}
	}

	return "", fmt.Errorf("Unable to find the public IPv4 address of the server")
}
