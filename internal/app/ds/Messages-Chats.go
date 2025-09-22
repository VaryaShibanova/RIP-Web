package ds

type MessageChat struct {
	ID uint `gorm:"primaryKey"`
	// здесь создаем Unique key, указывая общий uniqueIndex
	MessageID uint `gorm:"not null;uniqueIndex:idx_message_chat"`
	ChatID    uint `gorm:"not null;uniqueIndex:idx_message_chat"`

	Sound bool `gorm:"default:true"`

	Message Message `gorm:"foreignKey:MessageID"`
	Chat    Chat    `gorm:"foreignKey:ChatID"`
}
