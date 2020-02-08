package upcloud

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"

	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"golang.org/x/crypto/ssh"
	"io/ioutil"
)

// ObtainSSHKeys represents a Packer build step that either loads or generates ssh key pairs.
type ObtainSSHKeys struct {
	Debug        bool
	DebugKeyPath string
}

// Run executes the Packer build step that obtains SSH key pairs.
func (s *ObtainSSHKeys) Run(state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	config := state.Get("config").(Config)

	// If we have existing SSH key pairs, use them
	if config.SSHPrivateKeyFile != "" && config.SSHPublicKeyFile != "" {
		ui.Say("Using existing SSH keys")
		privateKey, err := ioutil.ReadFile(config.SSHPrivateKeyFile)
		if err != nil {
			return handleError(fmt.Errorf("Failed to read private key: %s", err), state)
		}
		publicKey, err := ioutil.ReadFile(config.SSHPublicKeyFile)
		if err != nil {
			return handleError(fmt.Errorf("Failed to read public key: %s", err), state)
		}

		state.Put("ssh_private_key", string(privateKey))
		state.Put("ssh_public_key", string(publicKey))
	} else {

		// Otherwise generate a new 2048-bit private key
		ui.Say("Creating temporary SSH key ...")
		privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			return handleError(fmt.Errorf("Error creating temporary SSH private key: %s", err), state)
		}

		// Convert the key to a PEM block
		privateBlock := pem.Block{
			Type:    "RSA PRIVATE KEY",
			Headers: nil,
			Bytes:   x509.MarshalPKCS1PrivateKey(privateKey),
		}

		// Create the corresponding public key
		publicKey, err := ssh.NewPublicKey(&privateKey.PublicKey)
		if err != nil {
			return handleError(fmt.Errorf("Error creating temporary SSH public key: %s", err), state)
		}

		// Save the keys in the state bag
		state.Put("ssh_private_key", string(pem.EncodeToMemory(&privateBlock)))
		state.Put("ssh_public_key", string(ssh.MarshalAuthorizedKey(publicKey)))

		// Optionally save the keys for debug purposes
		if s.Debug {
			ui.Message(fmt.Sprintf("Saving SSH key for debug purposes: %s", s.DebugKeyPath))

			file, err := os.Create(s.DebugKeyPath)
			if err != nil {
				return handleError(fmt.Errorf("Error saving debug key: %s", err), state)
			}

			// Write out the key
			err = pem.Encode(file, &privateBlock)
			file.Close()
			if err != nil {
				return handleError(fmt.Errorf("Error saving debug key: %s", err), state)
			}
		}
	}

	return multistep.ActionContinue
}

// Cleanup cleans up after the step. In this step we don't need to perform any cleanup since the SSH keys are only
// temporary.
func (s *ObtainSSHKeys) Cleanup(state multistep.StateBag) {}
