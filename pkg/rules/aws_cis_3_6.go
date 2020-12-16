package rules

import (
	"github.com/m-mizutani/uguisu/pkg/models"
)

// https://docs.aws.amazon.com/securityhub/latest/userguide/securityhub-cis-controls.html#cis-3.1-remediation
type awsCIS3_6 struct{}

func newAwsCIS3_6() models.Rule {
	return &awsCIS3_6{}
}

func (x *awsCIS3_6) ID() string                { return "aws_cis_3.6" }
func (x *awsCIS3_6) Title() string             { return "AWS Management Console authentication failures" }
func (x *awsCIS3_6) Severity() models.Severity { return models.SeverityMedium }
func (x *awsCIS3_6) Description() string {
	return "AWS CIS benchmark 3.6 recommend to ensure a log metric filter and alarm exist for AWS Management Console authentication failures"
}
func (x *awsCIS3_6) Match(record *models.CloudTrailRecord) bool {
	return record.EventName == "ConsoleLogin" &&
		record.ErrorMessage != nil &&
		*record.ErrorMessage == "Failed authentication"
}
