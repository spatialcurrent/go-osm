package s3util

import (
	"compress/bzip2"
	"compress/gzip"
	"io/ioutil"
	"strings"
)

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
)

import (
	"github.com/pkg/errors"
)

// GetObject retrieves the contents of the object in S3 at the provided bucket and key
// The contents of the file are read in full.
// If the key ends in ".gz" or ".bz2", the contents are automatically uncompressed.
func GetObject(s3_client *s3.S3, bucket string, key string) ([]byte, error) {
	input := &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}

	result, err := s3_client.GetObject(input)
	if err != nil {
		return make([]byte, 0), errors.Wrap(err, "Error fetching data from S3")
	}

	if strings.HasSuffix(key, ".gz") {

		gr, err := gzip.NewReader(result.Body)
		if err != nil {
			return make([]byte, 0), errors.Wrap(err, "Error creating gizp reader for AWS s3 object at s3://" + bucket + "/" + key + ".")
		}
		defer gr.Close()

		obj, err := ioutil.ReadAll(gr)
		if err != nil {
			return make([]byte, 0), errors.Wrap(err, "ERror reading from gzip reader for AWS s3 object at s3://" + bucket + "/" + key + ".")
		}
		return obj, nil
	}

	if strings.HasSuffix(key, ".bz2") {

		br := bzip2.NewReader(result.Body)

		obj, err := ioutil.ReadAll(br)
		if err != nil {
			return make([]byte, 0), err
		}
		return obj, nil
	}

	obj, err := ioutil.ReadAll(result.Body)
	if err != nil {
		return make([]byte, 0), err
	}

	return obj, nil

}
