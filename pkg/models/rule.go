package models

import "github.com/m-mizutani/golambda"

type Rule interface {
	ID() string
	Title() string
	Description() string
	Severity() Severity
	Match(record *CloudTrailRecord) bool
}

// RuleSet is collection of Rule and has Detect method for bulk evaluation
type RuleSet struct {
	Rules []Rule
}

// Detect is bulk evaluation method of rules in the RuleSet. It returns set of Alert that is matched with a rule. It returns nil (0 length array) if no rule is matched
func (x *RuleSet) Detect(record *CloudTrailRecord) []*Alert {
	var alerts []*Alert
	for _, rule := range x.Rules {
		if rule.Match(record) {
			alerts = append(alerts, &Alert{
				DetectedBy:  rule,
				Title:       rule.Title(),
				RuleID:      rule.ID(),
				Sev:         rule.Severity(),
				Description: rule.Description(),
				Events:      []*CloudTrailRecord{record},
			})
		}
	}

	return alerts
}

// Diagnosis checks consistency of RuleSet. Checking conflict Rule ID for now.
func (x *RuleSet) Diagnosis() error {
	ruleMap := make(map[string]Rule)
	for _, rule := range x.Rules {
		if dup, ok := ruleMap[rule.ID()]; ok {
			return golambda.NewError("Rule.ID in RuleSet is conflicted").
				With("id", rule.ID()).
				With("titleA", rule.Title()).
				With("titleB", dup.Title())
		}

		ruleMap[rule.ID()] = rule
	}

	return nil
}
