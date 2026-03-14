package handlers

import (
	"context"
	"net/http"
	"strings"
	"time"

	"url-shortener-go/database"
	"url-shortener-go/models"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func getDeviceType(userAgent string) string {
	userAgent = strings.ToLower(userAgent)
	if strings.Contains(userAgent, "mobile") {
		return "Mobile"
	} else if strings.Contains(userAgent, "tablet") || strings.Contains(userAgent, "ipad") {
		return "Tablet"
	}
	return "Desktop"
}

func RedirectURL(c *gin.Context) {
	shortCode := c.Param("code")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var url models.URL
	err := database.URLsCollection.FindOne(ctx, bson.M{"short_code": shortCode}).Decode(&url)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"error": "URL not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	// Track Click
	ipAddress := c.ClientIP()
	userAgent := c.Request.UserAgent()
	deviceType := getDeviceType(userAgent)

	go func() {
		clickCtx, clickCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer clickCancel()

		click := models.Click{
			ShortCode:  shortCode,
			IPAddress:  ipAddress,
			DeviceType: deviceType,
			ClickedAt:  time.Now(),
		}
		_, _ = database.ClicksCollection.InsertOne(clickCtx, click)
	}()

	c.Redirect(http.StatusFound, url.LongURL)
}
