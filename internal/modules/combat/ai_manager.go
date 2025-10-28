// file: internal/modules/combat/ai_manager.go
package combat

import (
	"sage-of-elements-backend/internal/domain"
	"sort"
)

// ==================== AI Manager ====================
// ไฟล์นี้เป็น main orchestrator สำหรับ AI turn
// ทำหน้าที่ประสานงานระหว่าง decision making และ action execution
//
// โครงสร้าง AI System:
// - ai_manager.go: Main orchestrator (ไฟล์นี้)
// - ai_decision.go: Decision making (condition checking, action selection)
// - ai_execution.go: Action execution (resource deduction, effect application)

// processAllAITurns ประมวลผลเทิร์นของ AI ทั้งหมดจนกว่าจะกลับมาเป็นเทิร์นผู้เล่น
// หรือเกมจบลง (ป้องกัน infinite loop โดยจำกัดจำนวนเทิร์นสูงสุด)
func (s *combatService) processAllAITurns(match *domain.CombatMatch) (*domain.CombatMatch, error) {
	const maxConsecutiveAITurns = 20 // ป้องกัน infinite loop
	turnsProcessed := 0

	for turnsProcessed < maxConsecutiveAITurns {
		// ตรวจสอบว่าเกมจบแล้วหรือยัง
		if match.Status != domain.MatchInProgress {
			s.appLogger.Info("Match ended during AI turns",
				"status", match.Status,
				"turns_processed", turnsProcessed,
			)
			return match, nil
		}

		// ดูว่าเทิร์นปัจจุบันเป็นของ AI หรือไม่
		currentCombatant := s.findCombatantByID(match, match.CurrentTurn)
		if currentCombatant == nil {
			s.appLogger.Error("Current combatant not found",
				nil,
				"current_turn_id", match.CurrentTurn,
			)
			break
		}

		// ถ้าเป็นเทิร์นผู้เล่น (CharacterID != nil) ให้หยุด loop
		if currentCombatant.CharacterID != nil {
			s.appLogger.Debug("Player turn reached, stopping AI loop",
				"turns_processed", turnsProcessed,
			)
			break
		}

		// ถ้าเป็น AI ให้ประมวลผลเทิร์น
		if currentCombatant.EnemyID != nil {
			s.appLogger.Info("Processing AI turn",
				"ai_id", currentCombatant.ID,
				"turn_count", turnsProcessed+1,
			)

			var err error
			match, err = s.processAITurn(match, currentCombatant)
			if err != nil {
				return nil, err
			}

			// ตรวจสอบจบเกมหลังจาก AI เล่นเสร็จ
			match = s.checkMatchEndCondition(match)

			turnsProcessed++
		}
	}

	if turnsProcessed >= maxConsecutiveAITurns {
		s.appLogger.Warn("Max consecutive AI turns reached",
			"max_turns", maxConsecutiveAITurns,
		)
	}

	return match, nil
}

// processAITurn คือฟังก์ชันหลักในการประมวลผลเทิร์นของ AI **1 เทิร์น**
// จะทำงานต่อเนื่องจนกว่า AP จะหมด หรือไม่มี action ที่ทำได้
func (s *combatService) processAITurn(match *domain.CombatMatch, aiCombatant *domain.Combatant) (*domain.CombatMatch, error) {
	s.appLogger.Info("🤖 AI turn started",
		"ai_id", aiCombatant.ID,
		"match_id", match.ID,
		"turn_number", match.TurnNumber,
	)

	// 1. Validate AI combatant data
	if !s._ValidateAICombatant(aiCombatant) {
		s.appLogger.Warn("Invalid AI combatant, skipping turn",
			"ai_id", aiCombatant.ID,
		)
		return s._EndAITurn(match)
	}

	// 2. Prepare AI decision context
	ctx := s._PrepareDecisionContext(match, aiCombatant)
	if ctx == nil {
		s.appLogger.Error("Failed to prepare AI decision context",
			nil,
			"ai_id", aiCombatant.ID,
		)
		return s._EndAITurn(match)
	}

	// 3. Execute AI action loop
	err := s._ExecuteActionLoop(ctx, aiCombatant)
	if err != nil {
		s.appLogger.Error("AI action loop failed", err,
			"ai_id", aiCombatant.ID,
		)
		return nil, err
	}

	// 4. End AI turn and start next turn
	s.appLogger.Info("🤖 AI turn ended",
		"ai_id", aiCombatant.ID,
		"final_ap", aiCombatant.CurrentAP,
		"final_mp", aiCombatant.CurrentMP,
	)

	return s._EndAITurn(match)
}

// ==================== AI Turn Management ====================

// _ValidateAICombatant ตรวจสอบว่า AI combatant มีข้อมูลครบถ้วนหรือไม่
func (s *combatService) _ValidateAICombatant(aiCombatant *domain.Combatant) bool {
	if aiCombatant.Enemy == nil {
		s.appLogger.Warn("AI combatant has no enemy data",
			"ai_id", aiCombatant.ID,
		)
		return false
	}

	if len(aiCombatant.Enemy.AI) == 0 {
		s.appLogger.Warn("AI combatant has no AI rules",
			"ai_id", aiCombatant.ID,
		)
		return false
	}

	return true
}

// _PrepareDecisionContext สร้าง context สำหรับการตัดสินใจของ AI
func (s *combatService) _PrepareDecisionContext(
	match *domain.CombatMatch,
	aiCombatant *domain.Combatant,
) *AIDecisionContext {
	// หา player target
	playerTarget := s.findPlayerCombatant(match)
	if playerTarget == nil {
		s.appLogger.Error("Player combatant not found in match",
			nil,
			"match_id", match.ID,
		)
		return nil
	}

	// เรียง AI rules ตาม priority (น้อยไปมาก = ทำก่อน)
	aiRules := aiCombatant.Enemy.AI
	sort.Slice(aiRules, func(i, j int) bool {
		return aiRules[i].Priority < aiRules[j].Priority
	})

	return &AIDecisionContext{
		AICombatant:    aiCombatant,
		PlayerTarget:   playerTarget,
		Match:          match,
		AvailableRules: aiRules,
	}
}

// _ExecuteActionLoop วน loop ในการเลือกและ execute actions
func (s *combatService) _ExecuteActionLoop(
	ctx *AIDecisionContext,
	aiCombatant *domain.Combatant,
) error {
	const maxActionsPerTurn = 5
	actionsTaken := 0

	s.appLogger.Debug("Starting AI action loop",
		"ai_id", aiCombatant.ID,
		"initial_ap", aiCombatant.CurrentAP,
		"max_actions", maxActionsPerTurn,
	)

	// Loop จนกว่า AP จะหมด หรือไม่มี action ที่ทำได้
	for aiCombatant.CurrentAP > 0 && actionsTaken < maxActionsPerTurn {
		actionsTaken++

		s.appLogger.Debug("AI action loop iteration",
			"iteration", actionsTaken,
			"current_ap", aiCombatant.CurrentAP,
		)

		// เลือก action ถัดไป
		selectedAction := s.SelectNextAction(ctx)

		// ถ้าไม่เจอ action ที่ทำได้ ให้หยุด loop
		if selectedAction == nil {
			s.appLogger.Info("No valid action found, ending AI turn",
				"ai_id", aiCombatant.ID,
				"ap_remaining", aiCombatant.CurrentAP,
			)
			break
		}

		// Execute action ที่เลือก
		err := s.ExecuteAIAction(aiCombatant, selectedAction)
		if err != nil {
			return err
		}

		// ตรวจสอบว่ายัง loop ต่อได้ไหม
		if aiCombatant.CurrentAP <= 0 {
			s.appLogger.Info("AI AP depleted, ending action loop",
				"ai_id", aiCombatant.ID,
			)
			break
		}
	}

	// ตรวจสอบว่าถึง max actions หรือไม่
	if actionsTaken >= maxActionsPerTurn {
		s.appLogger.Warn("AI reached max actions per turn",
			"ai_id", aiCombatant.ID,
			"max_actions", maxActionsPerTurn,
		)
	}

	s.appLogger.Info("AI action loop completed",
		"ai_id", aiCombatant.ID,
		"actions_taken", actionsTaken,
		"final_ap", aiCombatant.CurrentAP,
	)

	return nil
}

// _EndAITurn จบเทิร์นของ AI และเริ่มเทิร์นถัดไป
func (s *combatService) _EndAITurn(match *domain.CombatMatch) (*domain.CombatMatch, error) {
	s.appLogger.Debug("Ending AI turn", "match_id", match.ID)

	// จบเทิร์นปัจจุบัน
	match = s.endTurn(match)

	// เริ่มเทิร์นถัดไป
	updatedMatch, err := s.startNewTurn(match)
	if err != nil {
		s.appLogger.Error("Failed to start new turn after AI", err,
			"match_id", match.ID,
		)
		return nil, err
	}

	return updatedMatch, nil
}
