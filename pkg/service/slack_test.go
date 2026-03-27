package service_test

import (
	"io"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/cookpad/uguisu/pkg/mock"
	"github.com/cookpad/uguisu/pkg/models"
	"github.com/cookpad/uguisu/pkg/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func baseAlert() *models.Alert {
	return &models.Alert{
		Title:       "Test alert",
		RuleID:      "test_rule_id",
		Sev:         models.SeverityMedium,
		Description: "test description",
		Events: []*models.CloudTrailRecord{
			{
				EventTime:          "2020-01-02T15:04:05",
				EventName:          "TestEvent",
				SourceIPAddress:    "10.1.2.3",
				UserAgent:          "my-user-agent",
				RecipientAccountID: "123456789012",
				AwsRegion:          "ap-northeast-1",
				UserIdentity: models.CloudTrailUserIdentity{
					ARN: "arn:aws:sts::123456789012:assumed-role/TestRole/session",
				},
			},
		},
	}
}

func TestSlackNotify_PostsToWebhook(t *testing.T) {
	httpClient := &mock.HTTPClient{}
	svc := service.NewSlack(httpClient, "https://hooks.example.com/webhook", "test")

	require.NoError(t, svc.Notify(baseAlert()))
	require.Equal(t, 1, httpClient.RequestNum())

	req := httpClient.Requests[0]
	assert.Equal(t, "POST", req.Method)
	assert.Equal(t, "https://hooks.example.com/webhook", req.URL.String())
}

func TestSlackNotify_PayloadContainsAlertFields(t *testing.T) {
	httpClient := &mock.HTTPClient{}
	svc := service.NewSlack(httpClient, "https://hooks.example.com/webhook", "test")

	require.NoError(t, svc.Notify(baseAlert()))

	body := httpClient.Body(0)
	assert.Contains(t, body, "Test alert")
	assert.Contains(t, body, "test_rule_id")
	assert.Contains(t, body, "test description")
	assert.Contains(t, body, "TestEvent")
	assert.Contains(t, body, "10.1.2.3")
}

func TestSlackNotify_SeverityColors(t *testing.T) {
	cases := []struct {
		sev   models.Severity
		color string
	}{
		{models.SeverityHigh, "#A30200"},
		{models.SeverityMedium, "#F2C744"},
		{models.SeverityLow, "#2EB886"},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(string(tc.sev), func(t *testing.T) {
			httpClient := &mock.HTTPClient{}
			svc := service.NewSlack(httpClient, "https://hooks.example.com/webhook", "test")
			alert := baseAlert()
			alert.Sev = tc.sev
			require.NoError(t, svc.Notify(alert))
			assert.Contains(t, httpClient.Body(0), tc.color)
		})
	}
}

func TestSlackNotify_IncludesErrorCode(t *testing.T) {
	httpClient := &mock.HTTPClient{}
	svc := service.NewSlack(httpClient, "https://hooks.example.com/webhook", "test")

	alert := baseAlert()
	alert.Events[0].ErrorCode = aws.String("UnauthorizedOperation")
	require.NoError(t, svc.Notify(alert))

	assert.Contains(t, httpClient.Body(0), "UnauthorizedOperation")
}

func TestSlackNotify_IncludesErrorMessage(t *testing.T) {
	httpClient := &mock.HTTPClient{}
	svc := service.NewSlack(httpClient, "https://hooks.example.com/webhook", "test")

	alert := baseAlert()
	alert.Events[0].ErrorMessage = aws.String("Failed authentication")
	require.NoError(t, svc.Notify(alert))

	assert.Contains(t, httpClient.Body(0), "Failed authentication")
}

func TestSlackNotify_IncludesRequestParameters(t *testing.T) {
	httpClient := &mock.HTTPClient{}
	svc := service.NewSlack(httpClient, "https://hooks.example.com/webhook", "test")

	alert := baseAlert()
	alert.Events[0].RequestParameters = map[string]interface{}{"bucketName": "my-bucket"}
	require.NoError(t, svc.Notify(alert))

	assert.Contains(t, httpClient.Body(0), "bucketName")
	assert.Contains(t, httpClient.Body(0), "my-bucket")
}

func TestSlackNotify_TruncatesLongRequestParameters(t *testing.T) {
	httpClient := &mock.HTTPClient{}
	svc := service.NewSlack(httpClient, "https://hooks.example.com/webhook", "test")

	alert := baseAlert()
	alert.Events[0].RequestParameters = map[string]interface{}{
		"key": strings.Repeat("x", 2000),
	}
	require.NoError(t, svc.Notify(alert))

	body := httpClient.Body(0)
	assert.Contains(t, body, "RequestParameters")

	// The truncation limit is 1000 chars total for the param block.
	// Without truncation the payload would contain ~2000 consecutive x's.
	// With truncation it will contain fewer than 1001 consecutive x's.
	assert.NotContains(t, body, strings.Repeat("x", 1001), "RequestParameters value should be truncated to 1000 chars")
	assert.Contains(t, body, strings.Repeat("x", 100), "RequestParameters value should not be empty after truncation")
}

func TestSlackNotify_ErrorOnNilHTTPClient(t *testing.T) {
	svc := service.NewSlack(nil, "https://hooks.example.com/webhook", "test")
	err := svc.Notify(baseAlert())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "HTTPClient is required")
}

func TestSlackNotify_ErrorOnEmptyWebhookURL(t *testing.T) {
	svc := service.NewSlack(&mock.HTTPClient{}, "", "test")
	err := svc.Notify(baseAlert())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "webhookURL is required")
}

func TestSlackNotify_ErrorOnNon200Response(t *testing.T) {
	httpClient := &mock.HTTPClient{
		RespCode: http.StatusInternalServerError,
		RespBody: io.NopCloser(strings.NewReader("internal error")),
	}
	svc := service.NewSlack(httpClient, "https://hooks.example.com/webhook", "test")
	err := svc.Notify(baseAlert())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to post message to slack API")
}

func TestSlackNotify_RetriesOn429ThenSucceeds(t *testing.T) {
	// Two 429s followed by a 200 — should succeed and make 3 requests total.
	httpClient := &mock.HTTPClient{
		Responses: []mock.Response{
			{Code: http.StatusTooManyRequests, Headers: http.Header{"Retry-After": []string{"0"}}},
			{Code: http.StatusTooManyRequests, Headers: http.Header{"Retry-After": []string{"0"}}},
			{Code: http.StatusOK},
		},
	}
	svc := service.NewSlack(httpClient, "https://hooks.example.com/webhook", "test")
	err := svc.Notify(baseAlert())
	require.NoError(t, err)
	assert.Equal(t, 3, httpClient.RequestNum())
}

func TestSlackNotify_ErrorAfterMaxRetries429(t *testing.T) {
	// All responses are 429 — should exhaust retries and return an error.
	httpClient := &mock.HTTPClient{
		Responses: []mock.Response{
			{Code: http.StatusTooManyRequests, Headers: http.Header{"Retry-After": []string{"0"}}},
		},
	}
	svc := service.NewSlack(httpClient, "https://hooks.example.com/webhook", "test")
	err := svc.Notify(baseAlert())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "max retries exceeded")
	assert.Equal(t, 4, httpClient.RequestNum()) // 1 initial + 3 retries
}

func TestSlackNotify_ErrorOnExcessiveRetryAfter(t *testing.T) {
	// A Retry-After value exceeding the cap should return an error immediately
	// without sleeping, rather than blocking the Lambda invocation.
	httpClient := &mock.HTTPClient{
		Responses: []mock.Response{
			{Code: http.StatusTooManyRequests, Headers: http.Header{"Retry-After": []string{"3600"}}},
		},
	}
	svc := service.NewSlack(httpClient, "https://hooks.example.com/webhook", "test")
	err := svc.Notify(baseAlert())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "excessive Retry-After")
	assert.Equal(t, 1, httpClient.RequestNum()) // gave up after the first 429
}

// TestSlackIntegration sends a real Slack notification when TEST_SLACK_URL is set.
func TestSlackIntegration(t *testing.T) {
	url, ok := os.LookupEnv("TEST_SLACK_URL")
	if !ok {
		t.Skip("TEST_SLACK_URL is not set")
	}

	svc := service.NewSlack(&http.Client{}, url, "test")
	alert := baseAlert()
	alert.Title = "Integration test alert"
	alert.Events[0].ErrorCode = aws.String("some-error")
	alert.Events[0].RequestParameters = map[string]interface{}{"test": "message"}
	require.NoError(t, svc.Notify(alert))
}
