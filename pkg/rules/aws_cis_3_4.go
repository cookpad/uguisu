package rules

import (
	"github.com/m-mizutani/uguisu/pkg/models"
)

// https://docs.aws.amazon.com/securityhub/latest/userguide/securityhub-cis-controls.html#cis-3.4-remediation
type awsCIS3_4 struct {
	targetEvents map[string]bool
}

func newAwsCIS3_4() models.Rule {
	return &awsCIS3_4{
		targetEvents: map[string]bool{
			"DeleteGroupPolicy":   true,
			"DeleteRolePolicy":    true,
			"DeleteUserPolicy":    true,
			"PutGroupPolicy":      true,
			"PutRolePolicy":       true,
			"PutUserPolicy":       true,
			"CreatePolicy":        true,
			"DeletePolicy":        true,
			"CreatePolicyVersion": true,
			"DeletePolicyVersion": true,
			"AttachRolePolicy":    true,
			"DetachRolePolicy":    true,
			"AttachUserPolicy":    true,
			"DetachUserPolicy":    true,
			"AttachGroupPolicy":   true,
			"DetachGroupPolicy":   true,
		},
	}
}

func (x *awsCIS3_4) ID() string                { return "aws_cis_3.4" }
func (x *awsCIS3_4) Title() string             { return "IAM policy changes" }
func (x *awsCIS3_4) Severity() models.Severity { return models.SeverityMedium }
func (x *awsCIS3_4) Description() string {
	return "AWS CIS benchmark 3.4 recommend to ensure a log metric filter and alarm exist for IAM policy changes"
}
func (x *awsCIS3_4) Match(record *models.CloudTrailRecord) bool {
	return x.targetEvents[record.EventName] == true
}
