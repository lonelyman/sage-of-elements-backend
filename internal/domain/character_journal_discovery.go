package domain

import "time"

// CharacterJournalDiscovery บันทึกสูตรที่ผู้เล่นค้นพบแล้ว
type CharacterJournalDiscovery struct {
	CharacterID  uint       `gorm:"primaryKey;autoIncrement:false;comment:ID ของตัวละคร (Composite PK)" json:"-"`
	RecipeID     uint       `gorm:"primaryKey;autoIncrement:false;comment:ID ของสูตรที่ค้นพบ (Composite PK)" json:"recipe_id"`
	DiscoveredAt time.Time  `gorm:"comment:วันเวลาที่ค้นพบ" json:"discovered_at"`
	Character    *Character `gorm:"foreignKey:CharacterID"`
	Recipe       *Recipe    `gorm:"foreignKey:RecipeID" json:"recipe"`
}
