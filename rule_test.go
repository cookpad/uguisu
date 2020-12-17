package uguisu

import (
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/google/uuid"
	"github.com/m-mizutani/golambda"
	"github.com/m-mizutani/uguisu/pkg/mock"
	"github.com/m-mizutani/uguisu/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func runRuleTest(records []*models.CloudTrailRecord) ([]string, *mock.HTTPClient) {
	newS3, s3Client := mock.NewS3Mock()
	httpClient := &mock.HTTPClient{}

	ug := New()
	ug.NewS3 = newS3
	ug.HTTPClient = httpClient
	ug.SlackWebhookURL = "https://test.example.com/endpoint"

	s3Region := "us-east-0"
	s3Bucket := "your-ct-logs"
	s3Key := "some/object/" + uuid.New().String()

	eventIDs := make([]string, len(records))
	for i := range records {
		eventIDs[i] = uuid.New().String()
		records[i].EventID = eventIDs[i]
	}

	putData(s3Client, s3Region, s3Bucket, s3Key, records)

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

	if err := ug.Run(event); err != nil {
		panic(err)
	}

	return eventIDs, httpClient
}

func TestAwsCIS3_1(t *testing.T) {
	t.Run("detect with UnauthorizedOperation", func(t *testing.T) {
		eventIDs, httpClient := runRuleTest([]*models.CloudTrailRecord{
			{
				ErrorCode: aws.String("SomeUnauthorizedOperation"),
			},
		})
		require.Equal(t, 1, httpClient.RequestNum())
		assert.Contains(t, httpClient.Body(0), eventIDs[0])
	})

	t.Run("detect with AccessDenied", func(t *testing.T) {
		eventIDs, httpClient := runRuleTest([]*models.CloudTrailRecord{
			{
				ErrorCode: aws.String("AccessDeniedSomething"),
			},
		})
		require.Equal(t, 1, httpClient.RequestNum())
		assert.Contains(t, httpClient.Body(0), eventIDs[0])
	})

	t.Run("not detect when errorCode is nil", func(t *testing.T) {
		_, httpClient := runRuleTest([]*models.CloudTrailRecord{
			{
				ErrorCode: nil,
			},
		})
		require.Equal(t, 0, httpClient.RequestNum())
	})

	t.Run("not detect when errorCode contains neither UnauthorizedOperation nor AccessDenied", func(t *testing.T) {
		_, httpClient := runRuleTest([]*models.CloudTrailRecord{
			{
				ErrorCode: aws.String("hoge"),
			},
		})
		require.Equal(t, 0, httpClient.RequestNum())
	})
}
