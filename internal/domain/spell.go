package domain

import "gorm.io/datatypes"

// Spell เก็บรายชื่อเวทมนตร์ทั้งหมด
type Spell struct {
	ID           uint              `gorm:"primaryKey"`
	Name         string            `gorm:"size:255;not null;unique;comment:ชื่อในระบบ (Eng)"`
	DisplayNames datatypes.JSONMap `gorm:"type:jsonb;comment:ชื่อที่แสดงผลในแต่ละภาษา"`
	Descriptions datatypes.JSONMap `gorm:"type:jsonb;comment:คำอธิบายในแต่ละภาษา"`
	ElementID    uint              `gorm:"not null"`
	MasteryID    uint              `gorm:"not null"`
	APCost       int               `gorm:"not null"`
	MPCost       int               `gorm:"not null"`
	Element      *Element          `gorm:"foreignKey:ElementID;references:ID"`
	Mastery      *Mastery          `gorm:"foreignKey:MasteryID;references:ID"`
	Effects      []*SpellEffect    `gorm:"foreignKey:SpellID"`
}
