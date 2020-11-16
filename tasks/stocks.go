package tasks

import (
	"context"
	"log"

	"github.com/zhenzou/richman"
	"github.com/zhenzou/richman/pkg/stock"
)

type StockHandler = func(ctx context.Context, stock stock.Stock) error

type StockConfig struct {
	Provider string   `yaml:"provider"`
	Stocks   []string `yaml:"stocks"`
}

func NewStockTask(config StockConfig, handler StockHandler) richman.Task {
	p := stock.GetProvider(config.Provider)
	return &StockTask{
		provider: p,
		stocks:   config.Stocks,
		handler:  handler,
	}
}

type StockTask struct {
	provider stock.Provider
	stocks   []string
	handler  StockHandler
}

func (s *StockTask) Run(ctx context.Context) error {
	stocks, err := s.provider.List(ctx, s.stocks...)
	if err != nil {
		return err
	}
	for _, stock := range stocks {
		log.Printf("stock:%s increse:%s", stock.Name, stock.IncreaseRate())
		if s.handler != nil {
			err = s.handler(ctx, stock)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
