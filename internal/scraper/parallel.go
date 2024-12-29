package scraper

// 90% taken from here
// https://storj.dev/blog/production-concurrency#batch-processing-a-slice

import (
	"context"
	"sync"
)

type Limiter struct {
	limit   chan struct{}
	working sync.WaitGroup
}

func NewLimiter(n int) *Limiter {
	return &Limiter{limit: make(chan struct{}, n)}
}

func (lim *Limiter) Go(ctx context.Context, fn func()) bool {
	// ensure that we aren't trying to start when the
	// context has been cancelled.
	if ctx.Err() != nil {
		return false
	}

	// wait until we can start a goroutine:
	select {
	case lim.limit <- struct{}{}:
	case <-ctx.Done():
		// maybe the user got tired of waiting?
		return false
	}

	lim.working.Add(1)
	go func() {
		defer func() {
			<-lim.limit
			lim.working.Done()
		}()

		fn()
	}()

	return true
}

func (lim *Limiter) Wait() {
	lim.working.Wait()
}

type Parallel struct {
	Concurrency int
	BatchSize   int
}

func (p Parallel) Process(ctx context.Context, total int, process func(start, end int)) error {
	if p.Concurrency < 1 {
		p.Concurrency = 1
	}
	if p.BatchSize < 1 {
		p.BatchSize = 1
	}

	lim := NewLimiter(p.Concurrency)
	defer lim.Wait()

	for start := 0; start < total; start += p.BatchSize {
		start, end := start, start+p.BatchSize
		if end > total {
			end = total
		}

		started := lim.Go(ctx, func() {
			process(start, end)
		})
		if !started {
			return ctx.Err()
		}
	}

	return nil
}
