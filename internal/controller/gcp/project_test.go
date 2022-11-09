package gcp

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseprojectCSV(mainT *testing.T) {
	mainT.Run("parse valid csv should succeed", func(t *testing.T) {
		buf := bytes.NewBufferString(`project_id
project-1
some-project
yeah
123-project
`)
		project, err := parseListProjectsCSV(buf)
		assert.NoError(t, err)
		assert.Equal(t, "project-1", project)
	})

	mainT.Run("parse empty csv should fail", func(t *testing.T) {
		buf := bytes.NewBufferString("")
		_, err := parseListProjectsCSV(buf)
		assert.Error(t, err)
	})

	mainT.Run("parse valid empty csv should error", func(t *testing.T) {
		buf := bytes.NewBufferString(`project_id
`)
		_, err := parseListProjectsCSV(buf)
		assert.Error(t, err)
	})

	mainT.Run("parse variable entries should succeed", func(t *testing.T) {
		buf := bytes.NewBufferString(`something,project_id
,project-1
123,project-2
	`)
		project, err := parseListProjectsCSV(buf)
		assert.NoError(t, err)
		assert.Equal(t, "project-1", project)
	})

	mainT.Run("parse invalid csv should error", func(t *testing.T) {
		buf := bytes.NewBufferString(`ERROR: imagine something went wrong`)
		_, err := parseListProjectsCSV(buf)
		assert.Error(t, err)
	})
}
