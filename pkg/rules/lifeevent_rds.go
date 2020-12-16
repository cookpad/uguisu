package rules

import (
	"github.com/m-mizutani/uguisu/pkg/models"
)

type lifeEventRDS struct {
	targetEvents map[string]bool
}

func newLifeEventRDS() models.Rule {
	return &lifeEventRDS{
		targetEvents: map[string]bool{
			"CreateDBInstance": true,
			"DeleteDBInstance": true,
		},
	}
}

func (x *lifeEventRDS) ID() string                { return "resource_lifeevent_rds" }
func (x *lifeEventRDS) Title() string             { return "RDS Instance Life Event" }
func (x *lifeEventRDS) Severity() models.Severity { return models.SeverityLow }
func (x *lifeEventRDS) Description() string {
	return "Monitoring events of RDS instance creation and destruction"
}

func (x *lifeEventRDS) Match(record *models.CloudTrailRecord) bool {
	return record.EventSource == "rds.amazonaws.com" &&
		x.targetEvents[record.EventName] == true
}
