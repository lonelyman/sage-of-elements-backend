package domain

import "gorm.io/datatypes"

// Element เก็บข้อมูลธาตุทั้งหมด
type Element struct {
	ID           uint              `gorm:"primaryKey;comment:ID เฉพาะของธาตุ (PK)" json:"id"`
	Name         string            `gorm:"size:255;not null;unique;comment:ชื่อธาตุ (เช่น Magma)" json:"name"`
	DisplayNames datatypes.JSONMap `gorm:"type:jsonb; comment:ชื่อที่แสดงผลในแต่ละภาษา" json:"display_names"`
	Description  datatypes.JSONMap `gorm:"type:jsonb; comment:คำอธิบายในแต่ละภาษา" json:"description"`
	Tier         int               `gorm:"not null;comment:ระดับ Tier ของธาตุ (0, 1, 2...)" json:"tier"`
}
