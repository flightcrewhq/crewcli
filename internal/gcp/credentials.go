package gcp

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"os/exec"

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

func GetProjectsFromEnvironment() ([]string, error) {
	var stdout, stderr bytes.Buffer
	c := exec.Command("bash", "-c", "gcloud projects list --format='csv(PROJECT_ID)'")
	c.Stdout = &stdout
	c.Stderr = &stderr
	if err := c.Run(); err != nil {
		return nil, fmt.Errorf("gcloud projects list: %w", err)
	}

	projectIDs := make([]string, 0)
	r := csv.NewReader(&stdout)
	if _, err := r.Read(); err != nil {
		return nil, fmt.Errorf("no header when reading csv for project list: %w", err)
	}

	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}

		if len(record) == 1 {
			projectIDs = append(projectIDs, record[0])
		}
	}

	return projectIDs, nil
}
