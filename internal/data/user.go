package data

import (
	"database/sql"
	"time"
)

type User struct {
	ID            int64     `json:"id"`
	Email         string    `json:"email"`
	Name          string    `json:"name"`
	OAuthProvider string    `json:"oauth_provider"`
	OAuthID       string    `json:"-"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type UserModel struct {
	DB *sql.DB
}

// GetByOAuth retrieves a user by their OAuth provider and ID
func (m UserModel) GetByOAuth(provider, oauthID string) (*User, error) {
	user := &User{}
	err := m.DB.QueryRow(`
        SELECT id, email, name, oauth_provider, oauth_id, created_at, updated_at
        FROM users 
        WHERE oauth_provider = ? AND oauth_id = ?`,
		provider, oauthID,
	).Scan(
		&user.ID,
		&user.Email,
		&user.Name,
		&user.OAuthProvider,
		&user.OAuthID,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrRecordNotFound
		}
		return nil, err
	}

	return user, nil
}

// CreateOrUpdate creates a new user or updates an existing one
func (m UserModel) CreateOrUpdate(user *User) error {
	// Try to get existing user
	existing, err := m.GetByOAuth(user.OAuthProvider, user.OAuthID)
	if err != nil {
		if err == ErrRecordNotFound {
			// Create new user
			result, err := m.DB.Exec(`
                INSERT INTO users (email, name, oauth_provider, oauth_id)
                VALUES (?, ?, ?, ?)`,
				user.Email, user.Name, user.OAuthProvider, user.OAuthID,
			)
			if err != nil {
				return err
			}

			id, err := result.LastInsertId()
			if err != nil {
				return err
			}
			user.ID = id
			return nil
		}
		return err
	}

	// Update existing user
	_, err = m.DB.Exec(`
        UPDATE users 
        SET email = ?, name = ?, updated_at = CURRENT_TIMESTAMP
        WHERE id = ?`,
		user.Email, user.Name, existing.ID,
	)
	user.ID = existing.ID
	return err
}
