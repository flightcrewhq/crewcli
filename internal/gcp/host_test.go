package gcp_test

import (
	"testing"

	"flightcrew.io/cli/internal/gcp"
	"github.com/stretchr/testify/assert"
)

func TestGetAPIHostName(t *testing.T) {
	assert.Equal(t, "api.flightcrew.io", gcp.GetAPIHostName("", ""))
	assert.Equal(t, "api.flightcrew.io", gcp.GetAPIHostName("", "something-dev"))
	assert.Equal(t, "api.flightcrew.io", gcp.GetAPIHostName("some-project", "something-dev"))
	assert.Equal(t, "api.flightcrew.io", gcp.GetAPIHostName("yay-yay-yay", "something-dev"))
}
