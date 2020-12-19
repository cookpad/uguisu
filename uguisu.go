package uguisu

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"

	env "github.com/Netflix/go-env"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/google/uuid"
	"github.com/m-mizutani/golambda"

	"github.com/m-mizutani/uguisu/pkg/adaptor"
	"github.com/m-mizutani/uguisu/pkg/mock"
	"github.com/m-mizutani/uguisu/pkg/models"
	"github.com/m-mizutani/uguisu/pkg/rules"
	"github.com/m-mizutani/uguisu/pkg/service"
)

var logger = golambda.Logger

// Uguisu is main procedure of the package
type Uguisu struct {
	NewS3      adaptor.S3ClientFactory
	HTTPClient adaptor.HTTPClient

	Rules           *models.RuleSet
	Filters         AlertFilters
	SlackWebhookURL string `env:"SLACK_WEBHOOK_URL"`
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

	return u
}

// Start invokes lambda.Start via golambda.Start. Start() manage not only main procedure but also error handling. Then a developer to use uguisu need to configure uguisu before calling Start().
func (x *Uguisu) Start() {
	golambda.Start(func(event golambda.Event) (interface{}, error) {
		if err := x.run(event); err != nil {
			return nil, err
		}
		return nil, nil
	})
}

// run is invoked without golambda.Start. It's exported for testing
func (x *Uguisu) run(event golambda.Event) error {
	messages, err := event.DecapSNSonSQSMessage()
	if err != nil {
		return err
	}

	ctSvc := service.NewCloudTrailLogs(x.NewS3)
	slackSvc := service.NewSlack(x.HTTPClient, x.SlackWebhookURL)

	for _, event := range messages {
		logger.With("event", string(event)).Info("event proessing")
		var s3Event events.S3Event
		if err := event.Bind(&s3Event); err != nil {
			return err
		}
		logger.With("s3Event", s3Event).Info("Binding s3Event")

		for _, s3Record := range s3Event.Records {
			if err := handleS3Object(
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

func handleS3Object(ctSvc *service.CloudTrailLogs, slackSvc *service.Slack, rules *models.RuleSet, filters AlertFilters, region, bucket, key string) error {
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

	logger.With("processed", len(records)).Info("handleS3Object completed")

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
	s3Key := "some/object/" + uuid.New().String()

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
	gz.Close()

	_, err = s3Client.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(s3Bucket),
		Key:    aws.String(s3Key),
		Body:   bytes.NewReader(buf.Bytes()),
	})
	if err != nil {
		panic(err)
	}

	var event golambda.Event
	err = event.EncapSNSonSQSMessage(events.S3Event{
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
	if err != nil {
		panic(err)
	}

	if err := x.run(event); err != nil {
		panic(err)
	}

	var detected []*models.CloudTrailRecord
	for _, req := range httpClient.Requests {
		data, err := ioutil.ReadAll(req.Body)
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
