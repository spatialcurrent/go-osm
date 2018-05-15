package s3util

import (
	"bytes"
	"compress/gzip"
	"net/http"
	"strings"
)

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
)

import (
	"github.com/pkg/errors"
)

// PutObject uploads a file to S3 at the given bucket and key
// If the key ends in ".gz", then the contents are compressed before upload.
// If the key ends in ".bz2", then errors out as Go does not support writing to bzip2 (see https://github.com/golang/go/issues/4828)
// Returns an error if any.
func PutObject(s3_client *s3.S3, bucket string, key string, data []byte) error {

	input := &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}

	if strings.HasSuffix(key, ".gz") {

		var buf bytes.Buffer
		gw := gzip.NewWriter(&buf)
		gw.Write(data)
		gw.Close()
		body := buf.Bytes()
		input.Body = bytes.NewReader(body)
		input.ContentLength = aws.Int64(int64(len(body)))
		input.ContentType = aws.String(http.DetectContentType(body))

	} else if strings.HasSuffix(key, ".bz2") {

		return errors.New("Go does not support writing bzip2 files.")

	} else {

		input.Body = bytes.NewReader(data)
		input.ContentLength = aws.Int64(int64(len(data)))
		input.ContentType = aws.String(http.DetectContentType(data))

	}

	_, err := s3_client.PutObject(input)
	if err != nil {
		return err
	}

	//fmt.Println("Put Object Result:", result)

	return nil

}
