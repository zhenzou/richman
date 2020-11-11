package stock

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStock_Increase(t *testing.T) {
	stock := Stock{
		BeginPrice: parsePrice("10.00"),
		NowPrice:   parsePrice("5"),
	}

	assert.Equal(t, "-50", stock.IncreaseRate())
}
