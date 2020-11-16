package utils

import (
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"gopkg.in/yaml.v2"
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

func EnsureDirExists(path string) error {
	err := os.MkdirAll(filepath.Dir(path), os.ModePerm)
	if err != nil {
		log.Println("write conf file error:", err.Error())
		Die()
	}
	return err
}

func ReadYamlFile(err error, path string, ptr interface{}) error {
	file, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(file, ptr)
	if err != nil {
		log.Println("read ptr file error:", err.Error())
		return err
	}
	return nil
}
