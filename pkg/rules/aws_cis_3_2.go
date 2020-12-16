package rules

import (
	"github.com/m-mizutani/uguisu/pkg/models"
)

// https://docs.aws.amazon.com/securityhub/latest/userguide/securityhub-cis-controls.html#cis-3.2-remediation
type awsCIS3_2 struct{}

func newAwsCIS3_2() models.Rule {
	return &awsCIS3_2{}
}

func (x *awsCIS3_2) ID() string                { return "aws_cis_3.2" }
func (x *awsCIS3_2) Title() string             { return "AWS Management Console sign-in without MFA" }
func (x *awsCIS3_2) Severity() models.Severity { return models.SeverityHigh }
func (x *awsCIS3_2) Description() string {
	return "AWS CIS benchmark 3.2 recommend to ensure a log metric filter and alarm exist for AWS Management Console sign-in without MFA"
}
func (x *awsCIS3_2) Match(record *models.CloudTrailRecord) bool {
	return record.EventName == "ConsoleLogin" &&
		record.AdditionalEventData != nil &&
		record.AdditionalEventData.MFAUsed != "Yes" &&
		record.AdditionalEventData.SamlProviderArn == nil &&
		(record.UserIdentity.SessionContext == nil ||
			record.UserIdentity.SessionContext.SessionIssuer == nil ||
			record.UserIdentity.SessionContext.SessionIssuer.Type != "Role")

}
