package main

import (
	"context"
	"github.com/rs/zerolog/log"
	"testing"
	"time"
)

func newTask(i int, nap time.Duration, completedCh chan<- TaskTrace) Task {
	created := time.Now()
	return func(ctx context.Context, trace TaskTrace) {
		trace.Created = created
		trace.Started = time.Now()
		log.Info().
			Int("task", i).
			Int("worker", trace.WorkerID).
			Msg("task started")
		time.Sleep(nap)
		trace.Completed = time.Now()
		log.Info().
			Int("task", i).
			Int("worker", trace.WorkerID).
			Interface("trace", trace).
			Msg("task completed")
		completedCh <- trace
	}
}

func TestTaskController_Start(t *testing.T) {
	const n = 10
	ctrl := NewController(5, 5)
	traceCh := make(chan TaskTrace, 1)
	defer close(traceCh)

	ctx, cancel := context.WithCancel(context.Background())
	ctrl.Start(ctx)

	go func() {
		for i := 1; i <= n; i++ {
			log.Info().
				Int("task", i).
				Msg("submit task")
			ctrl.taskCh <- newTask(i, 2*time.Second, traceCh)
			log.Info().
				Int("task", i).
				Msg("accepted task")
		}
	}()

	j := 0
	for range traceCh {
		j++
		log.Info().
			Int("completed", j).
			Msg("task ended")
		if j >= n {
			break
		}
	}
	cancel()
	ctrl.End()
}
