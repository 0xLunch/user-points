package db

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID       uuid.UUID
	Username string
	Password string
	Points   int
}

func (db *DB) RegisterUser(ctx context.Context, username, password string) error {
	// Hash the password
	hashedPassword, err := hashPassword(password)
	if err != nil {
		return err
	}

	// Insert the user into the database
	_, err = db.Pool.Exec(
		ctx,
		`INSERT INTO users (username, password) VALUES ($1, $2) ON CONFLICT (username) DO NOTHING`,
		username,
		hashedPassword,
	)
	if err != nil {
		return err
	}

	return nil
}

func (db *DB) LoginUser(ctx context.Context, username, password string) (uuid.UUID, error) {
	var userID uuid.UUID
	var hashedPassword string

	// Query the database for the user
	err := db.Pool.QueryRow(ctx, `SELECT id, password FROM users WHERE username = $1`, username).
		Scan(&userID, &hashedPassword)
	if err != nil {
		if err == pgx.ErrNoRows {
			return uuid.Nil, errors.New("invalid username or password")
		}
		return uuid.Nil, err
	}

	// Verify the password
	if err := verifyPassword(hashedPassword, password); err != nil {
		return uuid.Nil, errors.New("invalid username or password")
	}

	return userID, nil
}

// hashPassword securely hashes the password.
func hashPassword(password string) (string, error) {
	// increase bcrypt cost
	cost := 13

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), cost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}

// verifyPassword checks if the provided password matches the hashed password.
func verifyPassword(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

// Add GetUserPoints, UpdateUserPoints, etc.
