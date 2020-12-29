package main

import (
	"fmt"
	"godis/src/config"
	"godis/src/lib/logger"
	"godis/src/tcp"
	RedisServer "godis/src/redis/server"
	"os"
)

func main() {
	configFilename := os.Getenv("CONFIG")
	if configFilename == "" {
		configFilename = "redis.conf"
	}
	// config.SetupConfig(configFilename)
	logger.Setup(&logger.Settings{
		Path:       "logs",
		Name:       "godis",
		Ext:        ".log",
		TimeFormat: "2006-01-02",
	})

	tcp.ListenAndServe(&tcp.Config{
		Address: fmt.Sprintf("%s:%d", config.Properties.Bind, config.Properties.Port),
	}, RedisServer.NewHandler())
}
