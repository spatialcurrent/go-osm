package s3util

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
)

func DeleteObject(s3_client *s3.S3, bucket string, key string) error {
	input := &s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}

	_, err := s3_client.DeleteObject(input)
	if err != nil {
		return err
	}

	return nil
}
