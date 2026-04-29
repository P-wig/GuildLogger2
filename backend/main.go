package main

import (
	"log"
	"os"

	"github.com/P-wig/GuildLogger2/backend/app"
)

func main() {
	e, cleanup, err := app.CreateApp()
	if err != nil {
		log.Fatalf("failed to create app: %v", err)
	}
	defer func() {
		if err := cleanup(); err != nil {
			log.Printf("cleanup error: %v", err)
		}
	}()

	port := os.Getenv("PORT")
	if port == "" {
		port = "5001"
	}

	if err := e.Start(":" + port); err != nil {
		log.Fatalf("server stopped: %v", err)
	}
}
