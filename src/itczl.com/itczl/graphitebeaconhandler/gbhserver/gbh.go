//Copyright 201606 itczl. All rights reserved.

//gbhserver is used to specify slack beacon
package main

import (
	"flag"
	"log"
	"os"

	"itczl.com/itczl/graphitebeaconhandler/config"
	"itczl.com/itczl/graphitebeaconhandler/httpservice"
)

const (
	Ok int = iota
	ConfigFileEmpty
	ConfigParseFailed
	HttpserviceRunFailed
)

var configFile = flag.String("conf", "conf.json", "config file")

func main() {
	flag.Parse()

	if *configFile == "" {
		log.Println("configuration file should be specified")
		os.Exit(ConfigFileEmpty)
	}

	//parse config
	conf, err := config.Parse(*configFile)
	if err != nil {
		log.Printf("configuration file parse error: %v\n", err)
		os.Exit(ConfigParseFailed)
	}

	//invoke httpserver
	if err := httpservice.Run(conf); err != nil {
		log.Printf("httpservice.Run error: %v\n", err)
		os.Exit(HttpserviceRunFailed)
	}
}
