// file: internal/modules/combat/ai_execution.go
package combat

import (
	"encoding/json"
	"sage-of-elements-backend/internal/domain"
)

// ==================== AI Action Execution ====================
// ไฟล์นี้รวบรวมฟังก์ชันที่เกี่ยวกับการ execute action ของ AI
// - การหักทรัพยากร (resource deduction)
// - การ apply effects
// - การ log action results

// ExecuteAIAction ทำการ execute action ที่เลือกแล้ว
// จะหักทรัพยากร, apply effects, และ log ผลลัพธ์
func (s *combatService) ExecuteAIAction(
	aiCombatant *domain.Combatant,
	action *AISelectedAction,
) error {
	s.appLogger.Info("🎯 AI executing action",
		"ai_id", aiCombatant.ID,
		"ability", action.Ability.Name,
		"target", action.Target.ID,
	)

	// 1. หักทรัพยากร (AP, MP)
	s._DeductResources(aiCombatant, action.Ability)

	// 2. Apply effects ของ ability
	err := s._ApplyAbilityEffects(aiCombatant, action)
	if err != nil {
		s.appLogger.Error("Failed to apply AI ability effects", err,
			"ai_id", aiCombatant.ID,
			"ability", action.Ability.Name,
		)
		return err
	}

	// 3. Log สถานะหลัง execute
	s.appLogger.Info("✅ AI action executed successfully",
		"ai_id", aiCombatant.ID,
		"ability", action.Ability.Name,
		"ap_remaining", aiCombatant.CurrentAP,
		"mp_remaining", aiCombatant.CurrentMP,
	)

	return nil
}

// ==================== Resource Management ====================

// _DeductResources หักทรัพยากร (AP, MP) จาก AI combatant
func (s *combatService) _DeductResources(
	aiCombatant *domain.Combatant,
	ability *domain.EnemyAbility,
) {
	// หัก AP
	aiCombatant.CurrentAP -= ability.APCost
	if aiCombatant.CurrentAP < 0 {
		aiCombatant.CurrentAP = 0
	}

	// หัก MP (ถ้ามี)
	if ability.MPCost > 0 {
		aiCombatant.CurrentMP -= ability.MPCost
		if aiCombatant.CurrentMP < 0 {
			aiCombatant.CurrentMP = 0
		}
	}

	s.appLogger.Debug("AI resources deducted",
		"ap_cost", ability.APCost,
		"mp_cost", ability.MPCost,
		"ap_remaining", aiCombatant.CurrentAP,
		"mp_remaining", aiCombatant.CurrentMP,
	)
}

// ==================== Effect Application ====================

// _ApplyAbilityEffects apply effects ทั้งหมดของ ability
func (s *combatService) _ApplyAbilityEffects(
	aiCombatant *domain.Combatant,
	action *AISelectedAction,
) error {
	// Parse effects จาก JSON
	var effects []map[string]interface{}
	err := json.Unmarshal(action.Ability.EffectsJSON, &effects)
	if err != nil {
		s.appLogger.Error("Failed to unmarshal ability effects", err,
			"ability", action.Ability.Name,
		)
		return err
	}

	// สร้าง dummy spell สำหรับ effect manager
	// (ใช้ธาตุของ enemy เป็น element ของเวท)
	dummySpell := &domain.Spell{
		ElementID: aiCombatant.Enemy.ElementID,
	}

	// Apply แต่ละ effect
	s.appLogger.Debug("Applying AI ability effects",
		"ability", action.Ability.Name,
		"effects_count", len(effects),
	)

	for i, effectData := range effects {
		s.appLogger.Debug("Applying effect",
			"effect_index", i,
			"effect_data", effectData,
		)

		// เรียกใช้ effect manager ที่มีอยู่แล้ว
		s.applyEffect(aiCombatant, action.Target, effectData, dummySpell)
	}

	return nil
}
