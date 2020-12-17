package rules

import (
	"github.com/m-mizutani/uguisu/pkg/models"
)

// https://docs.aws.amazon.com/securityhub/latest/userguide/securityhub-cis-controls.html#cis-3.9-remediation
type awsCIS3_9 struct {
	targetEvents map[string]bool
}

func newAwsCIS3_9() models.Rule {
	return &awsCIS3_9{
		targetEvents: map[string]bool{
			"StopConfigurationRecorder": true,
			"DeleteDeliveryChannel":     true,
			"PutDeliveryChannel":        true,
			"PutConfigurationRecorder":  true,
		}}
}

func (x *awsCIS3_9) ID() string                { return "aws_cis_3.9" }
func (x *awsCIS3_9) Title() string             { return "AWS Config configuration changes" }
func (x *awsCIS3_9) Severity() models.Severity { return models.SeverityMedium }
func (x *awsCIS3_9) Description() string {
	return "AWS CIS benchmark 3.9 recommend to ensure a log metric filter and alarm exist for AWS Config configuration changes"
}
func (x *awsCIS3_9) Match(record *models.CloudTrailRecord) bool {
	return x.targetEvents[record.EventName]
}
