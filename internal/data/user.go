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
	OAuthID       string    `json:"oauth_id"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type UserModel struct {
	DB *sql.DB
}
