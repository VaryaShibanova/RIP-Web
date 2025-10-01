package repository

import (
	"RIP-WEB/internal/app/ds"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
)

func (r *Repository) GetOrCreateDraftTree(userID uint) (*ds.Tree, error) {
	var tree ds.Tree
	err := r.db.Where("creator_id = ? AND status = ?", userID, "черновик").First(&tree).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// Создаем новое дерево
			newTree := ds.Tree{
				Status:     "черновик",
				DateCreate: time.Now(),
				DateUpdate: time.Now(),
				CreatorID:  userID,
			}
			err = r.db.Create(&newTree).Error
			if err != nil {
				return nil, err
			}
			return &newTree, nil
		}
		return nil, err
	}
	return &tree, nil
}

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

	// Создаем новую запись
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
	err := r.db.First(&tree, treeID).Error
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

func (r *Repository) DeleteTree(treeID uint) error {
	// Прямой SQL UPDATE - без курсора
	result := r.db.Exec("UPDATE trees SET status = 'удалён', date_update = NOW() WHERE id = ?", treeID)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("дерево не найдено")
	}
	return nil
}

func (r *Repository) GetTreeByID(treeID uint) (*ds.Tree, error) {
	var tree ds.Tree
	err := r.db.First(&tree, treeID).Error
	if err != nil {
		return nil, err
	}
	return &tree, nil
}

func (r *Repository) GetDraftTree(userID uint) (*ds.Tree, error) {
	var tree ds.Tree
	err := r.db.Where("creator_id = ? AND status = ?", userID, "черновик").First(&tree).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &tree, nil
}
