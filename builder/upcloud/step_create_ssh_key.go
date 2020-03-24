package upcloud

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"github.com/hashicorp/packer/helper/multistep"
	"os"

	"github.com/hashicorp/packer/packer"
	"golang.org/x/crypto/ssh"
)

// StepCreateSSHKey represents a Packer build step that generates SSH key pairs.
type StepCreateSSHKey struct {
	Debug        bool
	DebugKeyPath string
}

// Run executes the Packer build step that generates SSH key pairs.
func (s *StepCreateSSHKey) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	// Extract state
	ui := state.Get("ui").(packer.Ui)

	// Generate a new 2048-bit private key
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

	return multistep.ActionContinue
}

// Cleanup cleans up after the step. In this step we don't need to perform any cleanup since the SSH keys are only
// temporary.
func (s *StepCreateSSHKey) Cleanup(state multistep.StateBag) {}
