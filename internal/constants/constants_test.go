package constants_test

import (
	"testing"

	"flightcrew.io/cli/internal/constants"
	"github.com/stretchr/testify/assert"
)

func TestGetPlatformKey(t *testing.T) {
	assert.Equal(t, constants.GoogleAppEngineStdKey, constants.GetPlatformKey(constants.GoogleAppEngineStdKey))
	assert.Equal(t, constants.GoogleAppEngineStdKey, constants.GetPlatformKey(constants.GoogleAppEngineStdDisplay))
	assert.Equal(t, constants.GoogleAppEngineStdKey, constants.GetPlatformKey(constants.GoogleAppEngineStdPlatform))

	assert.Equal(t, constants.GoogleComputeEngineKey, constants.GetPlatformKey(constants.GoogleComputeEngineKey))
	assert.Equal(t, constants.GoogleComputeEngineKey, constants.GetPlatformKey(constants.GoogleComputeEngineDisplay))
	assert.Equal(t, constants.GoogleComputeEngineKey, constants.GetPlatformKey(constants.GoogleComputeEnginePlatform))
}
