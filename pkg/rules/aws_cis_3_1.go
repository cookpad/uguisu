package rules

import (
	"regexp"

	"github.com/m-mizutani/uguisu/pkg/models"
)

// CIS 3.1 – Ensure a log metric filter and alarm exist for unauthorized API calls
// https://docs.aws.amazon.com/securityhub/latest/userguide/securityhub-cis-controls.html#cis-3.1-remediation
type awsCIS3_1 struct {
	unauth *regexp.Regexp
	denied *regexp.Regexp
}

func newAwsCIS3_1() models.Rule {
	return &awsCIS3_1{
		unauth: regexp.MustCompile(`UnauthorizedOperation$`),
		denied: regexp.MustCompile(`^AccessDenied`),
	}
}

func (x *awsCIS3_1) ID() string                { return "aws_cis_3.1" }
func (x *awsCIS3_1) Title() string             { return "Unauthorized API calls monitoring" }
func (x *awsCIS3_1) Severity() models.Severity { return models.SeverityMedium }
func (x *awsCIS3_1) Description() string {
	return "AWS CIS benchmark 3.1 recommend to ensure a log metric filter and alarm exist for unauthorized API calls"
}
func (x *awsCIS3_1) Match(record *models.CloudTrailRecord) bool {
	return record.ErrorCode != nil &&
		(x.unauth.MatchString(*record.ErrorCode) ||
			x.denied.MatchString(*record.ErrorCode))
}
