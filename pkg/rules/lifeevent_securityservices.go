package rules

import (
	"github.com/cookpad/uguisu/pkg/models"
)

type lifeEventSecurityServices struct {
	targetEvents map[string]bool
}

func newLifeEventSecurityServices() models.Rule {
	return &lifeEventSecurityServices{
		targetEvents: map[string]bool{
			// GuardDuty
			"DeleteDetector":                  true,
			"DisassociateFromMasterAccount":   true,
			"DisassociateFromAdministratorAccount": true,
			// Security Hub
			"DisableSecurityHub":              true,
			"DeleteInsight":                   true,
			// CloudWatch alarms
			"DeleteAlarms":                    true,
			"DisableAlarmActions":             true,
		},
	}
}

func (x *lifeEventSecurityServices) ID() string                { return "resource_lifeevent_security_services" }
func (x *lifeEventSecurityServices) Title() string             { return "Security Service Disabled" }
func (x *lifeEventSecurityServices) Severity() models.Severity { return models.SeverityHigh }
func (x *lifeEventSecurityServices) Description() string {
	return "Monitoring disabling or deletion of security services including GuardDuty, Security Hub, and CloudWatch alarms"
}

func (x *lifeEventSecurityServices) Match(record *models.CloudTrailRecord) bool {
	return x.targetEvents[record.EventName]
}
