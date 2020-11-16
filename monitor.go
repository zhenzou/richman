package richman

import (
	"context"
	"github.com/zhenzou/richman/utils"
	"log"
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
		config: config,
		jobs:   map[string]Job{},
		tasks:  map[string]Task{},
	}
	return monitor
}

type Monitor interface {
	RegisterTask(name string, task Task) error
	Start()
	Stop(ctx context.Context)
}

type Job interface {
	Start()
	Stop(context.Context)
}

type monitor struct {
	config Config
	tasks  map[string]Task
	jobs   map[string]Job
}

func (m *monitor) RegisterTask(name string, task Task) error {
	m.tasks[name] = task
	return nil
}

func (m *monitor) Start() {
	m.initJobs()

	for name, job := range m.jobs {
		job.Start()
		log.Println("[monitor] started job ", name)
	}
}

func (m *monitor) Stop(ctx context.Context) {
	for _, j := range m.jobs {
		j.Stop(ctx)
	}
}

func (m *monitor) initJobs() {
	tasks := m.tasks
	conf := m.config
	for name, job := range conf.Jobs {
		switch job.Schedule.Type {
		case "cron":
			cfg := CronSchedulerConfig{}
			err := job.Schedule.Params.Unmarshal(&cfg)
			if err != nil {
				log.Printf("[job] %s read config error \n", name)
				utils.Die()
			}
			scheduler := NewCronScheduler(cfg)
			task, ok := tasks[job.Task]
			if !ok {
				log.Printf("[job] task %s not found for %s\n", job.Task, name)
				utils.Die()
			}
			m.jobs[name] = NewJob(name, scheduler, task)
		default:
			log.Printf("%s scheduler does not support for now\n", job.Schedule.Type)
		}
	}
}

func NewJob(name string, scheduler Scheduler, task Task) Job {
	return &job{
		name: name,
		sch:  scheduler,
		task: task,
		ch:   make(chan struct{}),
	}
}

type job struct {
	name string
	sch  Scheduler
	task Task
	ch   chan struct{}
}

func (j *job) Start() {

	go func() {

		signal := j.sch.Start()
		for {
			select {
			case _, ok := <-signal:
				if !ok {
					log.Printf("[%s] job scheduler stopped\n", j.name)
					return
				}
				err := j.task.Run(context.Background())
				if err != nil {
					log.Printf("[%s] job run error for %s\n", j.name, err.Error())
				}
			case <-j.ch:
				log.Printf("[%s] job to stop\n", j.name)
				close(j.ch)
				return
			}
		}
	}()
}

func (j *job) Stop(ctx context.Context) {
	j.ch <- struct{}{}
	<-j.ch
	select {
	case <-ctx.Done():
		log.Printf("[job] stoped %s timeout\n", j.name)
	case <-j.ch:
		log.Printf("[job] stoped %s success\n", j.name)
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
