package routes

import (
	"url-shortener-go/handlers"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(router *gin.Engine) {
	// Serve static files
	router.Static("/static", "./static")

	// HTML rendering routes
	router.GET("/", func(c *gin.Context) {
		c.HTML(200, "index.html", gin.H{})
	})

	router.GET("/dashboard", func(c *gin.Context) {
		c.HTML(200, "dashboard.html", gin.H{})
	})

	// API routes
	api := router.Group("/api")
	{
		api.POST("/shorten", handlers.ShortenURL)
		api.GET("/analytics/:code", handlers.GetAnalytics)
	}

	// Redirect route (must be last to not catch /dashboard or /api)
	router.GET("/:code", handlers.RedirectURL)
}
