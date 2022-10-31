package timeconv_test

import (
	"testing"
	"time"

	"flightcrew.io/cli/internal/timeconv"
	"github.com/stretchr/testify/assert"
)

func TestParseDurationPasses(t *testing.T) {
	type testCase struct {
		input    string
		expected time.Duration
	}

	for _, tc := range []testCase{
		{"-1.5ns", time.Duration(-1)},
		{"-1.5us", time.Duration(-1.5 * float64(time.Microsecond))},
		{"-1.5µs", time.Duration(-1.5 * float64(time.Microsecond))},
		{"-1.5μs", time.Duration(-1.5 * float64(time.Microsecond))},
		{"-1.5ms", time.Duration(-1.5 * float64(time.Millisecond))},
		{"-1.5s", time.Duration(-1.5 * float64(time.Second))},
		{"-1.5m", time.Duration(-1.5 * float64(time.Minute))},
		{"-1.5h", time.Duration(-1.5 * float64(time.Hour))},
		{"-1.5d", time.Duration(-1.5 * float64(time.Hour*24))},
		{"-1.5w", time.Duration(-1.5 * float64(time.Hour*24*7))},
		{"-1.5mo", time.Duration(-1.5 * float64(time.Hour*24*7*30))},
		{"1w2d3h4m5s", time.Duration(time.Hour*((7+2)*24) + time.Hour*3 + time.Minute*4 + time.Second*5)},
		{"5s30m10s12s", time.Duration((5+10+12)*time.Second + time.Minute*30)},
	} {
		actual, err := timeconv.ParseDuration(tc.input)
		assert.NoError(t, err)
		assert.Equal(t, tc.expected, actual)
	}
}

func TestParseDurationFails(t *testing.T) {
	for _, tc := range []string{
		"notaduration",
		"--15.12h",
		"",
	} {
		_, err := timeconv.ParseDuration(tc)
		assert.Error(t, err)
	}
}
