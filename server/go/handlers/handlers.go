package handlers

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/0xlunch/user-service/db"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
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
func (h *Handlers) RegisterHandler(c *gin.Context) {
	var user struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := context.Background()
	err := h.DB.RegisterUser(ctx, user.Username, user.Password)
	if err != nil {
		fmt.Printf("%s\n", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to register user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User registered successfully"})
}

// LoginHandler handles user login and returns a JWT auth token
func (h *Handlers) LoginHandler(c *gin.Context) {
	var user struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := context.Background()
	userID, err := h.DB.LoginUser(ctx, user.Username, user.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
		return
	}

	token, err := generateJWT(userID, h.JWTSeed)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}

// GetPoints retrieves the points of the authenticated user
func (h *Handlers) GetPointsHandler(c *gin.Context) {
	// pending based on authentication method
}

// UpdatePoints updates the points of a user
func (h *Handlers) UpdatePointsHandler(c *gin.Context) {
	// pending based on authentication method
}

// generate JWT token
func generateJWT(userID uuid.UUID, seed string) (string, error) {

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
