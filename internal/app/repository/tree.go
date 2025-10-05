package repository

import (
	"RIP-WEB/internal/app/ds"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
)

func (r *Repository) AddAnomalyToTree(treeID, anomalyID uint, anomalousRings string, calculatedYear int) error {
	// Проверяем, не добавлена ли уже эта аномалия в дерево
	var existingItem ds.TreeItem
	err := r.db.Where("tree_id = ? AND anomaly_id = ?", treeID, anomalyID).First(&existingItem).Error

	if err == nil {
		// Аномалия уже есть в дереве - обновляем существующую запись
		return r.db.Model(&existingItem).Updates(map[string]interface{}{
			"anomalous_rings": anomalousRings,
			"calculated_year": calculatedYear,
		}).Error
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	// Создаем новую запись БЕЗ указания ID (база данных сама сгенерирует)
	treeItem := ds.TreeItem{
		TreeID:         treeID,
		AnomalyID:      anomalyID,
		AnomalousRings: anomalousRings,
		CalculatedYear: calculatedYear,
	}

	return r.db.Create(&treeItem).Error
}

func (r *Repository) GetTreeWithItems(treeID uint) (*ds.Tree, []ds.TreeItem, error) {
	var tree ds.Tree
	err := r.db.Preload("Creator").Preload("Moderator").First(&tree, treeID).Error
	if err != nil {
		return nil, nil, err
	}

	var treeItems []ds.TreeItem
	err = r.db.Where("tree_id = ?", treeID).Preload("Anomaly").Find(&treeItems).Error
	if err != nil {
		return nil, nil, err
	}

	return &tree, treeItems, nil
}

func (r *Repository) GetTreesWithFilters(status string, dateFrom, dateTo time.Time) ([]ds.Tree, error) {
	var trees []ds.Tree
	query := r.db.Preload("Creator").Preload("Moderator") // УБИРАЕМ ФИЛЬТРАЦИЮ ПО СТАТУСУ

	if status != "" {
		query = query.Where("status = ?", status)
	} else {
		// ПОКАЗЫВАЕМ ВСЕ ЗАЯВКИ КРОМЕ УДАЛЕННЫХ
		query = query.Where("status != ?", "удалён")
	}

	if !dateFrom.IsZero() {
		query = query.Where("date_create >= ?", dateFrom)
	}

	if !dateTo.IsZero() {
		query = query.Where("date_create <= ?", dateTo)
	}

	err := query.Order("date_create DESC").Find(&trees).Error
	if err != nil {
		return nil, err
	}

	return trees, nil
}

func (r *Repository) UpdateTree(tree *ds.Tree) error {
	tree.DateUpdate = time.Now()

	return r.db.Save(tree).Error
}

func (r *Repository) FormTree(treeID uint) error {
	var tree ds.Tree
	if err := r.db.First(&tree, treeID).Error; err != nil {
		return err
	}

	// Проверяем обязательные поля
	if tree.Description == "" || tree.TotalRings == 0 {
		return fmt.Errorf("описание и общее количество колец обязательны для формирования заявки")
	}

	// Проверяем что есть аномалии в заявке
	var itemCount int64
	r.db.Model(&ds.TreeItem{}).Where("tree_id = ?", treeID).Count(&itemCount)
	if itemCount == 0 {
		return fmt.Errorf("заявка должна содержать хотя бы одну аномалию")
	}

	// НЕ рассчитываем final_year при формировании
	return r.db.Model(&tree).Updates(map[string]interface{}{
		"status":      "сформирован",
		"date_update": time.Now(),
	}).Error
}

// repository/tree.go - обновим метод CompleteTree
func (r *Repository) CompleteTree(treeID uint, moderatorID uint, action string) error {
	var tree ds.Tree
	if err := r.db.First(&tree, treeID).Error; err != nil {
		return err
	}

	if tree.Status != "сформирован" {
		return fmt.Errorf("можно завершать только сформированные заявки")
	}

	// ✅ РАСЧЕТ final_year ТОЛЬКО ПРИ ЗАВЕРШЕНИИ
	finalYear := 0
	if action == "complete" {
		finalYear = r.calculateFinalYear(treeID)
	} else if action == "reject" {
		finalYear = 0 // Для отклоненных заявок FinalYear = 0
	}

	updates := map[string]interface{}{
		"moderator_id": moderatorID,
		"date_update":  time.Now(),
		"date_finish":  time.Now(),
		"final_year":   finalYear, // УСТАНАВЛИВАЕМ РАССЧИТАННОЕ ЗНАЧЕНИЕ
	}

	switch action {
	case "complete":
		updates["status"] = "завершён"
	case "reject":
		updates["status"] = "отклонён"
	default:
		return fmt.Errorf("неверное действие: %s", action)
	}

	return r.db.Model(&tree).Updates(updates).Error
}

// repository/tree.go - улучшенная функция calculateFinalYear
func (r *Repository) calculateFinalYear(treeID uint) int {
	var treeItems []ds.TreeItem
	r.db.Where("tree_id = ?", treeID).Preload("Anomaly").Find(&treeItems)

	if len(treeItems) == 0 {
		return 0
	}

	// Вариант A - среднее значение CalculatedYear
	total := 0
	validCount := 0

	for _, item := range treeItems {
		if item.CalculatedYear > 0 {
			total += item.CalculatedYear
			validCount++
		}
	}

	if validCount == 0 {
		return 0
	}

	return total / validCount
}

func (r *Repository) DeleteTree(treeID uint) error {
	var tree ds.Tree
	if err := r.db.First(&tree, treeID).Error; err != nil {
		return err
	}

	if tree.Status == "удалён" {
		return fmt.Errorf("заявка уже удалена")
	}

	return r.db.Model(&tree).Updates(map[string]interface{}{
		"status":      "удалён",
		"date_update": time.Now(),
	}).Error
}

func (r *Repository) GetTreeByID(treeID uint) (*ds.Tree, error) {
	var tree ds.Tree
	err := r.db.Preload("Creator").Preload("Moderator").First(&tree, treeID).Error
	if err != nil {
		return nil, err
	}
	return &tree, nil
}

func (r *Repository) UpdateTreeItem(treeID, anomalyID uint, anomalousRings string, calculatedYear int) error {
	var treeItem ds.TreeItem
	err := r.db.Where("tree_id = ? AND anomaly_id = ?", treeID, anomalyID).First(&treeItem).Error
	if err != nil {
		return err
	}

	// Обновляем элемент заявки
	err = r.db.Model(&treeItem).Updates(map[string]interface{}{
		"anomalous_rings": anomalousRings,
		"calculated_year": calculatedYear,
	}).Error

	if err != nil {
		return err
	}

	return nil
}

func (r *Repository) RemoveFromTree(treeID, anomalyID uint) error {
	// Удаляем элемент из заявки
	err := r.db.Where("tree_id = ? AND anomaly_id = ?", treeID, anomalyID).Delete(&ds.TreeItem{}).Error
	if err != nil {
		return err
	}

	return nil
}
