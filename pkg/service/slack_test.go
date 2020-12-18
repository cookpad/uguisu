package service_test

import (
	"net/http"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/m-mizutani/uguisu/pkg/models"
	"github.com/m-mizutani/uguisu/pkg/service"
	"github.com/stretchr/testify/require"
)

func TestSlackIntegrtion(t *testing.T) {
	url, ok := os.LookupEnv("TEST_SLACK_URL")
	if !ok {
		t.Skip("TEST_SLACK_URL is not set")
	}

	slack := service.NewSlack(&http.Client{}, url)
	require.NoError(t, slack.Notify(&models.Alert{
		Title:       "Test alert",
		RuleID:      "test_rule_id",
		Sev:         models.SeverityMedium,
		Description: "this is test. please ignore me",
		Events: []*models.CloudTrailRecord{
			{
				EventTime:          "2020-01-02T15:04:05",
				SourceIPAddress:    "10.1.2.3",
				EventName:          "TestEvent",
				UserAgent:          "my-user-agent",
				ErrorCode:          aws.String("some-error"),
				RecipientAccountID: "1111111111111",
				AwsRegion:          "ap-northeast-1",
				RequestParameters: map[string]interface{}{
					"test": "message",
				},
				UserIdentity: models.CloudTrailUserIdentity{
					ARN: `arn:aws:sts::11111111111111:assumed-role/TestRole/xxxxxxxxxxxxx`,
				},
			},
		},
	}))
}
