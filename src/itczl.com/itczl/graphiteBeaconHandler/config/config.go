//Copyright 201606 itczl. All rights reserved.

//config is used to parse config file
package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

type alter struct {
	MailFrom string   `json:"mail_from"`
	MailTo   []string `json:"mail_to"`
	Slack    struct {
		Webhook  string
		Username string
	}
	Host string
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
