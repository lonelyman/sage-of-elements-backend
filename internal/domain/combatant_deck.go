// file: internal/domain/combatant_deck.go
package domain

import "github.com/gofrs/uuid"

// CombatantDeck แทน "กระสุน" T1 หนึ่งนัด ที่ Combatant พกเข้ามาในสนามรบ
type CombatantDeck struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CombatantID uuid.UUID `gorm:"type:uuid;not null;index"`
	ElementID   uint      `gorm:"not null"`
	IsConsumed  bool      `gorm:"not null;default:false"`
}
