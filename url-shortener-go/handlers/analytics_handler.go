package handlers

import (
	"context"
	"math"
	"net/http"
	"time"

	"url-shortener-go/database"
	"url-shortener-go/models"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type RecentClick struct {
	IPAddress  string    `json:"ip_address"`
	DeviceType string    `json:"device_type"`
	ClickedAt  time.Time `json:"clicked_at"`
}

type AnalyticsResponse struct {
	TotalClicks    int            `json:"total_clicks"`
	UniqueVisits   int            `json:"unique_visitors"`
	AvgClicksDay   float64        `json:"avg_clicks_per_day"`
	DaysActive     int            `json:"days_active"`
	ClicksOverTime map[string]int `json:"clicks_over_time"`
	DeviceTypes    map[string]int `json:"device_types"`
	RecentClicks   []RecentClick  `json:"recent_clicks"`
}

func GetAnalytics(c *gin.Context) {
	shortCode := c.Param("code")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Ensure URL exists
	var url models.URL
	err := database.URLsCollection.FindOne(ctx, bson.M{"short_code": shortCode}).Decode(&url)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"error": "URL not found"})
			return
		}
		c.JSON(http.StatusNotFound, gin.H{"error": "URL not found"})
		return
	}

	response := AnalyticsResponse{
		ClicksOverTime: make(map[string]int),
		DeviceTypes:    make(map[string]int),
		RecentClicks:   []RecentClick{},
	}

	filter := bson.M{"short_code": shortCode}

	// 1. Total Clicks
	totalClicks, err := database.ClicksCollection.CountDocuments(ctx, filter)
	if err == nil {
		response.TotalClicks = int(totalClicks)
	}

	// 2. Unique Visitors (Distinct IPs) — use aggregation instead
	uniquePipeline := bson.A{
		bson.M{"$match": bson.M{"short_code": shortCode}},
		bson.M{"$group": bson.M{"_id": "$ip_address"}},
		bson.M{"$count": "count"},
	}
	uniqueCursor, err := database.ClicksCollection.Aggregate(ctx, uniquePipeline)
	if err == nil {
		defer uniqueCursor.Close(ctx)
		if uniqueCursor.Next(ctx) {
			var result struct {
				Count int `bson:"count"`
			}
			if uniqueCursor.Decode(&result) == nil {
				response.UniqueVisits = result.Count
			}
		}
	}

	// 3. Clicks Over Time (aggregation: group by date)
	clicksOverTimePipeline := bson.A{
		bson.M{"$match": bson.M{"short_code": shortCode}},
		bson.M{"$group": bson.M{
			"_id": bson.M{
				"$dateToString": bson.M{"format": "%Y-%m-%d", "date": "$clicked_at"},
			},
			"count": bson.M{"$sum": 1},
		}},
		bson.M{"$sort": bson.M{"_id": 1}},
	}

	cursor, err := database.ClicksCollection.Aggregate(ctx, clicksOverTimePipeline)
	if err == nil {
		defer cursor.Close(ctx)
		for cursor.Next(ctx) {
			var result struct {
				Date  string `bson:"_id"`
				Count int    `bson:"count"`
			}
			if cursor.Decode(&result) == nil {
				response.ClicksOverTime[result.Date] = result.Count
			}
		}
	}

	// 4. Device Types (aggregation: group by device_type)
	devicePipeline := bson.A{
		bson.M{"$match": bson.M{"short_code": shortCode}},
		bson.M{"$group": bson.M{
			"_id":   "$device_type",
			"count": bson.M{"$sum": 1},
		}},
	}

	deviceCursor, err := database.ClicksCollection.Aggregate(ctx, devicePipeline)
	if err == nil {
		defer deviceCursor.Close(ctx)
		for deviceCursor.Next(ctx) {
			var result struct {
				Device string `bson:"_id"`
				Count  int    `bson:"count"`
			}
			if deviceCursor.Decode(&result) == nil {
				response.DeviceTypes[result.Device] = result.Count
			}
		}
	}

	// 5. Days Active
	response.DaysActive = len(response.ClicksOverTime)

	// 6. Avg Clicks per Day
	if response.DaysActive > 0 {
		response.AvgClicksDay = float64(response.TotalClicks) / float64(response.DaysActive)
		response.AvgClicksDay = math.Round(response.AvgClicksDay*100) / 100
	} else if response.TotalClicks > 0 {
		response.DaysActive = 1
		response.AvgClicksDay = float64(response.TotalClicks)
	}

	// 7. Recent Clicks (last 10, sorted by clicked_at desc)
	findOpts := options.Find().SetSort(bson.M{"clicked_at": -1}).SetLimit(10)
	recentCursor, err := database.ClicksCollection.Find(ctx, filter, findOpts)
	if err == nil {
		defer recentCursor.Close(ctx)
		for recentCursor.Next(ctx) {
			var click models.Click
			if recentCursor.Decode(&click) == nil {
				response.RecentClicks = append(response.RecentClicks, RecentClick{
					IPAddress:  click.IPAddress,
					DeviceType: click.DeviceType,
					ClickedAt:  click.ClickedAt,
				})
			}
		}
	}

	c.JSON(http.StatusOK, response)
}
