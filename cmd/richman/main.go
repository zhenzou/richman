package main

import (
	"context"
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v2"

	"github.com/zhenzou/richman"
	"github.com/zhenzou/richman/utils"
)

var (
	release = "unknown"
	repo    = "unknown"
	commit  = "unknown"
	debug   = "true"
)

func init() {

	if strings.EqualFold(debug, "true") {
		log.SetFlags(log.LstdFlags | log.Lshortfile)
	} else {
		log.SetFlags(0)
	}
}

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
      params:
        cron: '*/6 * * * * *'
    task: "get-stocks"
`

func buildConfPath() string {
	u, err := user.Current()
	if err != nil {
		log.Println("get current user error:", err.Error())
		utils.Die()
	}
	return u.HomeDir + confFile
}

func loadConf() Config {
	path := buildConfPath()
	err := os.MkdirAll(filepath.Dir(path), os.ModePerm)
	if err != nil {
		log.Println("write conf file error:", err.Error())
		utils.Die()
	}
	_, err = os.Stat(path)
	if os.IsNotExist(err) {
		err := ioutil.WriteFile(path, []byte(confSample), 0644)
		if err != nil {
			log.Println("write conf file error:", err.Error())
			utils.Die()
		}
	}
	file, err := ioutil.ReadFile(path)
	if err != nil {
		log.Println("read conf file error:", err.Error())
		utils.Die()
	}
	conf := Config{}
	err = yaml.Unmarshal(file, &conf)
	if err != nil {
		log.Println("read conf file error:", err.Error())
		utils.Die()
	}
	return conf
}

type Config struct {
	Tasks map[string]Task
	Jobs  map[string]Job
}

type Task struct {
	Type   string         `yaml:"type"`
	Params richman.Params `yaml:"params"`
}

type Job struct {
	Schedule struct {
		Type   string         `yaml:"type"`
		Params richman.Params `yaml:"params"`
	} `yaml:"schedule"`
	Task string `yaml:"task"`
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
				utils.Die()
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
			cfg := richman.CronSchedulerConfig{}
			err := job.Schedule.Params.Unmarshal(&cfg)
			if err != nil {
				log.Printf("[job] %s read config error \n", name)
			}
			scheduler := richman.NewCronScheduler(cfg)
			task, ok := tasks[job.Task]
			if !ok {
				log.Printf("[job] task %s not found for %s\n", job.Task, name)
			}
			jobs[name] = richman.NewJob(name, scheduler, task)
		default:
			log.Printf("%s scheduler does not support for now\n", job.Schedule.Type)
		}
	}

	return jobs
}

func main() {

	log.Printf("release: %s, repo: %s, commit: %s\n", release, repo, commit)

	conf := loadConf()

	tasks := initTasks(conf)

	jobs := initJobs(conf, tasks)

	monitor := richman.NewMonitor()

	for name, job := range jobs {
		err := monitor.AddJob(name, job)
		if err != nil {
			log.Println("add job error:", err.Error())
			utils.Die()
		}
	}

	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()
	monitor.Start(ctx)

	utils.WaitStopSignal()

	ctx, cancelFun2 := context.WithTimeout(ctx, 10*time.Second)
	defer cancelFun2()
	monitor.Stop(ctx)
}
