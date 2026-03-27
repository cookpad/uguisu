package uguisu

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"

	env "github.com/Netflix/go-env"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"

	"github.com/cookpad/uguisu/pkg/adaptor"
	"github.com/cookpad/uguisu/pkg/log"
	"github.com/cookpad/uguisu/pkg/mock"
	"github.com/cookpad/uguisu/pkg/models"
	"github.com/cookpad/uguisu/pkg/rules"
	"github.com/cookpad/uguisu/pkg/service"
	"github.com/cookpad/uguisu/pkg/sqs"
)

// Version is set at build time via -ldflags "-X github.com/cookpad/uguisu.Version=..."
var Version = "dev"

// Uguisu is main procedure of the package
type Uguisu struct {
	NewS3      adaptor.S3ClientFactory
	HTTPClient adaptor.HTTPClient

	Rules           *models.RuleSet
	Filters         AlertFilters
	SlackWebhookURL string `env:"SLACK_WEBHOOK_URL"`
	DisabledRules   string `env:"DISABLED_RULES"`
}

// New is constructor of Uguisu
func New() *Uguisu {
	u := &Uguisu{
		NewS3:      adaptor.NewS3Client,
		HTTPClient: &http.Client{},
		Rules:      rules.NewDefaultRuleSet(),
	}

	if _, err := env.UnmarshalFromEnviron(u); err != nil {
		panic(err)
	}

	u.Rules.Disable(u.DisabledRules)

	return u
}

// Start runs the Lambda handler (logging, request ID on default logger for this invocation, then processing).
func (x *Uguisu) Start() {
	lambda.Start(func(ctx context.Context, event *events.SQSEvent) error {
		defer flushSentry()

		slog.SetDefault(slog.New(log.Handler(ctx)))
		evts, err := sqs.ExtractEvents(event)
		if err != nil {
			logError(ctx, err)
			return err
		}
		if err := x.run(ctx, evts); err != nil {
			logError(ctx, err)
			return err
		}
		return nil
	})
}

func logError(ctx context.Context, err error) {
	if id := captureSentryError(ctx, err); id != nil {
		slog.Error(err.Error(), "sentry_event_id", string(*id))
	} else {
		slog.Error(err.Error())
	}
}

// run is invoked by Start & Test - Test is exported to make testing this easy
func (x *Uguisu) run(ctx context.Context, events []events.S3Event) error {
	for _, filter := range x.Filters {
		slog.Debug("Set filter", "filter(addr)", fmt.Sprintf("%v", filter))
	}
	ctSvc := service.NewCloudTrailLogs(x.NewS3)
	slackSvc := service.NewSlack(x.HTTPClient, x.SlackWebhookURL, Version)

	for _, event := range events {
		for _, s3Record := range event.Records {
			if err := handleS3Object(
				ctx,
				ctSvc,
				slackSvc,
				x.Rules,
				x.Filters,
				s3Record.AWSRegion,
				s3Record.S3.Bucket.Name,
				s3Record.S3.Object.Key,
			); err != nil {
				return err
			}
		}
	}

	return nil
}

func handleS3Object(ctx context.Context, ctSvc *service.CloudTrailLogs, slackSvc *service.Slack, rules *models.RuleSet, filters AlertFilters, region, bucket, key string) error {
	records, err := ctSvc.Read(region, bucket, key)
	if err != nil {
		return err
	}

	for _, record := range records {
		alerts := rules.Detect(record)

		for _, alert := range alerts {
			if !filters.filter(alert) {
				continue
			}

			if err := slackSvc.Notify(alert); err != nil {
				return err
			}
		}
	}

	slog.Log(ctx, slog.LevelDebug, "handleS3Object completed", "processed", len(records))

	return nil
}

// Test invokes uguisu.run to make rule test easy
func (x *Uguisu) Test(records []*models.CloudTrailRecord) []*models.CloudTrailRecord {
	newS3, s3Client := mock.NewS3Mock()
	httpClient := &mock.HTTPClient{}

	x.NewS3 = newS3
	x.HTTPClient = httpClient
	x.SlackWebhookURL = "https://test.example.com/endpoint"

	s3Region := "us-east-0"
	s3Bucket := "your-ct-logs"
	s3Key := "some/object/" + uuid.New().String() + ".json.gz"

	eventIDs := make([]string, len(records))
	for i := range records {
		eventIDs[i] = uuid.New().String()
		records[i].EventID = eventIDs[i]
	}

	// Put data
	data := models.CloudTrailLogObject{
		Records: records,
	}

	raw, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}

	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	_, err = gz.Write(raw)
	if err != nil {
		panic(err)
	}
	if err = gz.Close(); err != nil {
		panic(err)
	}

	_, err = s3Client.PutObject(context.Background(), &s3.PutObjectInput{
		Bucket: aws.String(s3Bucket),
		Key:    aws.String(s3Key),
		Body:   bytes.NewReader(buf.Bytes()),
	})
	if err != nil {
		panic(err)
	}

	s3Events := []events.S3Event{

		{
			Records: []events.S3EventRecord{
				{
					AWSRegion: s3Region,
					S3: events.S3Entity{
						Bucket: events.S3Bucket{Name: s3Bucket},
						Object: events.S3Object{Key: s3Key},
					},
				},
			},
		},
	}

	if err := x.run(context.Background(), s3Events); err != nil {
		panic(err)
	}

	var detected []*models.CloudTrailRecord
	for _, req := range httpClient.Requests {
		data, err := io.ReadAll(req.Body)
		if err != nil {
			panic(err)
		}

		for _, record := range records {
			if strings.Contains(string(data), record.EventID) {
				detected = append(detected, record)
			}
		}
	}

	return detected
}
