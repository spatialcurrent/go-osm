package osm

import (
	"strings"
)

func ParseSliceString(in string) []string {
	out := make([]string, 0)
	if len(in) > 0 {
		out = strings.Split(in, ",")
	}
	return out
}
