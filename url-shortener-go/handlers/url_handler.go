package handlers

import (
	"crypto/rand"
	"math/big"
	"net/http"
	"regexp"
	"strings"
	"time"

	"url-shortener-go/database"
	"url-shortener-go/models"

	"github.com/gin-gonic/gin"
)

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func generateShortCode() string {
	b := make([]byte, 6)
	for i := range b {
		num, _ := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		b[i] = charset[num.Int64()]
	}
	return string(b)
}

func ShortenURL(c *gin.Context) {
	var body struct {
		LongURL    string `json:"long_url"`
		CustomCode string `json:"custom_code"`
	}

	if err := c.BindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	if body.LongURL == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "URL cannot be empty"})
		return
	}

	var shortCode string
	
	customCode := strings.TrimSpace(body.CustomCode)
	if customCode != "" {
		// 1. Validation
		if len(customCode) < 3 || len(customCode) > 20 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Custom keyword must be between 3 and 20 characters"})
			return
		}
		
		matched, _ := regexp.MatchString(`^[a-zA-Z0-9-]+$`, customCode)
		if !matched {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Custom keyword can only contain letters, numbers, and hyphens"})
			return
		}

		// 2. Duplicate Check
		var exists bool
		err := database.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM urls WHERE short_code = ?)", customCode).Scan(&exists)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error checking keyword"})
			return
		}
		
		if exists {
			c.JSON(http.StatusConflict, gin.H{"error": "This keyword is already taken"})
			return
		}

		shortCode = customCode
	} else {
		shortCode = generateShortCode()
	}

	url := models.URL{
		LongURL:   body.LongURL,
		ShortCode: shortCode,
		CreatedAt: time.Now(),
	}

	query := `INSERT INTO urls (long_url, short_code, created_at) VALUES (?, ?, ?)`
	result, err := database.DB.Exec(query, url.LongURL, url.ShortCode, url.CreatedAt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save URL"})
		return
	}

	id, _ := result.LastInsertId()
	url.ID = int(id)

	// Build the public base URL
	baseURL := getPublicBaseURL(c)

	c.JSON(http.StatusOK, gin.H{
		"short_url":  baseURL + "/" + shortCode,
		"short_code": shortCode,
	})
}

// getPublicBaseURL detects the public URL using proxy headers (works with Render, ngrok, etc.)
func getPublicBaseURL(c *gin.Context) string {
	scheme := "http"
	host := c.Request.Host
	if proto := c.GetHeader("X-Forwarded-Proto"); proto != "" {
		scheme = proto
	}
	if fHost := c.GetHeader("X-Forwarded-Host"); fHost != "" {
		host = fHost
	}
	return scheme + "://" + host
}

