package richman

import (
	"context"
	"log"
	"time"

	"github.com/zhenzou/richman/utils"
)

type Config struct {
	Tasks map[string]struct {
		Type   string `yaml:"type"`
		Params Params `yaml:"params"`
	} `yaml:"tasks"`
	Jobs map[string]struct {
		Schedule struct {
			Type   string `yaml:"type"`
			Params Params `yaml:"params"`
		} `yaml:"schedule"`
		Task string `yaml:"task"`
	} `yaml:"jobs"`
}

func NewMonitor(config Config) Monitor {
	tasks := initTasks(config)
	jobs := initJobs(config, tasks)
	monitor := &monitor{
		jobs:        map[string]Job{},
		cancelFuncs: map[string]context.CancelFunc{},
	}

	for name, job := range jobs {
		err := monitor.AddJob(name, job)
		if err != nil {
			log.Println("add job error:", err.Error())
			utils.Die()
		}
	}
	return monitor
}

type Monitor interface {
	AddJob(name string, job Job) error
	Start()
	Stop(ctx context.Context)
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

func initTasks(conf Config) map[string]Task {
	tasks := make(map[string]Task)

	for name, task := range conf.Tasks {
		switch task.Type {
		case "stocks":
			cfg := StockTaskConfig{}
			err := task.Params.Unmarshal(&cfg)
			if err != nil {
				log.Println("read stock config error:", err.Error())
				utils.Die()
			}
			stockTask := NewStockTask(cfg)
			tasks[name] = stockTask
		default:
			log.Printf("%s task does not support for now\n", task.Type)
		}
	}

	return tasks
}

func initJobs(conf Config, tasks map[string]Task) map[string]Job {

	jobs := map[string]Job{}

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
			jobs[name] = NewJob(name, scheduler, task)
		default:
			log.Printf("%s scheduler does not support for now\n", job.Schedule.Type)
		}
	}

	return jobs
}
