package rules

import (
	"github.com/m-mizutani/uguisu/pkg/models"
)

// https://docs.aws.amazon.com/securityhub/latest/userguide/securityhub-cis-controls.html#cis-3.10-remediation
type awsCIS3_10 struct {
	targetEvents map[string]bool
}

func newAwsCIS3_10() models.Rule {
	return &awsCIS3_10{
		targetEvents: map[string]bool{
			"AuthorizeSecurityGroupIngress": true,
			"AuthorizeSecurityGroupEgress":  true,
			"RevokeSecurityGroupIngress":    true,
			"RevokeSecurityGroupEgress":     true,
			"CreateSecurityGroup":           true,
			"DeleteSecurityGroup":           true,
		},
	}
}

func (x *awsCIS3_10) ID() string                { return "aws_cis_3.10" }
func (x *awsCIS3_10) Title() string             { return "Security group changes" }
func (x *awsCIS3_10) Severity() models.Severity { return models.SeverityMedium }
func (x *awsCIS3_10) Description() string {
	return "AWS CIS benchmark 3.10 recommend to ensure a log metric filter and alarm exist for security group changes"
}
func (x *awsCIS3_10) Match(record *models.CloudTrailRecord) bool {
	return x.targetEvents[record.EventName] == true
}
