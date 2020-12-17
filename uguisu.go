package uguisu

import (
	"net/http"

	env "github.com/Netflix/go-env"
	"github.com/aws/aws-lambda-go/events"
	"github.com/m-mizutani/golambda"

	"github.com/m-mizutani/uguisu/pkg/adaptor"
	"github.com/m-mizutani/uguisu/pkg/models"
	"github.com/m-mizutani/uguisu/pkg/rules"
	"github.com/m-mizutani/uguisu/pkg/service"
)

// Uguisu is main procedure of the package
type Uguisu struct {
	NewS3      adaptor.S3ClientFactory
	HTTPClient adaptor.HTTPClient

	Rules           *models.RuleSet
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
		if err := x.Run(event); err != nil {
			return nil, err
		}
		return nil, nil
	})
}

// Run is invoked without golambda.Start. It's exported for testing
func (x *Uguisu) Run(event golambda.Event) error {
	messages, err := event.DecapSNSonSQSMessage()
	if err != nil {
		return err
	}

	ctSvc := service.NewCloudTrailLogs(x.NewS3)
	slackSvc := service.NewSlack(x.HTTPClient, x.SlackWebhookURL)

	for _, event := range messages {
		var s3Event events.S3Event
		if err := event.Bind(&s3Event); err != nil {
			return err
		}

		for _, s3Record := range s3Event.Records {
			if err := handleS3Object(
				ctSvc,
				slackSvc,
				x.Rules,
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

func handleS3Object(ctSvc *service.CloudTrailLogs, slackSvc *service.Slack, rules *models.RuleSet, region, bucket, key string) error {
	records, err := ctSvc.Read(region, bucket, key)
	if err != nil {
		return err
	}

	for _, record := range records {
		alerts := rules.Detect(record)

		for _, alert := range alerts {
			if err := slackSvc.Notify(alert); err != nil {
				return err
			}
		}
	}

	return nil
}
