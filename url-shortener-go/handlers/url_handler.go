package handlers

import (
	"context"
	"crypto/rand"
	"math/big"
	"net/http"
	"regexp"
	"strings"
	"time"

	"url-shortener-go/database"
	"url-shortener-go/models"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
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
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		var existing models.URL
		err := database.URLsCollection.FindOne(ctx, bson.M{"short_code": customCode}).Decode(&existing)
		if err == nil {
			c.JSON(http.StatusConflict, gin.H{"error": "This keyword is already taken"})
			return
		}
		if err != mongo.ErrNoDocuments {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error checking keyword"})
			return
		}

		shortCode = customCode
	} else {
		shortCode = generateShortCode()
	}

	// Build the public base URL
	baseURL := getPublicBaseURL(c)
	fullShortURL := baseURL + "/" + shortCode

	url := models.URL{
		LongURL:   body.LongURL,
		ShortCode: shortCode,
		ShortURL:  fullShortURL,
		CreatedAt: time.Now(),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := database.URLsCollection.InsertOne(ctx, url)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save URL"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"short_url":  fullShortURL,
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
