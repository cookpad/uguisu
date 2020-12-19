package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

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
		slack.NewContextBlock("", slack.NewTextBlockObject("mrkdwn", "RuleID: "+alert.RuleID, false, false)),
	}

	for _, record := range alert.Events {
		eventDescription := strings.Join([]string{
			fmt.Sprintf("*%s*", record.EventName),
			fmt.Sprintf("- %s @ %s", record.RecipientAccountID, record.AwsRegion),
			fmt.Sprintf("- by `%s`", record.UserIdentity.ARN),
			fmt.Sprintf("- from %s", record.SourceIPAddress),
		}, "\n")

		var errors []*slack.TextBlockObject
		if record.ErrorCode != nil {
			errors = append(errors, newField("ErrorCode", *record.ErrorCode))
		}
		if record.ErrorMessage != nil {
			errors = append(errors, newField("ErrorMessage", *record.ErrorMessage))
		}

		blocks = append(blocks, []slack.Block{
			slack.NewDividerBlock(),
			slack.NewSectionBlock(slack.NewTextBlockObject("mrkdwn", eventDescription, false, false), errors, nil),
		}...)

		if record.RequestParameters != nil {
			raw, err := json.MarshalIndent(record.RequestParameters, "", "  ")
			var param string
			if err == nil {
				param = string(raw)
			} else {
				param = fmt.Sprintf("%v", record.RequestParameters)
			}

			field := fmt.Sprintf("*RequestParameters*:\n```%s```", param)
			blocks = append(blocks, slack.NewSectionBlock(
				slack.NewTextBlockObject("mrkdwn", field, false, false), nil, nil))
		}

		footer := strings.Join([]string{
			fmt.Sprintf("ID: %s (%s)", record.EventID, record.EventTime),
			fmt.Sprintf("UserAgent: %s", record.UserAgent),
		}, "\n")
		blocks = append(blocks, slack.NewContextBlock("", slack.NewTextBlockObject("mrkdwn", footer, false, false)))

		/*
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
		*/

	}

	colorMap := map[models.Severity]string{
		models.SeverityHigh:   "#A30200",
		models.SeverityMedium: "#F2C744",
		models.SeverityLow:    "#2EB886",
	}

	msg := slack.WebhookMessage{
		Attachments: []slack.Attachment{
			{
				Color: colorMap[alert.Sev],
				Blocks: slack.Blocks{
					BlockSet: blocks,
				},
			},
		},
	}

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
		fmt.Println(string(raw))
		return golambda.NewError("Failed to post message to slack in API").
			With("msg", msg).
			With("code", resp.StatusCode).
			With("body", string(body))
	}

	return nil
}
