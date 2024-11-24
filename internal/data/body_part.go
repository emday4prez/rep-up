package data

import (
	"database/sql"
	"time"
)

// BodyPart represents a body part record from the database
type BodyPart struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// BodyPartModel wraps the database connection pool
type BodyPartModel struct {
	DB *sql.DB
}

// GetByID retrieves a single body part by its ID
func (m BodyPartModel) GetByID(id int64) (*BodyPart, error) {
	if id < 1 {
		return nil, ErrInvalidInput
	}

	bodyPart := &BodyPart{}

	err := m.DB.QueryRow(`
		SELECT id, name, created_at, updated_at
		FROM body_parts
		WHERE id = ?`, id,
	).Scan(
		&bodyPart.ID,
		&bodyPart.Name,
		&bodyPart.CreatedAt,
		&bodyPart.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrRecordNotFound
		}
		return nil, err
	}

	return bodyPart, nil
}

// GetAll retrieves all body parts from the database
func (m BodyPartModel) GetAll() ([]*BodyPart, error) {
	rows, err := m.DB.Query(`
		SELECT id, name, created_at, updated_at
		FROM body_parts
		ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var bodyParts []*BodyPart

	for rows.Next() {
		bodyPart := &BodyPart{}
		err := rows.Scan(
			&bodyPart.ID,
			&bodyPart.Name,
			&bodyPart.CreatedAt,
			&bodyPart.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		bodyParts = append(bodyParts, bodyPart)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return bodyParts, nil
}

// Create inserts a new body part into the database
func (m BodyPartModel) Create(bodyPart *BodyPart) error {
	if bodyPart.Name == "" {
		return ErrInvalidInput
	}

	// Check if a body part with this name already exists
	var exists bool
	err := m.DB.QueryRow(`
		SELECT EXISTS(
			SELECT 1 FROM body_parts WHERE name = ?
		)`, bodyPart.Name,
	).Scan(&exists)
	if err != nil {
		return err
	}
	if exists {
		return ErrDuplicateRecord
	}

	result, err := m.DB.Exec(`
		INSERT INTO body_parts (name)
		VALUES (?)`,
		bodyPart.Name,
	)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	bodyPart.ID = id
	return nil
}

// Update modifies an existing body part in the database
func (m BodyPartModel) Update(bodyPart *BodyPart) error {
	if bodyPart.ID < 1 || bodyPart.Name == "" {
		return ErrInvalidInput
	}

	// Check if another body part with this name already exists
	var exists bool
	err := m.DB.QueryRow(`
		SELECT EXISTS(
			SELECT 1 FROM body_parts 
			WHERE name = ? AND id != ?
		)`, bodyPart.Name, bodyPart.ID,
	).Scan(&exists)
	if err != nil {
		return err
	}
	if exists {
		return ErrDuplicateRecord
	}

	result, err := m.DB.Exec(`
		UPDATE body_parts 
		SET name = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?`,
		bodyPart.Name, bodyPart.ID,
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

// Delete removes a body part from the database
func (m BodyPartModel) Delete(id int64) error {
	if id < 1 {
		return ErrInvalidInput
	}

	tx, err := m.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Check if the body part is referenced by any exercises
	var exists bool
	err = tx.QueryRow(`
		SELECT EXISTS(
			SELECT 1 FROM exercises WHERE body_part_id = ?
		)`, id,
	).Scan(&exists)
	if err != nil {
		return err
	}
	if exists {
		return ErrReferentialIntegrity
	}

	// If not referenced by any exercises, proceed with deletion
	result, err := tx.Exec("DELETE FROM body_parts WHERE id = ?", id)
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
