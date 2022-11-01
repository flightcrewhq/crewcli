package timeconv_test

import (
	"testing"
	"time"

	"flightcrew.io/cli/internal/timeconv"
	"github.com/stretchr/testify/assert"
)

func TestFormatDurationPasses(t *testing.T) {
	convert := timeconv.GetDurationFormatter([]string{"m", "h", "s"})
	type testCase struct {
		input    time.Duration
		expected string
	}

	for _, tc := range []testCase{
		{time.Duration(0), ""},
		{time.Duration(15 * time.Second), `15s`},
		{time.Duration(15 * time.Minute), `15m`},
		{time.Duration(15 * time.Hour), `15h`},
		{time.Duration(64 * time.Second), `1m4s`},
		{time.Duration(64 * time.Minute), `1h4m`},
		{time.Duration(10*time.Second + (24*7*15+15)*time.Hour + time.Minute), `2535h1m10s`},
	} {
		actual, err := convert(tc.input)
		assert.NoError(t, err)
		assert.Equal(t, tc.expected, actual)
	}
}

func TestFormatDurationFails(t *testing.T) {
	convert := timeconv.GetDurationFormatter([]string{"m", "h", "s"})

	for _, tc := range []time.Duration{
		time.Duration(1 * time.Microsecond),
		time.Duration(1*time.Minute + time.Nanosecond),
	} {
		_, err := convert(tc)
		assert.Error(t, err)
	}
}
