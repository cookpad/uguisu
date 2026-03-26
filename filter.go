package uguisu

import (
	"fmt"

	"github.com/cookpad/uguisu/pkg/logx"
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
			logx.Logger.With("filter", fmt.Sprintf("%v", filter)).Debug("Alert filtered")
			return false
		}
	}

	return true
}
