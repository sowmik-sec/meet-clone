package main

import (
	"log"
	"os"

	"meet-clone/internal/modules/auth"
	"meet-clone/internal/shared/infrastructure/database"
	"meet-clone/pkg/jwt"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Load env
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found in root, checking parent directories might be needed or relying on system envs")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Initialize Infrastructure
	db := database.NewMongoClient()
	jwtService := jwt.NewService()

	// Initialize Modules
	authModule := auth.NewModule(db, jwtService)

	// Setup Router
	router := gin.Default()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// Register Routes
	authModule.RegisterRoutes(router)

	// Health Check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	})

	log.Printf("Server starting on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatal("Failed to start server: ", err)
	}
}
