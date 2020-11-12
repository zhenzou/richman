package stock

import (
	"context"
	"errors"
	"time"

	"github.com/shopspring/decimal"
)

// TODO factory
func GetProvider(name string) Provider {
	switch name {
	case "sina":
		return NewSina()
	default:
		panic(errors.New("unsupported provider"))
	}
}

type Stock struct {
	Name         string
	OpenPrice    decimal.Decimal // 开盘价格
	BeginPrice   decimal.Decimal // 昨日收盘价格
	NowPrice     decimal.Decimal
	HighestPrice decimal.Decimal
	LowestPrice  decimal.Decimal
	Time         time.Time
}

func (s Stock) IncreaseRate() string {
	return s.NowPrice.
		Sub(s.BeginPrice).
		Div(s.BeginPrice).
		Mul(decimal.NewFromInt(100)).
		Round(2).
		String() + "%"
}

func (s Stock) Increase() string {
	return s.NowPrice.Sub(s.OpenPrice).Round(4).String()
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
