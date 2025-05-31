package models

import "time"

type WithdrawnInfo struct {
	Number      string  `json:"order"`
	Sum         float64 `json:"sum"`
	WithdrawnAt string  `json:"withdrawn_at"`
}

type Withdraw struct {
	Number      int64     `json:"order"`
	UserID      string    `json:"user_id"`
	Sum         float64   `json:"sum"`
	WithdrawnAt time.Time `json:"withdrawn_at"`
}
