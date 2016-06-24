//Copyright 201606 itczl. All rights reserved.

//config is used to parse config file
package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

type alter struct {
	//Mail  []string `json:"nail"`
	Mail  []string
	Slack struct {
		Webhook  string
		Promote  string
		Username string
	}
}

type Config map[string]alter

func Parse(configFile string) (Config, error) {
	c, err := ioutil.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("config.Parse read file error: %v\n", err)
	}
	var conf Config
	if err := json.Unmarshal(c, &conf); err != nil {
		return nil, fmt.Errorf("config.Parse json error: %v\n", err)
	}
	return conf, nil
}
