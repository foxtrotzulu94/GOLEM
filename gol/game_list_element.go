package gol

import (
	"time"
)

type GameListElement struct {
	ID   int
	Base ListElementFields `gorm:"polymorphic:Owner;"`

	Platform string
	Release  time.Time
	GameType string
}

// TODO: Implement the interface methods for each struct
