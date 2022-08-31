package wpool

import (
	"context"
	"sync"

	"github.com/rs/zerolog/log"
)

type job struct {
	execute func(ctx context.Context)
}

type WorkerPool struct {
	workersCount int
	jobs         chan job
	done         chan struct{}
}

func New(count int) *WorkerPool {
	return &WorkerPool{
		workersCount: count,
		jobs:         make(chan job, count),
		done:         make(chan struct{}),
	}
}

func worker(ctx context.Context, wg *sync.WaitGroup, jobs <-chan job) {
	defer wg.Done()
	for {
		select {
		case j, ok := <-jobs:
			if !ok {
				return
			}
			j.execute(ctx)
		case <-ctx.Done():
			log.Debug().Err(ctx.Err()).Msg("Worker cancelled")
			return
		}
	}
}

func (wp *WorkerPool) Run(ctx context.Context) {
	var wg sync.WaitGroup

	for i := 0; i < wp.workersCount; i++ {
		wg.Add(1)
		go worker(ctx, &wg, wp.jobs)
	}

	wg.Wait()
	close(wp.jobs)
	close(wp.done)
}

func (wp *WorkerPool) Do(f func(ctx context.Context)) {
	wp.jobs <- job{execute: f}
}

func (wp *WorkerPool) Wait() {
	<-wp.done
}
