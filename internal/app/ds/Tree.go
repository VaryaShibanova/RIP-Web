package ds

import (
	"database/sql"
	"time"
)

type Tree struct {
	ID          uint         `gorm:"primaryKey"`
	Status      string       `gorm:"type:varchar(20);not null;default:'черновик';check:status IN ('черновик', 'удалён', 'сформирован', 'завершён', 'отклонён')"`
	Description string       `gorm:"type:text"`
	TotalRings  int          `gorm:"not null;default:0"`
	FinalYear   int          `gorm:"default:0"`
	DateCreate  time.Time    `gorm:"not null;default:now()"`
	DateUpdate  time.Time    `gorm:"default:now()"`
	DateFinish  sql.NullTime `gorm:"default:null"`
	CreatorID   uint         `gorm:"not null"`
	ModeratorID sql.NullInt64

	Creator   Users `gorm:"foreignKey:CreatorID"`
	Moderator Users `gorm:"foreignKey:ModeratorID"`
	TreeItems []TreeItem
}
