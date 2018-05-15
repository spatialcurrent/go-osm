package osm

import (
	"strings"
)

// SplitUri splits a uri string and returns the scheme and path.
// If no scheme is specified, then returns scheme "file" and the input uri as the path.
func SplitUri(uri string, schemes []string) (string, string) {
	for _, scheme := range schemes {
		if strings.HasPrefix(strings.ToLower(uri), scheme+"://") {
			return scheme, uri[len(scheme+"://"):]
		}
	}
	return "file", uri
}
