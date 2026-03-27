package sqs

import (
	"encoding/json"
	"fmt"

	"github.com/aws/aws-lambda-go/events"
)

// ExtractEvents decodes S3 event notifications from an SQS batch whose records wrap SNS payloads.
func ExtractEvents(event *events.SQSEvent) ([]events.S3Event, error) {
	if event == nil {
		return nil, nil
	}
	var output []events.S3Event
	for i, record := range event.Records {
		var snsEntity events.SNSEntity
		if err := json.Unmarshal([]byte(record.Body), &snsEntity); err != nil {
			return nil, fmt.Errorf("sqs record %d: unmarshal SNS entity: %w", i, err)
		}
		var s3Event events.S3Event
		if err := json.Unmarshal([]byte(snsEntity.Message), &s3Event); err != nil {
			return nil, fmt.Errorf("sqs record %d: unmarshal S3 event: %w", i, err)
		}
		output = append(output, s3Event)
	}
	return output, nil
}
