package config

import (
	"io/ioutil"
	"time"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

var (
	log = logrus.WithField("component", "config")
)

type Conf struct {
	Tick                 int           `yaml:"tick"`
	LambdaWindowElements int           `yaml:"lambdaWindowElements"` // number of requests which must be withing the lambdaThreshold time period
	LambdaThreshold      time.Duration `yaml:"lambdaThreshold"`      // lambdaWindowElements within this time period, start lambda-runtime
	RedisThreshold       int           `yaml:"redisThreshold"`       // if lambdaCount reaches this threshold, start redis-server
	RedisUtilization     time.Duration `yaml:"redisUtilization"`     // last 5 requests not in this duration, shutdown redis server
}

func (c *Conf) PrintConfig() {
	log.Infof("tick: %d", c.Tick)
	log.Infof("lambdaWindowElements: %d", c.LambdaWindowElements)
	log.Infof("lambdaThreshold: %s", c.LambdaThreshold)
	log.Infof("redisThreshold: %d", c.RedisThreshold)
	log.Infof("redisUtilization: %s", c.RedisUtilization)

}

func (c *Conf) ReadConf() {
	yamlFile, err := ioutil.ReadFile("conf.yaml")
	if err != nil {
		log.Errorf("yamlFile.Get err   #%v ", err)
	}
	err = yaml.Unmarshal(yamlFile, c)
	if err != nil {
		log.Fatalf("Unmarshal: %v", err)
	}

}

func WriteConf(c *Conf) {

	data, err := yaml.Marshal(c)

	if err != nil {
		log.Fatal(err)
	}

	err2 := ioutil.WriteFile("conf.yaml", data, 0777)

	if err2 != nil {

		log.Fatal(err2)
	}

	log.Info("data written")
}
