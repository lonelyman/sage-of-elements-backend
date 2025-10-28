// file: internal/modules/combat/match_utils.go
package combat

import (
	"sage-of-elements-backend/internal/domain"

	"github.com/gofrs/uuid"
)

// ==================== Match Utilities ====================
// ไฟล์นี้รวบรวม utility functions สำหรับจัดการ combat match
// - การค้นหา combatants
// - การกรอง combatants
// - อื่นๆ ที่เกี่ยวกับ match state

// ==================== Combatant Lookup ====================

// findCombatantByID ค้นหา combatant จาก match ด้วย UUID
// Returns nil ถ้าไม่เจอ
func (s *combatService) findCombatantByID(match *domain.CombatMatch, id uuid.UUID) *domain.Combatant {
	for _, c := range match.Combatants {
		if c.ID == id {
			return c
		}
	}
	return nil
}

// findPlayerCombatant ค้นหา combatant ที่เป็น player (ตัวแรกที่เจอ)
// Returns nil ถ้าไม่เจอ player ในแมตช์
func (s *combatService) findPlayerCombatant(match *domain.CombatMatch) *domain.Combatant {
	for _, c := range match.Combatants {
		if c.CharacterID != nil {
			return c
		}
	}
	return nil
}

// ==================== Combatant Filtering ====================

// findPlayerCombatants ค้นหา combatants ที่เป็น player ทั้งหมด
func (s *combatService) findPlayerCombatants(match *domain.CombatMatch) []*domain.Combatant {
	var players []*domain.Combatant
	for _, c := range match.Combatants {
		if c.CharacterID != nil {
			players = append(players, c)
		}
	}
	return players
}

// findEnemyCombatants ค้นหา combatants ที่เป็น enemy ทั้งหมด
func (s *combatService) findEnemyCombatants(match *domain.CombatMatch) []*domain.Combatant {
	var enemies []*domain.Combatant
	for _, c := range match.Combatants {
		if c.EnemyID != nil {
			enemies = append(enemies, c)
		}
	}
	return enemies
}

// findAliveCombatants ค้นหา combatants ที่ยังมีชีวิต (HP > 0)
func (s *combatService) findAliveCombatants(match *domain.CombatMatch) []*domain.Combatant {
	var alive []*domain.Combatant
	for _, c := range match.Combatants {
		if c.CurrentHP > 0 {
			alive = append(alive, c)
		}
	}
	return alive
}
