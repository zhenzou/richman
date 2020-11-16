package conf

import (
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"time"

	"github.com/zhenzou/richman"
	"github.com/zhenzou/richman/utils"
)

const confFile = "/.richman/conf.yaml"

const confSample = `
refresh: 1s
monitor:
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
	Refresh time.Duration  `yaml:"refresh"`
	Monitor richman.Config `yaml:"monitor"`
}

func buildConfPath() string {
	u, err := user.Current()
	if err != nil {
		log.Println("get current user error:", err.Error())
		utils.Die()
	}
	return u.HomeDir + confFile
}

func Load() Config {
	path := buildConfPath()
	err := utils.EnsureDirExists(path)
	_, err = os.Stat(path)
	if os.IsNotExist(err) {
		err := ioutil.WriteFile(path, []byte(confSample), 0644)
		if err != nil {
			log.Println("write conf file error:", err.Error())
			utils.Die()
		}
	}
	conf := Config{}
	if err := utils.ReadYamlFile(err, path, &conf); err != nil {
		log.Println("read conf file error:", err.Error())
		utils.Die()
	}
	return conf
}
