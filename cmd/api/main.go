package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	// Create a new Chi router
	r := chi.NewRouter()

	// Basic middleware stack
	r.Use(middleware.Logger)    // Log all requests
	r.Use(middleware.Recoverer) // Recover from panics without crashing server

	// Basic test route
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Welcome to RepUp API"))
	})

	// Get port from environment variable
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = ":8080" // Default port if not specified
	}

	// Start the server
	fmt.Printf("Server starting on port %s\n", port)
	if err := http.ListenAndServe(port, r); err != nil {
		log.Fatal(err)
	}
}
