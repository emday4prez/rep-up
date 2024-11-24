package data

import (
	"database/sql"
	"time"
)

// WorkoutExercise represents a junction between workouts and exercises with additional metadata
type WorkoutExercise struct {
	ID         int64     `json:"id"`
	WorkoutID  int64     `json:"workout_id"`
	ExerciseID int64     `json:"exercise_id"`
	Sets       int       `json:"sets"`
	Reps       int       `json:"reps"`
	Weight     *float64  `json:"weight,omitempty"` // Pointer to allow NULL in database
	Notes      string    `json:"notes,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	// Include nested structs for related data
	Exercise *Exercise `json:"exercise,omitempty"`
}

// WorkoutExerciseModel wraps the database connection pool
type WorkoutExerciseModel struct {
	DB *sql.DB
}

// GetByWorkoutID retrieves all exercises for a specific workout
func (m WorkoutExerciseModel) GetByWorkoutID(workoutID int64) ([]*WorkoutExercise, error) {
	if workoutID < 1 {
		return nil, ErrInvalidInput
	}

	// Join with exercises table to get exercise details
	rows, err := m.DB.Query(`
		SELECT 
			we.id, we.workout_id, we.exercise_id, we.sets, we.reps, 
			we.weight, we.notes, we.created_at, we.updated_at,
			e.name, e.description, e.body_part_id
		FROM workout_exercises we
		JOIN exercises e ON we.exercise_id = e.id
		WHERE we.workout_id = ?
		ORDER BY we.id`, workoutID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var workoutExercises []*WorkoutExercise

	for rows.Next() {
		we := &WorkoutExercise{
			Exercise: &Exercise{},
		}
		err := rows.Scan(
			&we.ID,
			&we.WorkoutID,
			&we.ExerciseID,
			&we.Sets,
			&we.Reps,
			&we.Weight,
			&we.Notes,
			&we.CreatedAt,
			&we.UpdatedAt,
			&we.Exercise.Name,
			&we.Exercise.Description,
			&we.Exercise.BodyPartID,
		)
		if err != nil {
			return nil, err
		}
		we.Exercise.ID = we.ExerciseID
		workoutExercises = append(workoutExercises, we)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return workoutExercises, nil
}

// Create adds a new exercise to a workout
func (m WorkoutExerciseModel) Create(we *WorkoutExercise) error {
	if we.WorkoutID < 1 || we.ExerciseID < 1 || we.Sets < 1 || we.Reps < 1 {
		return ErrInvalidInput
	}

	result, err := m.DB.Exec(`
		INSERT INTO workout_exercises (workout_id, exercise_id, sets, reps, weight, notes)
		VALUES (?, ?, ?, ?, ?, ?)`,
		we.WorkoutID, we.ExerciseID, we.Sets, we.Reps, we.Weight, we.Notes,
	)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	we.ID = id
	return nil
}

// Update modifies an existing workout exercise
func (m WorkoutExerciseModel) Update(we *WorkoutExercise) error {
	if we.ID < 1 || we.Sets < 1 || we.Reps < 1 {
		return ErrInvalidInput
	}

	result, err := m.DB.Exec(`
		UPDATE workout_exercises 
		SET sets = ?, reps = ?, weight = ?, notes = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?`,
		we.Sets, we.Reps, we.Weight, we.Notes, we.ID,
	)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrRecordNotFound
	}

	return nil
}

// Delete removes an exercise from a workout
func (m WorkoutExerciseModel) Delete(id int64) error {
	if id < 1 {
		return ErrInvalidInput
	}

	result, err := m.DB.Exec("DELETE FROM workout_exercises WHERE id = ?", id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrRecordNotFound
	}

	return nil
}

// DeleteAllForWorkout removes all exercises from a specific workout
func (m WorkoutExerciseModel) DeleteAllForWorkout(workoutID int64) error {
	if workoutID < 1 {
		return ErrInvalidInput
	}

	result, err := m.DB.Exec("DELETE FROM workout_exercises WHERE workout_id = ?", workoutID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrRecordNotFound
	}

	return nil
}
