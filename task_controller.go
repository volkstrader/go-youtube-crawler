package main

import (
	"context"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"sync"
	"time"
)

// TaskController rate limit the concurrent task execution
type TaskController struct {
	rate         time.Duration
	concurrency  int
	taskCh       chan Task
	doneCh       chan struct{}
	activeWorker int
	mux          sync.Mutex
	ended        bool
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
		doneCh:      make(chan struct{}),
	}
	return &controller
}

// End to release resources
func (ctrl *TaskController) End() {
	if !ctrl.ended {
		close(ctrl.taskCh)
	}
	ctrl.ended = true
}

type dispatcher struct {
	t TaskTrace
	w Task
}

// Start task controller processing
func (ctrl *TaskController) Start(ctx context.Context) {
	dispatchCh := make(chan dispatcher, ctrl.concurrency)
	done := false

	wg.Add(1)
	go func() {
		defer func() {
			close(dispatchCh)
			wg.Done()
		}()

		for !done {
			t := time.Now()
			log.Debug().Msg("TaskController tick")
			select {
			case <-ctx.Done():
				done = true
				log.Debug().Msg("TaskController exit")
				return
			case work := <-ctrl.taskCh:
				if work == nil {
					done = true
					log.Info().Msg("TaskController taskCh closed, end of task")
					return
				}

				log.Debug().Msg("TaskController received task")
				tt := TaskTrace{
					TraceID:    uuid.New().String(),
					Ticked:     t,
					Dispatched: time.Now(),
				}

				dispatchCh <- dispatcher{t: tt, w: work}
				log.Debug().Msg("TaskController dispatched task")
			}
			<-time.Tick(ctrl.rate)
		}
	}()

	wg.Add(ctrl.concurrency)

	for i := 1; i <= ctrl.concurrency; i++ {
		ctrl.activeWorker++
		go func(id int) {
			defer func() {
				ctrl.mux.Lock()
				ctrl.activeWorker--
				if ctrl.activeWorker <= 0 {
					ctrl.doneCh <- struct{}{}
					close(ctrl.doneCh)
				}
				ctrl.mux.Unlock()

				wg.Done()
			}()

			const worker = "worker"
			for !done {
				log.Debug().Int(worker, id).Msg("wait for task")
				select {
				case <-ctx.Done():
					done = true
					break
				case task := <-dispatchCh:
					if task.w == nil {
						log.Info().Int("worker", id).Msg("end of dispatch")
						return
					}
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
