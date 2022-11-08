package gcp

import (
	"testing"

	"github.com/stretchr/testify/assert"
	registry "google.golang.org/api/artifactregistry/v1"
)

func TestGetImageVersionWithoutArtifactRegistry(t *testing.T) {
	keepARS := ArtifactRegistryService
	ArtifactRegistryService = nil
	defer func() {
		ArtifactRegistryService = keepARS
	}()

	var version string
	var err error

	version, err = GetTowerImageVersion("2.1.15")
	assert.NoError(t, err)
	assert.Equal(t, "2.1.15", version)

	_, err = GetTowerImageVersion("stable")
	assert.Error(t, err)
}

func TestGetDesiredImageVersion(mainT *testing.T) {
	images := []*registry.DockerImage{
		{
			Name: "blahblah",
			Tags: []string{
				"stable",
				"0.1.2",
			},
		},
		{
			Name: "blehbleh",
			Tags: []string{
				"latest",
				"12.2.123",
			},
		},
		{
			Name: "blubblub",
			Tags: []string{
				"12.1.1351-test-not-stable",
			},
		},
		{
			Name: "blabblab",
			Tags: []string{
				"labeled",
			},
		},
	}

	mainT.Run("no existing version in numbered format should fail", func(t *testing.T) {
		_, err := getDesiredImageVersion(images, "1.2.3")
		assert.Errorf(t, err, "tag 1.2.3 should not exist")
	})

	mainT.Run("no existing version in string format should fail", func(t *testing.T) {
		_, err := getDesiredImageVersion(images, "some_tag")
		assert.Errorf(t, err, "some_tag should not exist")
	})

	mainT.Run("stable should get version format", func(t *testing.T) {
		version, err := getDesiredImageVersion(images, "stable")
		assert.NoError(t, err)
		assert.Equal(t, "0.1.2", version)
	})

	mainT.Run("latest should get version format", func(t *testing.T) {
		version, err := getDesiredImageVersion(images, "latest")
		assert.NoError(t, err)
		assert.Equal(t, "12.2.123", version)
	})

	mainT.Run("get exact tag should get exact tag", func(t *testing.T) {
		version, err := getDesiredImageVersion(images, "12.2.123")
		assert.NoError(t, err)
		assert.Equal(t, "12.2.123", version)
	})

	mainT.Run("get image with no version label fails", func(t *testing.T) {
		_, err := getDesiredImageVersion(images, "labeled")
		assert.Error(t, err)
	})
}
