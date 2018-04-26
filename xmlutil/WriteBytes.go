package xmlutil

import (
  "encoding/xml"
)

import (
  "github.com/pkg/errors"
)

import (
	"github.com/aws/aws-sdk-go/service/s3"
)

import (
  "github.com/spatialcurrent/go-osm/s3util"
)


func WriteBytes(output_uri string, output_scheme string, output_path string, output_bytes []byte, s3_client *s3.S3, output_s3_bucket string, output_s3_key string) error {

  if output_uri == "stdout" {
    WriteToStdout(output_bytes)
    return nil
  }

  if output_uri == "stderr" {
    WriteToStderr(output_bytes)
    return nil
  }

  if output_scheme == "s3" {
    err := s3util.PutObject(
      s3_client,
      output_s3_bucket,
      output_s3_key,
      append(append([]byte(xml.Header), output_bytes...), []byte("\n")...))
    if err != nil {
      return errors.Wrap(err, "Error uploading object to AWS S3 at output location " + output_uri + ".")
    }
    return nil
  }

  if output_scheme == "file" {
    err := WriteToDisk(output_path, output_bytes)
    if err != nil {
      return errors.Wrap(err, "Error writing xml file to disk at " + output_uri + ".")
    }
    return nil

  }

  return nil

}
