package document

import (
	"errors"
	"time"
)

type Link struct {
	URL       string    `json:"url"`
	ExpiresAt time.Time `json:"expiresAt"`
}

var ErrInvalidPreview = errors.New("invalid preview")
