package data

import (
	"database/sql"
	"time"
)

type Models struct {
	Workouts *WorkoutModel
	// Add other models as needed:
	BodyParts *BodyPartModel
	Exercises *ExerciseModel
}

type BodyPart struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type Workout struct {
	ID      int64             `json:"id"`
	UserID  int64             `json:"user_id"`
	Name    string            `json:"name"`
	Date    time.Time         `json:"date"`
	Notes   string            `json:"notes"`
	Details []WorkoutExercise `json:"details,omitempty"`
}

type WorkoutExercise struct {
	ID         int64   `json:"id"`
	WorkoutID  int64   `json:"workout_id"`
	ExerciseID int64   `json:"exercise_id"`
	Sets       int     `json:"sets"`
	Reps       int     `json:"reps"`
	Weight     float64 `json:"weight"`
}

// BodyPartModel handles database operations for body parts
type BodyPartModel struct {
	DB *sql.DB
}

// WorkoutModel handles database operations for workouts
type WorkoutModel struct {
	DB *sql.DB
}

// GetBodyPart retrieves a single body part by ID
func (m BodyPartModel) GetByID(id int64) (*BodyPart, error) {
	if id < 1 {
		return nil, ErrInvalidInput
	}

	query := `
		SELECT id, name
		FROM body_parts
		WHERE id = ?`

	var bodyPart BodyPart
	err := m.DB.QueryRow(query, id).Scan(&bodyPart.ID, &bodyPart.Name)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrRecordNotFound
		}
		return nil, err
	}

	return &bodyPart, nil
}

// Similar Get methods need to be implemented for Exercise and Workout models

// Additional methods like Create, Update, Delete, and List also needed
