package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/m-mizutani/golambda"
	"github.com/m-mizutani/uguisu/pkg/adaptor"
	"github.com/m-mizutani/uguisu/pkg/models"
	"github.com/slack-go/slack"
)

type Slack struct {
	httpClient adaptor.HTTPClient
	webhookURL string
}

func NewSlack(httpClient adaptor.HTTPClient, webhookURL string) *Slack {
	return &Slack{
		httpClient: httpClient,
		webhookURL: webhookURL,
	}
}

func newField(title, value string) *slack.TextBlockObject {
	return slack.NewTextBlockObject("mrkdwn", fmt.Sprintf("*%s*\n%s", title, value), false, false)
}

func (x *Slack) Notify(alert *models.Alert) error {
	if x.httpClient == nil {
		return golambda.NewError("HTTPClient is required to emit Slack, but not set")
	}
	if x.webhookURL == "" {
		return golambda.NewError("webhookURL is required to emit Slack, but not set")
	}

	blocks := []slack.Block{
		slack.NewHeaderBlock(slack.NewTextBlockObject("plain_text", alert.Title, true, false)),
		slack.NewSectionBlock(slack.NewTextBlockObject("mrkdwn", alert.Description, false, false), nil, nil),
	}

	for _, record := range alert.Events {
		objects := []*slack.TextBlockObject{
			newField("EventName", record.EventName),
			newField("EventTime", record.EventTime),
			newField("EventID", record.EventID),
			newField("Region", record.AwsRegion),
			newField("AccountID", record.UserIdentity.AccountID),
			newField("SourceIPAddress", record.SourceIPAddress),
			newField("User", record.UserIdentity.ARN),
			newField("UserAgent", record.UserAgent),
		}
		if record.ErrorCode != nil {
			objects = append(objects, newField("ErrorCode", *record.ErrorCode))
		}
		if record.ErrorMessage != nil {
			objects = append(objects, newField("ErrorMessage", *record.ErrorMessage))
		}

		blocks = append(blocks, slack.NewDividerBlock())
		blocks = append(blocks, slack.NewSectionBlock(
			slack.NewTextBlockObject("mrkdwn", "", false, false), objects, nil),
		)

		if record.RequestParameters != nil {
			raw, err := json.MarshalIndent(record.RequestParameters, "", "  ")
			var param string
			if err == nil {
				param = string(raw)
			} else {
				param = fmt.Sprintf("%v", record.RequestParameters)
			}

			blocks = append(blocks,
				slack.NewTextBlockObject("mrkdwn", fmt.Sprintf("*RequestParameters*:\n```%s```", param), false, false))
		}
	}

	msg := slack.NewBlockMessage(blocks...)
	raw, err := json.Marshal(msg)
	if err != nil {
		return golambda.WrapError(err, "Failed to unmarshal slack message").With("msg", msg)
	}

	req, err := http.NewRequest("POST", x.webhookURL, bytes.NewBuffer(raw))
	if err != nil {
		return golambda.WrapError(err, "Failed to create a new HTTP request to Slack")
	}

	resp, err := x.httpClient.Do(req)
	if err != nil {
		return golambda.WrapError(err, "Failed to post message to slack in communication").With("msg", msg)
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return golambda.NewError("Failed to post message to slack in API").
			With("msg", msg).With("code", resp.StatusCode).With("body", string(body))
	}

	return nil
}
