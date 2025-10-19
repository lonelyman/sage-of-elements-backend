// file: internal/modules/combat/helper.go
package combat

import (
	"sage-of-elements-backend/internal/domain"

	"github.com/gofrs/uuid"
)

// --- นี่คือบ้านใหม่ของผู้ช่วย ---

func (s *combatService) findCombatantByID(match *domain.CombatMatch, id uuid.UUID) *domain.Combatant {
	for _, c := range match.Combatants {
		if c.ID == id {
			return c
		}
	}
	return nil
}

func (s *combatService) findPlayerCombatant(match *domain.CombatMatch) *domain.Combatant {
	for _, c := range match.Combatants {
		if c.CharacterID != nil {
			return c
		}
	}
	return nil
}
