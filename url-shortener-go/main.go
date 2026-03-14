package main

import (
	"log"
	
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

	// Start server
	log.Println("Server running on http://localhost:8080")
	err := router.Run(":8080")
	if err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
