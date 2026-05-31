package storage

import (
	"time"

	"gorm.io/gorm"
)

type Check struct {
	gorm.Model
	Name     string        `gorm:"not null"`
	Target   string        `gorm:"not null"`
	Interval time.Duration `gorm:"not null"`
	Results  []Result      `gorm:"foreignKey:CheckID"`
}

type Status string

const (
	StatusUp      Status = "up"
	StatusDown    Status = "down"
	StatusUnknown Status = "unknown"
)

type Result struct {
	gorm.Model
	CheckID   uint   `gorm:"not null"`
	Status    Status `gorm:"not null"`
	Success   bool   `gorm:"not null"`
	LatencyMS int64  `gorm:"not null"`
	Error     string
}
