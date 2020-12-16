package rules

import (
	"github.com/m-mizutani/uguisu/pkg/models"
)

// CIS 3.7 – Ensure a log metric filter and alarm exist for unauthorized API calls
// https://docs.aws.amazon.com/securityhub/latest/userguide/securityhub-cis-controls.html#cis-3.7-remediation
type awsCIS3_7 struct{}

func newAwsCIS3_7() models.Rule {
	return &awsCIS3_7{}
}

func (x *awsCIS3_7) ID() string                { return "aws_cis_3.7" }
func (x *awsCIS3_7) Title() string             { return "Disabling or scheduled deletion of customer created CMKs" }
func (x *awsCIS3_7) Severity() models.Severity { return models.SeverityMedium }
func (x *awsCIS3_7) Description() string {
	return "AWS CIS benchmark 3.7 recommend to ensure a log metric filter and alarm exist for disabling or scheduled deletion of customer created CMKs"
}
func (x *awsCIS3_7) Match(record *models.CloudTrailRecord) bool {
	return record.EventSource == "kms.amazonaws.com" &&
		(record.EventName == "DisableKey" ||
			record.EventName == "ScheduleKeyDeletion")
}
