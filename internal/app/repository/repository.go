package repository

import (
	"RIP-WEB/internal/app/ds"
	"errors"
	"time"

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

// GetDB возвращает экземпляр базы данных
func (r *Repository) GetDB() *gorm.DB {
	return r.db
}

// GetUserByLogin возвращает пользователя по логину
func (r *Repository) GetUserByLogin(login string) (*ds.Users, error) {
	var user ds.Users
	err := r.db.Where("login = ?", login).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// GetUserByID возвращает пользователя по ID
func (r *Repository) GetUserByID(id uint) (*ds.Users, error) {
	var user ds.Users
	err := r.db.First(&user, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// CreateUser создает нового пользователя
func (r *Repository) CreateUser(user *ds.Users) error {
	return r.db.Create(user).Error
}

// GetSystemUser возвращает системного пользователя (для обратной совместимости)
func (r *Repository) GetSystemUser() *ds.Users {
	return &ds.Users{
		ID:          1,
		Login:       "research_user",
		Password:    "password123",
		IsModerator: false,
	}
}

// GetModerator возвращает модератора (для обратной совместимости)
func (r *Repository) GetModerator() *ds.Users {
	return &ds.Users{
		ID:          2,
		Login:       "moderator_user",
		Password:    "password123",
		IsModerator: true,
	}
}

// Получить черновую заявку текущего пользователя
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

// Создать или получить черновую заявку
func (r *Repository) GetOrCreateDraftTree(userID uint) (*ds.Tree, error) {
	tree, err := r.GetDraftTree(userID)
	if err != nil {
		return nil, err
	}

	if tree == nil {
		// Создаем новую заявку
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

	return tree, nil
}

// Получить количество элементов в корзине
func (r *Repository) GetCartCount(userID uint) int64 {
	tree, err := r.GetDraftTree(userID)
	if err != nil || tree == nil {
		return 0
	}

	var count int64
	r.db.Model(&ds.TreeItem{}).Where("tree_id = ?", tree.ID).Count(&count)
	return count
}

// Проверка существования дерева с элементами
func (r *Repository) TreeExistsWithItems(treeID uint) (bool, error) {
	var count int64
	err := r.db.Model(&ds.TreeItem{}).Where("tree_id = ?", treeID).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
