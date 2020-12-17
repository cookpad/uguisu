package uguisu

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/google/uuid"
	"github.com/m-mizutani/golambda"
	"github.com/m-mizutani/uguisu/pkg/mock"
	"github.com/m-mizutani/uguisu/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func putData(client *mock.S3Client, region, bucket, key string, records []*models.CloudTrailRecord) {
	data := models.CloudTrailLogObject{
		Records: records,
	}

	raw, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}

	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	gz.Write(raw)
	gz.Close()

	_, err = client.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   bytes.NewReader(buf.Bytes()),
	})
	if err != nil {
		panic(err)
	}
}

func TestUguisuBasic(t *testing.T) {
	newS3, s3Client := mock.NewS3Mock()
	httpClient := &mock.HTTPClient{}

	ug := New()
	ug.NewS3 = newS3
	ug.HTTPClient = httpClient
	ug.SlackWebhookURL = "https://test.example.com/endpoint"

	s3Region := "us-east-0"
	s3Bucket := "your-ct-logs"
	s3Key := "some/object/" + uuid.New().String()

	putData(s3Client, s3Region, s3Bucket, s3Key, []*models.CloudTrailRecord{
		{
			ErrorCode: aws.String("UnauthorizedOperation"),
		},
	})

	var event golambda.Event
	event.EncapSNSonSQSMessage(events.S3Event{
		Records: []events.S3EventRecord{
			{
				AWSRegion: s3Region,
				S3: events.S3Entity{
					Bucket: events.S3Bucket{Name: s3Bucket},
					Object: events.S3Object{Key: s3Key},
				},
			},
		},
	})

	require.NoError(t, ug.run(event))
	require.Equal(t, 1, len(httpClient.Requests))
	assert.Equal(t, "test.example.com", httpClient.Requests[0].URL.Host)
	assert.Equal(t, "/endpoint", httpClient.Requests[0].URL.Path)

	sentData, err := ioutil.ReadAll(httpClient.Requests[0].Body)
	require.NoError(t, err)
	assert.Contains(t, string(sentData), "AWS CIS benchmark 3.1 ")
}
