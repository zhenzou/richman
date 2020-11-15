package richman

import (
	"context"
	"log"
	"time"
)

type Monitor interface {
	AddJob(name string, job Job) error
	Start()
	Stop(ctx context.Context)
}

func NewMonitor() Monitor {
	return &monitor{
		jobs:        map[string]Job{},
		cancelFuncs: map[string]context.CancelFunc{},
	}
}

type monitor struct {
	jobs        map[string]Job
	cancelFuncs map[string]context.CancelFunc
}

func (m *monitor) AddJob(name string, job Job) error {
	m.jobs[name] = job
	return nil
}

func (m *monitor) Start() {
	for name, job := range m.jobs {
		m.cancelFuncs[name] = job.Schedule(context.Background())
	}
}

// TODO graceful
func (m *monitor) Stop(ctx context.Context) {
	for _, cancelFunc := range m.cancelFuncs {
		cancelFunc()
	}
	select {
	case <-ctx.Done():
		log.Println("[monitor] stop timeout")
	case <-time.After(5 * time.Second):
		log.Println("[monitor] stopped")
	}
}

type Job interface {
	Schedule(context.Context) context.CancelFunc
}

func NewJob(name string, scheduler Scheduler, task Task) Job {
	return &job{
		name: name,
		sch:  scheduler,
		task: task,
	}
}

type job struct {
	name string
	sch  Scheduler
	task Task
}

func (j *job) Schedule(parent context.Context) context.CancelFunc {

	ctx, cancelFunc := context.WithCancel(parent)

	go func() {
		defer cancelFunc()

		signal := j.sch.Start(ctx)
		for {
			select {
			case _, ok := <-signal:
				if !ok {
					log.Printf("[%s] job scheduler stopped\n", j.name)
					return
				}
				err := j.task.Run(ctx)
				if err != nil {
					log.Printf("[%s] job run error for %s\n", j.name, err.Error())
				}
			case <-ctx.Done():
				log.Printf("[%s] job canceled\n", j.name)
				return
			}
		}
	}()

	return cancelFunc
}
