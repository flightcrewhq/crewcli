package gcp

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"regexp"

	registry "google.golang.org/api/artifactregistry/v1"
)

var (
	ArtifactRegistryService *registry.Service
	ImagePath               = "us-west1-docker.pkg.dev/flightcrew-artifacts/client/tower"
	versionRE               = regexp.MustCompile(`[0-9]+\.[0-9]+\.[0-9]+`)
)

const (
	parent   = "projects/flightcrew-artifacts/locations/us-west1/repositories/client"
	pageSize = 50
)

func init() {
	ctx := context.Background()
	service, err := registry.NewService(ctx)
	if err != nil {
		return
	}

	ArtifactRegistryService = service
}

// GetTowerImageVersion returns the associated image tag in the form of x.x.x
// so that it can be passed into the Tower.
func GetTowerImageVersion(version string) (string, error) {
	if ArtifactRegistryService == nil {
		if versionRE.MatchString(version) {
			return version, nil
		}

		return "", errors.New("want `x.x.x` format: failed to lookup image")
	}

	images := make([]*registry.DockerImage, 0)

	var resp []*registry.DockerImage
	var pageToken string
	var err error
	for {
		resp, pageToken, err = queryDockerImageAPI(pageToken)
		if err != nil {
			return "", fmt.Errorf("query docker image api: %w", err)
		}

		images = append(images, resp...)

		if pageToken == "" {
			break
		}
	}

	return getDesiredImageVersion(images, version)
}

func queryDockerImageAPI(pageToken string) ([]*registry.DockerImage, string, error) {
	dockerImageSvc := ArtifactRegistryService.Projects.Locations.Repositories.DockerImages
	call := dockerImageSvc.List(parent).PageSize(pageSize).PageToken(pageToken)
	resp, err := call.Do()
	if err != nil {
		return nil, "", fmt.Errorf("list docker images: %w", err)
	}

	if resp.HTTPStatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("google artifact registry api returned not OK: %d", resp.HTTPStatusCode)
	}

	return resp.DockerImages, resp.NextPageToken, nil
}

func getDesiredImageVersion(images []*registry.DockerImage, version string) (string, error) {
	var desiredImage *registry.DockerImage
	for i, image := range images {
		for _, tag := range image.Tags {
			if tag == version {
				desiredImage = images[i]
				break
			}
		}

		if desiredImage != nil {
			break
		}
	}

	if desiredImage == nil {
		return "", fmt.Errorf("unable to find tower version: %s", version)
	}

	if versionRE.MatchString(version) {
		return version, nil
	}

	for _, tag := range desiredImage.Tags {
		if versionRE.MatchString(tag) {
			return tag, nil
		}
	}

	return "", fmt.Errorf("no valid tower version tag found from %s", version)
}
