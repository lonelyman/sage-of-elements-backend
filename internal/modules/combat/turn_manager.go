// file: internal/modules/combat/turn_manager.go
package combat

import (
	"errors"
	"math"
	"sage-of-elements-backend/internal/domain"
	"sage-of-elements-backend/pkg/apperrors"
	"strconv"
	"time"
)

// --- นี่คือบ้านใหม่ของผู้เชี่ยวชาญด้านเทิร์น ---

func (s *combatService) endTurn(match *domain.CombatMatch) *domain.CombatMatch {
	// --- หา Index ของเทิร์นปัจจุบัน และ เทิร์นถัดไป ---
	var currentTurnIndex int = -1 // Initialize with -1 to detect if not found
	for i, c := range match.Combatants {
		if c.ID == match.CurrentTurn {
			currentTurnIndex = i
			break
		}
	}
	// เพิ่มการตรวจสอบเผื่อหา Index ไม่เจอ (ไม่ควรเกิด แต่ปลอดภัยไว้ก่อน)
	if currentTurnIndex == -1 {
		s.appLogger.Error("Could not find current turn index in endTurn", errors.New("current turn index not found"), match.ID, "current_turn_id", match.CurrentTurn)
		// อาจจะ return match เดิม หรือ panic? ตอนนี้แค่ Log ไว้ก่อน
		return match // Return original match to avoid further issues
	}

	// เก็บ ID ของคนที่เพิ่งจบเทิร์นไว้ก่อนเปลี่ยน
	endedTurnCombatantID := match.CurrentTurn

	nextTurnIndex := (currentTurnIndex + 1) % len(match.Combatants)
	nextCombatant := match.Combatants[nextTurnIndex] // หา Combatant ของเทิร์นถัดไป

	// --- อัปเดต Match State ---
	match.CurrentTurn = nextCombatant.ID // เปลี่ยน CurrentTurn เป็น ID ของคนถัดไป
	if nextTurnIndex == 0 {              // ถ้าวนกลับมาที่คนแรก (Index 0)
		match.TurnNumber++ // ให้เพิ่มเลขรอบ (Round Number)
		s.appLogger.Info("New Round Started", "round_number", match.TurnNumber)
	}

	// Log โดยใช้ ID ที่เก็บไว้ และ ID ของคนถัดไป
	s.appLogger.Info("Ending turn", "ended_by_id", endedTurnCombatantID, "next_turn_id", nextCombatant.ID)
	return match
}
func (s *combatService) startNewTurn(match *domain.CombatMatch) (*domain.CombatMatch, error) {
	currentCombatant := s.findCombatantByID(match, match.CurrentTurn)
	if currentCombatant == nil {
		return nil, apperrors.SystemError("failed to find current combatant")
	}

	s.appLogger.Info("Processing effect ticks and expiry at start of turn", "combatant_id", currentCombatant.ID)
	s.processEffectTicksAndExpiry(currentCombatant)

	s.recalculateStats(currentCombatant)

	apPerTurnStr, _ := s.gameDataRepo.GetGameConfigValue("COMBAT_AP_PER_TURN")
	apPerTurn, _ := strconv.Atoi(apPerTurnStr)

	baseApCapStr, _ := s.gameDataRepo.GetGameConfigValue("COMBAT_BASE_AP_CAP")
	baseApCap, _ := strconv.Atoi(baseApCapStr)
	apCapBonus := 0
	currentMaxAP := baseApCap + apCapBonus

	currentCombatant.CurrentAP += apPerTurn

	if currentCombatant.CurrentAP > currentMaxAP {
		currentCombatant.CurrentAP = currentMaxAP
	}

	// 1. เช็คว่าเป็น Player หรือไม่
	if currentCombatant.CharacterID != nil && currentCombatant.Character != nil {

		// --- 2. Logic เด้ง MP (15%) ---
		maxMP := currentCombatant.Character.CurrentMP
		mpRegenAmount := int(math.Round(float64(maxMP) * 0.02))

		currentCombatant.CurrentMP += mpRegenAmount
		if currentCombatant.CurrentMP > maxMP {
			currentCombatant.CurrentMP = maxMP
		}
		s.appLogger.Info("Player MP Regen", "combatant_id", currentCombatant.ID, "regen_amount", mpRegenAmount, "new_mp", currentCombatant.CurrentMP)

		// --- 3. Logic เด้ง HP (10%) [เพิ่มใหม่!] ---
		// maxHP := currentCombatant.Character.CurrentHP
		// hpRegenAmount := int(math.Round(float64(maxHP) * 0.10)) // 10%

		// currentCombatant.CurrentHP += hpRegenAmount
		// if currentCombatant.CurrentHP > maxHP {
		// 	currentCombatant.CurrentHP = maxHP
		// }
		// s.appLogger.Info("Player HP Regen", "combatant_id", currentCombatant.ID, "regen_amount", hpRegenAmount, "new_hp", currentCombatant.CurrentHP)
	}
	// --- ⭐️ สิ้นสุดส่วนที่แก้ไข ⭐️ ---
	s.appLogger.Info("New turn started", "combatant_id", currentCombatant.ID, "new_ap", currentCombatant.CurrentAP)
	return match, nil
}

func (s *combatService) checkMatchEndCondition(match *domain.CombatMatch) *domain.CombatMatch {
	var playerTeam []*domain.Combatant
	var enemyTeam []*domain.Combatant
	for _, c := range match.Combatants {
		if c.CharacterID != nil {
			playerTeam = append(playerTeam, c)
		} else if c.EnemyID != nil {
			enemyTeam = append(enemyTeam, c)
		}
	}

	playerTeamDefeated := true
	for _, p := range playerTeam {
		if p.CurrentHP > 0 {
			playerTeamDefeated = false
			break
		}
	}

	enemyTeamDefeated := true
	for _, e := range enemyTeam {
		if e.CurrentHP > 0 {
			enemyTeamDefeated = false
			break
		}
	}

	now := time.Now()
	if playerTeamDefeated {
		s.appLogger.Info("Match ended: Player team defeated.", "match_id", match.ID)
		match.Status = domain.MatchFinished
		match.FinishedAt = &now
	} else if enemyTeamDefeated {
		s.appLogger.Info("Match ended: Enemy team defeated. Player wins!", "match_id", match.ID)
		match.Status = domain.MatchFinished
		match.FinishedAt = &now
	}

	return match
}
