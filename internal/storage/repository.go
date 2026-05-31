package storage

import (
	"time"

	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) InsertCheck(name, target string, interval time.Duration) (*Check, error) {
	check := &Check{Name: name, Target: target, Interval: interval}
	if err := r.db.Create(check).Error; err != nil {
		return nil, err
	}
	return check, nil
}

func (r *Repository) InsertResult(checkID uint, status Status, success bool, latencyMS int64, checkErr string) (*Result, error) {
	result := &Result{
		CheckID:   checkID,
		Status:    status,
		Success:   success,
		LatencyMS: latencyMS,
		Error:     checkErr,
	}
	if err := r.db.Create(result).Error; err != nil {
		return nil, err
	}
	return result, nil
}

func (r *Repository) GetChecks() ([]Check, error) {
	var checks []Check
	if err := r.db.Find(&checks).Error; err != nil {
		return nil, err
	}
	return checks, nil
}

func (r *Repository) GetLatestResults() ([]Result, error) {
	var results []Result
	if err := r.db.
		Where("id IN (?)", r.db.
			Model(&Result{}).
			Select("MAX(id)").
			Group("check_id"),
		).
		Find(&results).Error; err != nil {
		return nil, err
	}
	return results, nil
}
