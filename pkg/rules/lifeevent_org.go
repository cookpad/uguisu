package rules

import (
	"github.com/m-mizutani/uguisu/pkg/models"
)

type lifeEventOrg struct {
	targetEvents map[string]bool
}

func newLifeEventOrg() models.Rule {
	return &lifeEventOrg{
		targetEvents: map[string]bool{
			"CreateAccount":      true,
			"CreateOrganization": true,
			"DeleteOrganization": true,
			"AcceptHandshake":    true,
			"LeaveOrganization":  true,
		},
	}
}

func (x *lifeEventOrg) ID() string                { return "resource_lifeevent_org" }
func (x *lifeEventOrg) Title() string             { return "Organization/Account Life Event" }
func (x *lifeEventOrg) Severity() models.Severity { return models.SeverityHigh }
func (x *lifeEventOrg) Description() string {
	return "Monitoring events of organization/account"
}

func (x *lifeEventOrg) Match(record *models.CloudTrailRecord) bool {
	return x.targetEvents[record.EventName]
}
