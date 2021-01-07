package rules

import (
	"github.com/m-mizutani/uguisu/pkg/models"
)

type lifeEventVPC struct {
	targetEvents map[string]bool
}

func newLifeEventVPC() models.Rule {
	return &lifeEventVPC{
		targetEvents: map[string]bool{
			"CreateVpc": true,
		},
	}
}

func (x *lifeEventVPC) ID() string                { return "resource_lifeevent_vpc" }
func (x *lifeEventVPC) Title() string             { return "VPC Life Event" }
func (x *lifeEventVPC) Severity() models.Severity { return models.SeverityHigh }
func (x *lifeEventVPC) Description() string {
	return "New VPC is created, check security settings"
}

func (x *lifeEventVPC) Match(record *models.CloudTrailRecord) bool {
	return x.targetEvents[record.EventName]
}
