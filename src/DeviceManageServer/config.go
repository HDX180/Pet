package DeviceManageServer

import (
	"fmt"
	yaml "gopkg.in/yaml.v2"
	"io/ioutil"
)

var config *StruConfig = &StruConfig{}

func GetConfigInstance() *StruConfig {
	return config
}

type StruConfig struct {
	Mysql Mysql `yaml:"mysql"`
}

type Mysql struct {
	User     string `yaml:"user"`
	Host     string `yaml:"host"`
	Password string `yaml:"password"`
	Port     string `yaml:"port"`
	Name     string `yaml:"name"`
}

func (c *StruConfig) Init(yamlPath string) bool {

	yamlFile, err := ioutil.ReadFile(yamlPath)
	if err != nil {
		logger.Error(fmt.Sprintf("yamlFile.Get: %s", err.Error()))
		return false
	}

	if err = yaml.Unmarshal(yamlFile, c); err != nil {
		logger.Error(fmt.Sprintf("yaml.Unmarshal fail: %s", err.Error()))
		return false
	}
	return true
}

func (c *StruConfig) GetMySqlUri() string {
	//"root:123456@tcp(127.0.0.1:3306)/test"
	mysql := c.Mysql
	return mysql.User + ":" + mysql.Password + "@tcp(" + mysql.Host + ":" + mysql.Port + ")/" + mysql.Name
}
