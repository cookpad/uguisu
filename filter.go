package uguisu

import (
	"fmt"
	"log/slog"

	"github.com/cookpad/uguisu/pkg/models"
)

// Filter is interface to modify or drop detected alert before notifying
type Filter func(alert *models.Alert) bool

// AlertFilters is set of AlertFilter
type AlertFilters []Filter

func (x AlertFilters) filter(alert *models.Alert) bool {
	if x == nil {
		return true // All alert should be passed if no filter
	}

	for _, filter := range x {
		if !filter(alert) {
			slog.Debug("Alert filtered", "filter", fmt.Sprintf("%v", filter))
			return false
		}
	}

	return true
}
