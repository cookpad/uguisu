package service_test

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"io"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/cookpad/uguisu/pkg/adaptor"
	"github.com/cookpad/uguisu/pkg/models"
	"github.com/cookpad/uguisu/pkg/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// stubS3Client returns an S3ClientFactory that serves a single object at the
// given bucket/key with the provided raw bytes as the body.
func stubS3Client(bucket, key string, body []byte) adaptor.S3ClientFactory {
	return func(region string) (adaptor.S3Client, error) {
		return &fixedS3Client{bucket: bucket, key: key, body: body}, nil
	}
}

type fixedS3Client struct {
	bucket string
	key    string
	body   []byte
}

func (c *fixedS3Client) GetObject(_ context.Context, input *s3.GetObjectInput, _ ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
	if *input.Bucket != c.bucket {
		return nil, &types.NoSuchBucket{}
	}
	if *input.Key != c.key {
		return nil, &types.NoSuchKey{}
	}
	return &s3.GetObjectOutput{Body: io.NopCloser(bytes.NewReader(c.body))}, nil
}

func (c *fixedS3Client) PutObject(_ context.Context, _ *s3.PutObjectInput, _ ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
	return nil, errors.New("not implemented")
}

// gzipJSON compresses a CloudTrailLogObject to gzip-encoded JSON bytes.
func gzipJSON(t *testing.T, obj models.CloudTrailLogObject) []byte {
	t.Helper()
	raw, err := json.Marshal(obj)
	require.NoError(t, err)
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	_, err = gz.Write(raw)
	require.NoError(t, err)
	require.NoError(t, gz.Close())
	return buf.Bytes()
}

// plainJSON encodes a CloudTrailLogObject as plain (uncompressed) JSON bytes.
func plainJSON(t *testing.T, obj models.CloudTrailLogObject) []byte {
	t.Helper()
	raw, err := json.Marshal(obj)
	require.NoError(t, err)
	return raw
}

func TestCloudTrailLogs_ReadsGzipObject(t *testing.T) {
	records := []*models.CloudTrailRecord{
		{EventName: "ConsoleLogin", AwsRegion: "us-east-1"},
		{EventName: "CreateBucket", AwsRegion: "ap-northeast-1"},
	}
	body := gzipJSON(t, models.CloudTrailLogObject{Records: records})

	svc := service.NewCloudTrailLogs(stubS3Client("my-bucket", "logs/test.json.gz", body))
	got, err := svc.Read("us-east-1", "my-bucket", "logs/test.json.gz")

	require.NoError(t, err)
	require.Len(t, got, 2)
	assert.Equal(t, "ConsoleLogin", got[0].EventName)
	assert.Equal(t, "CreateBucket", got[1].EventName)
}

func TestCloudTrailLogs_ReadsPlainJSONObject(t *testing.T) {
	records := []*models.CloudTrailRecord{
		{EventName: "PutObject", AwsRegion: "eu-west-1"},
	}
	body := plainJSON(t, models.CloudTrailLogObject{Records: records})

	svc := service.NewCloudTrailLogs(stubS3Client("my-bucket", "logs/test.json", body))
	got, err := svc.Read("eu-west-1", "my-bucket", "logs/test.json")

	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, "PutObject", got[0].EventName)
}

func TestCloudTrailLogs_ReturnsEmptySliceForNoRecords(t *testing.T) {
	body := gzipJSON(t, models.CloudTrailLogObject{Records: nil})

	svc := service.NewCloudTrailLogs(stubS3Client("my-bucket", "logs/empty.json.gz", body))
	got, err := svc.Read("us-east-1", "my-bucket", "logs/empty.json.gz")

	require.NoError(t, err)
	assert.Empty(t, got)
}

func TestCloudTrailLogs_ErrorOnMissingBucket(t *testing.T) {
	body := gzipJSON(t, models.CloudTrailLogObject{})
	svc := service.NewCloudTrailLogs(stubS3Client("other-bucket", "logs/test.json.gz", body))

	_, err := svc.Read("us-east-1", "my-bucket", "logs/test.json.gz")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to download CloudTrail log object")
}

func TestCloudTrailLogs_ErrorOnMissingKey(t *testing.T) {
	body := gzipJSON(t, models.CloudTrailLogObject{})
	svc := service.NewCloudTrailLogs(stubS3Client("my-bucket", "logs/other.json.gz", body))

	_, err := svc.Read("us-east-1", "my-bucket", "logs/test.json.gz")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to download CloudTrail log object")
}

func TestCloudTrailLogs_ErrorOnInvalidGzip(t *testing.T) {
	svc := service.NewCloudTrailLogs(stubS3Client("my-bucket", "logs/bad.json.gz", []byte("not gzip data")))

	_, err := svc.Read("us-east-1", "my-bucket", "logs/bad.json.gz")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create gzip reader for CloudTrail log")
}

func TestCloudTrailLogs_ErrorOnInvalidJSON(t *testing.T) {
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	_, err := gz.Write([]byte("not valid json"))
	require.NoError(t, err)
	require.NoError(t, gz.Close())

	svc := service.NewCloudTrailLogs(stubS3Client("my-bucket", "logs/bad.json.gz", buf.Bytes()))

	_, err = svc.Read("us-east-1", "my-bucket", "logs/bad.json.gz")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to decode CloudTrail logs")
}

func TestCloudTrailLogs_PassesRegionToS3Factory(t *testing.T) {
	var capturedRegion string
	records := []*models.CloudTrailRecord{{EventName: "DescribeInstances"}}
	body := gzipJSON(t, models.CloudTrailLogObject{Records: records})

	factory := func(region string) (adaptor.S3Client, error) {
		capturedRegion = region
		return &fixedS3Client{bucket: "my-bucket", key: "logs/test.json.gz", body: body}, nil
	}

	svc := service.NewCloudTrailLogs(factory)
	_, err := svc.Read("ap-southeast-2", "my-bucket", "logs/test.json.gz")

	require.NoError(t, err)
	assert.Equal(t, "ap-southeast-2", capturedRegion)
}

func TestCloudTrailLogs_ErrorOnFactoryFailure(t *testing.T) {
	factory := func(region string) (adaptor.S3Client, error) {
		return nil, errors.New("no credentials")
	}

	svc := service.NewCloudTrailLogs(factory)
	_, err := svc.Read("us-east-1", "my-bucket", "logs/test.json.gz")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "no credentials")
}
