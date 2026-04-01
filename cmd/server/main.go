package main

import (
	"log"
	"mini-redis/internal/config"
)

func main() {
	cfg := config.Default()
	log.Printf("mini-redis boot config: addr=%s aof=%s", cfg.Address, cfg.AOFPath)
}
