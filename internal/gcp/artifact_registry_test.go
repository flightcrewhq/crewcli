package gcp

import (
	"testing"

	"github.com/stretchr/testify/assert"
	registry "google.golang.org/api/artifactregistry/v1"
)

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
}
