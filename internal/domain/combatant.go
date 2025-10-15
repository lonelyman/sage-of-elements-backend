// file: internal/domain/combatant.go
package domain

import (
	"github.com/gofrs/uuid"
	"gorm.io/datatypes"
)

// Combatant แทนตัวละคร/ศัตรู 1 ตัวที่อยู่ในสนามรบ
type Combatant struct {
	ID      uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	MatchID uuid.UUID `gorm:"type:uuid;not null;index" json:"-"`

	// --- ระบุว่าเป็น "ใคร" ---
	CharacterID *uint      `gorm:"comment:ID ของตัวละคร (ถ้าเป็นผู้เล่น)" json:"characterId,omitempty"`
	Character   *Character `gorm:"foreignKey:CharacterID" json:"character,omitempty"`
	EnemyID     *uint      `gorm:"comment:ID ของศัตรู (ถ้าเป็น AI)" json:"-"`
	Enemy       *Enemy     `gorm:"foreignKey:EnemyID" json:"enemy,omitempty"`

	// --- Stats ชั่วคราวใน Match นี้ ---
	Initiative int `gorm:"not null" json:"initiative"`
	CurrentHP  int `gorm:"not null" json:"currentHp"`
	CurrentMP  int `gorm:"not null" json:"currentMp"`
	CurrentAP  int `gorm:"not null" json:"currentAp"`

	Hand          datatypes.JSON `gorm:"type:jsonb" json:"hand"`
	ActiveEffects datatypes.JSON `gorm:"type:jsonb" json:"activeEffects"`
}

// type ActiveEffect struct {
// 	EffectID       uint      `json:"effectId"`
// 	Value          int       `json:"value"`
// 	TurnsRemaining int       `json:"turnsRemaining"`
// 	SourceID       uuid.UUID `json:"sourceId"`
// }
