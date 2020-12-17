package rules

import (
	"github.com/m-mizutani/uguisu/pkg/models"
)

// https://docs.aws.amazon.com/securityhub/latest/userguide/securityhub-cis-controls.html#cis-3.14-remediation
type awsCIS3_14 struct {
	targetEvents map[string]bool
}

func newAwsCIS3_14() models.Rule {
	return &awsCIS3_14{
		targetEvents: map[string]bool{
			"CreateVpc":                  true,
			"DeleteVpc":                  true,
			"ModifyVpcAttribute":         true,
			"AcceptVpcPeeringConnection": true,
			"CreateVpcPeeringConnection": true,
			"DeleteVpcPeeringConnection": true,
			"RejectVpcPeeringConnection": true,
			"AttachClassicLinkVpc":       true,
			"DetachClassicLinkVpc":       true,
			"DisableVpcClassicLink":      true,
			"EnableVpcClassicLink":       true},
	}
}

func (x *awsCIS3_14) ID() string                { return "aws_cis_3.14" }
func (x *awsCIS3_14) Title() string             { return "Route table changes" }
func (x *awsCIS3_14) Severity() models.Severity { return models.SeverityMedium }
func (x *awsCIS3_14) Description() string {
	return "AWS CIS benchmark 3.14 recommend to ensure a log metric filter and alarm exist for VPC changes"
}
func (x *awsCIS3_14) Match(record *models.CloudTrailRecord) bool {
	return x.targetEvents[record.EventName]
}
