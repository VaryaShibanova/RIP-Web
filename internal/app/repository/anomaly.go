package repository

import (
	"RIP-WEB/internal/app/ds"
	"errors"

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

func (r *Repository) SearchAnomalies(name string, year string) ([]ds.Anomaly, error) {
	var anomalies []ds.Anomaly

	query := r.db.Model(&ds.Anomaly{})

	// Добавляем условия поиска по названию ИЛИ году
	conditions := []string{}
	args := []interface{}{}

	if name != "" {
		conditions = append(conditions, "name ILIKE ?")
		args = append(args, "%"+name+"%")
	}

	if year != "" {
		// ЧАСТИЧНЫЙ поиск по году - ищем вхождение строки в год
		conditions = append(conditions, "CAST(year AS TEXT) ILIKE ?")
		args = append(args, "%"+year+"%")
	}

	// Если есть условия, применяем их
	if len(conditions) > 0 {
		// Объединяем условия через OR
		whereClause := ""
		for i, condition := range conditions {
			if i > 0 {
				whereClause += " OR "
			}
			whereClause += condition
		}
		query = query.Where(whereClause, args...)
	}

	// Убираем фильтр is_delete
	err := query.Order("id ASC").Find(&anomalies).Error

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
