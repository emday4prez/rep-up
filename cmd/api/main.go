package main

import (
	"net/http"
	"os"

	"repup/internal/auth" // Add this import
	"repup/internal/data"
	"repup/internal/handlers"
	"repup/internal/logger"
	customMiddleware "repup/internal/middleware"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog/log"
)

func main() {
	// Load environment variables
	logger.Init()
	logger := log.With().Str("component", "main").Logger()

	if err := godotenv.Load(); err != nil {
		log.Error().Err(err).Msg("Error loading .env file")
	}

	// Initialize database connection
	dbConfig := data.Config{
		URL:   os.Getenv("TURSO_DATABASE_URL"),
		Token: os.Getenv("TURSO_AUTH_TOKEN"),
	}

	if err := data.Initialize(dbConfig); err != nil {
		logger.Fatal().Err(err).Msg("Database initialization error")
	}
	defer data.Close()

	// Create a new router instance
	r := chi.NewRouter()

	// Middleware
	r.Use(customMiddleware.RequestLogger)
	r.Use(customMiddleware.LoggingMiddleware)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)

	// Initialize handlers with database connection
	db := data.GetDB()
	mainHandlers := handlers.NewHandlers(db)

	// Routes
	r.Route("/api", func(r chi.Router) {
		// Public routes
		//r.Post("/auth/google", handlers.GoogleAuth)
		//r.Get("/auth/google/callback", handlers.GoogleCallback)

		// Protected routes
		r.Group(func(r chi.Router) {
			r.Use(auth.RequireAuth)

			// Body Parts
			r.Route("/body-parts", func(r chi.Router) {
				r.Get("/", mainHandlers.ListBodyParts)
				r.Post("/", mainHandlers.CreateBodyPart)
				r.Get("/{id}", mainHandlers.GetBodyPart)
				r.Put("/{id}", mainHandlers.UpdateBodyPart)
				r.Delete("/{id}", mainHandlers.DeleteBodyPart)
			})

			r.Route("/exercises", func(r chi.Router) {
				r.Get("/", mainHandlers.ListExercises)
				r.Post("/", mainHandlers.CreateExercise)
				r.Get("/{id}", mainHandlers.GetExercise)
				r.Put("/{id}", mainHandlers.UpdateExercise)
				r.Delete("/{id}", mainHandlers.DeleteExercise)
			})

			r.Route("/workouts", func(r chi.Router) {
				r.Get("/", mainHandlers.ListWorkouts)
				r.Post("/", mainHandlers.CreateWorkout)
				r.Get("/{id}", mainHandlers.GetWorkout)
				r.Put("/{id}", mainHandlers.UpdateWorkout)
				r.Delete("/{id}", mainHandlers.DeleteWorkout)
			})
		})
	})

	// Debug routes - only in development
	if os.Getenv("ENV") != "production" {
		debugHandlers := handlers.NewDebugHandlers(mainHandlers)

		r.Route("/debug", func(r chi.Router) {
			r.Get("/health", debugHandlers.TestHealthCheck)
			r.Get("/tables", debugHandlers.TestListTables)
			r.Post("/test-workout", debugHandlers.TestCreateWorkout)
		})
	}

	// Start the server
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}

	logger.Info().Str("port", port).Msg("Starting server")
	if err := http.ListenAndServe(":"+port, r); err != nil {
		logger.Fatal().Err(err).Msg("Server failed to start")
	}
}
