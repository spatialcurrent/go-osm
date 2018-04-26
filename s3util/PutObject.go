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
