package rules

import (
	"github.com/cookpad/uguisu/pkg/models"
)

type lifeEventSecurityServices struct {
	guardDutyEvents   map[string]bool
	securityHubEvents map[string]bool
	cloudWatchEvents  map[string]bool
}

func newLifeEventSecurityServices() models.Rule {
	return &lifeEventSecurityServices{
		guardDutyEvents: map[string]bool{
			"DeleteDetector":                       true,
			"DisassociateFromMasterAccount":        true,
			"DisassociateFromAdministratorAccount": true,
		},
		securityHubEvents: map[string]bool{
			"DisableSecurityHub": true,
			"DeleteInsight":      true,
		},
		cloudWatchEvents: map[string]bool{
			"DeleteAlarms":        true,
			"DisableAlarmActions": true,
		},
	}
}

func (x *lifeEventSecurityServices) ID() string                { return "resource_lifeevent_security_services" }
func (x *lifeEventSecurityServices) Title() string             { return "Security Services Disabled/Deleted" }
func (x *lifeEventSecurityServices) Severity() models.Severity { return models.SeverityHigh }
func (x *lifeEventSecurityServices) Description() string {
	return "Monitoring disabling or deletion of security services including GuardDuty, Security Hub, and CloudWatch alarms"
}

func (x *lifeEventSecurityServices) Match(record *models.CloudTrailRecord) bool {
	switch record.EventSource {
	case "guardduty.amazonaws.com":
		return x.guardDutyEvents[record.EventName]
	case "securityhub.amazonaws.com":
		return x.securityHubEvents[record.EventName]
	case "monitoring.amazonaws.com":
		return x.cloudWatchEvents[record.EventName]
	}
	return false
}
