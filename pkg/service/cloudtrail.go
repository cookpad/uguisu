package service

import (
	"encoding/json"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/m-mizutani/golambda"
	"github.com/m-mizutani/uguisu/pkg/adaptor"
	"github.com/m-mizutani/uguisu/pkg/models"
)

type CloudTrailLogs struct {
	NewS3 adaptor.S3ClientFactory
}

func NewCloudTrailLogs(newS3 adaptor.S3ClientFactory) *CloudTrailLogs {
	return &CloudTrailLogs{
		NewS3: newS3,
	}
}

func (x *CloudTrailLogs) Read(s3Region, s3Bucket, s3Key string) ([]*models.CloudTrailRecord, error) {
	s3Client, err := x.NewS3(s3Region)
	if err != nil {
		return nil, err
	}

	input := &s3.GetObjectInput{
		Bucket: aws.String(s3Bucket),
		Key:    aws.String(s3Key),
	}

	output, err := s3Client.GetObject(input)
	if err != nil {
		return nil, golambda.WrapError(err, "Failed to download cloudtrail log object").With("input", input)
	}

	decoder := json.NewDecoder(output.Body)
	var object models.CloudTrailLogObject
	if err := decoder.Decode(&object); err != nil {
		return nil, golambda.WrapError(err, "Failed to decode CloudTrail logs").With("input", input)
	}

	return object.Records, nil
}
