// Package mpmc provides a Multi-Producer Multi-Consumer implementation with different fanout strategies.
package mpmc

import (
	"context"
	"errors"
	"math/rand"
	"sort"
	"sync"
	"time"

	"github.com/Moonlight-Companies/gompmc/logger"
)

var (
	// ErrProducerClosed is returned when trying to write to a closed producer.
	ErrProducerClosed = errors.New("producer is closed")
	// ErrBufferFull is returned when the producer's buffer is full and can't accept more items.
	ErrBufferFull = errors.New("buffer is full")
)

// ProducerKind defines the type of fanout strategy used by the producer.
type ProducerKind int

const (
	// ProducerKind_Single sends each item to a single randomly selected consumer.
	ProducerKind_Single ProducerKind = iota
	// ProducerKind_LRU sends each item to the least recently used consumer.
	ProducerKind_LRU
	// ProducerKind_All sends each item to all consumers.
	ProducerKind_All
)

// Producer manages the distribution of items to consumers based on a specified strategy.
type Producer[T any] struct {
	logger               *logger.Logger
	input                chan T
	consumer_buffer_size uint
	consumers            ConsumerList[T]
	consumers_mu         sync.Mutex
	done                 chan struct{}
	closeOnce            sync.Once
}

// NewProducer creates a new Producer with the specified fanout strategy and buffer sizes.
// It returns a pointer to the new Producer.
func NewProducer[T any](kind ProducerKind, input_buffer_size, consumer_buffer_size uint) (result *Producer[T]) {
	result = &Producer[T]{
		logger:               logger.NewLogger(logger.LogLevelDebug, TypeName[T]()),
		input:                make(chan T, input_buffer_size),
		consumer_buffer_size: consumer_buffer_size,
		consumers:            ConsumerList[T]{},
		consumers_mu:         sync.Mutex{},
		done:                 make(chan struct{}),
	}

	result.logger.Debugln("Producer created")

	switch kind {
	case ProducerKind_Single:
		go result.goroutine_Producer_single()
	case ProducerKind_LRU:
		go result.goroutine_Producer_lru()
	case ProducerKind_All:
		go result.goroutine_Producer_all()
	}

	go func() {
		<-result.done
		result.logger.Debugln("Producer closing, closing all consumers")
		result.consumers_mu.Lock()
		for _, consumer := range result.consumers {
			consumer.Close()
		}
		result.consumers_mu.Unlock()
		result.logger.Debugln("Producer closed")
	}()

	return
}

// Write sends an item to the Producer's input channel.
// It returns an error if the Producer is closed or if the buffer is full.
func (f *Producer[T]) Write(item T) error {
	select {
	case f.input <- item:
	case <-f.done:
		f.logger.Warnln("Producer is closed, dropping item")
		return ErrProducerClosed
	default:
		f.logger.Warnln("Producer buffer is full, dropping item")
		return ErrBufferFull
	}
	return nil
}

// CreateConsumer creates a new Consumer associated with this Producer.
// It takes a context for cancellation and returns a pointer to the new Consumer.
func (f *Producer[T]) CreateConsumer(ctx context.Context) (result *Consumer[T]) {
	result = newConsumer(f, ctx, f.consumer_buffer_size)

	f.consumers_mu.Lock()
	f.consumers = append(f.consumers, result)
	f.consumers_mu.Unlock()

	f.logger.Debugln("Consumer", result.id, "created, adding to Producer")

	go func() {
		<-result.ctx.Done()
		f.logger.Debugln("Consumer", result.id, "closed, removing from Producer")
		f.consumers_mu.Lock()
		for i, consumer := range f.consumers {
			if consumer == result {
				f.consumers = append(f.consumers[:i], f.consumers[i+1:]...)
				break
			}
		}
		f.consumers_mu.Unlock()
	}()

	return
}

// Close shuts down the Producer and all associated Consumers.
func (f *Producer[T]) Close() {
	f.closeOnce.Do(func() {
		close(f.done)
	})
}

// goroutine_Producer_single implements the single consumer fanout strategy.
func (f *Producer[T]) goroutine_Producer_single() {
	f.logger.Debugln("goroutine producer single started")
	for {
		select {
		case item := <-f.input:
			f.consumers_mu.Lock()
			if len(f.consumers) > 0 {
				selected := f.consumers[rand.Intn(len(f.consumers))]
				select {
				case selected.input <- item:
					selected.lastUsed = time.Now()
				default:
					f.logger.Warnln("Consumer buffer is full, dropping item")
				}
			} else {
				f.logger.Warnln("No consumers available, dropping item")
			}
			f.consumers_mu.Unlock()
		case <-f.done:
			f.logger.Debugln("goroutine Producer single closing")
			return
		}
	}
}

// goroutine_Producer_lru implements the least recently used consumer fanout strategy.
func (f *Producer[T]) goroutine_Producer_lru() {
	f.logger.Debugln("goroutine producer lru started")
	for {
		select {
		case item := <-f.input:
			f.consumers_mu.Lock()
			if len(f.consumers) > 0 {
				sort.Sort(f.consumers)
				lru := f.consumers[0]
				select {
				case lru.input <- item:
					lru.lastUsed = time.Now()
				default:
					f.logger.Warnln("Consumer buffer is full, dropping item")
				}
			} else {
				f.logger.Warnln("No consumers available, dropping item")
			}
			f.consumers_mu.Unlock()
		case <-f.done:
			f.logger.Debugln("goroutine Producer lru closing")
			return
		}
	}
}

// goroutine_Producer_all implements the all consumers fanout strategy.
func (f *Producer[T]) goroutine_Producer_all() {
	f.logger.Debugln("goroutine producer all started")
	for {
		select {
		case item := <-f.input:
			f.consumers_mu.Lock()
			for _, consumer := range f.consumers {
				select {
				case consumer.input <- item:
					consumer.lastUsed = time.Now()
				default:
					f.logger.Warnln("Consumer buffer is full, dropping item")
				}
			}
			f.consumers_mu.Unlock()
		case <-f.done:
			f.logger.Debugln("goroutine Producer all closing")
			return
		}
	}
}
