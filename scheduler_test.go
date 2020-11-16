package richman

import (
	"context"
	"time"
)

func Example_cronScheduler() {
	scheduler := NewCronScheduler(CronSchedulerConfig{
		Cron: "*/3 * * * * *",
	})

	signals := scheduler.Start()
	defer scheduler.Stop(context.Background())

	timer := time.After(5 * time.Second)
	select {
	case <-signals:
		println("signal")
	case <-timer:
		println("timeout")
	}
}
