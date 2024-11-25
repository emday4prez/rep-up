// internal/handlers/debug.go

package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"
)

// TestCreateWorkout creates a test workout with sample exercises
func (h *DebugHandlers) TestCreateWorkout(w http.ResponseWriter, r *http.Request) {
	// Sample workout data
	workout := struct {
		UserID  int64  `json:"user_id"`
		Name    string `json:"name"`
		Date    string `json:"date"`
		Notes   string `json:"notes"`
		Details []struct {
			ExerciseID int64   `json:"exercise_id"`
			Sets       int     `json:"sets"`
			Reps       int     `json:"reps"`
			Weight     float64 `json:"weight"`
			Notes      string  `json:"notes"`
		} `json:"details"`
	}{
		UserID: 1,
		Name:   "Test Full Body Workout",
		Date:   time.Now().Format("2006-01-02"),
		Notes:  "Test workout created via debug endpoint",
		Details: []struct {
			ExerciseID int64   `json:"exercise_id"`
			Sets       int     `json:"sets"`
			Reps       int     `json:"reps"`
			Weight     float64 `json:"weight"`
			Notes      string  `json:"notes"`
		}{
			{
				ExerciseID: 1, // Bench Press
				Sets:       3,
				Reps:       10,
				Weight:     135.5,
				Notes:      "Warmup set included",
			},
			{
				ExerciseID: 5, // Squats
				Sets:       4,
				Reps:       8,
				Weight:     185.0,
				Notes:      "Focus on form",
			},
		},
	}

	// Convert to JSON
	jsonData, err := json.Marshal(workout)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Error creating test data")
		return
	}

	// Create a new request with our test data
	req, err := http.NewRequest("POST", "/api/workouts", bytes.NewBuffer(jsonData))
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Error creating test request")
		return
	}

	// Copy original headers and set content type
	req.Header = r.Header
	req.Header.Set("Content-Type", "application/json")

	// Call the actual workout creation handler
	h.CreateWorkout(w, req)
}

// TestHealthCheck provides basic health check info
func (h *DebugHandlers) TestHealthCheck(w http.ResponseWriter, r *http.Request) {
	// Get database status
	err := h.db.Ping()
	dbStatus := "healthy"
	if err != nil {
		dbStatus = "unhealthy"
	}

	health := struct {
		Status    string    `json:"status"`
		DBStatus  string    `json:"db_status"`
		Timestamp time.Time `json:"timestamp"`
	}{
		Status:    "ok",
		DBStatus:  dbStatus,
		Timestamp: time.Now(),
	}

	h.respondWithJSON(w, http.StatusOK, health)
}

// TestListTables lists all tables in the database
func (h *DebugHandlers) TestListTables(w http.ResponseWriter, r *http.Request) {
	rows, err := h.db.Query(`
        SELECT name FROM sqlite_master 
        WHERE type='table' 
        ORDER BY name
    `)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Error querying tables")
		return
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			h.respondWithError(w, http.StatusInternalServerError, "Error scanning table names")
			return
		}
		tables = append(tables, name)
	}

	h.respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"tables": tables,
	})
}
