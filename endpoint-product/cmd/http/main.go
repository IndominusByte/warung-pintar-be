package main

import (
	"log"

	"github.com/IndominusByte/warung-pintar-be/endpoint-product/internal/config"
)

func main() {
	// init config
	cfg, err := config.New()
	if err != nil {
		log.Fatalf("failed to init the config: %v", err)
	}

	err = startApp(cfg)
	if err != nil {
		log.Fatalf("failed to start app: %v", err)
	}
}
