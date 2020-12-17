package rules

import (
	"github.com/m-mizutani/uguisu/pkg/models"
)

// https://docs.aws.amazon.com/securityhub/latest/userguide/securityhub-cis-controls.html#cis-3.11-remediation
type awsCIS3_11 struct {
	targetEvents map[string]bool
}

func newAwsCIS3_11() models.Rule {
	return &awsCIS3_11{
		targetEvents: map[string]bool{
			"CreateNetworkAcl":             true,
			"CreateNetworkAclEntry":        true,
			"DeleteNetworkAcl":             true,
			"DeleteNetworkAclEntry":        true,
			"ReplaceNetworkAclEntry":       true,
			"ReplaceNetworkAclAssociation": true,
		},
	}
}

func (x *awsCIS3_11) ID() string                { return "aws_cis_3.11" }
func (x *awsCIS3_11) Title() string             { return "Network Access Control Lists (NACL)" }
func (x *awsCIS3_11) Severity() models.Severity { return models.SeverityMedium }
func (x *awsCIS3_11) Description() string {
	return "AWS CIS benchmark 3.11 recommend to ensure a log metric filter and alarm exist for Network Access Control Lists (NACL)"
}
func (x *awsCIS3_11) Match(record *models.CloudTrailRecord) bool {
	return x.targetEvents[record.EventName]
}
