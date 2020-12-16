package rules

import (
	"github.com/m-mizutani/uguisu/pkg/models"
)

type lifeEventACM struct {
	targetEvents map[string]bool
}

func newLifeEventACM() models.Rule {
	return &lifeEventACM{
		targetEvents: map[string]bool{
			"ExportCertificate": true,
			"ImportCertificate": true,
			"RenewCertificate":  true,
			"DeleteCertificate": true,
		},
	}
}

func (x *lifeEventACM) ID() string                { return "resource_lifeevent_acm.5" }
func (x *lifeEventACM) Title() string             { return "ACM certification life event" }
func (x *lifeEventACM) Severity() models.Severity { return models.SeverityMedium }
func (x *lifeEventACM) Description() string {
	return "Monitoring events of ACM creation and destruction"
}

func (x *lifeEventACM) Match(record *models.CloudTrailRecord) bool {
	return x.targetEvents[record.EventName] == true
}
