package main

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/zhenzou/richman"
	"github.com/zhenzou/richman/conf"
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

func initTasks(conf conf.Config) map[string]richman.Task {
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

func initJobs(conf conf.Config, tasks map[string]richman.Task) map[string]richman.Job {

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

	config := conf.Load()

	tasks := initTasks(config)

	jobs := initJobs(config, tasks)

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
