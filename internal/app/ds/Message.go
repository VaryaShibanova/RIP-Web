package ds

import (
	"database/sql"
	"time"
)

type Message struct {
	ID          uint           `gorm:"primaryKey"`
	Status      string         `gorm:"type:varchar(15);not null"`
	Text        sql.NullString `gorm:"type:text;default:null"`
	DateCreate  time.Time      `gorm:"not null"`
	DateUpdate  time.Time
	DateFinish  sql.NullTime `gorm:"default:null"`
	CreatorID   uint         `gorm:"not null"`
	ModeratorID uint

	Creator   Users `gorm:"foreignKey:CreatorID"`
	Moderator Users `gorm:"foreignKey:ModeratorID"`
}
