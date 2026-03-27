package service

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/cookpad/uguisu/pkg/adaptor"
	"github.com/cookpad/uguisu/pkg/models"
	"github.com/slack-go/slack"
)

const (
	maxRetries        = 3
	defaultRetryAfter = 1 * time.Second
	maxRetryAfter     = 30 * time.Second
)

type Slack struct {
	httpClient adaptor.HTTPClient
	webhookURL string
	version    string
}

func NewSlack(httpClient adaptor.HTTPClient, webhookURL, version string) *Slack {
	if version == "" {
		version = "dev"
	}
	return &Slack{
		httpClient: httpClient,
		webhookURL: webhookURL,
		version:    version,
	}
}

func newField(title, value string) *slack.TextBlockObject {
	return slack.NewTextBlockObject("mrkdwn", fmt.Sprintf("*%s*\n%s", title, value), false, false)
}

func (x *Slack) Notify(alert *models.Alert) error {
	if x.httpClient == nil {
		return errors.New("HTTPClient is required to emit Slack, but not set")
	}
	if x.webhookURL == "" {
		return errors.New("webhookURL is required to emit Slack, but not set")
	}

	blocks := []slack.Block{
		slack.NewHeaderBlock(slack.NewTextBlockObject("plain_text", alert.Title, true, false)),
		slack.NewSectionBlock(slack.NewTextBlockObject("mrkdwn", alert.Description, false, false), nil, nil),
		slack.NewContextBlock("", slack.NewTextBlockObject("mrkdwn", "RuleID: "+alert.RuleID+" • uguisu "+x.version, false, false)),
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

			if len(param) > 1000 {
				param = param[:1000]
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
		return fmt.Errorf("failed to unmarshal slack message: %w", err)
	}

	// doAttempt performs a single POST.
	// Returns (rateLimited=true, wait, nil) on 429.
	// Returns (false, 0, nil) on success.
	// Returns (false, 0, err) on any other failure.
	doAttempt := func() (bool, time.Duration, error) {
		req, err := http.NewRequest("POST", x.webhookURL, bytes.NewBuffer(raw))
		if err != nil {
			return false, 0, fmt.Errorf("failed to create a new HTTP request to Slack: %w", err)
		}

		resp, err := x.httpClient.Do(req)
		if err != nil {
			return false, 0, fmt.Errorf("failed to post message to slack in communication: %w", err)
		}
		defer resp.Body.Close() //nolint:errcheck

		if resp.StatusCode == http.StatusTooManyRequests {
			wait := defaultRetryAfter
			if retryAfter := resp.Header.Get("Retry-After"); retryAfter != "" {
				if secs, err := strconv.Atoi(retryAfter); err == nil && secs >= 0 {
					wait = time.Duration(secs) * time.Second
				}
			}
			return true, wait, nil
		}

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			return false, 0, fmt.Errorf("failed to post message to slack API: code=%d body=%s", resp.StatusCode, string(body))
		}

		return false, 0, nil
	}

	var lastWait time.Duration
	for attempt := 0; attempt <= maxRetries; attempt++ {
		rateLimited, wait, err := doAttempt()
		if err != nil {
			return err
		}
		if !rateLimited {
			return nil
		}
		lastWait = wait
		if wait > maxRetryAfter {
			return fmt.Errorf("rate limited by Slack API with excessive Retry-After, giving up: retry_after=%s max_retry_after=%s", wait, maxRetryAfter)
		}
		if attempt < maxRetries {
			slog.Info("Rate limited by Slack, retrying", "attempt", attempt+1, "wait", wait.String())
			time.Sleep(wait)
		}
	}

	return fmt.Errorf("rate limited by Slack API, max retries exceeded: retry_after=%s attempts=%d", lastWait, maxRetries+1)
}
