package models

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type URL struct {
	ID        bson.ObjectID `json:"id" bson:"_id,omitempty"`
	LongURL   string        `json:"long_url" bson:"long_url"`
	ShortCode string        `json:"short_code" bson:"short_code"`
	ShortURL  string        `json:"short_url" bson:"short_url"`
	CreatedAt time.Time     `json:"created_at" bson:"created_at"`
}
