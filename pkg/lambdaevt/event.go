package lambdaevt

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"

	"github.com/aws/aws-lambda-go/events"
	"github.com/google/uuid"
)

// Event wraps the raw Lambda payload and optional context (mirrors prior golambda.Event usage).
type Event struct {
	Ctx    context.Context
	Origin interface{}
}

// Bind JSON-marshals the original event and unmarshals into v.
func (x *Event) Bind(v interface{}) error {
	raw, err := json.Marshal(x.Origin)
	if err != nil {
		return fmt.Errorf("marshal original lambda event: %w", err)
	}

	if err := json.Unmarshal(raw, v); err != nil {
		return fmt.Errorf("unmarshal to v: %w (raw=%s)", err, string(raw))
	}

	return nil
}

// EventRecord is a decapsulated payload (e.g. SNS Message bytes inside SQS).
type EventRecord []byte

// Bind unmarshals the record into ev.
func (x EventRecord) Bind(ev interface{}) error {
	if err := json.Unmarshal(x, ev); err != nil {
		return fmt.Errorf("json unmarshal event record: %w (raw=%s)", err, string(x))
	}
	return nil
}

// DecapSNSonSQSMessage unwraps SNS payloads carried in SQS records (SNS → SQS → Lambda).
func (x *Event) DecapSNSonSQSMessage() ([]EventRecord, error) {
	var sqsEvent events.SQSEvent
	if err := x.Bind(&sqsEvent); err != nil {
		return nil, err
	}

	if len(sqsEvent.Records) == 0 {
		return nil, errors.New("no SQS event records")
	}

	var output []EventRecord
	for _, record := range sqsEvent.Records {
		var snsEntity events.SNSEntity
		if err := json.Unmarshal([]byte(record.Body), &snsEntity); err != nil {
			return nil, fmt.Errorf("unmarshal SNS entity in SQS msg: %w (body=%s)", err, record.Body)
		}

		output = append(output, EventRecord(snsEntity.Message))
	}

	return output, nil
}

// EncapSNSonSQSMessage sets the event origin to an SQSEvent whose bodies wrap SNS entities around v.
// Used for tests and local harnesses.
func (x *Event) EncapSNSonSQSMessage(v interface{}) error {
	snsEntities, err := encapSNSEntity(v)
	if err != nil {
		return err
	}

	sqsMessages, err := encapSQSMessage(snsEntities)
	if err != nil {
		return err
	}

	x.Origin = events.SQSEvent{
		Records: sqsMessages,
	}

	return nil
}

func encapSQSMessage(v interface{}) ([]events.SQSMessage, error) {
	value := reflect.ValueOf(v)

	switch value.Kind() {
	case reflect.Array, reflect.Slice:
		var messages []events.SQSMessage

		for i := 0; i < value.Len(); i++ {
			msg, err := encapSQSMessage(value.Index(i).Interface())
			if err != nil {
				return nil, err
			}
			messages = append(messages, msg...)
		}

		return messages, nil

	case reflect.Ptr, reflect.UnsafePointer:
		if value.IsZero() || value.Elem().IsZero() {
			return nil, nil
		}

		return encapSQSMessage(value.Elem().Interface())

	default:
		raw, err := json.Marshal(v)
		if err != nil {
			return nil, fmt.Errorf("marshal value for SQS body: %w", err)
		}
		return []events.SQSMessage{
			{
				MessageId: uuid.New().String(),
				Body:      string(raw),
			},
		}, nil
	}
}

func encapSNSEntity(v interface{}) ([]events.SNSEntity, error) {
	value := reflect.ValueOf(v)

	switch value.Kind() {
	case reflect.Array, reflect.Slice:
		var messages []events.SNSEntity

		for i := 0; i < value.Len(); i++ {
			msg, err := encapSNSEntity(value.Index(i).Interface())
			if err != nil {
				return nil, err
			}
			messages = append(messages, msg...)
		}

		return messages, nil

	case reflect.Ptr, reflect.UnsafePointer:
		if value.IsZero() || value.Elem().IsZero() {
			return nil, nil
		}

		return encapSNSEntity(value.Elem().Interface())

	default:
		raw, err := json.Marshal(v)
		if err != nil {
			return nil, fmt.Errorf("marshal value for SNS message: %w", err)
		}
		return []events.SNSEntity{
			{
				Message: string(raw),
			},
		}, nil
	}
}
