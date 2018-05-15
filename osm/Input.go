package osm

import (
	"bufio"
	"bytes"
	"compress/bzip2"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"strings"
)

import (
	"github.com/aws/aws-sdk-go/service/s3"
	//"github.com/mitchellh/go-homedir"
	"github.com/colinmarc/hdfs"
	"github.com/pkg/errors"
)

import (
	"github.com/spatialcurrent/go-osm/s3util"
)

import (
	"github.com/spatialcurrent/go-dfl/dfl"
)

// Input is a struct for holding all the configuration describing an input destination
type Input struct {
	*PlanetResource `hcl:"resource"`
	File            *os.File  `hcl:"-"`
	Reader          io.Reader `hcl:"-"`
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
		i.Reader = bufio.NewReader(os.Stdin)
	} else {
		switch i.GetType() {
		case "file":
			return i.OpenFile(read_buffer_size)
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

		f, err := os.Open(i.PathExpanded)
		if err != nil {
			return errors.New("Error opening input file at " + i.Uri + ".")
		}
		i.File = f

		gr, err := gzip.NewReader(bufio.NewReaderSize(i.File, read_buffer_size))
		if err != nil {
			return errors.New("Error creating gzip reader for file at " + i.Uri + ".")
		}
		i.Reader = gr

	} else if strings.HasSuffix(i.PathExpanded, ".osm.bz2") {

		f, err := os.Open(i.PathExpanded)
		if err != nil {
			return errors.New("Error opening input file at " + i.Uri + ".")
		}
		i.File = f
		i.Reader = bzip2.NewReader(bufio.NewReaderSize(i.File, read_buffer_size))

	} else if strings.HasSuffix(i.PathExpanded, ".osm") {

		f, err := os.Open(i.PathExpanded)
		if err != nil {
			fmt.Println("Error opening input file at " + i.Uri + ".")
			fmt.Println(err)
			os.Exit(1)
		}
		i.File = f
		i.Reader = bufio.NewReaderSize(i.File, read_buffer_size)

	} else if strings.HasSuffix(i.PathExpanded, ".osm.pbf") {
		return errors.New("The OSM PBF format is not supported yet.")
	} else if strings.HasSuffix(i.PathExpanded, ".o5m") {
		return errors.New("The o5m format is not supported yet.")
	} else {
		return errors.New("Unknown file extension for input at " + i.Uri + ".")
	}

	return nil
}

func (i *Input) OpenFileOnHDFS(hdfs_client *hdfs.Client, read_buffer_size int) error {

	if strings.HasSuffix(i.PathExpanded, ".osm.gz") {

		fileReader, err := hdfs_client.Open(i.PathExpanded)
		if err != nil {
			return errors.New("Error opening input file at " + i.Uri + ".")
		}

		gr, err := gzip.NewReader(fileReader)
		if err != nil {
			return errors.New("Error creating gzip reader for file at " + i.Uri + ".")
		}
		i.Reader = gr

	} else if strings.HasSuffix(i.PathExpanded, ".osm.bz2") {

		fileReader, err := hdfs_client.Open(i.PathExpanded)
		if err != nil {
			return errors.New("Error opening input file at " + i.Uri + ".")
		}
		i.Reader = bzip2.NewReader(fileReader)

	} else if strings.HasSuffix(i.PathExpanded, ".osm") {

		fileReader, err := hdfs_client.Open(i.PathExpanded)
		if err != nil {
			return errors.Wrap(err, "Error opening input file at "+i.Uri+".")
		}
		i.Reader = bufio.NewReaderSize(fileReader, read_buffer_size)

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

	if strings.HasSuffix(i.Key, ".osm.gz") || strings.HasSuffix(i.Key, ".osm.bz2") || strings.HasSuffix(i.Key, ".osm") {

		in, err := s3util.GetObject(s3_client, i.Bucket, i.Key)
		if err != nil {
			return errors.Wrap(err, "Error reading from AWS S3 uri "+i.Uri+".")
		}
		i.Reader = bytes.NewReader(in)

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
	if i.File != nil {
		return i.File.Close()
	}
	return nil
}
