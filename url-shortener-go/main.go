package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"url-shortener-go/database"
	"url-shortener-go/routes"

	"github.com/gin-gonic/gin"
)

func main() {
	// Connect to MongoDB
	database.ConnectMongo()

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

	// Graceful shutdown to disconnect MongoDB
	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit
		log.Println("Shutting down... disconnecting MongoDB")
		database.DisconnectMongo()
		os.Exit(0)
	}()

	// Start server
	log.Printf("Server running on http://localhost:%s\n", port)
	err := router.Run(":" + port)
	if err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
