package gcp_test

import (
	"testing"

	"flightcrew.io/cli/internal/controller/gcp"
	"github.com/stretchr/testify/assert"
)

func TestGetAPIHostName(t *testing.T) {
	assert.Equal(t, "flightcrew.io", gcp.GetHostBaseURL("", ""))
	assert.Equal(t, "flightcrew.io", gcp.GetHostBaseURL("", "something-dev"))
	assert.Equal(t, "flightcrew.io", gcp.GetHostBaseURL("some-project", "something-dev"))
	assert.Equal(t, "flightcrew.io", gcp.GetHostBaseURL("yay-yay-yay", "something-dev"))
}
