package utils

import (
	"os"
	"os/signal"
	"syscall"
)

func Die() {
	os.Exit(-1)
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
