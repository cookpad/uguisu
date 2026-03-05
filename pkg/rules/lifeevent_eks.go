package rules

import (
	"github.com/cookpad/uguisu/pkg/models"
)

type lifeEventEKS struct {
	targetEvents map[string]bool
}

func newLifeEventEKS() models.Rule {
	return &lifeEventEKS{
		targetEvents: map[string]bool{
			"CreateCluster": true,
			"DeleteCluster": true,
		},
	}
}

func (x *lifeEventEKS) ID() string                { return "resource_lifeevent_eks" }
func (x *lifeEventEKS) Title() string             { return "EKS Cluster Life Event" }
func (x *lifeEventEKS) Severity() models.Severity { return models.SeverityHigh }
func (x *lifeEventEKS) Description() string {
	return "Monitoring creation and deletion of EKS clusters"
}

func (x *lifeEventEKS) Match(record *models.CloudTrailRecord) bool {
	return record.EventSource == "eks.amazonaws.com" &&
		x.targetEvents[record.EventName]
}
