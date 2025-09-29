package repository

import (
	"RIP-WEB/internal/app/ds"
	"errors"
	"fmt"
	"strconv"

	"github.com/sirupsen/logrus"
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

func (r *Repository) GetCartCount() int64 {
	var tree ds.Tree
	var count int64

	// Находим черновик текущего пользователя
	creatorID := uint(1)
	err := r.db.Where("creator_id = ? AND status = ?", creatorID, "черновик").First(&tree).Error
	if err != nil {
		// Если черновика нет, возвращаем 0
		return 0
	}

	// Считаем количество элементов в дереве
	err = r.db.Model(&ds.TreeItem{}).Where("tree_id = ?", tree.ID).Count(&count).Error
	if err != nil {
		logrus.Error("Error counting tree items:", err)
		return 0
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
