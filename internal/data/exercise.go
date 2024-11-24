package data

import (
	"database/sql"
	"time"
)

// Exercise represents an exercise record from the database
type Exercise struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	BodyPartID  int64     `json:"body_part_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ExerciseModel wraps the database connection pool
type ExerciseModel struct {
	DB *sql.DB
}

// GetByID retrieves a single exercise by its ID
func (m ExerciseModel) GetByID(id int64) (*Exercise, error) {
	if id < 1 {
		return nil, ErrInvalidInput
	}

	exercise := &Exercise{}

	err := m.DB.QueryRow(`
		SELECT id, name, description, body_part_id, created_at, updated_at
		FROM exercises
		WHERE id = ?`, id,
	).Scan(
		&exercise.ID,
		&exercise.Name,
		&exercise.Description,
		&exercise.BodyPartID,
		&exercise.CreatedAt,
		&exercise.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrRecordNotFound
		}
		return nil, err
	}

	return exercise, nil
}

// GetByBodyPart retrieves all exercises for a specific body part
func (m ExerciseModel) GetByBodyPart(bodyPartID int64) ([]*Exercise, error) {
	if bodyPartID < 1 {
		return nil, ErrInvalidInput
	}

	rows, err := m.DB.Query(`
		SELECT id, name, description, body_part_id, created_at, updated_at
		FROM exercises
		WHERE body_part_id = ?
		ORDER BY name`, bodyPartID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var exercises []*Exercise

	for rows.Next() {
		exercise := &Exercise{}
		err := rows.Scan(
			&exercise.ID,
			&exercise.Name,
			&exercise.Description,
			&exercise.BodyPartID,
			&exercise.CreatedAt,
			&exercise.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		exercises = append(exercises, exercise)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return exercises, nil
}

// GetAll retrieves all exercises from the database
func (m ExerciseModel) GetAll() ([]*Exercise, error) {
	rows, err := m.DB.Query(`
		SELECT id, name, description, body_part_id, created_at, updated_at
		FROM exercises
		ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var exercises []*Exercise

	for rows.Next() {
		exercise := &Exercise{}
		err := rows.Scan(
			&exercise.ID,
			&exercise.Name,
			&exercise.Description,
			&exercise.BodyPartID,
			&exercise.CreatedAt,
			&exercise.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		exercises = append(exercises, exercise)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return exercises, nil
}

// Create inserts a new exercise into the database
func (m ExerciseModel) Create(exercise *Exercise) error {
	if exercise.Name == "" || exercise.BodyPartID < 1 {
		return ErrInvalidInput
	}

	result, err := m.DB.Exec(`
		INSERT INTO exercises (name, description, body_part_id)
		VALUES (?, ?, ?)`,
		exercise.Name, exercise.Description, exercise.BodyPartID,
	)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	exercise.ID = id
	return nil
}

// Update modifies an existing exercise in the database
func (m ExerciseModel) Update(exercise *Exercise) error {
	if exercise.ID < 1 || exercise.Name == "" || exercise.BodyPartID < 1 {
		return ErrInvalidInput
	}

	result, err := m.DB.Exec(`
		UPDATE exercises 
		SET name = ?, description = ?, body_part_id = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?`,
		exercise.Name, exercise.Description, exercise.BodyPartID, exercise.ID,
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

// Delete removes an exercise from the database
func (m ExerciseModel) Delete(id int64) error {
	if id < 1 {
		return ErrInvalidInput
	}

	tx, err := m.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// First check if the exercise is used in any workouts
	var exists bool
	err = tx.QueryRow(`
		SELECT EXISTS(
			SELECT 1 FROM workout_exercises WHERE exercise_id = ?
		)`, id,
	).Scan(&exists)
	if err != nil {
		return err
	}
	if exists {
		return ErrReferentialIntegrity
	}

	// If not used in any workouts, proceed with deletion
	result, err := tx.Exec("DELETE FROM exercises WHERE id = ?", id)
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

	return tx.Commit()
}
