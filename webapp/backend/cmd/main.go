package main

import (
	"backend/internal/server"
	"log"

	"github.com/kaz/pprotein/integration/standalone"
)

func main() {
	go standalone.Integrate(":19001")

	srv, dbConn, err := server.NewServer()
	if err != nil {
		log.Fatalf("Failed to initialize server: %v", err)
	}
	if dbConn != nil {
		defer dbConn.Close()
	}

	srv.Run()
}
