package rules

import (
	"github.com/cookpad/uguisu/pkg/models"
)

type lifeEventIAM struct {
	targetEvents map[string]bool
}

func newLifeEventIAM() models.Rule {
	return &lifeEventIAM{
		targetEvents: map[string]bool{
			"CreateUser":                true,
			"DeleteUser":                true,
			"CreateRole":                true,
			"DeleteRole":                true,
			"CreateAccessKey":           true,
			"DeleteAccessKey":           true,
			"CreateLoginProfile":        true,
			"DeleteLoginProfile":        true,
			"AddUserToGroup":            true,
			"RemoveUserFromGroup":       true,
		},
	}
}

func (x *lifeEventIAM) ID() string                { return "resource_lifeevent_iam" }
func (x *lifeEventIAM) Title() string             { return "IAM Principal Life Event" }
func (x *lifeEventIAM) Severity() models.Severity { return models.SeverityHigh }
func (x *lifeEventIAM) Description() string {
	return "Monitoring creation and deletion of IAM users, roles, access keys, and group membership changes"
}

func (x *lifeEventIAM) Match(record *models.CloudTrailRecord) bool {
	return x.targetEvents[record.EventName]
}
