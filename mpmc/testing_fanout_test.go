package mpmc

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestFanoutAll(t *testing.T) {
	fanout := NewProducer[int](ProducerKind_All, 65535, 65535)
	defer fanout.Close()

	numProducers := 3
	numConsumers := 5
	itemsPerProducer := 10000
	totalItems := numProducers * itemsPerProducer

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	var wg sync.WaitGroup
	results := make([][]int, numConsumers)

	// Create consumers
	for i := 0; i < numConsumers; i++ {
		consumer := fanout.CreateConsumer(ctx)
		results[i] = make([]int, 0, totalItems)

		wg.Add(1)
		go func(c *Consumer[int], resultSlice *[]int) {
			defer wg.Done()
			for {
				select {
				case item := <-c.Messages:
					*resultSlice = append(*resultSlice, item)
				case <-ctx.Done():
					return
				}
			}
		}(consumer, &results[i])
	}

	// Create producers
	for i := 0; i < numProducers; i++ {
		wg.Add(1)
		go func(producerID int) {
			defer wg.Done()
			for j := 0; j < itemsPerProducer; j++ {
				select {
				case fanout.input <- producerID*itemsPerProducer + j:
				case <-ctx.Done():
					return
				}
			}
		}(i)
	}

	wg.Wait()

	// Verify results
	for i, result := range results {
		if len(result) != totalItems {
			t.Errorf("Consumer %d received %d items, expected %d", i, len(result), totalItems)
		}
	}

	// Check if all consumers received the same items
	for i := 1; i < numConsumers; i++ {
		if !equalSlices(results[0], results[i]) {
			t.Errorf("Consumer %d received different items than Consumer 0", i)
		}
	}
}

func TestFanoutSingle(t *testing.T) {
	fanout := NewProducer[int](ProducerKind_Single, 65535, 65535)
	defer fanout.Close()

	numProducers := 3
	numConsumers := 5
	itemsPerProducer := 10000
	totalItems := numProducers * itemsPerProducer

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	var wg sync.WaitGroup
	results := make([][]int, numConsumers)

	// Create consumers
	for i := 0; i < numConsumers; i++ {
		consumer := fanout.CreateConsumer(ctx)
		results[i] = make([]int, 0, totalItems/numConsumers)

		wg.Add(1)
		go func(c *Consumer[int], resultSlice *[]int) {
			defer wg.Done()
			for {
				select {
				case item := <-c.Messages:
					*resultSlice = append(*resultSlice, item)
				case <-ctx.Done():
					return
				}
			}
		}(consumer, &results[i])
	}

	// Create producers
	tp := time.Now()
	for i := 0; i < numProducers; i++ {
		wg.Add(1)
		go func(producerID int) {
			defer wg.Done()
			t.Logf("Producer %d started %v", producerID, time.Since(tp))
			for j := 0; j < itemsPerProducer; j++ {
				select {
				case fanout.input <- producerID*itemsPerProducer + j:
				case <-ctx.Done():
					return
				}
			}
			t.Logf("Producer %d done %v", producerID, time.Since(tp))
		}(i)
	}

	wg.Wait()
	t.Logf("Producer Time taken: %v", time.Since(tp))

	// Verify results
	totalReceived := 0
	for _, result := range results {
		totalReceived += len(result)
	}

	if totalReceived != totalItems {
		t.Errorf("Total received items: %d, expected: %d", totalReceived, totalItems)
	}

	// Check if items are distributed somewhat evenly
	expectedPerConsumer := totalItems / numConsumers
	tolerance := expectedPerConsumer / 2

	for i, result := range results {
		if len(result) < expectedPerConsumer-tolerance || len(result) > expectedPerConsumer+tolerance {
			t.Errorf("Consumer %d received %d items, expected around %d (tolerance: ±%d)", i, len(result), expectedPerConsumer, tolerance)
		}
	}
}
func TestFanoutLRU(t *testing.T) {
	fanout := NewProducer[int](ProducerKind_LRU, 65535, 65535)
	defer fanout.Close()

	numProducers := 3
	numConsumers := 5
	itemsPerProducer := 10000
	totalItems := numProducers * itemsPerProducer

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	var wg sync.WaitGroup
	results := make([][]int, numConsumers)

	// Create consumers
	for i := 0; i < numConsumers; i++ {
		consumer := fanout.CreateConsumer(ctx)
		results[i] = make([]int, 0, totalItems/numConsumers)

		wg.Add(1)
		go func(c *Consumer[int], resultSlice *[]int) {
			defer wg.Done()
			for {
				select {
				case item := <-c.Messages:
					*resultSlice = append(*resultSlice, item)
				case <-ctx.Done():
					return
				}
			}
		}(consumer, &results[i])
	}

	// Create producers
	tp := time.Now()
	for i := 0; i < numProducers; i++ {
		wg.Add(1)
		go func(producerID int) {
			defer wg.Done()
			t.Logf("Producer %d started %v", producerID, time.Since(tp))
			for j := 0; j < itemsPerProducer; j++ {
				select {
				case fanout.input <- producerID*itemsPerProducer + j:
				case <-ctx.Done():
					return
				}
			}
			t.Logf("Producer %d done %v", producerID, time.Since(tp))
		}(i)
	}

	wg.Wait()
	t.Logf("Producer Time taken: %v", time.Since(tp))

	// Verify results
	totalReceived := 0
	for _, result := range results {
		totalReceived += len(result)
	}

	if totalReceived != totalItems {
		t.Errorf("Total received items: %d, expected: %d", totalReceived, totalItems)
	}

	// Check if items are distributed somewhat evenly
	expectedPerConsumer := totalItems / numConsumers
	tolerance := expectedPerConsumer / 2

	for i, result := range results {
		if len(result) < expectedPerConsumer-tolerance || len(result) > expectedPerConsumer+tolerance {
			t.Errorf("Consumer %d received %d items, expected around %d (tolerance: ±%d)", i, len(result), expectedPerConsumer, tolerance)
		}
	}
}
