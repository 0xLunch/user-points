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
	tag, err := db.Pool.Exec(
		ctx,
		`INSERT INTO users (username, password) VALUES ($1, $2) ON CONFLICT (username) DO NOTHING`,
		username,
		hashedPassword,
	)
	if err != nil {
		return err
	}

	// duplicate username found
	if tag.RowsAffected() == 0 {
		return errors.New("username already exists")
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
			return uuid.Nil, errors.New("invalid username")
		}
		return uuid.Nil, err
	}

	// Verify the password
	if err := verifyPassword(hashedPassword, password); err != nil {
		return uuid.Nil, errors.New("invalid password")
	}

	return userID, nil
}

// GetUserPoints retrieves the points of the authenticated user.
func (db *DB) GetUserPoints(ctx context.Context, userID uuid.UUID) (int, error) {
	var points int
	err := db.Pool.QueryRow(ctx, `SELECT points FROM users WHERE id = $1`, userID).Scan(&points)
	if err != nil {
		return 0, err
	}
	return points, nil
}

// AddUserPoints adds points to a user.
func (db *DB) AddUserPoints(ctx context.Context, userID uuid.UUID, points int) error {
	if points < 1 {
		return errors.New("points value must be >=1")
	}
	_, err := db.Pool.Exec(
		ctx,
		`UPDATE users SET points = points + $1 WHERE id = $2`,
		points,
		userID,
	)
	return err
}

// UpdateUserPoints updates the points of a user.
func (db *DB) UpdateUserPoints(ctx context.Context, userID uuid.UUID, points int) error {
	if points < 0 {
		return errors.New("points value must be positive")
	}
	_, err := db.Pool.Exec(ctx, `UPDATE users SET points = $1 WHERE id = $2`, points, userID)
	return err
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
