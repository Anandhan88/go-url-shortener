package models

import "time"

type Click struct {
	ID         int       `json:"id"`
	ShortCode  string    `json:"short_code"`
	IPAddress  string    `json:"ip_address"`
	DeviceType string    `json:"device_type"`
	ClickedAt  time.Time `json:"clicked_at"`
}
