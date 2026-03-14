package handlers

import (
	"math"
	"net/http"

	"time"

	"url-shortener-go/database"

	"github.com/gin-gonic/gin"
)

type RecentClick struct {
	IPAddress  string    `json:"ip_address"`
	DeviceType string    `json:"device_type"`
	ClickedAt  time.Time `json:"clicked_at"`
}

type AnalyticsResponse struct {
	TotalClicks   int            `json:"total_clicks"`
	UniqueVisits  int            `json:"unique_visitors"`
	AvgClicksDay  float64        `json:"avg_clicks_per_day"`
	DaysActive    int            `json:"days_active"`
	ClicksOverTime map[string]int `json:"clicks_over_time"`
	DeviceTypes   map[string]int `json:"device_types"`
	RecentClicks  []RecentClick  `json:"recent_clicks"`
}

func GetAnalytics(c *gin.Context) {
	shortCode := c.Param("code")
	
	// Ensure URL exists
	var exists bool
	err := database.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM urls WHERE short_code = ?)", shortCode).Scan(&exists)
	if err != nil || !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "URL not found"})
		return
	}

	response := AnalyticsResponse{
		ClicksOverTime: make(map[string]int),
		DeviceTypes:    make(map[string]int),
	}

	// 1. Total Clicks
	database.DB.QueryRow(`SELECT COUNT(*) FROM clicks WHERE short_code = ?`, shortCode).Scan(&response.TotalClicks)

	// 2. Unique Visitors (Unique IPs)
	database.DB.QueryRow(`SELECT COUNT(DISTINCT ip_address) FROM clicks WHERE short_code = ?`, shortCode).Scan(&response.UniqueVisits)

	// 3. Days Active
	database.DB.QueryRow(`
		SELECT COUNT(DISTINCT date(clicked_at)) 
		FROM clicks 
		WHERE short_code = ?`, shortCode).Scan(&response.DaysActive)

	// 4. Avg Clicks per day
	if response.DaysActive > 0 {
		response.AvgClicksDay = float64(response.TotalClicks) / float64(response.DaysActive)
        response.AvgClicksDay = math.Round(response.AvgClicksDay*100) / 100
	} else if response.TotalClicks > 0 {
		response.DaysActive = 1
        response.AvgClicksDay = float64(response.TotalClicks)
    }

	// 5. Clicks Over Time
	rows, err := database.DB.Query(`
		SELECT date(clicked_at) as date, COUNT(*) as count 
		FROM clicks 
		WHERE short_code = ? 
		GROUP BY date(clicked_at) 
		ORDER BY date(clicked_at) ASC`, shortCode)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var date string
			var count int
			rows.Scan(&date, &count)
			response.ClicksOverTime[date] = count
		}
	}

	// 6. Device Types
	deviceRows, err := database.DB.Query(`
		SELECT device_type, COUNT(*) as count 
		FROM clicks 
		WHERE short_code = ? 
		GROUP BY device_type`, shortCode)
	if err == nil {
		defer deviceRows.Close()
		for deviceRows.Next() {
			var device string
			var count int
			deviceRows.Scan(&device, &count)
			response.DeviceTypes[device] = count
		}
	}

	// 7. Recent Clicks
	response.RecentClicks = []RecentClick{}
	recentRows, err := database.DB.Query(`
		SELECT ip_address, device_type, clicked_at 
		FROM clicks 
		WHERE short_code = ? 
		ORDER BY clicked_at DESC LIMIT 10`, shortCode)
	if err == nil {
		defer recentRows.Close()
		for recentRows.Next() {
			var click RecentClick
			recentRows.Scan(&click.IPAddress, &click.DeviceType, &click.ClickedAt)
			response.RecentClicks = append(response.RecentClicks, click)
		}
	}

	c.JSON(http.StatusOK, response)
}
