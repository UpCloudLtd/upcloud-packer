package upcloud

import (
	"fmt"
	"golang.org/x/crypto/ssh"

	"github.com/Jalle19/upcloud-go-sdk/upcloud"
	"github.com/mitchellh/multistep"
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

// sshConfigCallback
func sshConfigCallback(state multistep.StateBag) (*ssh.ClientConfig, error) {
	config := state.Get("config").(Config)
	privateKey := state.Get("ssh_private_key").(string)

	signer, err := ssh.ParsePrivateKey([]byte(privateKey))
	if err != nil {
		return nil, fmt.Errorf("Error creating SSH config: %s", err)
	}

	return &ssh.ClientConfig{
		User: config.Comm.SSHUsername,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
	}, nil
}
