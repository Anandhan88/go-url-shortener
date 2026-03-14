package handlers

import (
	"database/sql"
	"net/http"
	"strings"
	"time"

	"url-shortener-go/database"

	"github.com/gin-gonic/gin"
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

	var longURL string
	query := `SELECT long_url FROM urls WHERE short_code = ?`
	err := database.DB.QueryRow(query, shortCode).Scan(&longURL)

	if err != nil {
		if err == sql.ErrNoRows {
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
		insertClickQuery := `INSERT INTO clicks (short_code, ip_address, device_type, clicked_at) VALUES (?, ?, ?, ?)`
		_, _ = database.DB.Exec(insertClickQuery, shortCode, ipAddress, deviceType, time.Now())
	}()

	c.Redirect(http.StatusFound, longURL)
}
