package main

import (
	"context"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"time"
)

// TaskController rate limit the concurrent task execution
type TaskController struct {
	rate        time.Duration
	concurrency int
	taskCh      chan Task
}

// TaskTrace trace task execution
type TaskTrace struct {
	TraceID    string
	Created    time.Time
	Ticked     time.Time
	Dispatched time.Time
	Engaged    time.Time
	Started    time.Time
	Completed  time.Time
	WorkerID   int
}

// Task function type
type Task func(context.Context, TaskTrace)

// NewController create new TaskController
func NewController(callsPerMinute int, concurrency int) *TaskController {
	mspc := int64(time.Minute) / int64(callsPerMinute)
	controller := TaskController{
		rate:        time.Duration(mspc),
		concurrency: concurrency,
		taskCh:      make(chan Task, concurrency),
	}
	return &controller
}

// End to release resources
func (ctrl *TaskController) End() {
	close(ctrl.taskCh)
}

// Start task controller processing
func (ctrl *TaskController) Start(ctx context.Context) {
	dispatchCh := make(chan struct {
		t TaskTrace
		w Task
	}, ctrl.concurrency)
	done := false

	wg.Add(1)
	go func() {
		defer wg.Done()

		for !done {
			t := time.Now()
			log.Debug().Msg("TaskController tick")
			select {
			case <-ctx.Done():
				done = true
				close(dispatchCh)
				log.Debug().Msg("TaskController exit")
				return
			case work := <-ctrl.taskCh:
				log.Debug().Msg("TaskController received task")

				tt := TaskTrace{
					TraceID:    uuid.New().String(),
					Ticked:     t,
					Dispatched: time.Now(),
				}
				dispatchCh <- struct {
					t TaskTrace
					w Task
				}{t: tt, w: work}

				log.Debug().Msg("TaskController dispatched task")
			}
			<-time.Tick(ctrl.rate)
		}
	}()

	wg.Add(ctrl.concurrency)
	for i := 1; i <= ctrl.concurrency; i++ {
		go func(id int) {
			defer wg.Done()

			const worker = "worker"
			for !done {
				log.Debug().Int(worker, id).Msg("wait for task")
				select {
				case <-ctx.Done():
					done = true
					break
				case task := <-dispatchCh:
					task.t.WorkerID = id
					task.t.Engaged = time.Now()
					log.Debug().Int(worker, id).Msg("start exec task")
					task.w(ctx, task.t)
					log.Debug().Int(worker, id).Msg("complete exec task")
				}
			}
			log.Info().Int(worker, id).Msg("worker exit")
		}(i)
	}
}
