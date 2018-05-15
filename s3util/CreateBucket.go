package s3util

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
)

// CreateBucket creates a bucket in AWS S3 with simplistic error checking
func CreateBucket(s3_client *s3.S3, region string, bucket string) error {
	input := &s3.CreateBucketInput{
		Bucket: aws.String(bucket),
		CreateBucketConfiguration: &s3.CreateBucketConfiguration{
			LocationConstraint: aws.String(region),
		},
	}

	_, err := s3_client.CreateBucket(input)
	if err != nil {
		return err
	}

	return nil

}
