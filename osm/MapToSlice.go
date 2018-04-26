package osm

import (
	"fmt"
)

func MapToSlice(m map[string]interface{}, keys []string) []string {
	s := make([]string, len(keys))
	for i, k := range keys {
		if v, ok := m[k]; ok {
			s[i] = fmt.Sprint(v)
		}
	}
	return s
}
