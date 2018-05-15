package s3util

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
)

// BucketExists checks in a bucket exists in AWS S3 with simplistic error checking
func BucketExists(s3_client *s3.S3, bucket string) bool {

	input := &s3.HeadBucketInput{
		Bucket: aws.String(bucket),
	}

	_, err := s3_client.HeadBucket(input)
	if err != nil {
		return false
	}

	return true
}
