package osm

import (
	"os"
)

import (
	"github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"
)

var SUPPORTED_SCHEMES = []string{
	"file",
	"http",
	"https",
	"s3",
	"hdfs",
}

// Resource is an abstract struct for URI addressable resources
type Resource struct {
	Uri          string `hcl:"uri"` // resoure URI
	Type         string `hcl:"-"`   // type of resource inferred from uri
	Scheme       string `hcl:"-"`   // scheme, .e.g., http, https, file, s3
	Path         string `hcl:"-"`   // path
	PathExpanded string `hcl:"-"`   // path with home directory expanded
	NameNode     string `hcl:"-"`   // FQDN of Name Node
	Bucket       string `hcl:"-"`   // S3 Bucket
	Key          string `hcl:"-"`   // S3 Key
	Exists       bool   `hcl:"-"`   // resource exists
}

func (r *Resource) GetType() string {
	return r.Type
}

func (r *Resource) IsType(t string) bool {
	return r.Type == t
}

func (r *Resource) Init(ctx map[string]interface{}) error {

	uri, err := RenderTemplate(r.Uri, ctx)
	if err != nil {
		return errors.Wrap(err, "Error rendering uri template "+r.Uri)
	}
	r.Uri = uri

	if r.Uri == "stdin" || r.Uri == "stdout" || r.Uri == "stderr" {

		r.Type = "stream"

	} else {
		scheme, fullpath := SplitUri(r.Uri, SUPPORTED_SCHEMES)
		r.Scheme = scheme

		if scheme == "file" {

			r.Type = "file"
			r.Path = fullpath
			p, err := homedir.Expand(r.Path)
			if err != nil {
				return errors.Wrap(err, "Error expanding resource file path")
			}
			r.PathExpanded = p

		} else if scheme == "hdfs" {

			r.Type = "hdfs"
			nameNode, path, err := ParsePath(fullpath)
			if err != nil {
				return errors.Wrap(err, "Error parsing HDFS path")
			}
			r.NameNode = nameNode
			r.Path = path
			r.PathExpanded = r.Path

		} else if scheme == "s3" {

			r.Type = "s3"
			b, k, err := ParsePath(r.Path)
			if err != nil {
				return errors.Wrap(err, "Error parsing AWS S3 path")
			}
			r.Bucket = b
			r.Key = k

		} else {
			return errors.New("Unknown resource scheme " + scheme)
		}
	}

	return nil
}

func (r *Resource) FileExists() bool {
	if _, err := os.Stat(r.PathExpanded); os.IsNotExist(err) {
		return false
	} else {
		return true
	}
}

func (r Resource) HasUri() bool {
	return len(r.Uri) > 0
}
