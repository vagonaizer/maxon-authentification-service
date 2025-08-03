package main

import (
	"log"
	"os"

	"github.com/vagonaizer/authenitfication-service/internal/app"
)

var (
	version   = "dev"
	buildTime = "unknown"
)

func main() {
	// Print version info
	log.Printf("Auth Service %s (built at %s)", version, buildTime)

	// Initialize application
	application, err := app.NewApp()
	if err != nil {
		log.Fatalf("Failed to initialize application: %v", err)
	}

	if err := application.Run(); err != nil {
		log.Fatalf("Application error: %v", err)
		os.Exit(1)
	}
}
