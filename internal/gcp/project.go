package gcp

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"os/exec"
)

func GetProjectsFromEnvironment() ([]string, error) {
	var stdout, stderr bytes.Buffer
	if err := bashListProjects(&stdout, &stderr); err != nil {
		return nil, err
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

func parseListProjectsCSV(output *bytes.Buffer) ([]string, error) {
	projectIDs := make([]string, 0)
	r := csv.NewReader(output)
	columns, err := r.Read()
	if err != nil {
		return nil, fmt.Errorf("no header when reading csv for project list: %w", err)
	}

	projectIndex := -1
	for i, column := range columns {
		if column == "project_id" {
			projectIndex = i
			break
		}
	}

	if projectIndex < 0 {
		return nil, fmt.Errorf("no project_id in csv headers: %+v", columns)
	}

	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}

		if len(record) > projectIndex {
			projectIDs = append(projectIDs, record[projectIndex])
		}
	}

	return projectIDs, nil
}
