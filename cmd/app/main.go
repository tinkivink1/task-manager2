package main

import (
	"TaskManager/internal/cache/redis"
	apiserver "TaskManager/internal/delivery/http_server"
	"TaskManager/internal/storage/postgres"
	"flag"
	"log"

	"github.com/BurntSushi/toml"
)

var (
	configPath string
	caching    int
)

func init() {
	flag.StringVar(&configPath, "config-path", "configs/taskmanager.toml", "path to config file")
	flag.IntVar(&caching, "caching", -1, "should the server cache requests, 0 - false, 1 - true")
}

func main() {
	flag.Parse()

	// Server config
	c := apiserver.NewConfig()
	_, err := toml.DecodeFile(configPath, c)
	if err != nil {
		log.Fatal(err)
	}

	// New server with config
	s, err := apiserver.New(c)

	if err != nil {
		log.Fatal(err)
	}

	// Postgres config
	pgConfig := postgres.NewConfig()
	_, err = toml.DecodeFile(configPath, pgConfig)
	if err != nil {
		log.Fatal(err)
	}

	// New postgres client
	db := postgres.New(pgConfig)
	s.UseDB(db)

	if caching == -1 && c.Caching ||
		caching == 1 {
		redisConfig := redis.NewConfig()
		_, err = toml.DecodeFile(configPath, redisConfig)
		if err != nil {
			log.Fatal(err)
		}

		cache := redis.New(redisConfig)
		s.UseCache(cache)
	}

	e := s.Start()

	if e != nil {
		log.Fatal(e)
	}
}
