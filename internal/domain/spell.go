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
	TargetType   TargetType        `gorm:"size:50;not null;default:'ENEMY';comment:ประเภทเป้าหมาย (SELF, ENEMY, ALLY, etc.)"`
	Element      *Element          `gorm:"foreignKey:ElementID;references:ID"`
	Mastery      *Mastery          `gorm:"foreignKey:MasteryID;references:ID"`
	Effects      []*SpellEffect    `gorm:"foreignKey:SpellID"`
}

// TargetType defines the possible targeting modes for spells and abilities.
type TargetType string

const (
	TargetTypeSelf       TargetType = "SELF"        // Targets the caster : "ตัวเอง"
	TargetTypeEnemy      TargetType = "ENEMY"       // Targets a single enemy : "ศัตรู"
	TargetTypeAlly       TargetType = "ALLY"        // Targets a single ally (for future team battles) : "พันธมิตร"
	TargetTypeEnemyAOE   TargetType = "ENEMY_AOE"   // Targets multiple enemies (Area of Effect) : "ศัตรูหมู่"
	TargetTypeAllyAOE    TargetType = "ALLY_AOE"    // Targets multiple allies (Area of Effect) : "พันธมิตรหมู่"
	TargetTypeAllEnemies TargetType = "ALL_ENEMIES" // Targets all enemies on the field : "ศัตรูทั้งหมด"
	TargetTypeAllAllies  TargetType = "ALL_ALLIES"  // Targets all allies on the field : "พันธมิตรทั้งหมด"
)
