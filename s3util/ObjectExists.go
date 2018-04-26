package s3util

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
)

func ObjectExists(s3_client *s3.S3, bucket string, key string) bool {

	input := &s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}

	_, err := s3_client.HeadObject(input)
	if err != nil {
		return false
	}

	return true
}
