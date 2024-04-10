package main

import (
	"github.com/0xlunch/user-service/db"
	"github.com/0xlunch/user-service/handlers"
	"github.com/gin-gonic/gin"
)

func setupRoutes(r *gin.Engine, db *db.DB) {

	h := handlers.NewHandlers(db)
	// User routes
	r.POST("/register", h.RegisterHandler)
	r.POST("/login", h.LoginHandler)
	r.GET("/points", h.GetPointsHandler)
	r.POST("/points", h.UpdatePointsHandler)
}
