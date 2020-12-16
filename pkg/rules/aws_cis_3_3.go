package rules

import (
	"github.com/m-mizutani/uguisu/pkg/models"
)

// https://docs.aws.amazon.com/securityhub/latest/userguide/securityhub-cis-controls.html#cis-3.1-remediation
type awsCIS3_3 struct {
}

func newAwsCIS3_3() models.Rule {
	return &awsCIS3_3{}
}

func (x *awsCIS3_3) ID() string                { return "aws_cis_3.3" }
func (x *awsCIS3_3) Title() string             { return "Usage of root account" }
func (x *awsCIS3_3) Severity() models.Severity { return models.SeverityMedium }
func (x *awsCIS3_3) Description() string {
	return "AWS CIS benchmark 3.3 recommend to ensure a log metric filter and alarm exist for usage of root account"
}
func (x *awsCIS3_3) Match(record *models.CloudTrailRecord) bool {
	return record.UserIdentity.Type == "Root" &&
		record.UserIdentity.InvokedBy == nil &&
		record.EventType != "AwsServiceEvent"
}
