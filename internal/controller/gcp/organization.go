package gcp

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

func GetOrganizationID(projectID string) (string, error) {
	var stdout, stderr bytes.Buffer
	err := bashGetAncestors(projectID, &stdout, &stderr)
	if err != nil {
		return "", err
	}

	orgID := strings.Trim(stdout.String(), " \n\t")
	if len(orgID) == 0 {
		return "", errors.New("not found: Google organization ID")
	}

	return orgID, nil
}

func bashGetAncestors(projectID string, stdout, stderr *bytes.Buffer) error {
	cmdStr := strings.Replace("gcloud projects get-ancestors ${PROJECT_ID} | awk '/organization/ {print $1}'", "${PROJECT_ID}", projectID, 1)
	c := exec.Command("bash", "-c", cmdStr)
	c.Stdout = stdout
	c.Stderr = stderr
	if err := c.Run(); err != nil {
		return fmt.Errorf("gcloud projects get-ancestors: %w", err)
	}

	return nil
}
