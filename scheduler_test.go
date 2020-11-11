package richman

import (
	"context"
	"time"
)

func Example_cronScheduler() {
	scheduler := NewCronScheduler("*/3 * * * * *")

	signals := scheduler.Start(context.Background())

	timer := time.After(5 * time.Second)
	select {
	case <-signals:
		println("signal")
	case <-timer:
		println("timeout")
	}
}
