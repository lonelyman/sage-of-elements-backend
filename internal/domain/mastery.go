package domain

import "gorm.io/datatypes"

// Mastery คือตาราง Master Data ที่เก็บชื่อศาสตร์ทั้ง 4
type Mastery struct {
	ID           uint              `gorm:"primaryKey;comment:ID เฉพาะของศาสตร์ (PK)" json:"id"`
	Name         string            `gorm:"size:255;not null;unique;comment:ชื่อในระบบ (Eng)" json:"name"`
	DisplayNames datatypes.JSONMap `gorm:"type:jsonb;comment:ชื่อที่แสดงผลในแต่ละภาษา" json:"displayNames"`
	Descriptions datatypes.JSONMap `gorm:"type:jsonb;comment:คำอธิบายในแต่ละภาษา" json:"descriptions"`
}
