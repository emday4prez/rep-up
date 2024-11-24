package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

// bodyPartRequest represents the expected request body for creating/updating a body part
type bodyPartRequest struct {
	Name string `json:"name"`
}

// ////////////////////////////////////////////////////////
// GetBodyPart handles GET requests for a single body part
func (h *Handlers) GetBodyPart(w http.ResponseWriter, r *http.Request) {
	// Extract ID from URL using Chi router
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid ID format")
		return
	}

	// Query the database
	var bodyPart struct {
		ID   int64  `json:"id"`
		Name string `json:"name"`
	}

	err = h.db.QueryRow(
		"SELECT id, name FROM body_parts WHERE id = ?",
		id,
	).Scan(&bodyPart.ID, &bodyPart.Name)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			h.respondWithError(w, http.StatusNotFound, "Body part not found")
			return
		}
		h.respondWithError(w, http.StatusInternalServerError, "Database error")
		return
	}

	h.respondWithJSON(w, http.StatusOK, bodyPart)
}

// ListBodyParts handles GET requests for all body parts
func (h *Handlers) ListBodyParts(w http.ResponseWriter, r *http.Request) {
	// Query all body parts
	rows, err := h.db.Query("SELECT id, name FROM body_parts ORDER BY name")
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Database error")
		return
	}
	defer rows.Close() // Important: always close rows!

	var bodyParts []struct {
		ID   int64  `json:"id"`
		Name string `json:"name"`
	}

	// Iterate through the rows
	for rows.Next() {
		var bp struct {
			ID   int64  `json:"id"`
			Name string `json:"name"`
		}
		if err := rows.Scan(&bp.ID, &bp.Name); err != nil {
			h.respondWithError(w, http.StatusInternalServerError, "Row scanning error")
			return
		}
		bodyParts = append(bodyParts, bp)
	}

	// Check for errors from iterating over rows
	if err := rows.Err(); err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Row iteration error")
		return
	}

	h.respondWithJSON(w, http.StatusOK, bodyParts)
}

// CreateBodyPart handles POST requests to create a new body part
func (h *Handlers) CreateBodyPart(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	var req bodyPartRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate input
	if req.Name == "" {
		h.respondWithError(w, http.StatusBadRequest, "Name is required")
		return
	}

	// Insert into database
	result, err := h.db.Exec(
		"INSERT INTO body_parts (name) VALUES (?)",
		req.Name,
	)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Database error")
		return
	}

	// Get the ID of the newly inserted record
	id, err := result.LastInsertId()
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Error getting new ID")
		return
	}

	// Return the created body part
	bodyPart := struct {
		ID   int64  `json:"id"`
		Name string `json:"name"`
	}{
		ID:   id,
		Name: req.Name,
	}

	h.respondWithJSON(w, http.StatusCreated, bodyPart)
}

// ////////////////////////////////////////////////////////////////////
// UpdateBodyPart handles PUT requests to update an existing body part
func (h *Handlers) UpdateBodyPart(w http.ResponseWriter, r *http.Request) {
	// Extract ID from URL
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid ID format")
		return
	}

	// Parse request body
	var req bodyPartRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate input
	if req.Name == "" {
		h.respondWithError(w, http.StatusBadRequest, "Name is required")
		return
	}

	// First check if the record exists
	var exists bool
	err = h.db.QueryRow("SELECT EXISTS(SELECT 1 FROM body_parts WHERE id = ?)", id).Scan(&exists)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Database error")
		return
	}

	if !exists {
		h.respondWithError(w, http.StatusNotFound, "Body part not found")
		return
	}

	// Update the record
	result, err := h.db.Exec(
		"UPDATE body_parts SET name = ? WHERE id = ?",
		req.Name,
		id,
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
		h.respondWithError(w, http.StatusNotFound, "Body part not found")
		return
	}

	// Return the updated body part
	bodyPart := struct {
		ID   int64  `json:"id"`
		Name string `json:"name"`
	}{
		ID:   id,
		Name: req.Name,
	}

	h.respondWithJSON(w, http.StatusOK, bodyPart)
}

// /////////////////////////////////////////////////////////////
// DeleteBodyPart handles DELETE requests to remove a body part
func (h *Handlers) DeleteBodyPart(w http.ResponseWriter, r *http.Request) {
	// Extract ID from URL
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid ID format")
		return
	}

	// Begin a transaction since there are multiple operations
	tx, err := h.db.Begin()
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Transaction error")
		return
	}
	defer tx.Rollback() // Rollback if we don't commit

	// Check for exercises using this body part
	var exerciseCount int
	err = tx.QueryRow("SELECT COUNT(*) FROM exercises WHERE body_part_id = ?", id).Scan(&exerciseCount)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Database error")
		return
	}

	if exerciseCount > 0 {
		h.respondWithError(w, http.StatusConflict,
			"Cannot delete body part: it is referenced by existing exercises")
		return
	}

	// Delete the body part
	result, err := tx.Exec("DELETE FROM body_parts WHERE id = ?", id)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Database error")
		return
	}

	// Check if anything was actually deleted
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Error checking delete status")
		return
	}

	if rowsAffected == 0 {
		h.respondWithError(w, http.StatusNotFound, "Body part not found")
		return
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Transaction commit error")
		return
	}

	// Return a success message with no content
	w.WriteHeader(http.StatusNoContent)
}

// /////////////////////////////////////////////
// Helper function to check if a body part exists
func (h *Handlers) bodyPartExists(id int64) (bool, error) {
	var exists bool
	err := h.db.QueryRow("SELECT EXISTS(SELECT 1 FROM body_parts WHERE id = ?)", id).Scan(&exists)
	return exists, err
}
