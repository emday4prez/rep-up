package data

import (
	"database/sql"
)

// GetByID retrieves a single workout with its exercises
func (m WorkoutModel) GetByID(id int64) (*Workout, error) {
	if id < 1 {
		return nil, ErrInvalidInput
	}

	// Start a transaction since we need to query multiple tables
	tx, err := m.DB.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// Get workout
	workout := &Workout{}
	err = tx.QueryRow(`
        SELECT id, user_id, name, date, notes
        FROM workouts
        WHERE id = ?`, id,
	).Scan(&workout.ID, &workout.UserID, &workout.Name, &workout.Date, &workout.Notes)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrRecordNotFound
		}
		return nil, err
	}

	// Get workout exercises
	rows, err := tx.Query(`
        SELECT id, exercise_id, sets, reps, weight
        FROM workout_exercises
        WHERE workout_id = ?`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var detail WorkoutExercise
		err := rows.Scan(
			&detail.ID,
			&detail.ExerciseID,
			&detail.Sets,
			&detail.Reps,
			&detail.Weight,
		)
		if err != nil {
			return nil, err
		}
		detail.WorkoutID = id
		workout.Details = append(workout.Details, detail)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return workout, tx.Commit()
}

// Create inserts a new workout and its exercises
func (m WorkoutModel) Create(workout *Workout) error {
	if workout.UserID < 1 || workout.Name == "" {
		return ErrInvalidInput
	}

	tx, err := m.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Insert workout
	result, err := tx.Exec(`
        INSERT INTO workouts (user_id, name, date, notes)
        VALUES (?, ?, ?, ?)`,
		workout.UserID, workout.Name, workout.Date, workout.Notes,
	)
	if err != nil {
		return err
	}

	// Get the workout ID
	workoutID, err := result.LastInsertId()
	if err != nil {
		return err
	}
	workout.ID = workoutID

	// Insert workout exercises
	for i := range workout.Details {
		workout.Details[i].WorkoutID = workoutID
		result, err := tx.Exec(`
            INSERT INTO workout_exercises (workout_id, exercise_id, sets, reps, weight)
            VALUES (?, ?, ?, ?, ?)`,
			workoutID,
			workout.Details[i].ExerciseID,
			workout.Details[i].Sets,
			workout.Details[i].Reps,
			workout.Details[i].Weight,
		)
		if err != nil {
			return err
		}

		exerciseID, err := result.LastInsertId()
		if err != nil {
			return err
		}
		workout.Details[i].ID = exerciseID
	}

	return tx.Commit()
}

// Update modifies an existing workout and its exercises
func (m WorkoutModel) Update(workout *Workout) error {
	if workout.ID < 1 || workout.UserID < 1 || workout.Name == "" {
		return ErrInvalidInput
	}

	tx, err := m.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Check if workout exists
	var exists bool
	err = tx.QueryRow("SELECT EXISTS(SELECT 1 FROM workouts WHERE id = ?)", workout.ID).Scan(&exists)
	if err != nil {
		return err
	}
	if !exists {
		return ErrRecordNotFound
	}

	// Update workout
	result, err := tx.Exec(`
        UPDATE workouts 
        SET user_id = ?, name = ?, date = ?, notes = ?
        WHERE id = ?`,
		workout.UserID, workout.Name, workout.Date, workout.Notes, workout.ID,
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

	// Delete existing workout exercises
	_, err = tx.Exec("DELETE FROM workout_exercises WHERE workout_id = ?", workout.ID)
	if err != nil {
		return err
	}

	// Insert new workout exercises
	for i := range workout.Details {
		workout.Details[i].WorkoutID = workout.ID
		result, err := tx.Exec(`
            INSERT INTO workout_exercises (workout_id, exercise_id, sets, reps, weight)
            VALUES (?, ?, ?, ?, ?)`,
			workout.ID,
			workout.Details[i].ExerciseID,
			workout.Details[i].Sets,
			workout.Details[i].Reps,
			workout.Details[i].Weight,
		)
		if err != nil {
			return err
		}

		exerciseID, err := result.LastInsertId()
		if err != nil {
			return err
		}
		workout.Details[i].ID = exerciseID
	}

	return tx.Commit()
}

// Delete removes a workout and its exercises
func (m WorkoutModel) Delete(id int64) error {
	if id < 1 {
		return ErrInvalidInput
	}

	tx, err := m.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Delete workout exercises first (due to foreign key)
	_, err = tx.Exec("DELETE FROM workout_exercises WHERE workout_id = ?", id)
	if err != nil {
		return err
	}

	// Delete workout
	result, err := tx.Exec("DELETE FROM workouts WHERE id = ?", id)
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

// GetAll retrieves all workouts for a user
func (m WorkoutModel) GetAll(userID int64) ([]*Workout, error) {
	if userID < 1 {
		return nil, ErrInvalidInput
	}

	// Query all workouts for the user
	rows, err := m.DB.Query(`
        SELECT id, user_id, name, date, notes
        FROM workouts
        WHERE user_id = ?
        ORDER BY date DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var workouts []*Workout

	for rows.Next() {
		workout := &Workout{}
		err := rows.Scan(
			&workout.ID,
			&workout.UserID,
			&workout.Name,
			&workout.Date,
			&workout.Notes,
		)
		if err != nil {
			return nil, err
		}
		workouts = append(workouts, workout)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return workouts, nil
}
