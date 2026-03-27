package sqs

import (
	"encoding/json"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func sampleS3Event(region, bucket, key string) events.S3Event {
	return events.S3Event{
		Records: []events.S3EventRecord{
			{
				AWSRegion: region,
				S3: events.S3Entity{
					Bucket: events.S3Bucket{Name: bucket},
					Object: events.S3Object{Key: key},
				},
			},
		},
	}
}

// wrapS3EventsInSQSEvent builds an SQS event like SNS→SQS→Lambda: each inner S3 event is JSON inside SNSEntity.Message.
func wrapS3EventsInSQSEvent(t *testing.T, s3events ...events.S3Event) *events.SQSEvent {
	t.Helper()
	var records []events.SQSMessage
	for _, ev := range s3events {
		s3Bytes, err := json.Marshal(ev)
		require.NoError(t, err)
		sns := events.SNSEntity{Message: string(s3Bytes)}
		snsBytes, err := json.Marshal(sns)
		require.NoError(t, err)
		records = append(records, events.SQSMessage{Body: string(snsBytes)})
	}
	return &events.SQSEvent{Records: records}
}

func TestExtractEvents_nilEvent(t *testing.T) {
	got, err := ExtractEvents(nil)
	require.NoError(t, err)
	assert.Nil(t, got)
}

func TestExtractEvents_emptyRecords(t *testing.T) {
	got, err := ExtractEvents(&events.SQSEvent{Records: nil})
	require.NoError(t, err)
	assert.Empty(t, got)

	got, err = ExtractEvents(&events.SQSEvent{Records: []events.SQSMessage{}})
	require.NoError(t, err)
	assert.Empty(t, got)
}

func TestExtractEvents_successSingleRecord(t *testing.T) {
	want := sampleS3Event("eu-west-1", "ct-bucket", "logs/obj.json.gz")
	in := wrapS3EventsInSQSEvent(t, want)

	got, err := ExtractEvents(in)
	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, want, got[0])
}

func TestExtractEvents_successMultipleRecords(t *testing.T) {
	a := sampleS3Event("us-east-1", "bucket-a", "a/key")
	b := sampleS3Event("us-west-2", "bucket-b", "b/key")
	in := wrapS3EventsInSQSEvent(t, a, b)

	got, err := ExtractEvents(in)
	require.NoError(t, err)
	require.Len(t, got, 2)
	assert.Equal(t, a, got[0])
	assert.Equal(t, b, got[1])
}

func TestExtractEvents_invalidSNSBody(t *testing.T) {
	in := &events.SQSEvent{
		Records: []events.SQSMessage{
			{Body: `{this is not valid json`},
		},
	}

	_, err := ExtractEvents(in)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "sqs record 0: unmarshal SNS entity")
}

func TestExtractEvents_invalidS3Message(t *testing.T) {
	sns := events.SNSEntity{Message: `not-json-at-all`}
	snsBytes, err := json.Marshal(sns)
	require.NoError(t, err)
	in := &events.SQSEvent{
		Records: []events.SQSMessage{{Body: string(snsBytes)}},
	}

	_, err = ExtractEvents(in)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "sqs record 0: unmarshal S3 event")
}

func TestExtractEvents_secondRecordInvalidSNS(t *testing.T) {
	first := wrapS3EventsInSQSEvent(t, sampleS3Event("us-east-1", "b", "k"))
	in := &events.SQSEvent{
		Records: append(first.Records, events.SQSMessage{Body: `{{{`}),
	}

	_, err := ExtractEvents(in)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "sqs record 1: unmarshal SNS entity")
}
