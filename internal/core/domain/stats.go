package domain

import (
	"time"

	"github.com/google/uuid"
)

type URLStats struct {
	ID         string    `json:"id"`
	URLID      string    `json:"url_id"`
	AccessedAt time.Time `json:"accessed_at"`
	IPAddress  string    `json:"ip_address,omitempty"`
	UserAgent  string    `json:"user_agent,omitempty"`
	Referer    string    `json:"referer,omitempty"`
}

func NewURLStats(urlID, ipAddress, userAgent, referer string) *URLStats {
	return &URLStats{
		ID:         uuid.New().String(),
		URLID:      urlID,
		AccessedAt: time.Now().UTC(),
		IPAddress:  ipAddress,
		UserAgent:  userAgent,
		Referer:    referer,
	}
}
