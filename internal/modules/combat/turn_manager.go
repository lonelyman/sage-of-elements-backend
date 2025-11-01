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

// ==================== Turn Manager ====================
// ไฟล์นี้จัดการ turn-based combat flow:
// - การหมุนเทิร์น (turn rotation)
// - การเริ่มเทิร์นใหม่ (resource regeneration, effect processing)
// - การตรวจสอบเงื่อนไขจบเกม (win/lose conditions)

// ==================== Turn Rotation ====================

// endTurn จบเทิร์นปัจจุบันและเลื่อนไปยังเทิร์นถัดไป
func (s *combatService) endTurn(match *domain.CombatMatch) *domain.CombatMatch {
	s.appLogger.Debug("🔄 Ending turn",
		"match_id", match.ID,
		"current_turn", match.CurrentTurn,
		"turn_number", match.TurnNumber,
	)

	// หา combatant ปัจจุบันและถัดไป
	currentIndex := -1
	for i, c := range match.Combatants {
		if c.ID == match.CurrentTurn {
			currentIndex = i
			break
		}
	}

	if currentIndex == -1 {
		s.appLogger.Error("Could not find current turn index",
			errors.New("current turn index not found"),
			"match_id", match.ID,
			"current_turn_id", match.CurrentTurn,
		)
		return match // Return original match to prevent further issues
	}

	// เก็บ ID ของคนที่เพิ่งจบเทิร์น
	endedTurnCombatantID := match.CurrentTurn

	// คำนวณ index ถัดไป (wrap around)
	nextIndex := (currentIndex + 1) % len(match.Combatants)
	nextCombatant := match.Combatants[nextIndex]

	// อัปเดต match state
	match.CurrentTurn = nextCombatant.ID

	// เพิ่มรอบถ้าวนกลับมาคนแรก
	if nextIndex == 0 {
		match.TurnNumber++
		s.appLogger.Info("🔄 New round started",
			"match_id", match.ID,
			"round_number", match.TurnNumber,
		)
	}

	s.appLogger.Info("✅ Turn ended",
		"ended_by", endedTurnCombatantID,
		"next_turn", nextCombatant.ID,
		"turn_number", match.TurnNumber,
	)

	return match
}

// ==================== Turn Initialization ====================

// startNewTurn เริ่มเทิร์นใหม่และประมวลผลทุกอย่างที่ต้องทำต้นเทิร์น
func (s *combatService) startNewTurn(match *domain.CombatMatch) (*domain.CombatMatch, error) {
	// หา combatant ที่เป็นเทิร์นปัจจุบัน
	currentCombatant := s.findCombatantByID(match, match.CurrentTurn)
	if currentCombatant == nil {
		return nil, apperrors.SystemError("failed to find current combatant")
	}

	s.appLogger.Info("🎮 Starting new turn",
		"combatant_id", currentCombatant.ID,
		"turn_number", match.TurnNumber,
	)

	// 1. ประมวลผล effects (ticks และ expiry)
	s._ProcessTurnEffects(currentCombatant)

	// 2. คำนวณ stats ใหม่
	s.recalculateStats(currentCombatant)

	// 3. เพิ่ม AP
	s._RegenerateAP(currentCombatant)

	// 4. เพิ่ม MP (สำหรับ player เท่านั้น)
	if currentCombatant.CharacterID != nil && currentCombatant.Character != nil {
		s._RegeneratePlayerMP(currentCombatant)
	}

	s.appLogger.Info("✅ New turn ready",
		"combatant_id", currentCombatant.ID,
		"ap", currentCombatant.CurrentAP,
		"mp", currentCombatant.CurrentMP,
	)

	return match, nil
}

// ==================== Turn Start Helpers ====================

// _ProcessTurnEffects ประมวลผล active effects ต้นเทิร์น
func (s *combatService) _ProcessTurnEffects(combatant *domain.Combatant) {
	s.appLogger.Debug("⚡ Processing effects",
		"combatant_id", combatant.ID,
	)
	s.processEffectTicksAndExpiry(combatant)
}

// _RegenerateAP เพิ่ม AP ต้นเทิร์น (มี cap)
func (s *combatService) _RegenerateAP(combatant *domain.Combatant) {
	// โหลด config
	apPerTurn := s._GetAPPerTurn()
	maxAP := s._GetMaxAP()

	// เพิ่ม AP
	combatant.CurrentAP += apPerTurn

	// Cap ที่ max
	if combatant.CurrentAP > maxAP {
		combatant.CurrentAP = maxAP
	}

	s.appLogger.Debug("⚡ AP regenerated",
		"combatant_id", combatant.ID,
		"ap_gained", apPerTurn,
		"current_ap", combatant.CurrentAP,
		"max_ap", maxAP,
	)
}

// _RegeneratePlayerMP เพิ่ม MP ต้นเทิร์น (player เท่านั้น)
func (s *combatService) _RegeneratePlayerMP(combatant *domain.Combatant) {
	// โหลด config
	regenPercent := s._GetMPRegenPercent()

	// คำนวณ MP ที่จะเพิ่ม
	maxMP := combatant.Character.CurrentMP
	mpRegenAmount := int(math.Round(float64(maxMP) * regenPercent))

	// เพิ่ม MP
	combatant.CurrentMP += mpRegenAmount

	// Cap ที่ max
	if combatant.CurrentMP > maxMP {
		combatant.CurrentMP = maxMP
	}

	s.appLogger.Debug("⚡ MP regenerated",
		"combatant_id", combatant.ID,
		"mp_gained", mpRegenAmount,
		"current_mp", combatant.CurrentMP,
		"max_mp", maxMP,
		"regen_percent", regenPercent,
	)
}

// ==================== Config Helpers ====================

// _GetAPPerTurn ดึงค่า AP ที่จะได้รับต่อเทิร์น
func (s *combatService) _GetAPPerTurn() int {
	valueStr, _ := s.gameDataRepo.GetGameConfigValue("COMBAT_AP_PER_TURN")
	value, _ := strconv.Atoi(valueStr)
	if value <= 0 {
		return 3 // Default fallback
	}
	return value
}

// _GetMaxAP ดึงค่า AP สูงสุด
func (s *combatService) _GetMaxAP() int {
	valueStr, _ := s.gameDataRepo.GetGameConfigValue("COMBAT_BASE_AP_CAP")
	value, _ := strconv.Atoi(valueStr)
	if value <= 0 {
		return 12 // Default fallback
	}
	// TODO: เพิ่ม bonus จากอุปกรณ์/buff
	return value
}

// _GetMPRegenPercent ดึงค่าเปอร์เซ็นต์ MP ที่จะ regen ต่อเทิร์น
func (s *combatService) _GetMPRegenPercent() float64 {
	valueStr, _ := s.gameDataRepo.GetGameConfigValue("COMBAT_MP_REGEN_PERCENT")
	value, _ := strconv.ParseFloat(valueStr, 64)
	if value <= 0 {
		return 0.02 // Default 2%
	}
	return value
}

// ==================== Match End Condition ====================

// checkMatchEndCondition ตรวจสอบว่าเกมจบหรือยัง (ทีมใดทีมหนึ่งแพ้)
func (s *combatService) checkMatchEndCondition(match *domain.CombatMatch) *domain.CombatMatch {
	s.appLogger.Debug("🔍 Checking match end condition",
		"match_id", match.ID,
	)

	// แบ่งทีม
	playerTeam, enemyTeam := s._SeparateTeams(match)

	// เช็คว่าทีมไหนแพ้
	playerDefeated := s._IsTeamDefeated(playerTeam)
	enemyDefeated := s._IsTeamDefeated(enemyTeam)

	// ถ้ามีทีมแพ้ ให้จบเกม
	if playerDefeated || enemyDefeated {
		s._EndMatch(match, playerDefeated, enemyDefeated)
	}

	return match
}

// ==================== Match End Helpers ====================

// _SeparateTeams แยก combatants เป็น player team และ enemy team
func (s *combatService) _SeparateTeams(match *domain.CombatMatch) ([]*domain.Combatant, []*domain.Combatant) {
	playerTeam := s.findPlayerCombatants(match)
	enemyTeam := s.findEnemyCombatants(match)

	s.appLogger.Debug("Teams separated",
		"player_count", len(playerTeam),
		"enemy_count", len(enemyTeam),
	)

	return playerTeam, enemyTeam
}

// _IsTeamDefeated เช็คว่าทีมนี้แพ้หรือยัง (HP ของทุกคนเป็น 0)
func (s *combatService) _IsTeamDefeated(team []*domain.Combatant) bool {
	// สร้าง temporary match เพื่อใช้ findAliveCombatants
	tempMatch := &domain.CombatMatch{Combatants: team}
	aliveMembers := s.findAliveCombatants(tempMatch)
	return len(aliveMembers) == 0 // ถ้าไม่มีคนเหลือ = แพ้
}

// _EndMatch จบเกมและบันทึกผลลัพธ์
func (s *combatService) _EndMatch(match *domain.CombatMatch, playerDefeated bool, enemyDefeated bool) {
	now := time.Now()
	match.Status = domain.MatchFinished
	match.FinishedAt = &now

	if playerDefeated {
		s.appLogger.Info("💀 Match ended: Player defeated",
			"match_id", match.ID,
		)
	} else if enemyDefeated {
		s.appLogger.Info("🎉 Match ended: Player wins!",
			"match_id", match.ID,
		)

		// ⭐️ เพิ่ม EXP ให้ผู้เล่นเมื่อชนะ
		// หา player combatant จาก combatants list
		for _, combatant := range match.Combatants {
			if combatant.CharacterID != nil {
				// นี่คือ player combatant
				expAmount := s._CalculateExpReward(match.MatchType)
				if expAmount > 0 {
					characterID := *combatant.CharacterID

					// อัปเดต EXP โดยตรงผ่าน repo
					character, err := s.characterRepo.FindByID(characterID)
					if err == nil && character != nil {
						character.Exp += expAmount
						_, err = s.characterRepo.Save(character)
						if err != nil {
							s.appLogger.Error("failed to grant exp after victory", err,
								"character_id", characterID,
								"exp_amount", expAmount,
							)
						} else {
							s.appLogger.Info("granted exp after victory",
								"character_id", characterID,
								"exp_amount", expAmount,
								"new_total", character.Exp,
							)
						}
					}
				}
				break
			}
		}
	}
}

// _CalculateExpReward คำนวณ EXP ที่ได้รับตาม Match Type
func (s *combatService) _CalculateExpReward(matchType domain.MatchType) int {
	var configKey string
	switch matchType {
	case domain.MatchTypeTraining:
		configKey = "EXP_TRAINING_MATCH"
	case domain.MatchTypeStory:
		configKey = "EXP_STORY_MATCH"
	case domain.MatchTypePVP:
		configKey = "EXP_PVP_MATCH"
	default:
		return 0
	}

	expStr, err := s.gameDataRepo.GetGameConfigValue(configKey)
	if err != nil {
		s.appLogger.Error("failed to get exp config", err, "config_key", configKey)
		return 0
	}

	exp, err := strconv.Atoi(expStr)
	if err != nil {
		s.appLogger.Error("invalid exp config value", err, "config_key", configKey, "value", expStr)
		return 0
	}

	return exp
}
