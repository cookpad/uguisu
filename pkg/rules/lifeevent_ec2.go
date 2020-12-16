package rules

import (
	"github.com/m-mizutani/uguisu/pkg/models"
)

type lifeEventEC2 struct {
	targetEvents map[string]bool
}

func newLifeEventEC2() models.Rule {
	return &lifeEventEC2{
		targetEvents: map[string]bool{
			"RunInstances":       true,
			"TerminateInstances": true,
		},
	}
}

func (x *lifeEventEC2) ID() string                { return "resource_lifeevent_ec2" }
func (x *lifeEventEC2) Title() string             { return "EC2 Instance Life Event" }
func (x *lifeEventEC2) Severity() models.Severity { return models.SeverityLow }
func (x *lifeEventEC2) Description() string {
	return "Monitoring events of EC2 instance creation and destruction"
}

func (x *lifeEventEC2) Match(record *models.CloudTrailRecord) bool {
	if record.SourceIPAddress == "autoscaling.amazonaws.com" ||
		record.SourceIPAddress == "batch.amazonaws.com" {
		return false
	}

	return record.EventSource == "ec2.amazonaws.com" &&
		x.targetEvents[record.EventName] == true
}
