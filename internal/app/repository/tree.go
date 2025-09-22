package repository

import (
	"RIP-WEB/internal/app/ds"
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
	err := r.db.Preload("TreeItems").Preload("TreeItems.Anomaly").First(&tree, treeID).Error
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
	// Используем RAW SQL для обновления статуса
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
