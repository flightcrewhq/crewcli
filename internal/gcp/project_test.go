package gcp

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseProjectsCSV(mainT *testing.T) {
	mainT.Run("parse valid csv should succeed", func(t *testing.T) {
		buf := bytes.NewBufferString(`project_id
project-1
some-project
yeah
123-project
`)
		projects, err := parseListProjectsCSV(buf)
		assert.NoError(t, err)
		assert.ElementsMatch(t, []string{
			"project-1",
			"some-project",
			"yeah",
			"123-project",
		}, projects)
	})

	mainT.Run("parse empty csv should fail", func(t *testing.T) {
		buf := bytes.NewBufferString("")
		_, err := parseListProjectsCSV(buf)
		assert.Error(t, err)
	})

	mainT.Run("parse valid empty csv should succeed", func(t *testing.T) {
		buf := bytes.NewBufferString(`project_id
`)
		projects, err := parseListProjectsCSV(buf)
		assert.NoError(t, err)
		assert.Empty(t, projects)
	})

	mainT.Run("parse variable entries should succeed", func(t *testing.T) {
		buf := bytes.NewBufferString(`something,project_id
,project-1
123,project-2
	`)
		projects, err := parseListProjectsCSV(buf)
		assert.NoError(t, err)
		assert.ElementsMatch(t, []string{
			"project-1",
			"project-2",
		}, projects)
	})

	mainT.Run("parse invalid csv should error", func(t *testing.T) {
		buf := bytes.NewBufferString(`ERROR: imagine something went wrong`)
		_, err := parseListProjectsCSV(buf)
		assert.Error(t, err)
	})
}
