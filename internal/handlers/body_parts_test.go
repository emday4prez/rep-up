package handlers

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/go-chi/chi/v5"
	_ "github.com/mattn/go-sqlite3" // Import SQLite driver for testing
)

// setupTestHandler creates a new Handlers instance with a test database
func setupTestHandler(t *testing.T) *Handlers {
	// Create an in-memory SQLite database for testing
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	// Create test table
	_, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS body_parts (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            name TEXT NOT NULL
        )
    `)
	if err != nil {
		t.Fatalf("Failed to create test table: %v", err)
	}

	return NewHandlers(db)
}

// cleanup function to clear the test database after each test
func cleanup(t *testing.T, h *Handlers) {
	_, err := h.db.Exec("DELETE FROM body_parts")
	if err != nil {
		t.Fatalf("Failed to cleanup test database: %v", err)
	}
}

func TestCreateBodyPart(t *testing.T) {
	h := setupTestHandler(t)
	defer cleanup(t, h)

	tests := []struct {
		name           string
		requestBody    map[string]interface{}
		expectedStatus int
		expectedBody   bool
	}{
		{
			name: "Valid body part",
			requestBody: map[string]interface{}{
				"name": "Chest",
			},
			expectedStatus: http.StatusCreated,
			expectedBody:   true,
		},
		{
			name: "Empty name",
			requestBody: map[string]interface{}{
				"name": "",
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   true,
		},
		{
			name:           "Missing body",
			requestBody:    nil,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var body []byte
			var err error

			if tt.requestBody != nil {
				body, err = json.Marshal(tt.requestBody)
				if err != nil {
					t.Fatalf("Failed to marshal request body: %v", err)
				}
			}

			req := httptest.NewRequest("POST", "/api/body-parts", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()

			h.CreateBodyPart(rr, req)

			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v",
					status, tt.expectedStatus)
			}

			if tt.expectedBody && rr.Body.Len() == 0 {
				t.Error("handler returned no body when one was expected")
			}

			// For successful creation, verify the response structure
			if tt.expectedStatus == http.StatusCreated {
				var response struct {
					Data struct {
						ID   int64  `json:"id"`
						Name string `json:"name"`
					} `json:"data"`
				}

				if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
					t.Errorf("Failed to decode response body: %v", err)
				}

				if response.Data.Name != tt.requestBody["name"] {
					t.Errorf("handler returned wrong name: got %v want %v",
						response.Data.Name, tt.requestBody["name"])
				}
			}
		})
	}
}

func TestGetBodyPart(t *testing.T) {
	h := setupTestHandler(t)
	defer cleanup(t, h)

	// Insert test data
	result, err := h.db.Exec("INSERT INTO body_parts (name) VALUES (?)", "Chest")
	if err != nil {
		t.Fatalf("Failed to insert test data: %v", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("Failed to get last insert ID: %v", err)
	}

	idStr := strconv.FormatInt(id, 10)

	tests := []struct {
		name           string
		id             string
		expectedStatus int
	}{
		{
			name:           "Existing body part",
			id:             idStr,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Non-existent body part",
			id:             "999",
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "Invalid ID",
			id:             "invalid",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/api/body-parts/"+tt.id, nil)

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.id)
			req = req.WithContext(context.WithValue(req.Context(),
				chi.RouteCtxKey, rctx))

			rr := httptest.NewRecorder()

			h.GetBodyPart(rr, req)

			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v",
					status, tt.expectedStatus)
			}

			if tt.expectedStatus == http.StatusOK {
				var response struct {
					Data struct {
						ID   int64  `json:"id"`
						Name string `json:"name"`
					} `json:"data"`
				}

				if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
					t.Errorf("Failed to decode response body: %v", err)
				}

				if response.Data.ID != id {
					t.Errorf("handler returned wrong ID: got %v want %v",
						response.Data.ID, id)
				}

				if response.Data.Name != "Chest" {
					t.Errorf("handler returned wrong name: got %v want %v",
						response.Data.Name, "Chest")
				}
			}
		})
	}
}
