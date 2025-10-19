// file: internal/modules/combat/turn_manager.go
package combat

import (
	"sage-of-elements-backend/internal/domain"
	"sage-of-elements-backend/pkg/apperrors"
	"strconv"
	"time"
)

// --- นี่คือบ้านใหม่ของผู้เชี่ยวชาญด้านเทิร์น ---

func (s *combatService) endTurn(match *domain.CombatMatch) *domain.CombatMatch {
	currentCombatant := s.findCombatantByID(match, match.CurrentTurn)
	if currentCombatant != nil {
		s.processEndOfTurnEffects(currentCombatant)
	}
	var currentTurnIndex int
	for i, c := range match.Combatants {
		if c.ID == match.CurrentTurn {
			currentTurnIndex = i
			break
		}
	}
	nextTurnIndex := (currentTurnIndex + 1) % len(match.Combatants)
	nextCombatant := match.Combatants[nextTurnIndex]

	match.CurrentTurn = nextCombatant.ID
	if nextTurnIndex == 0 {
		match.TurnNumber++
	}
	return match
}

func (s *combatService) startNewTurn(match *domain.CombatMatch) (*domain.CombatMatch, error) {
	currentCombatant := s.findCombatantByID(match, match.CurrentTurn)
	if currentCombatant == nil {
		return nil, apperrors.SystemError("failed to find current combatant")
	}

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
