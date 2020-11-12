package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"os/user"
	"path/filepath"
	"syscall"
	"time"

	"gopkg.in/yaml.v2"

	"github.com/zhenzou/richman"
)

const confFile = "/.richman/conf.yaml"

const confSample = `
tasks:
  get-stocks:
    type: "stocks"
    params:
      provider: "sina"
      stocks: [ "sz002594" ]

jobs:
  monitor-stocks:
    schedule:
      type: "cron"
      cron: '*/6 * * * * *'
    task: "get-stocks"
`

func buildConfPath() string {
	user, err := user.Current()
	if err != nil {
		fmt.Println("get user error:", err.Error())
		os.Exit(-1)
	}
	return user.HomeDir + confFile
}

func loadConf() Config {
	path := buildConfPath()
	err := os.MkdirAll(filepath.Dir(path), os.ModePerm)
	if err != nil {
		fmt.Println("write conf file error:", err.Error())
		os.Exit(-1)
	}
	_, err = os.Stat(path)
	if os.IsNotExist(err) {
		err := ioutil.WriteFile(path, []byte(confSample), 0644)
		if err != nil {
			fmt.Println("write conf file error:", err.Error())
			os.Exit(-1)
		}
	}
	file, err := ioutil.ReadFile(path)
	if err != nil {
		fmt.Println("read conf file error:", err.Error())
		os.Exit(-1)
	}
	conf := Config{}
	err = yaml.Unmarshal(file, &conf)
	if err != nil {
		fmt.Println("read conf file error:", err.Error())
		os.Exit(-1)
	}
	return conf
}

type Config struct {
	Tasks map[string]Task
	Jobs  map[string]Job
}

type Task struct {
	Type   string     `yaml:"type"`
	Params RawMessage `yaml:"params"`
}

type Job struct {
	Schedule struct {
		Type string `yaml:"type"`
		Cron string `yaml:"cron"`
	} `yaml:"schedule"`
	Task string `yaml:"task"`
}

type RawMessage struct {
	unmarshal func(interface{}) error
}

func (msg *RawMessage) UnmarshalYAML(unmarshal func(interface{}) error) error {
	msg.unmarshal = unmarshal
	return nil
}

func (msg *RawMessage) Unmarshal(v interface{}) error {
	return msg.unmarshal(v)
}

func initTasks(conf Config) map[string]richman.Task {
	tasks := make(map[string]richman.Task)

	for name, task := range conf.Tasks {
		switch task.Type {
		case "stocks":
			cfg := richman.StockTaskConfig{}
			err := task.Params.Unmarshal(&cfg)
			if err != nil {
				log.Println("read stock config error:", err.Error())
				os.Exit(-1)
			}
			stockTask := richman.NewStockTask(cfg)
			tasks[name] = stockTask
		default:
			log.Printf("%s task does not support for now\n", task.Type)
		}
	}

	return tasks
}

func initJobs(conf Config, tasks map[string]richman.Task) map[string]richman.Job {

	jobs := map[string]richman.Job{}

	for name, job := range conf.Jobs {
		switch job.Schedule.Type {
		case "cron":
			scheduler := richman.NewCronScheduler(job.Schedule.Cron)

			task, ok := tasks[job.Task]
			if !ok {
				log.Printf("task %s not found for job %s\n", job.Task, name)
			}
			jobs[name] = richman.NewJob(name, scheduler, task)
		default:
			log.Printf("%s scheduler does not support for now\n", job.Schedule.Type)
		}
	}

	return jobs
}

func WaitStopSignal() {
	waitSignal(os.Interrupt, os.Kill, syscall.SIGTERM, syscall.SIGINT)
}

func waitSignal(signals ...os.Signal) {
	signalChan := make(chan os.Signal)
	defer func() {
		signal.Stop(signalChan)
		close(signalChan)
	}()
	signal.Notify(signalChan, signals...)
	<-signalChan
}

func main() {
	conf := loadConf()

	tasks := initTasks(conf)

	jobs := initJobs(conf, tasks)

	monitor := richman.NewMonitor()

	for name, job := range jobs {
		err := monitor.AddJob(name, job)
		if err != nil {
			log.Println("add job error:", err.Error())
			os.Exit(-1)
		}
	}

	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()
	monitor.Start(ctx)

	WaitStopSignal()

	ctx, cancelFun2 := context.WithTimeout(ctx, 10*time.Second)
	defer cancelFun2()
	monitor.Stop(ctx)
}
