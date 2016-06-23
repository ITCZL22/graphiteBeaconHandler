package main

import (
	"flag"
	"log"
	"os"

	"uve.io/uve/graphiteBeaconHandler/config"
	"uve.io/uve/graphiteBeaconHandler/httpservice"
)

const (
	ErrnoOk int = iota
	ErrorConfigFileEmpty
	ErrorConfigParseFailed
	ErrorHttpserviceRunFailed
)

var configFile = flag.String("conf", "conf.json", "config file")

//放到request中
//var name = flag.String("name", "", "name")
//var level = flag.String("level", "", "level")
//var value = flag.String("value", "", "value")

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
