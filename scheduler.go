package richman

import (
	"context"
	"log"

	"github.com/robfig/cron/v3"
)

type Scheduler interface {
	Start() <-chan struct{}
	Stop(ctx context.Context)
}

type CronSchedulerConfig struct {
	Cron string `yaml:"cron"`
}

// NewCronScheduler
func NewCronScheduler(config CronSchedulerConfig) Scheduler {
	c := &cronScheduler{
		ch:   make(chan struct{}),
		cron: cron.New(cron.WithSeconds()),
	}
	_, err := c.cron.AddFunc(config.Cron, c.signal)
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
	select {
	case c.ch <- struct{}{}:
	default:
		log.Printf("[schduler] signal timeout,drop current")
	}
}

func (c *cronScheduler) Start() <-chan struct{} {
	c.cron.Start()
	return c.ch
}

func (c *cronScheduler) Stop(ctx context.Context) {
	cronCtx := c.cron.Stop()
	select {
	case <-ctx.Done():
		log.Println("shutdown timeout,force exit")
	case <-cronCtx.Done():
		log.Println("shutdown success")
	}
	close(c.ch)
}
