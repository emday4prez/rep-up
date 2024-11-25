package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"

	"repup/internal/data"
	"repup/internal/handlers"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	// Initialize database connection
	dbConfig := data.Config{
		URL:   os.Getenv("TURSO_DATABASE_URL"),
		Token: os.Getenv("TURSO_AUTH_TOKEN"),
	}

	if err := data.Initialize(dbConfig); err != nil {
		log.Fatal(fmt.Sprintf("Database initialization error: %v", err))
	}
	defer data.Close()

	// Create a new router instance
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)

	// Initialize handlers with database connection
	db := data.GetDB()
	handlers := handlers.NewHandlers(db)

	// Routes
	r.Route("/api", func(r chi.Router) {
		// Body Parts
		r.Route("/body-parts", func(r chi.Router) {
			r.Get("/", handlers.ListBodyParts)
			r.Post("/", handlers.CreateBodyPart)
			r.Get("/{id}", handlers.GetBodyPart)
			r.Put("/{id}", handlers.UpdateBodyPart)
			r.Delete("/{id}", handlers.DeleteBodyPart)
		})

		r.Route("/exercises", func(r chi.Router) {
			r.Get("/", handlers.ListExercises)
			r.Post("/", handlers.CreateExercise)
			r.Get("/{id}", handlers.GetExercise)
			r.Put("/{id}", handlers.UpdateExercise)
			r.Delete("/{id}", handlers.DeleteExercise)
		})

		r.Route("/workouts", func(r chi.Router) {
			r.Get("/", handlers.ListWorkouts)
			r.Post("/", handlers.CreateWorkout)
			r.Get("/{id}", handlers.GetWorkout)
			r.Put("/{id}", handlers.UpdateWorkout)
			r.Delete("/{id}", handlers.DeleteWorkout)
		})
	})

	// Start the server
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Starting server on :%s", port)
	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatal(err)
	}
}
