package richman

import (
	"context"
	"log"
	"time"
)

type Config struct {
	Jobs map[string]struct {
		Schedule struct {
			Type   string `yaml:"type"`
			Params Params `yaml:"params"`
		} `yaml:"schedule"`
		Task string `yaml:"task"`
	} `yaml:"jobs"`
}

func NewMonitor(config Config) Monitor {
	monitor := &monitor{
		config:      config,
		jobs:        map[string]Job{},
		tasks:       map[string]Task{},
		cancelFuncs: map[string]context.CancelFunc{},
	}
	return monitor
}

type Monitor interface {
	RegisterTask(name string, task Task) error
	Start()
	Stop(ctx context.Context)
}

type monitor struct {
	config      Config
	tasks       map[string]Task
	jobs        map[string]Job
	cancelFuncs map[string]context.CancelFunc
}

func (m *monitor) RegisterTask(name string, task Task) error {
	m.tasks[name] = task
	return nil
}

func (m *monitor) Start() {
	m.initJobs(m.tasks)

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

func (m *monitor) initJobs(tasks map[string]Task) {
	conf := m.config
	for name, job := range conf.Jobs {
		switch job.Schedule.Type {
		case "cron":
			cfg := CronSchedulerConfig{}
			err := job.Schedule.Params.Unmarshal(&cfg)
			if err != nil {
				log.Printf("[job] %s read config error \n", name)
			}
			scheduler := NewCronScheduler(cfg)
			task, ok := tasks[job.Task]
			if !ok {
				log.Printf("[job] task %s not found for %s\n", job.Task, name)
			}
			m.jobs[name] = NewJob(name, scheduler, task)
		default:
			log.Printf("%s scheduler does not support for now\n", job.Schedule.Type)
		}
	}
}

type Params struct {
	unmarshal func(interface{}) error
}

func (msg *Params) UnmarshalYAML(unmarshal func(interface{}) error) error {
	msg.unmarshal = unmarshal
	return nil
}

func (msg *Params) Unmarshal(v interface{}) error {
	return msg.unmarshal(v)
}
