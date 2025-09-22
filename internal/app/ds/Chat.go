package ds

type Chat struct {
	ID          int    `gorm:"primaryKey"`
	IsDelete    bool   `gorm:"type:boolean not null;default:false"`
	Img         string `gorm:"type:varchar(100)"`
	Name        string `gorm:"type:varchar(25);not null"`
	Info        string `gorm:"type:varchar(100)"`
	Nickname    string `gorm:"type:varchar(15);not null"`
	Friends     int
	Subscribers int
}
