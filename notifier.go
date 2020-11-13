package richman

import "context"

type Notification struct {
	Title string
	Body  string
}

type Notifier interface {
	Send(ctx context.Context, notification Notification) error
}
