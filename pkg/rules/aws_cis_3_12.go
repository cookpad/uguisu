package rules

import (
	"github.com/m-mizutani/uguisu/pkg/models"
)

// https://docs.aws.amazon.com/securityhub/latest/userguide/securityhub-cis-controls.html#cis-3.12-remediation
type awsCIS3_12 struct {
	targetEvents map[string]bool
}

func newAwsCIS3_12() models.Rule {
	return &awsCIS3_12{
		targetEvents: map[string]bool{
			"CreateCustomerGateway": true,
			"DeleteCustomerGateway": true,
			"AttachInternetGateway": true,
			"CreateInternetGateway": true,
			"DeleteInternetGateway": true,
			"DetachInternetGateway": true,
		},
	}
}

func (x *awsCIS3_12) ID() string                { return "aws_cis_3.12" }
func (x *awsCIS3_12) Title() string             { return "Changes to network gateways" }
func (x *awsCIS3_12) Severity() models.Severity { return models.SeverityMedium }
func (x *awsCIS3_12) Description() string {
	return "AWS CIS benchmark 3.12 recommend to ensure a log metric filter and alarm exist for Changes to network gateways"
}
func (x *awsCIS3_12) Match(record *models.CloudTrailRecord) bool {
	return x.targetEvents[record.EventName] == true
}
