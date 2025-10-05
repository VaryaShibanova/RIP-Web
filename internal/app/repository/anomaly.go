package repository

import (
	"RIP-WEB/internal/app/ds"
	"errors"
	"strconv"

	"gorm.io/gorm"
)

func (r *Repository) GetAllAnomalies() ([]ds.Anomaly, error) {
	var anomalies []ds.Anomaly
	// Убираем фильтр is_delete, показываем все аномалии
	err := r.db.Order("id ASC").Find(&anomalies).Error
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

	// Убираем фильтр is_delete
	err := r.db.Where(
		"(name ILIKE ? OR description ILIKE ? OR year = ?)",
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
	return r.db.Create(anomaly).Error
}

func (r *Repository) UpdateAnomaly(anomaly *ds.Anomaly) error {
	return r.db.Save(anomaly).Error
}

// УДАЛЕНО: DeleteAnomaly с soft delete
// Вместо этого используем прямое удаление из БД

func (r *Repository) UpdateAnomalyImage(anomalyID uint, imageURL string) error {
	return r.db.Model(&ds.Anomaly{}).Where("id = ?", anomalyID).Update("image", imageURL).Error
}
