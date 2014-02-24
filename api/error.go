package api

import (
	"time"
)

type Error struct {
	Time  time.Time `json:"time"`
	Error string    `json:"error"`
}
