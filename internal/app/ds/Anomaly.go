package ds

type Anomaly struct {
	ID          uint   `gorm:"primaryKey"`
	IsDelete    bool   `gorm:"type:boolean;not null;default:false"`
	Image       string `gorm:"type:varchar(200)"`
	Name        string `gorm:"type:varchar(100);not null"`
	Description string `gorm:"type:text"`
	Year        int    `gorm:"not null"`
	Pattern     string `gorm:"type:varchar(200);not null"`
}
