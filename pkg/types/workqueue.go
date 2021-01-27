package types

import "context"

// WorkQueue define workqueue
type WorkQueue interface {
	Add(item interface{})
	Run(ctx context.Context) error
}
