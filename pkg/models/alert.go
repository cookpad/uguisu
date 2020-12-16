package models

// Alert represents notification data set to Slack
type Alert struct {
	DetectedBy  Rule
	Title       string
	RuleID      string
	Sev         Severity
	Description string
	Events      []*CloudTrailRecord
}

type Severity string

const (
	SeverityHigh   Severity = "high"
	SeverityMedium Severity = "medium"
	SeverityLow    Severity = "low"
)
