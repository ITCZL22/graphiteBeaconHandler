//Copyright 201606 itczl. All rights reserved.

//gbhserver is used to specify slack beacon
package main

import (
	"flag"
	"log"
	"os"

	"itczl.com/itczl/graphiteBeaconHandler/config"
	"itczl.com/itczl/graphiteBeaconHandler/httpservice"
)

const (
	ErrnoOk int = iota
	ErrorConfigFileEmpty
	ErrorConfigParseFailed
	ErrorHttpserviceRunFailed
)

var configFile = flag.String("conf", "conf.json", "config file")

func main() {
	flag.Parse()

	if *configFile == "" {
		log.Println("configuration file should be specified")
		os.Exit(ErrorConfigFileEmpty)
	}

	//parse config
	conf, err := config.Parse(*configFile)
	if err != nil {
		log.Printf("configuration file parse error: %v\n", err)
		os.Exit(ErrorConfigParseFailed)
	}

	//invoke httpserver
	if err := httpservice.Run(conf); err != nil {
		log.Printf("httpservice.Run error: %v\n", err)
		os.Exit(ErrorHttpserviceRunFailed)
	}
}
