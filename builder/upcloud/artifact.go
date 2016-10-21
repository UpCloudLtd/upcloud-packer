package upcloud

import (
	"fmt"
	"github.com/Jalle19/upcloud-go-sdk/upcloud/request"
	"github.com/Jalle19/upcloud-go-sdk/upcloud/service"
	"log"
)

// Artifact represents a template of a storage as the result of a Packer build
type Artifact struct {
	UUID    string
	Zone    string
	Title   string
	service *service.Service
}

// BuilderId returns the unique identifier of this builder
func (*Artifact) BuilderId() string {
	return BuilderId
}

// Destroy destroys the template
func (a *Artifact) Destroy() error {
	log.Printf("Deleting template \"%s\"", a.Title)

	return a.service.DeleteStorage(&request.DeleteStorageRequest{
		UUID: a.UUID,
	})
}

// Files returns the files represented by the artifact
func (*Artifact) Files() []string {
	return nil
}

func (a *Artifact) Id() string {
	return a.UUID
}

func (*Artifact) State(name string) interface{} {
	return nil
}

// String returns the string representation of the artifact. It is printed at the end of builds.
func (a *Artifact) String() string {
	return fmt.Sprintf("Private template (UUID: %s, Title: %s, Zone: %s)", a.UUID, a.Title, a.Zone)
}
