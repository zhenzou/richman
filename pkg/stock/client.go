package stock

import (
	"context"
	"time"

	"github.com/shopspring/decimal"
)

type Stock struct {
	Name         string
	BeginPrice   decimal.Decimal
	NowPrice     decimal.Decimal
	HighestPrice decimal.Decimal
	LowestPrice  decimal.Decimal
	Time         time.Time
}

func (s *Stock) IncreaseRate() string {
	return s.NowPrice.Sub(s.BeginPrice).Div(s.BeginPrice).Mul(decimal.NewFromFloat(100)).String()
}

func (s *Stock) Increase() string {
	return s.NowPrice.Sub(s.BeginPrice).String()
}

type Provider interface {
	List(ctx context.Context, codes ...string) ([]Stock, error)
}

func parsePrice(str string) decimal.Decimal {
	d, err := decimal.NewFromString(str)
	if err != nil {
		panic(err)
	}
	return d
}
