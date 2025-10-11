package domain

import "gorm.io/datatypes"

// DimensionalSealInventory เก็บธาตุทั้งหมดในคลังของผู้เล่น
type DimensionalSealInventory struct {
	ID          uint              `gorm:"primaryKey;comment:ID ของไอเทมชิ้นนี้ในคลัง (PK)" json:"id"`
	CharacterID uint              `gorm:"not null;comment:ID ของตัวละครที่เป็นเจ้าของ" json:"-"`
	ElementID   uint              `gorm:"not null;comment:ID ของธาตุที่เก็บไว้ (FK to elements)" json:"element_id"`
	Quantity    int               `gorm:"not null;default:1;comment:จำนวนที่มี" json:"quantity"`
	ItemType    ItemType          `gorm:"size:50;not null;default:'NORMAL';comment:ประเภทไอเทม (NORMAL, UNSTABLE, CHARGED)"`
	Metadata    datatypes.JSONMap `gorm:"type:jsonb;comment:ข้อมูลพิเศษสำหรับไอเทม (เช่น จำนวนชาร์จ)"`
	Element     *Element          `gorm:"foreignKey:ElementID" json:"element"`
}

type ItemType string

const (
	ItemTypeNormal   ItemType = "NORMAL"
	ItemTypeCharged  ItemType = "CHARGED"
	ItemTypeUnstable ItemType = "UNSTABLE"
)
