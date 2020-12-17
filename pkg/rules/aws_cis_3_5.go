package rules

import (
	"github.com/m-mizutani/uguisu/pkg/models"
)

// https://docs.aws.amazon.com/securityhub/latest/userguide/securityhub-cis-controls.html#cis-3.5-remediation
type awsCIS3_5 struct {
	targetEvents map[string]bool
}

func newAwsCIS3_5() models.Rule {
	return &awsCIS3_5{
		targetEvents: map[string]bool{
			"CreateTrail":  true,
			"UpdateTrail":  true,
			"DeleteTrail":  true,
			"StartLogging": true,
			"StopLogging":  true,
		},
	}
}

func (x *awsCIS3_5) ID() string                { return "aws_cis_3.5" }
func (x *awsCIS3_5) Title() string             { return "CloudTrail configuration changes" }
func (x *awsCIS3_5) Severity() models.Severity { return models.SeverityMedium }
func (x *awsCIS3_5) Description() string {
	return "AWS CIS benchmark 3.5 recommend to ensure a log metric filter and alarm exist for CloudTrail configuration changes"
}
func (x *awsCIS3_5) Match(record *models.CloudTrailRecord) bool {
	return x.targetEvents[record.EventName]
}
