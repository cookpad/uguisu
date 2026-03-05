package rules

import (
	"github.com/cookpad/uguisu/pkg/models"
)

type lifeEventS3 struct {
	targetEvents map[string]bool
}

func newLifeEventS3() models.Rule {
	return &lifeEventS3{
		targetEvents: map[string]bool{
			"CreateBucket": true,
			"DeleteBucket": true,
		},
	}
}

func (x *lifeEventS3) ID() string                { return "resource_lifeevent_s3" }
func (x *lifeEventS3) Title() string             { return "S3 Bucket Life Event" }
func (x *lifeEventS3) Severity() models.Severity { return models.SeverityMedium }
func (x *lifeEventS3) Description() string {
	return "Monitoring creation and deletion of S3 buckets"
}

func (x *lifeEventS3) Match(record *models.CloudTrailRecord) bool {
	return record.EventSource == "s3.amazonaws.com" &&
		x.targetEvents[record.EventName]
}
