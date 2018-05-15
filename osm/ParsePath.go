package osm

import (
	"strings"
)

import (
	"github.com/pkg/errors"
)

// ParsePath splits a path into a top directory and child path
// Returns the top directory and child path, or an error if any.
func ParsePath(path string) (string, string, error) {
	if !strings.Contains(path, "/") {
		return "", "", errors.New("Path does not include a directory.")
	}
	parts := strings.Split(path, "/")
	return parts[0], strings.Join(parts[1:], "/"), nil
}
