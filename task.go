package richman

import (
	"context"
)

type Task interface {
	Run(ctx context.Context) error
}
