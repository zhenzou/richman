package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/getlantern/systray"

	"github.com/zhenzou/richman"
	"github.com/zhenzou/richman/conf"
	"github.com/zhenzou/richman/pkg/stock"
	"github.com/zhenzou/richman/tasks"
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

var (
	monitor    richman.Monitor
	titleQueue chan string
	config     conf.Config
)

func displayStock(ctx context.Context, stock stock.Stock) error {
	title := fmt.Sprintf("%s %s", stock.Name, stock.IncreaseRate())
	enqueue(title)
	return nil
}

func enqueue(title string) {
	log.Println("enqueue:", title)
	select {
	case titleQueue <- title:
	default:
		log.Println("[queue] drop ", title)
	}
}

func loopUpdateTitle() {
	tick := time.Tick(config.Refresh)
	for _ = range tick {
		select {
		case title := <-titleQueue:
			systray.SetTitle(title)
		default:
		}
	}
}

func initTray() {
	systray.SetTitle("Richman")

	quitItem := systray.AddMenuItem("Quit", "quit")

	go func() {
		select {
		case _ = <-quitItem.ClickedCh:
			systray.Quit()
		}
	}()
}

func start() {
	initTray()

	config = conf.Load()

	monitor = richman.NewMonitor(config.Monitor)

	initAndRegisterTasks(config)

	titleQueue = make(chan string, config.Queue)
	go loopUpdateTitle()

	monitor.Start()
}

func initAndRegisterTasks(conf conf.Config) {
	for name, config := range conf.Tasks {
		switch config.Type {
		case "stocks":
			cfg := tasks.StockConfig{}
			err := config.Params.Unmarshal(&cfg)
			if err != nil {
				log.Println("read stock config error:", err.Error())
				utils.Die()
			}
			task := tasks.NewStockTask(cfg, displayStock)
			_ = monitor.RegisterTask(name, task)
		default:
			log.Printf("%s config does not support for now\n", config.Type)
		}
	}
}

func exit() {
	ctx, cancelFun2 := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelFun2()
	monitor.Stop(ctx)
}

func main() {

	log.Printf("release: %s, repo: %s, commit: %s\n", release, repo, commit)

	systray.Run(start, exit)
}
