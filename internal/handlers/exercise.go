package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

// exerciseRequest represents the expected request body for creating/updating an exercise
type exerciseRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	BodyPartID  int64  `json:"body_part_id"`
}

// GetExercise handles GET requests for a single exercise
func (h *Handlers) GetExercise(w http.ResponseWriter, r *http.Request) {
	// Extract ID from URL using Chi router
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid ID format")
		return
	}

	// Query the database
	var exercise struct {
		ID          int64  `json:"id"`
		Name        string `json:"name"`
		Description string `json:"description"`
		BodyPartID  int64  `json:"body_part_id"`
		BodyPart    struct {
			ID   int64  `json:"id"`
			Name string `json:"name"`
		} `json:"body_part"`
	}

	err = h.db.QueryRow(`
        SELECT e.id, e.name, e.description, e.body_part_id, 
               b.id, b.name
        FROM exercises e
        JOIN body_parts b ON e.body_part_id = b.id
        WHERE e.id = ?`,
		id,
	).Scan(
		&exercise.ID, &exercise.Name, &exercise.Description, &exercise.BodyPartID,
		&exercise.BodyPart.ID, &exercise.BodyPart.Name,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			h.respondWithError(w, http.StatusNotFound, "Exercise not found")
			return
		}
		h.respondWithError(w, http.StatusInternalServerError, "Database error")
		return
	}

	h.respondWithJSON(w, http.StatusOK, exercise)
}

// ListExercises handles GET requests for all exercises
func (h *Handlers) ListExercises(w http.ResponseWriter, r *http.Request) {
	rows, err := h.db.Query(`
        SELECT e.id, e.name, e.description, e.body_part_id,
               b.id, b.name
        FROM exercises e
        JOIN body_parts b ON e.body_part_id = b.id
        ORDER BY e.name`)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Database error")
		return
	}
	defer rows.Close()

	var exercises []struct {
		ID          int64  `json:"id"`
		Name        string `json:"name"`
		Description string `json:"description"`
		BodyPartID  int64  `json:"body_part_id"`
		BodyPart    struct {
			ID   int64  `json:"id"`
			Name string `json:"name"`
		} `json:"body_part"`
	}

	for rows.Next() {
		var exercise struct {
			ID          int64  `json:"id"`
			Name        string `json:"name"`
			Description string `json:"description"`
			BodyPartID  int64  `json:"body_part_id"`
			BodyPart    struct {
				ID   int64  `json:"id"`
				Name string `json:"name"`
			} `json:"body_part"`
		}

		if err := rows.Scan(
			&exercise.ID, &exercise.Name, &exercise.Description, &exercise.BodyPartID,
			&exercise.BodyPart.ID, &exercise.BodyPart.Name,
		); err != nil {
			h.respondWithError(w, http.StatusInternalServerError, "Row scanning error")
			return
		}
		exercises = append(exercises, exercise)
	}

	if err := rows.Err(); err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Row iteration error")
		return
	}

	h.respondWithJSON(w, http.StatusOK, exercises)
}

// CreateExercise handles POST requests to create a new exercise
func (h *Handlers) CreateExercise(w http.ResponseWriter, r *http.Request) {
	var req exerciseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate input
	if req.Name == "" {
		h.respondWithError(w, http.StatusBadRequest, "Name is required")
		return
	}

	// Check if body part exists
	exists, err := h.bodyPartExists(req.BodyPartID)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Database error")
		return
	}
	if !exists {
		h.respondWithError(w, http.StatusBadRequest, "Body part not found")
		return
	}

	// Insert into database
	result, err := h.db.Exec(
		"INSERT INTO exercises (name, description, body_part_id) VALUES (?, ?, ?)",
		req.Name, req.Description, req.BodyPartID,
	)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Database error")
		return
	}

	id, err := result.LastInsertId()
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Error getting new ID")
		return
	}

	// Return the created exercise
	exercise := struct {
		ID          int64  `json:"id"`
		Name        string `json:"name"`
		Description string `json:"description"`
		BodyPartID  int64  `json:"body_part_id"`
	}{
		ID:          id,
		Name:        req.Name,
		Description: req.Description,
		BodyPartID:  req.BodyPartID,
	}

	h.respondWithJSON(w, http.StatusCreated, exercise)
}

// UpdateExercise handles PUT requests to update an existing exercise
func (h *Handlers) UpdateExercise(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid ID format")
		return
	}

	var req exerciseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate input
	if req.Name == "" {
		h.respondWithError(w, http.StatusBadRequest, "Name is required")
		return
	}

	// Begin transaction
	tx, err := h.db.Begin()
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Transaction error")
		return
	}
	defer tx.Rollback()

	// Check if exercise exists
	var exists bool
	err = tx.QueryRow("SELECT EXISTS(SELECT 1 FROM exercises WHERE id = ?)", id).Scan(&exists)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Database error")
		return
	}
	if !exists {
		h.respondWithError(w, http.StatusNotFound, "Exercise not found")
		return
	}

	// Check if body part exists
	exists, err = h.bodyPartExists(req.BodyPartID)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Database error")
		return
	}
	if !exists {
		h.respondWithError(w, http.StatusBadRequest, "Body part not found")
		return
	}

	// Update the record
	result, err := tx.Exec(
		"UPDATE exercises SET name = ?, description = ?, body_part_id = ? WHERE id = ?",
		req.Name, req.Description, req.BodyPartID, id,
	)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Database error")
		return
	}

	// Check if the update was successful
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Error checking update status")
		return
	}

	if rowsAffected == 0 {
		h.respondWithError(w, http.StatusNotFound, "Exercise not found")
		return
	}
	if err := tx.Commit(); err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Transaction commit error")
		return
	}

	// Return the updated exercise
	exercise := struct {
		ID          int64  `json:"id"`
		Name        string `json:"name"`
		Description string `json:"description"`
		BodyPartID  int64  `json:"body_part_id"`
	}{
		ID:          id,
		Name:        req.Name,
		Description: req.Description,
		BodyPartID:  req.BodyPartID,
	}

	h.respondWithJSON(w, http.StatusOK, exercise)
}

// DeleteExercise handles DELETE requests to remove an exercise
func (h *Handlers) DeleteExercise(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid ID format")
		return
	}

	// Begin transaction
	tx, err := h.db.Begin()
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Transaction error")
		return
	}
	defer tx.Rollback()

	// Check for workouts using this exercise
	var workoutCount int
	err = tx.QueryRow("SELECT COUNT(*) FROM workout_exercises WHERE exercise_id = ?", id).Scan(&workoutCount)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Database error")
		return
	}

	if workoutCount > 0 {
		h.respondWithError(w, http.StatusConflict,
			"Cannot delete exercise: it is referenced by existing workouts")
		return
	}

	// Delete the exercise
	result, err := tx.Exec("DELETE FROM exercises WHERE id = ?", id)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Database error")
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Error checking delete status")
		return
	}

	if rowsAffected == 0 {
		h.respondWithError(w, http.StatusNotFound, "Exercise not found")
		return
	}

	if err := tx.Commit(); err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Transaction commit error")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
