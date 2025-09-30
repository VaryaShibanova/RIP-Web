package repository

import (
	"RIP-WEB/internal/app/ds"
	"errors"
	"strconv"

	"gorm.io/gorm"
)

func (r *Repository) GetAllAnomalies() ([]ds.Anomaly, error) {
	var anomalies []ds.Anomaly
	err := r.db.Where("is_delete = false").Order("id ASC").Find(&anomalies).Error
	if err != nil {
		return nil, err
	}
	return anomalies, nil
}

func (r *Repository) GetAnomalyByID(id int) (*ds.Anomaly, error) {
	var anomaly ds.Anomaly
	err := r.db.First(&anomaly, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &anomaly, nil
}

func (r *Repository) SearchAnomalies(query string) ([]ds.Anomaly, error) {
	var anomalies []ds.Anomaly

	// Пытаемся преобразовать запрос в число (для поиска по году)
	year := 0
	if yearValue, err := strconv.Atoi(query); err == nil {
		year = yearValue
	}

	// Поиск по названию, описанию ИЛИ году
	err := r.db.Where(
		"(name ILIKE ? OR description ILIKE ? OR year = ?) AND is_delete = false",
		"%"+query+"%",
		"%"+query+"%",
		year,
	).Order("id ASC").Find(&anomalies).Error

	if err != nil {
		return nil, err
	}
	return anomalies, nil
}

func (r *Repository) CreateAnomaly(anomaly *ds.Anomaly) error {
	anomaly.IsDelete = false
	return r.db.Create(anomaly).Error
}

func (r *Repository) UpdateAnomaly(anomaly *ds.Anomaly) error {
	return r.db.Save(anomaly).Error
}

func (r *Repository) DeleteAnomaly(anomalyID uint) error {
	return r.db.Model(&ds.Anomaly{}).Where("id = ?", anomalyID).Update("is_delete", true).Error
}

func (r *Repository) UpdateAnomalyImage(anomalyID uint, imageURL string) error {
	return r.db.Model(&ds.Anomaly{}).Where("id = ?", anomalyID).Update("image", imageURL).Error
}
