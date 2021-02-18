package upcloud

import (
	"fmt"
	"testing"

	"github.com/UpCloudLtd/upcloud-go-api/upcloud"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
)

func TestArtifact_impl(t *testing.T) {
	var _ packersdk.Artifact = new(Artifact)
}

func TestArtifact_Id(t *testing.T) {
	uuid1 := "some-uuid-1"
	uuid2 := "some-uuid-2"
	expected := fmt.Sprintf("%s,%s", uuid1, uuid2)

	templates := []*upcloud.Storage{}
	templates = append(templates, &upcloud.Storage{UUID: uuid1})
	templates = append(templates, &upcloud.Storage{UUID: uuid2})

	a := &Artifact{Templates: templates}
	result := a.Id()

	if result != expected {
		t.Errorf("Expected: %q, got: %q", expected, result)
	}
}

func TestArtifact_String(t *testing.T) {
	expected := `Storage template created, UUID: some-uuid`

	templates := []*upcloud.Storage{}
	templates = append(templates, &upcloud.Storage{UUID: "some-uuid"})

	a := &Artifact{Templates: templates}
	result := a.String()

	if result != expected {
		t.Errorf("Expected: %q, got: %q", expected, result)
	}
}
