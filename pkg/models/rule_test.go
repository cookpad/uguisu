package models_test

import (
	"testing"

	"github.com/m-mizutani/uguisu/pkg/models"
	"github.com/stretchr/testify/assert"
)

type rule1 struct{}

func (x *rule1) ID() string {
	return "rule1"
}

func (x *rule1) Title() string {
	return "this is rule1"
}

func (x *rule1) Description() string {
	return "test rule1"
}

func (x *rule1) Severity() models.Severity {
	return models.SeverityLow
}

func (x *rule1) Match(record *models.CloudTrailRecord) bool {
	return true
}

func TestRuleSet(t *testing.T) {
	ruleSet := models.RuleSet{
		Rules: []models.Rule{
			&rule1{},
		},
	}

	assert.NoError(t, ruleSet.Diagnosis())

	ruleSet.Rules = append(ruleSet.Rules, &rule1{})
	assert.Error(t, ruleSet.Diagnosis())
}
