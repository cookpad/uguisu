package models

import (
	"fmt"
	"strings"
)

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

// Disable removes rules whose ID appears in the comma-separated disabledIDs
// string (e.g. the DISABLED_RULES environment variable).
func (x *RuleSet) Disable(disabledIDs string) {
	if disabledIDs == "" {
		return
	}
	disabled := make(map[string]bool)
	for _, id := range strings.Split(disabledIDs, ",") {
		disabled[strings.TrimSpace(id)] = true
	}
	filtered := x.Rules[:0]
	for _, rule := range x.Rules {
		if !disabled[rule.ID()] {
			filtered = append(filtered, rule)
		}
	}
	x.Rules = filtered
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
			return fmt.Errorf("rule.ID conflict in RuleSet: id=%q titleA=%q titleB=%q",
				rule.ID(), rule.Title(), dup.Title())
		}

		ruleMap[rule.ID()] = rule
	}

	return nil
}
