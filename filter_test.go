package uguisu_test

import (
	"testing"

	"github.com/m-mizutani/uguisu"
	"github.com/m-mizutani/uguisu/pkg/models"
	"github.com/stretchr/testify/assert"
)

type dropCIS3_4 struct{}

func (x *dropCIS3_4) Filter(alert *models.Alert) bool {
	if alert.RuleID == "aws_cis_3.4" { // nolint
		return false
	}

	return true
}

func TestFilterDrop(t *testing.T) {
	t.Run("detect CIS 3.4 with no filter", func(t *testing.T) {
		detected := uguisu.New().Test([]*models.CloudTrailRecord{
			{
				EventName: "DeleteGroupPolicy",
			},
		})
		assert.Equal(t, 1, len(detected))
	})

	t.Run("dropCIS3_4", func(t *testing.T) {
		ug := uguisu.New()
		ug.Filters = append(ug.Filters, &dropCIS3_4{})

		t.Run("drops CIS 3.4 alert", func(t *testing.T) {
			detected := ug.Test([]*models.CloudTrailRecord{
				{
					EventName: "DeleteGroupPolicy",
				},
			})
			assert.Equal(t, 0, len(detected))
		})

		t.Run("does not drop detect CIS 3.5", func(t *testing.T) {

			detected := ug.Test([]*models.CloudTrailRecord{
				{
					EventName: "CreateTrail",
				},
			})
			assert.Equal(t, 1, len(detected))
		})
	})
}
