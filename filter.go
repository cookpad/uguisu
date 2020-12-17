package uguisu

import "github.com/m-mizutani/uguisu/pkg/models"

// AlertFilter is interface to modify or drop detected alert before notifying
type AlertFilter interface {
	// Filter enables to
	// - modify alert by changing values of `alert`
	// - drop alert if you do not want to notify the alert by returning false
	Filter(alert *models.Alert) bool
}

// AlertFilters is set of AlertFilter
type AlertFilters []AlertFilter

func (x AlertFilters) filter(alert *models.Alert) bool {
	if x == nil {
		return true // All alert should be passed if no filter
	}

	for _, filter := range x {
		if filter.Filter(alert) == false {
			return false
		}
	}

	return true
}
