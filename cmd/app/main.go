package main

import (
	apiserver "TaskManager/internal/delivery/http_server"
	"flag"
	"log"

	"github.com/BurntSushi/toml"
)

var (
	configPath string
)

func init() {
	flag.StringVar(&configPath, "config-path", "configs/taskmanager.toml", "path to config file")
}

func main() {
	flag.Parse()

	c := apiserver.NewConfig()
	_, err := toml.DecodeFile(configPath, c)
	if err != nil {
		log.Fatal(err)
	}
	
	s := apiserver.New(c)	
	e := s.Start()

	if e != nil { 
		log.Fatal(e)
	}
}