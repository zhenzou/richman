package richman

import (
	"context"
	"log"

	"github.com/robfig/cron/v3"
)

type Scheduler interface {
	Start(ctx context.Context) <-chan struct{}
	Shutdown(ctx context.Context)
}

// NewCronScheduler
func NewCronScheduler(spec string) Scheduler {
	c := &cronScheduler{
		ch:   make(chan struct{}),
		cron: cron.New(cron.WithSeconds()),
	}
	_, err := c.cron.AddFunc(spec, c.signal)
	if err != nil {
		panic(err)
	}
	return c
}

type cronScheduler struct {
	ch   chan struct{}
	cron *cron.Cron
}

func (c *cronScheduler) signal() {
	c.ch <- struct{}{}
}

func (c *cronScheduler) Start(ctx context.Context) <-chan struct{} {
	c.cron.Start()
	go func() {
		select {
		case <-ctx.Done():
			c.cron.Stop()
		}
	}()
	return c.ch
}

func (c *cronScheduler) Shutdown(ctx context.Context) {
	cronCtx := c.cron.Stop()
	select {
	case <-ctx.Done():
		log.Println("shutdown timeout,force exit")
	case <-cronCtx.Done():
		log.Println("shutdown success")
	}
	close(c.ch)
}
