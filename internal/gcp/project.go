package gcp

import (
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os/exec"
)

func GetProjectFromEnvironment() (string, error) {
	var stdout, stderr bytes.Buffer
	if err := bashListProjects(&stdout, &stderr); err != nil {
		return "", err
	}

	res, err := parseListProjectsCSV(&stdout)
	return res, err
}

func bashListProjects(stdout, stderr *bytes.Buffer) error {
	c := exec.Command("bash", "-c", "gcloud projects list --format='csv(PROJECT_ID)'")
	c.Stdout = stdout
	c.Stderr = stderr
	if err := c.Run(); err != nil {
		return fmt.Errorf("gcloud projects list: %w", err)
	}

	return nil
}

func parseListProjectsCSV(output *bytes.Buffer) (string, error) {
	r := csv.NewReader(output)
	columns, err := r.Read()
	if err != nil {
		return "", fmt.Errorf("no header when reading csv for project list: %w", err)
	}

	projectIndex := -1
	for i, column := range columns {
		if column == "project_id" {
			projectIndex = i
			break
		}
	}

	if projectIndex < 0 {
		return "", fmt.Errorf("no project_id in csv headers: %+v", columns)
	}

	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}

		if len(record) > projectIndex {
			if len(record[projectIndex]) > 0 {
				// Limit to only one
				return record[projectIndex], nil
			}
		}
	}

	return "", errors.New("not found")
}
