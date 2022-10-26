package gcp

import (
	"context"
	"fmt"

	"golang.org/x/oauth2/google"
	"google.golang.org/api/compute/v1"
)

func GetCurrentProject(ctx context.Context) (string, error) {
	credentials, err := google.FindDefaultCredentials(ctx, compute.ComputeScope)
	if err != nil {
		return "", fmt.Errorf("find default credentials: %w", err)
	}

	if credentials.ProjectID == "" {
		return "", fmt.Errorf("no project id in current scope")
	}

	return credentials.ProjectID, nil
}
