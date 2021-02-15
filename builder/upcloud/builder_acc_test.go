package upcloud

import (
	"fmt"
	"os"
	"strings"
	"testing"

	builderT "github.com/hashicorp/packer-plugin-sdk/acctest"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
)

// Run tests: PACKER_ACC=1 go test -count 1 -v ./...  -timeout=120m
func TestBuilderAcc_default(t *testing.T) {
	builderT.Test(t, builderT.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		Builder:  &Builder{},
		Template: testBuildBasic,
		Check:    checkTemplateDefaultSettings(),
	})
}

func TestBuilderAcc_storageUuid(t *testing.T) {
	builderT.Test(t, builderT.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		Builder:  &Builder{},
		Template: testBuilderAccStorageUuid,
	})
}

func TestBuilderAcc_storageName(t *testing.T) {
	builderT.Test(t, builderT.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		Builder:  &Builder{},
		Template: testBuilderAccStorageName,
	})
}

func testAccPreCheck(t *testing.T) {
	if v := os.Getenv("UPCLOUD_API_USER"); v == "" {
		t.Fatal("UPCLOUD_API_USER must be set for acceptance tests")
	}
	if v := os.Getenv("UPCLOUD_API_PASSWORD"); v == "" {
		t.Fatal("UPCLOUD_API_PASSWORD must be set for acceptance tests")
	}
}

const testBuildBasic = `
{
	"builders": [{
            "type": "test",
            "zone": "nl-ams1",
            "storage_uuid": "01000000-0000-4000-8000-000050010400"
	}]
}
`

const testBuilderAccStorageUuid = `
{
	"builders": [{
            "type": "test",
            "zone": "nl-ams1",
            "storage_uuid": "01000000-0000-4000-8000-000050010400",
            "ssh_username": "root",
            "template_prefix": "test-builder",
            "storage_size": "20"
	}]
}
`

const testBuilderAccStorageName = `
{
	"builders": [{
            "type": "test",
            "zone": "nl-ams1",
            "storage_name": "ubuntu server 20.04",
            "ssh_username": "root",
            "template_prefix": "test-builder",
            "storage_size": "20"
	}]
}
`

func checkTemplateDefaultSettings() builderT.TestCheckFunc {
	return func(artifacts []packersdk.Artifact) error {
		if len(artifacts) > 1 {
			return fmt.Errorf("more than 1 artifact")
		}

		artifactRaw := artifacts[0]
		artifact, ok := artifactRaw.(*Artifact)
		if !ok {
			return fmt.Errorf("unknown artifact: %#v", artifactRaw)
		}

		expectedSize := 25
		expectedTitle := "custom-image"

		if artifact.Template.Size != expectedSize {
			return fmt.Errorf("Wrong size. Expected %d, got %d", expectedSize, artifact.Template.Size)
		}

		if !strings.HasPrefix(artifact.Template.Title, expectedTitle) {
			return fmt.Errorf("Wrong title prefix. Expected %q, got %q", expectedTitle, artifact.Template.Title)
		}
		return nil
	}
}
