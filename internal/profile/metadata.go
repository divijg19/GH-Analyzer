package profile

import (
	"time"
)

type UserMetadata struct {
	Name      string    `json:"name"`
	Bio       string    `json:"bio"`
	Location  string    `json:"location"`
	Company   string    `json:"company"`
	Followers int       `json:"followers"`
	Following int       `json:"following"`
	CreatedAt time.Time `json:"created_at"`
}
