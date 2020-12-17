package rules

import (
	"github.com/m-mizutani/uguisu/pkg/models"
)

// https://docs.aws.amazon.com/securityhub/latest/userguide/securityhub-cis-controls.html#cis-3.8-remediation
type awsCIS3_8 struct {
	targetEvents map[string]bool
}

func newAwsCIS3_8() models.Rule {
	return &awsCIS3_8{
		targetEvents: map[string]bool{
			"PutBucketAcl":            true,
			"PutBucketPolicy":         true,
			"PutBucketCors":           true,
			"PutBucketLifecycle":      true,
			"PutBucketReplication":    true,
			"DeleteBucketPolicy":      true,
			"DeleteBucketCors":        true,
			"DeleteBucketLifecycle":   true,
			"DeleteBucketReplication": true,
		},
	}
}

func (x *awsCIS3_8) ID() string                { return "aws_cis_3.8" }
func (x *awsCIS3_8) Title() string             { return "S3 bucket policy changes" }
func (x *awsCIS3_8) Severity() models.Severity { return models.SeverityMedium }
func (x *awsCIS3_8) Description() string {
	return "AWS CIS benchmark 3.8 recommend to ensure a log metric filter and alarm exist for S3 bucket policy changes"
}
func (x *awsCIS3_8) Match(record *models.CloudTrailRecord) bool {
	return x.targetEvents[record.EventName]
}
