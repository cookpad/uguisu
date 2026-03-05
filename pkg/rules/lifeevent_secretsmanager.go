package rules

import (
	"github.com/cookpad/uguisu/pkg/models"
)

type lifeEventSecretsManager struct {
	targetEvents map[string]bool
}

func newLifeEventSecretsManager() models.Rule {
	return &lifeEventSecretsManager{
		targetEvents: map[string]bool{
			"CreateSecret":        true,
			"DeleteSecret":        true,
			"UpdateSecret":        true,
			"RotateSecret":        true,
			"PutResourcePolicy":   true,
			"DeleteResourcePolicy": true,
		},
	}
}

func (x *lifeEventSecretsManager) ID() string                { return "resource_lifeevent_secretsmanager" }
func (x *lifeEventSecretsManager) Title() string             { return "Secrets Manager Life Event" }
func (x *lifeEventSecretsManager) Severity() models.Severity { return models.SeverityHigh }
func (x *lifeEventSecretsManager) Description() string {
	return "Monitoring creation, deletion, updates, rotation, and resource policy changes on Secrets Manager secrets"
}

func (x *lifeEventSecretsManager) Match(record *models.CloudTrailRecord) bool {
	return record.EventSource == "secretsmanager.amazonaws.com" &&
		x.targetEvents[record.EventName]
}
