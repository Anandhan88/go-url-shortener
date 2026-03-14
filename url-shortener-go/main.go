package main

import (
	"log"
	"os"

	"url-shortener-go/database"
	"url-shortener-go/routes"

	"github.com/gin-gonic/gin"
)

func main() {
	// Initialize database
	database.InitDB()

	// Initialize Gin router
	router := gin.Default()

	// Load HTML templates
	router.LoadHTMLGlob("templates/*")

	// Setup routes
	routes.SetupRoutes(router)

	// Read port from environment (Render sets PORT automatically)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Start server
	log.Printf("Server running on http://localhost:%s\n", port)
	err := router.Run(":" + port)
	if err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
