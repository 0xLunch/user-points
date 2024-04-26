package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/0xlunch/user-service/internal/db"
	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
)

type Handlers struct {
	DB      *db.DB
	JWTSeed string
}

// JWT Seed
const SEEDEnvVar string = "JWT_SEED"

func NewHandlers(db *db.DB) *Handlers {
	// get seed
	seed := os.Getenv(SEEDEnvVar)
	if seed == "" {
		panic("JWT_SEED environment not set")
	}
	_jwtSeed := seed

	return &Handlers{
		DB:      db,
		JWTSeed: _jwtSeed,
	}
}

// Register handles user registration
func (h *Handlers) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	var user struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	if err := bindJSON(r, &user); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	err := h.DB.RegisterUser(ctx, user.Username, user.Password)
	if err != nil {
		fmt.Printf("%s\n", err.Error())
		http.Error(w, "Failed to register user", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "User registered successfully"})
}

// LoginHandler handles user login and returns a JWT auth token
func (h *Handlers) LoginHandler(w http.ResponseWriter, r *http.Request) {
	var user struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	if err := bindJSON(r, &user); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	userID, err := h.DB.LoginUser(ctx, user.Username, user.Password)
	if err != nil {
		fmt.Printf("%s\n", err.Error())
		http.Error(w, "Failed to register user", http.StatusInternalServerError)
		return
	}

	token, err := h.generateJWT(userID, h.JWTSeed)
	if err != nil {
		fmt.Printf("%s\n", err.Error())
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"token": token})
}

// GetPoints retrieves the points of the authenticated user
func (h *Handlers) GetPointsHandler(w http.ResponseWriter, r *http.Request) {

	userID, err := h.getUserIDFromRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	ctx := context.Background()
	points, err := h.DB.GetUserPoints(ctx, userID)
	if err != nil {
		http.Error(w, "Failed to get user points", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]int{"points": points})
}

// UpdatePoints updates the points of a user
func (h *Handlers) UpdatePointsHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := h.getUserIDFromRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	var update struct {
		Points int `json:"points"`
	}
	if err := bindJSON(r, &update); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	err = h.DB.UpdateUserPoints(ctx, userID, update.Points)
	if err != nil {
		http.Error(w, "Failed to update user points", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "User points updated successfully"})
}

// generate JWT token
func (h *Handlers) generateJWT(userID uuid.UUID, seed string) (string, error) {

	signingMethod := jwt.SigningMethodHS256

	token := jwt.NewWithClaims(signingMethod, jwt.MapClaims{
		"userID": userID,
		"exp":    time.Now().Add(time.Hour * 24).Unix(), // Token expires after 24 hours
	})

	tokenString, err := token.SignedString([]byte(seed))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// validateJWT validates the JWT token and returns the user ID if the token is valid.
func (h *Handlers) validateJWT(tokenString string) (uuid.UUID, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(h.JWTSeed), nil
	})
	if err != nil {
		fmt.Printf("%s\n", err.Error())
		return uuid.Nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userID, ok := claims["userID"].(string)
		if !ok {
			return uuid.Nil, fmt.Errorf("userID not found in token")
		}
		return uuid.Parse(userID)
	} else {
		return uuid.Nil, fmt.Errorf("invalid token")
	}
}

// getUserIDFromRequest extracts the JWT token from the request header, validates it,
// and returns the user ID if the token is valid.
func (h *Handlers) getUserIDFromRequest(r *http.Request) (uuid.UUID, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return uuid.Nil, fmt.Errorf("Authorization header is missing")
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	userID, err := h.validateJWT(tokenString)
	if err != nil {
		return uuid.Nil, fmt.Errorf("Invalid token: %v", err)
	}

	return userID, nil
}

// bindJSON to interface
func bindJSON(r *http.Request, v interface{}) error {
	return json.NewDecoder(r.Body).Decode(v)
}
