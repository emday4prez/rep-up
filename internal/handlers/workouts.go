package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"repup/internal/data"

	"github.com/go-chi/chi/v5"
)

// workoutRequest represents the expected request body for creating/updating a workout
type workoutRequest struct {
	UserID  int64                    `json:"user_id"`
	Name    string                   `json:"name"`
	Date    string                   `json:"date"` // Format: "2006-01-02"
	Notes   string                   `json:"notes"`
	Details []workoutExerciseRequest `json:"details"`
}

type workoutExerciseRequest struct {
	ExerciseID int64   `json:"exercise_id"`
	Sets       int     `json:"sets"`
	Reps       int     `json:"reps"`
	Weight     float64 `json:"weight"`
	Notes      string  `json:"notes"`
}

func (h *Handlers) GetWorkout(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid ID format")
		return
	}

	workout, err := h.models.Workouts.GetByID(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			h.respondWithError(w, http.StatusNotFound, "Workout not found")
		case errors.Is(err, data.ErrInvalidInput):
			h.respondWithError(w, http.StatusBadRequest, "Invalid input")
		default:
			h.respondWithError(w, http.StatusInternalServerError, "Database error")
		}
		return
	}

	h.respondWithJSON(w, http.StatusOK, workout)
}

func (h *Handlers) ListWorkouts(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.URL.Query().Get("user_id")
	if userIDStr == "" {
		h.respondWithError(w, http.StatusBadRequest, "user_id query parameter is required")
		return
	}

	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid user_id format")
		return
	}

	workouts, err := h.models.Workouts.GetAll(userID)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Database error")
		return
	}

	h.respondWithJSON(w, http.StatusOK, workouts)
}

func (h *Handlers) CreateWorkout(w http.ResponseWriter, r *http.Request) {
	var req workoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Parse date string
	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid date format")
		return
	}

	// Create workout object
	workout := &data.Workout{
		UserID: req.UserID,
		Name:   req.Name,
		Date:   date,
		Notes:  req.Notes,
	}

	// Add exercises
	for _, ex := range req.Details {
		var weight *float64
		if ex.Weight != 0 { // Assuming 0 means no weight provided
			weightVal := ex.Weight
			weight = &weightVal
		}

		workout.Details = append(workout.Details, data.WorkoutExercise{
			ExerciseID: ex.ExerciseID,
			Sets:       ex.Sets,
			Reps:       ex.Reps,
			Weight:     weight,
			Notes:      ex.Notes,
		})
	}
	err = h.models.Workouts.Create(workout)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrInvalidInput):
			h.respondWithError(w, http.StatusBadRequest, "Invalid input")
		default:
			h.respondWithError(w, http.StatusInternalServerError, "Database error")
		}
		return
	}

	h.respondWithJSON(w, http.StatusCreated, workout)
}

func (h *Handlers) UpdateWorkout(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid ID format")
		return
	}

	var req workoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Parse date string
	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid date format")
		return
	}

	workout := &data.Workout{
		ID:     id,
		UserID: req.UserID,
		Name:   req.Name,
		Date:   date,
		Notes:  req.Notes,
	}

	for _, ex := range req.Details {
		var weight *float64
		if ex.Weight != 0 { // Assuming 0 means no weight provided
			weightVal := ex.Weight
			weight = &weightVal
		}

		workout.Details = append(workout.Details, data.WorkoutExercise{
			ExerciseID: ex.ExerciseID,
			Sets:       ex.Sets,
			Reps:       ex.Reps,
			Weight:     weight,
			Notes:      ex.Notes, // If you have this in your request
		})
	}
	err = h.models.Workouts.Update(workout)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			h.respondWithError(w, http.StatusNotFound, "Workout not found")
		case errors.Is(err, data.ErrInvalidInput):
			h.respondWithError(w, http.StatusBadRequest, "Invalid input")
		default:
			h.respondWithError(w, http.StatusInternalServerError, "Database error")
		}
		return
	}

	h.respondWithJSON(w, http.StatusOK, workout)
}

func (h *Handlers) DeleteWorkout(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid ID format")
		return
	}

	err = h.models.Workouts.Delete(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			h.respondWithError(w, http.StatusNotFound, "Workout not found")
		case errors.Is(err, data.ErrInvalidInput):
			h.respondWithError(w, http.StatusBadRequest, "Invalid input")
		default:
			h.respondWithError(w, http.StatusInternalServerError, "Database error")
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
