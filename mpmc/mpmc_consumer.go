package mpmc

import (
	"context"
	"sync"
	"time"
)

// Consumer represents a consumer in the MPMC (Multi-Producer Multi-Consumer) system.
type Consumer[T any] struct {
	id        string
	owner     *Producer[T]
	Messages  chan T
	lastUsed  time.Time
	ctx       context.Context
	cancel    context.CancelFunc
	closeOnce sync.Once
}

// newConsumer creates a new Consumer with the given owner, context, and buffer size.
// It returns a pointer to the new Consumer.
func newConsumer[T any](owner *Producer[T], ctx context.Context, consumer_buffer_size uint) (result *Consumer[T]) {
	ctx, cancel := context.WithCancel(ctx)
	result = &Consumer[T]{
		id:        CreateID(),
		owner:     owner,
		Messages:  make(chan T, consumer_buffer_size),
		lastUsed:  time.Now(),
		ctx:       ctx,
		cancel:    cancel,
		closeOnce: sync.Once{},
	}
	owner.logger.Debugln("Consumer", result.id, "created")
	return
}

// Close shuts down the Consumer.
// It ensures that the close operation is performed only once.
func (c *Consumer[T]) Close() {
	c.closeOnce.Do(func() {
		c.owner.logger.Debugln("Consumer", c.id, "closing")
		c.cancel()
	})
}
