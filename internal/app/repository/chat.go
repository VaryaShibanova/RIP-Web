package repository

import (
	"RIP-WEB/internal/app/ds"
	"database/sql"
	"errors"
	"fmt"

	"github.com/sirupsen/logrus"
)

func (r *Repository) GetAllChats() ([]ds.Chat, error) {
	// тут мы пользуемся ORM
	var chats []ds.Chat
	err := r.db.Find(&chats).Error
	if err != nil {
		return nil, err
	}
	return chats, nil
}

func (r *Repository) GetChatByID(id int) (*ds.Chat, error) {
	// тут мы пользуемся "курсором" для примера
	query := "SELECT id, img, name, info, nickname, friends, subscribers FROM chats WHERE id = $1"

	// Создание курсора (строковый указатель)
	row := r.db.Raw(query, id).Row()

	// Создание объекта для хранения данных
	chat := &ds.Chat{}

	// Сканирование строки в структуру
	err := row.Scan(
		&chat.ID,
		&chat.Img,
		&chat.Name,
		&chat.Info,
		&chat.Nickname,
		&chat.Friends,
		&chat.Subscribers,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // Возвращаем nil, если записи нет
		}
		return nil, err
	}

	return chat, nil
}

func (r *Repository) SearchChatsByName(name string) ([]ds.Chat, error) {
	var chats []ds.Chat
	err := r.db.Where("name ILIKE ?", "%"+name+"%").Find(&chats).Error
	if err != nil {
		return nil, err
	}
	return chats, nil
}

// GetCartCount для получения количества услуг в заявке (чатов в сообщении в моем случае)
func (r *Repository) GetCartCount() int64 {
	var messageID uint
	var count int64
	creatorID := 1
	// пока что мы захардкодили id создателя заявки, в последующем вы сделаете авторизацию и будете получать его из JWT

	err := r.db.Model(&ds.Message{}).Where("creator_id = ? AND status = ?", creatorID, "черновик").Select("id").First(&messageID).Error
	if err != nil {
		return 0
	}

	err = r.db.Model(&ds.MessageChat{}).Where("message_id = ?", messageID).Count(&count).Error
	if err != nil {
		logrus.Println("Error counting records in lists_chats:", err)
	}

	return count
}

func (r *Repository) DeleteChat(chatID uint) error {
	err := r.db.Model(&ds.Chat{}).Where("id = ?", chatID).UpdateColumn("is_delete", true).Error
	fmt.Println(chatID)
	if err != nil {
		return fmt.Errorf("ошибка при удалении чата с id %d: %w", chatID, err)
	}

	return nil
}
