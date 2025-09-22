package repository

import (
	"RIP-WEB/internal/app/ds"
	"errors"
	"fmt"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

func (r *Repository) GetAllAnomalies() ([]ds.Anomaly, error) {
	var anomalies []ds.Anomaly
	err := r.db.Where("is_delete = false").Find(&anomalies).Error
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

func (r *Repository) SearchAnomaliesByName(name string) ([]ds.Anomaly, error) {
	var anomalies []ds.Anomaly
	err := r.db.Where("name ILIKE ? AND is_delete = false", "%"+name+"%").Find(&anomalies).Error
	if err != nil {
		return nil, err
	}
	return anomalies, nil
}

func (r *Repository) GetCartCount() int64 {
	var requestID uint
	var count int64
	creatorID := 1

	err := r.db.Model(&ds.Tree{}).Where("creator_id = ? AND status = ?", creatorID, "черновик").Select("id").First(&requestID).Error
	if err != nil {
		return 0
	}

	err = r.db.Model(&ds.TreeItem{}).Where("request_id = ?", requestID).Count(&count).Error
	if err != nil {
		logrus.Println("Error counting records in request_items:", err)
	}

	return count
}

func (r *Repository) DeleteAnomaly(anomalyID uint) error {
	err := r.db.Model(&ds.Anomaly{}).Where("id = ?", anomalyID).Update("is_delete", true).Error
	if err != nil {
		return fmt.Errorf("ошибка при удалении аномалии с id %d: %w", anomalyID, err)
	}
	return nil
}
