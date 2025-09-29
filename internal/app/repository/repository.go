package repository

import (
	"RIP-WEB/internal/app/ds"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func New(dsn string) (*Repository, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	return &Repository{
		db: db,
	}, nil
}

// Добавляем метод для проверки существования дерева с элементами
func (r *Repository) TreeExistsWithItems(treeID uint) (bool, error) {
	var count int64
	err := r.db.Model(&ds.TreeItem{}).Where("tree_id = ?", treeID).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// УДАЛЯЕМ дублированный метод GetTreeWithItems отсюда - он уже есть в tree.go
