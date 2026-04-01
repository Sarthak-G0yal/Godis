package config

import "time"

type Config struct {
	Address         string
	AOFPath         string
	AOFSyncInterval time.Duration
}

func Default() Config {
	return Config{
		Address:         ":6379",
		AOFPath:         "appendonly.aof",
		AOFSyncInterval: time.Second,
	}
}
