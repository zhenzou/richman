package stock

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_sina_convertToStock(t *testing.T) {
	str := `"大秦铁路,6.480,6.490,6.520,6.550,6.450,6.520,6.530,26048333,169706151.000,456838,6.520,843440,6.510,305800,6.500,334600,6.490,158700,6.480,395300,6.530,1535656,6.540,2976299,6.550,1109280,6.560,807600,6.570,2020-11-11,15:00:02,00,"`

	s := sina{}
	stock := s.convertToStock(str)

	assert.Equal(t, "大秦铁路", stock.Name)
	assert.Equal(t, "15:00:02", stock.Time.Format("15:04:05"))
}

func Test_sina_List(t *testing.T) {
	sina := NewSina()

	stocks, err := sina.List(context.Background(), "sh601006")
	if err != nil {
		assert.NoError(t, err)
	}
	stock := stocks[0]
	assert.Equal(t, "大秦铁路", stock.Name)
}
