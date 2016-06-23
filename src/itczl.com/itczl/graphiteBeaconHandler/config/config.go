package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

type alter struct {
	//Name  string `json:"name"`
	Name  string
	Mail  []string
	Slack struct {
		Channel  string
		Username string
	}
}

type Config []alter

func Parse(configFile string) (*Config, error) {
	c, err := ioutil.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("config.Parse read file error: %v\n", err)
	}
	var conf Config
	if err := json.Unmarshal(c, &conf); err != nil {
		return nil, fmt.Errorf("config.Parse json error: %v\n", err)
	}
	return &conf, nil
}
