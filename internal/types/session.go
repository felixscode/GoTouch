package types

import "time"

type TypingSession struct {
	Date     time.Time     `json:"date"`
	WPM      float32       `json:"wpm"`
	Accuracy float32       `json:"accuracy"`
	Errors   int           `json:"errors"`
	Duration time.Duration `json:"duration"`
}

type UserStats struct {
	Sessions []TypingSession `json:"sessions"`
}
