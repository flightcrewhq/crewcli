package timeconv

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

// byDurationDesc implements sort.Interface for []string based on
// the duration associated with the unit in descending order.
type byDurationDesc []string

func (s byDurationDesc) Len() int           { return len(s) }
func (s byDurationDesc) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s byDurationDesc) Less(i, j int) bool { return unitMap[s[j]] < unitMap[s[i]] }

func GetDurationFormatter(units []string) func(time.Duration) (string, error) {
	sort.Sort(byDurationDesc(units))

	return func(d time.Duration) (string, error) {
		var converted strings.Builder
		for _, unit := range units {
			unitDur := unitMap[unit]
			numUnits := int64(d / time.Duration(unitDur))
			if numUnits > 0 {
				converted.WriteString(fmt.Sprintf("%d%s", numUnits, unit))
			}
			d = d - time.Duration(numUnits)*time.Duration(unitDur)
		}

		if d > time.Duration(0) {
			return "", fmt.Errorf("units smaller than seconds (s) are not supported")
		}

		return converted.String(), nil
	}
}
