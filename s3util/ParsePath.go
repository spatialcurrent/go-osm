package s3util

import (
	"strings"
)

import (
	"github.com/pkg/errors"
)

func ParsePath(path string) (string, string, error) {
	if !strings.Contains(path, "/") {
		return "", "", errors.New("AWS S3 path does not include bucket.")
	}
	parts := strings.Split(path, "/")
	return parts[0], strings.Join(parts[1:], "/"), nil
}
