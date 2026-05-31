package storage

import (
	"github.com/krishnassh/picostatus/internal/config"
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) SyncChecks(cfgChecks []config.Check) error {
	for _, c := range cfgChecks {
		check := Check{Name: c.Name}
		err := r.db.
			Where(Check{Name: c.Name}).
			Assign(Check{Target: c.URL, Interval: c.Interval}).
			FirstOrCreate(&check).Error
		if err != nil {
			return err
		}
	}
	return nil
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

func (r *Repository) GetCheckHistory(checkID uint, limit int) ([]Result, error) {
	var results []Result
	if err := r.db.
		Where("check_id = ?", checkID).
		Order("created_at DESC").
		Limit(limit).
		Find(&results).Error; err != nil {
		return nil, err
	}
	return results, nil
}
