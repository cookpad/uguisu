package rules

import (
	"github.com/m-mizutani/uguisu/pkg/models"
)

// https://docs.aws.amazon.com/securityhub/latest/userguide/securityhub-cis-controls.html#cis-3.13-remediation
type awsCIS3_13 struct {
	targetEvents map[string]bool
}

func newAwsCIS3_13() models.Rule {
	return &awsCIS3_13{
		targetEvents: map[string]bool{
			"CreateRoute":                  true,
			"CreateRouteTable":             true,
			"ReplaceRoute":                 true,
			"ReplaceRouteTableAssociation": true,
			"DeleteRouteTable":             true,
			"DeleteRoute":                  true,
			"DisassociateRouteTable":       true,
		},
	}
}

func (x *awsCIS3_13) ID() string                { return "aws_cis_3.13" }
func (x *awsCIS3_13) Title() string             { return "Route table changes" }
func (x *awsCIS3_13) Severity() models.Severity { return models.SeverityMedium }
func (x *awsCIS3_13) Description() string {
	return "AWS CIS benchmark 3.13 recommend to ensure a log metric filter and alarm exist for route table changes"
}
func (x *awsCIS3_13) Match(record *models.CloudTrailRecord) bool {
	return x.targetEvents[record.EventName]
}
