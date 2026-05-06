package handlers

import (
	"context"
	"net/http"
	"time"

	"url-shortener-go/database"
	"url-shortener-go/models"

	"github.com/gin-gonic/gin"
	"github.com/skip2/go-qrcode"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func GenerateQR(c *gin.Context) {
	shortCode := c.Param("code")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var url models.URL
	err := database.URLsCollection.FindOne(ctx, bson.M{"short_code": shortCode}).Decode(&url)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"error": "Short URL not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	// Generate a 256x256 PNG QR code
	var png []byte
	png, err = qrcode.Encode(url.ShortURL, qrcode.Medium, 256)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate QR code"})
		return
	}

	c.Data(http.StatusOK, "image/png", png)
}
