// file: internal/modules/combat/ai_manager.go
package combat

import (
	"encoding/json"
	"sage-of-elements-backend/internal/domain"
	"sage-of-elements-backend/pkg/apperrors"
	"sort"
)

func (s *combatService) processAITurn(match *domain.CombatMatch, aiCombatant *domain.Combatant) (*domain.CombatMatch, error) {
	// 1. ตรวจสอบข้อมูลศัตรู (Nil check) - (เหมือนเดิม)
	if aiCombatant.Enemy == nil {
		s.appLogger.Warn("AI Combatant has no Enemy data loaded, skipping turn.", "ai_combatant_id", aiCombatant.ID)
		match = s.endTurn(match)
		match, err := s.startNewTurn(match)
		if err != nil {
			return nil, err
		}
		return match, nil
	}

	// 2. ดึงข้อมูล AI Rules และเรียงลำดับ - (เหมือนเดิม)
	aiRules := aiCombatant.Enemy.AI
	sort.Slice(aiRules, func(i, j int) bool {
		return aiRules[i].Priority < aiRules[j].Priority
	})

	// 3. หา "เป้าหมาย" (ผู้เล่น) - (เหมือนเดิม)
	target := s.findPlayerCombatant(match)
	if target == nil {
		return nil, apperrors.SystemError("player target not found in match")
	}

	// --- ⭐️ ณัชชาอัปเกรด: "Action Loop" (สมองใหม่!) ⭐️ ---
	// นี่คือ Loop หลักที่ AI จะ "คิด"
	// มันจะวน Loop ตราบใดที่ยังมี AP เหลือ
	// (เราใส่ maxActionsPerTurn (เช่น 5) ไว้... กันเหนียว เผื่อมี Bug แล้วมันวน Loop ไม่หยุด)
	const maxActionsPerTurn = 5
	actionsTakenThisTurn := 0

	for aiCombatant.CurrentAP > 0 && actionsTakenThisTurn < maxActionsPerTurn {
		actionsTakenThisTurn++
		s.appLogger.Info("AI starts thinking...", "ai_id", aiCombatant.ID, "current_ap", aiCombatant.CurrentAP, "action_loop_count", actionsTakenThisTurn)

		// 4. (ใน Loop) วน Loop เช็ค "กฎ" แต่ละข้อ เพื่อหา Action ที่จะทำ
		var actionToPerform *domain.EnemyAbility
		var targetCombatant *domain.Combatant // เป้าหมายของท่านี้

		for _, rule := range aiRules {
			conditionMet := false

			// --- "สมอง" ส่วนเช็คเงื่อนไข (เหมือนเดิม) ---
			switch rule.Condition {
			case domain.AIConditionAlways:
				conditionMet = true
			case domain.AIConditionTurnIs:
				if float64(match.TurnNumber) == rule.ConditionValue {
					conditionMet = true
				}
			case domain.AIConditionSelfHPBelow:
				if aiCombatant.Enemy.MaxHP == 0 {
					break
				}
				currentHPRatio := float64(aiCombatant.CurrentHP) / float64(aiCombatant.Enemy.MaxHP)
				if currentHPRatio <= rule.ConditionValue {
					conditionMet = true
				}
			}
			// --- สิ้นสุดการเช็คเงื่อนไข ---

			// --- ถ้าเงื่อนไขเป็นจริง... ⭐️ "เช็คทรัพยากร" ก่อน! ⭐️ ---
			if conditionMet {
				tempAction := rule.AbilityToUse
				if tempAction == nil {
					continue // กฎนี้ไม่มีท่าให้ใช้
				}

				// ⭐️ Check 1: AP พอไหม? (สำหรับท่านี้)
				canAffordAP := aiCombatant.CurrentAP >= tempAction.APCost
				// ⭐️ Check 2: MP พอไหม? (สำหรับท่านี้)
				canAffordMP := aiCombatant.CurrentMP >= tempAction.MPCost

				if canAffordAP && canAffordMP {
					// ⭐️ เย้! เจอท่าที่จะใช้แล้ว!
					actionToPerform = tempAction

					// กำหนดเป้าหมาย (เหมือนเดิม)
					switch rule.Target {
					case domain.AITargetPlayer:
						targetCombatant = target
					case domain.AITargetSelf:
						targetCombatant = aiCombatant
					default:
						targetCombatant = target
					}

					s.appLogger.Info("AI found a valid action to perform", "ability_name", actionToPerform.Name, "rule_priority", rule.Priority)
					break // ⭐️ หยุดหา "กฎ" (เพราะเจอกฎที่ Priority สูงสุดที่ทำได้แล้ว)

				} else {
					// ถ้าเงื่อนไขผ่าน... แต่ AP/MP ไม่พอ (สำหรับท่านี้)... ให้ข้ามไปดูกฎ Priority ถัดไป
					s.appLogger.Debug("AI condition met, but cannot afford ability. Checking next rule.", "ability_name", tempAction.Name, "rule_priority", rule.Priority, "need_ap", tempAction.APCost, "need_mp", tempAction.MPCost)
				}
			}
		} // --- สิ้นสุด For Loop ของ "กฎ" (aiRules) ---

		// 5. (ใน Loop) ถ้ามี Action ที่จะทำ... ก็ลงมือเลย!
		if actionToPerform != nil {
			// (เราเช็ค AP/MP ไปแล้วตอนเลือก)
			// หัก AP
			aiCombatant.CurrentAP -= actionToPerform.APCost

			// หัก MP (ถ้าใช้)
			if actionToPerform.MPCost > 0 {
				aiCombatant.CurrentMP -= actionToPerform.MPCost
				if aiCombatant.CurrentMP < 0 {
					aiCombatant.CurrentMP = 0
				}
			}
			s.appLogger.Info("AI EXECUTING action", "ability_name", actionToPerform.Name, "new_ap", aiCombatant.CurrentAP, "new_mp", aiCombatant.CurrentMP)

			// สร้าง "เวทจำลอง" (เหมือนเดิม)
			dummySpell := &domain.Spell{
				ElementID: aiCombatant.Enemy.ElementID,
			}

			// ประมวลผล "Effects" (เหมือนเดิม)
			var effects []map[string]interface{}
			json.Unmarshal(actionToPerform.EffectsJSON, &effects)

			for _, effectData := range effects {
				s.applyEffect(aiCombatant, targetCombatant, effectData, dummySpell)
			}

			// ⭐️⭐️⭐️ ถ้า AP หมดเกลี้ยง... ก็ไม่ต้องคิดต่อแล้ว ⭐️⭐️⭐️
			if aiCombatant.CurrentAP <= 0 {
				s.appLogger.Info("AI has 0 AP, stopping action loop.", "ai_id", aiCombatant.ID)
				break // ⭐️ ออกจาก "Action Loop" (for aiCombatant.CurrentAP > 0)
			}

		} else {
			// ⭐️ ถ้าวน "กฎ" จนครบแล้ว... แต่ "ไม่มีท่าไหนที่ทำได้เลย" (ทั้งที่ AP ยังเหลือ)
			s.appLogger.Info("AI has AP but found no valid action (all rules failed or unaffordable), stopping action loop.", "ai_id", aiCombatant.ID, "current_ap", aiCombatant.CurrentAP)
			break // ⭐️ ออกจาก "Action Loop" (for aiCombatant.CurrentAP > 0)
		}

	} // --- ⭐️ สิ้นสุด "Action Loop" (for aiCombatant.CurrentAP > 0) ⭐️ ---

	if actionsTakenThisTurn >= maxActionsPerTurn {
		s.appLogger.Warn("AI reached max actions per turn, force ending turn.", "ai_id", aiCombatant.ID)
	}

	// 6. จบเทิร์นของตัวเอง (ย้ายมาอยู่นอก Loop)
	s.appLogger.Info("AI ended its turn.", "match_id", match.ID)
	match = s.endTurn(match)

	// 7. เตรียมความพร้อมสำหรับเทิร์นของ "คนถัดไป" (เหมือนเดิม)
	match, err := s.startNewTurn(match)
	if err != nil {
		return nil, err
	}

	return match, nil
}
