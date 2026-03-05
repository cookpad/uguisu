package rules

import (
	"github.com/cookpad/uguisu/pkg/models"
)

type lifeEventLambda struct {
	targetEvents map[string]bool
}

func newLifeEventLambda() models.Rule {
	return &lifeEventLambda{
		targetEvents: map[string]bool{
			"CreateFunction20150331": true,
			"DeleteFunction20150331": true,
			"UpdateFunctionCode20150331v2": true,
			"AddPermission20150331v2": true,
		},
	}
}

func (x *lifeEventLambda) ID() string                { return "resource_lifeevent_lambda" }
func (x *lifeEventLambda) Title() string             { return "Lambda Function Life Event" }
func (x *lifeEventLambda) Severity() models.Severity { return models.SeverityMedium }
func (x *lifeEventLambda) Description() string {
	return "Monitoring creation, deletion, code updates, and permission changes on Lambda functions"
}

func (x *lifeEventLambda) Match(record *models.CloudTrailRecord) bool {
	return record.EventSource == "lambda.amazonaws.com" &&
		x.targetEvents[record.EventName]
}
