// +build !js
package osm

import (
	"strings"
)

import (
	"github.com/aws/aws-sdk-go/service/s3"
	//"github.com/mitchellh/go-homedir"
	"github.com/colinmarc/hdfs"
	"github.com/pkg/errors"
)

import (
	"github.com/spatialcurrent/go-dfl/dfl"
	"github.com/spatialcurrent/go-reader/reader"
)

// Input is a struct for holding all the configuration describing an input destination
type Input struct {
	*PlanetResource `hcl:"resource"`
	Reader          reader.ByteReadCloser `hcl:"-"`
}

func (i *Input) Init(globals map[string]interface{}, ctx map[string]interface{}, funcs *dfl.FunctionMap) error {

	err := i.PlanetResource.Init(globals, ctx, funcs)
	if err != nil {
		return err
	}

	return nil
}

func (i *Input) Open(read_buffer_size int, s3_client *s3.S3, hdfs_clients map[string]*hdfs.Client) error {

	if i.Uri == "stdin" {

		r, err := reader.OpenStdin("none", false)
		if err != nil {
			return errors.Wrap(err, "error reading from stdin")
		}
		i.Reader = r

	} else {
		switch i.GetType() {
		case "file":
			return i.OpenFile(read_buffer_size)
		case "web":
			return i.OpenWeb()
		case "hdfs":
			return i.OpenFileOnHDFS(hdfs_clients[i.NameNode], read_buffer_size)
		case "s3":
			return i.OpenS3Object(s3_client)
		}
	}

	return nil

}

func (i *Input) OpenFile(read_buffer_size int) error {

	if strings.HasSuffix(i.PathExpanded, ".osm.gz") {

		r, err := reader.OpenFile(i.Path, "gzip", false, read_buffer_size)
		if err != nil {
			return errors.Wrap(err, "error opening file at path "+i.Path)
		}
		i.Reader = r

	} else if strings.HasSuffix(i.PathExpanded, ".osm.bz2") {

		r, err := reader.OpenFile(i.Path, "bzip2", false, read_buffer_size)
		if err != nil {
			return errors.Wrap(err, "error opening file at path "+i.Path)
		}
		i.Reader = r

	} else if strings.HasSuffix(i.PathExpanded, ".osm") {

		r, err := reader.OpenFile(i.Path, "none", false, read_buffer_size)
		if err != nil {
			return errors.Wrap(err, "error opening file at path "+i.Path)
		}
		i.Reader = r

	} else if strings.HasSuffix(i.PathExpanded, ".osm.pbf") {
		return errors.New("The OSM PBF format is not supported yet.")
	} else if strings.HasSuffix(i.PathExpanded, ".o5m") {
		return errors.New("The o5m format is not supported yet.")
	} else {
		return errors.New("Unknown file extension for input at " + i.Uri + ".")
	}

	return nil
}

func (i *Input) OpenWeb() error {

	if strings.HasSuffix(i.Uri, ".osm.gz") {

		r, _, err := reader.OpenHTTPFile(i.Uri, "gzip", false)
		if err != nil {
			return errors.Wrap(err, "error opening file at uri "+i.Uri)
		}
		i.Reader = r

	} else if strings.HasSuffix(i.Uri, ".osm.bz2") {

		r, _, err := reader.OpenHTTPFile(i.Uri, "bzip2", false)
		if err != nil {
			return errors.Wrap(err, "error opening file at uri "+i.Uri)
		}
		i.Reader = r

	} else if strings.HasSuffix(i.Uri, ".osm") {

		r, _, err := reader.OpenHTTPFile(i.Uri, "none", false)
		if err != nil {
			return errors.Wrap(err, "error opening file at uri "+i.Uri)
		}
		i.Reader = r

	} else if strings.HasSuffix(i.Uri, ".osm.pbf") {
		return errors.New("The OSM PBF format is not supported yet.")
	} else if strings.HasSuffix(i.Uri, ".o5m") {
		return errors.New("The o5m format is not supported yet.")
	} else {
		return errors.New("Unknown file extension for input at " + i.Uri + ".")
	}

	return nil
}

func (i *Input) OpenFileOnHDFS(hdfs_client *hdfs.Client, read_buffer_size int) error {

	if strings.HasSuffix(i.PathExpanded, ".osm.gz") {

		r, err := reader.OpenHDFSFile(i.Path, "gzip", false, hdfs_client)
		if err != nil {
			return errors.Wrap(err, "error opening gzip file on HDFS at path "+i.Path)
		}
		i.Reader = r

	} else if strings.HasSuffix(i.PathExpanded, ".osm.bz2") {

		r, err := reader.OpenHDFSFile(i.Path, "bzip2", false, hdfs_client)
		if err != nil {
			return errors.Wrap(err, "error opening gzip file on HDFS at path "+i.Path)
		}
		i.Reader = r

	} else if strings.HasSuffix(i.PathExpanded, ".osm") {

		r, err := reader.OpenHDFSFile(i.Path, "none", false, hdfs_client)
		if err != nil {
			return errors.Wrap(err, "error opening gzip file on HDFS at path "+i.Path)
		}
		i.Reader = r

	} else if strings.HasSuffix(i.PathExpanded, ".osm.pbf") {
		return errors.New("The OSM PBF format is not supported yet.")
	} else if strings.HasSuffix(i.PathExpanded, ".o5m") {
		return errors.New("The o5m format is not supported yet.")
	} else {
		return errors.New("Unknown file extension for input at " + i.Uri + ".")
	}

	return nil
}

func (i *Input) OpenS3Object(s3_client *s3.S3) error {

	if strings.HasSuffix(i.Key, ".osm.gz") {

		r, _, err := reader.OpenS3Object(i.Bucket, i.Key, "gzip", false, s3_client)
		if err != nil {
			return errors.Wrap(err, "error opening s3 object at s3://"+i.Bucket+"/"+i.Key)
		}
		i.Reader = r

	} else if strings.HasSuffix(i.Key, ".osm.bz2") {

		r, _, err := reader.OpenS3Object(i.Bucket, i.Key, "bzip2", false, s3_client)
		if err != nil {
			return errors.Wrap(err, "error opening s3 object at s3://"+i.Bucket+"/"+i.Key)
		}
		i.Reader = r

	} else if strings.HasSuffix(i.Key, ".osm") {

		r, _, err := reader.OpenS3Object(i.Bucket, i.Key, "none", false, s3_client)
		if err != nil {
			return errors.Wrap(err, "error opening s3 object at s3://"+i.Bucket+"/"+i.Key)
		}
		i.Reader = r

	} else if strings.HasSuffix(i.Key, ".osm.pbf") {
		return errors.New("The OSM PBF format is not supported yet.")
	} else if strings.HasSuffix(i.Key, ".o5m") {
		return errors.New("The o5m format is not supported yet.")
	} else {
		return errors.New("Unknown file extension for input at " + i.Uri + ".")
	}

	return nil
}

func (i *Input) Close() error {

	if i.Reader != nil {
		return i.Reader.Close()
	}

	return nil
}
