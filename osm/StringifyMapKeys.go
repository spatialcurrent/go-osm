package osm

import (
	"fmt"
)

// https://github.com/gohugoio/hugo/pull/4138
func StringifyMapKeys(in interface{}) interface{} {
	switch in := in.(type) {
	case []interface{}:
		res := make([]interface{}, len(in))
		for i, v := range in {
			res[i] = StringifyMapKeys(v)
		}
		return res
	case map[interface{}]interface{}:
		res := make(map[string]interface{})
		for k, v := range in {
			res[fmt.Sprintf("%v", k)] = StringifyMapKeys(v)
		}
		return res
	default:
		return in
	}
}
