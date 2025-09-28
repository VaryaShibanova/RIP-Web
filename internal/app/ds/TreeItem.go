package ds

type TreeItem struct {
	ID             uint   `gorm:"primaryKey"`
	TreeID         uint   `gorm:"not null;uniqueIndex:idx_tree_anomaly"`
	AnomalyID      uint   `gorm:"not null;uniqueIndex:idx_tree_anomaly"`
	AnomalousRings string `gorm:"type:varchar(50)"`
	CalculatedYear int    `gorm:"not null;default:0"`

	Tree    Tree    `gorm:"foreignKey:TreeID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
	Anomaly Anomaly `gorm:"foreignKey:AnomalyID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
}
