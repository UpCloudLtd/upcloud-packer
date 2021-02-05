package upcloud

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"golang.org/x/crypto/ssh"
)

type StepCreateSSHKey struct {
	Debug        bool
	DebugKeyPath string
}

func (s *StepCreateSSHKey) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packersdk.Ui)
	config := state.Get("config").(*Config)

	ui.Say("Creating temporary ssh key...")

	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return StepHaltWithError(state, fmt.Errorf("Error generating SSH key: %s", err))
	}

	// ASN.1 DER encoded form
	privDer := x509.MarshalPKCS1PrivateKey(priv)
	privBlk := pem.Block{
		Type:    "RSA PRIVATE KEY",
		Headers: nil,
		Bytes:   privDer,
	}

	// Marshal the public key into SSH compatible format
	pub, err := ssh.NewPublicKey(&priv.PublicKey)
	if err != nil {
		return StepHaltWithError(state, fmt.Errorf("Error creating public ssh key: %s", err))
	}

	// Remember some state for the future
	pubSSHFormat := string(ssh.MarshalAuthorizedKey(pub))
	state.Put("ssh_key_public", pubSSHFormat)

	// Set the private key in the config for later
	config.Comm.SSHPrivateKey = pem.EncodeToMemory(&privBlk)
	config.Comm.SSHPublicKey = ssh.MarshalAuthorizedKey(pub)

	// If we're in debug mode, output the private key to the working directory.
	if s.Debug {
		ui.Say(fmt.Sprintf("Saving key for debug purposes: %s", s.DebugKeyPath))
		err := ioutil.WriteFile(s.DebugKeyPath, config.Comm.SSHPrivateKey, 0600)
		if err != nil {
			return StepHaltWithError(state, fmt.Errorf("Error saving debug key: %s", err))
		}
	}

	return multistep.ActionContinue
}

func (s *StepCreateSSHKey) Cleanup(state multistep.StateBag) {}
