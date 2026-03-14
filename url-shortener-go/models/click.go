package models

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type Click struct {
	ID         bson.ObjectID `json:"id" bson:"_id,omitempty"`
	ShortCode  string        `json:"short_code" bson:"short_code"`
	IPAddress  string        `json:"ip_address" bson:"ip_address"`
	DeviceType string        `json:"device_type" bson:"device_type"`
	ClickedAt  time.Time     `json:"clicked_at" bson:"clicked_at"`
}
