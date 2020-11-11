package richman

import (
	"context"
)

type Monitor interface {
	Schedule(ctx context.Context, scheduler Scheduler, task Task) error
}


