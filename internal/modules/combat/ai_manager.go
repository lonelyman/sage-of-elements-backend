// file: internal/modules/combat/ai_manager.go
package combat

import (
	"encoding/json"
	"sage-of-elements-backend/internal/domain"
	"sage-of-elements-backend/pkg/apperrors"
	"sort"
)

func (s *combatService) processAITurn(match *domain.CombatMatch, aiCombatant *domain.Combatant) (*domain.CombatMatch, error) {
	// 1. ตรวจสอบข้อมูลศัตรู (Nil check)
	if aiCombatant.Enemy == nil {
		s.appLogger.Warn("AI Combatant has no Enemy data loaded, skipping turn.", "ai_combatant_id", aiCombatant.ID)
		// ถ้าไม่มีข้อมูล AI ก็จบเทิร์นไปเลย
		match = s.endTurn(match)
		match, err := s.startNewTurn(match)
		if err != nil {
			return nil, err
		}
		return match, nil
	}

	// 2. ดึงข้อมูล AI Rules และเรียงลำดับ
	aiRules := aiCombatant.Enemy.AI
	sort.Slice(aiRules, func(i, j int) bool {
		return aiRules[i].Priority < aiRules[j].Priority
	})

	// 3. หา "เป้าหมาย" (ตอนนี้มีแค่ผู้เล่นคนเดียว)
	target := s.findPlayerCombatant(match)
	if target == nil {
		return nil, apperrors.SystemError("player target not found in match")
	}

	// 4. วน Loop เช็ค "กฎ" แต่ละข้อ เพื่อหา Action ที่จะทำ
	var actionToPerform *domain.EnemyAbility
	var targetCombatant *domain.Combatant // เป้าหมายของท่านี้

	for _, rule := range aiRules {
		conditionMet := false

		// --- ✨⭐️ "สมอง" ฉบับอัปเกรด! ⭐️✨ ---
		switch rule.Condition {
		case domain.AIConditionAlways:
			conditionMet = true

		case domain.AIConditionTurnIs:
			if float64(match.TurnNumber) == rule.ConditionValue {
				conditionMet = true
			}

		case domain.AIConditionSelfHPBelow:
			// 1. คำนวณ % HP ปัจจุบัน
			if aiCombatant.Enemy.MaxHP == 0 {
				break // ป้องกันการหารด้วย 0
			}
			currentHPRatio := float64(aiCombatant.CurrentHP) / float64(aiCombatant.Enemy.MaxHP)

			// 2. เปรียบเทียบกับ "กฎ" (เช่น 0.5 สำหรับ 50%)
			if currentHPRatio <= rule.ConditionValue {
				conditionMet = true
			}

			// TODO: เพิ่ม case สำหรับ Condition อื่นๆ ในอนาคต
		}

		// --- ถ้าเงื่อนไขเป็นจริง... ให้กำหนดเป้าหมาย! ---
		if conditionMet {
			actionToPerform = rule.AbilityToUse

			// ตรวจสอบ "เป้าหมาย" ของท่า!
			switch rule.Target {
			case domain.AITargetPlayer:
				targetCombatant = target
			case domain.AITargetSelf:
				targetCombatant = aiCombatant // ⭐️ เป้าหมายคือ "ตัวเอง"!
			default:
				targetCombatant = target // ถ้าไม่ระบุ ให้โจมตีผู้เล่น
			}

			s.appLogger.Info("AI decided to use ability", "ability_name", actionToPerform.Name, "rule_priority", rule.Priority, "target", rule.Target)
			break // หยุดหาทันที
		}
	}

	// 5. ถ้ามี Action ที่จะทำ... ก็ลงมือเลย!
	if actionToPerform != nil {
		if aiCombatant.CurrentAP < actionToPerform.APCost {
			s.appLogger.Warn("AI has insufficient AP for chosen action, passing turn.", "ability", actionToPerform.Name)
		} else {
			// หัก AP ก่อน
			aiCombatant.CurrentAP -= actionToPerform.APCost

			// สร้าง "เวทจำลอง" ขึ้นมา เพื่อส่งข้อมูลธาตุของผู้โจมตี
			dummySpell := &domain.Spell{
				ElementID: aiCombatant.Enemy.ElementID,
			}

			// ประมวลผล "Effects" ทั้งหมดของท่า
			var effects []map[string]interface{}
			json.Unmarshal(actionToPerform.EffectsJSON, &effects)

			for _, effectData := range effects {
				// ส่ง "เป้าหมาย" ที่ถูกต้องเข้าไป!
				s.applyEffect(aiCombatant, targetCombatant, effectData, dummySpell)
			}
		}
	} else {
		s.appLogger.Info("AI has no action to perform, passing turn.", "ai_combatant_id", aiCombatant.ID)
	}

	// 6. จบเทิร์นของตัวเองโดยอัตโนมัติ!
	s.appLogger.Info("AI ended its turn.", "match_id", match.ID)
	match = s.endTurn(match)

	// 7. เตรียมความพร้อมสำหรับเทิร์นของ "คนถัดไป" (ซึ่งก็คือผู้เล่น)
	match, err := s.startNewTurn(match)
	if err != nil {
		return nil, err
	}

	return match, nil
}
