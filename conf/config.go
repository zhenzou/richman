package conf

import (
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v2"

	"github.com/zhenzou/richman/utils"
)

const confFile = "/.richman/conf.yaml"

const confSample = `
refresh: 1s

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

type Config struct {
	Refresh time.Duration   `yaml:"refresh"`
	Tasks   map[string]Task `yaml:"tasks"`
	Jobs    map[string]Job  `yaml:"jobs"`
}

type Task struct {
	Type   string `yaml:"type"`
	Params Params `yaml:"params"`
}

type Job struct {
	Schedule struct {
		Type   string `yaml:"type"`
		Params Params `yaml:"params"`
	} `yaml:"schedule"`
	Task string `yaml:"task"`
}

func Load() Config {
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

func buildConfPath() string {
	u, err := user.Current()
	if err != nil {
		log.Println("get current user error:", err.Error())
		utils.Die()
	}
	return u.HomeDir + confFile
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
