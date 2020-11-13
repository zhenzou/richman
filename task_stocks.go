package richman

import (
	"context"
	"log"

	"github.com/zhenzou/richman/pkg/stock"
)

type StockTaskConfig struct {
	Provider string   `yaml:"provider"`
	Stocks   []string `yaml:"stocks"`
}

func NewStockTask(config StockTaskConfig) Task {
	p := stock.GetProvider(config.Provider)
	return &StockTask{
		provider: p,
		stocks:   config.Stocks,
	}
}

type StockTask struct {
	provider stock.Provider
	stocks   []string
}

func (s *StockTask) Run(ctx context.Context) error {
	stocks, err := s.provider.List(ctx, s.stocks...)
	if err != nil {
		return err
	}
	for _, stock := range stocks {
		log.Printf("stock:%s increse:%s", stock.Name, stock.IncreaseRate())
	}
	return nil
}
