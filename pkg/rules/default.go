package rules

import "github.com/m-mizutani/uguisu/pkg/models"

// NewDefaultRuleSet returns default rule set of
func NewDefaultRuleSet() *models.RuleSet {
	return &models.RuleSet{
		Rules: []models.Rule{
			// AWS CIS monitoring rules
			newAwsCIS3_1(),
			newAwsCIS3_2(),
			newAwsCIS3_3(),
			newAwsCIS3_4(),
			newAwsCIS3_5(),
			newAwsCIS3_6(),
			newAwsCIS3_7(),
			newAwsCIS3_8(),
			newAwsCIS3_9(),
			newAwsCIS3_10(),
			newAwsCIS3_11(),
			newAwsCIS3_12(),
			newAwsCIS3_13(),
			newAwsCIS3_14(),

			// AWS resource life events
			newLifeEventACM(),
			newLifeEventEC2(),
			newLifeEventRDS(),
		},
	}
}
