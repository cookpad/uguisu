package models_test

import (
	"testing"

	"github.com/cookpad/uguisu/pkg/models"
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

type rule2 struct{}

func (x *rule2) ID() string          { return "rule2" }
func (x *rule2) Title() string       { return "this is rule2" }
func (x *rule2) Description() string { return "test rule2" }
func (x *rule2) Severity() models.Severity { return models.SeverityLow }
func (x *rule2) Match(record *models.CloudTrailRecord) bool { return true }

func makeRuleSet() models.RuleSet {
	return models.RuleSet{Rules: []models.Rule{&rule1{}, &rule2{}}}
}

func TestDisable(t *testing.T) {
	t.Run("empty string leaves all rules intact", func(t *testing.T) {
		rs := makeRuleSet()
		rs.Disable("")
		assert.Len(t, rs.Rules, 2)
	})

	t.Run("disables a single rule by ID", func(t *testing.T) {
		rs := makeRuleSet()
		rs.Disable("rule1")
		assert.Len(t, rs.Rules, 1)
		assert.Equal(t, "rule2", rs.Rules[0].ID())
	})

	t.Run("disables multiple rules", func(t *testing.T) {
		rs := makeRuleSet()
		rs.Disable("rule1,rule2")
		assert.Empty(t, rs.Rules)
	})

	t.Run("trims whitespace around IDs", func(t *testing.T) {
		rs := makeRuleSet()
		rs.Disable(" rule1 , rule2 ")
		assert.Empty(t, rs.Rules)
	})

	t.Run("unknown ID is ignored", func(t *testing.T) {
		rs := makeRuleSet()
		rs.Disable("unknown_rule")
		assert.Len(t, rs.Rules, 2)
	})
}
