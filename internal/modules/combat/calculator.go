// file: internal/modules/combat/calculator.go
package combat

import (
	"encoding/json"
	"math"
	"sage-of-elements-backend/internal/domain"
	"strconv"
)

// --- นี่คือบ้านของฟังก์ชันช่วยเหลือสำหรับการคำนวณ ---

// ==================== DEPRECATED FUNCTIONS ====================
// calculateEffectValue - DEPRECATED: ใช้ใน effect_direct.go เท่านั้น (ระบบเก่า)
// ระบบใหม่ใช้ CalculateInitialEffectValues + CalculateCombinedModifiers ใน spell_calculation.go แทน
func (s *combatService) calculateEffectValue(caster *domain.Combatant, target *domain.Combatant, spell *domain.Spell, effect *domain.SpellEffect, powerModifier float64) (float64, error) {
	baseValue := effect.BaseValue
	masteryBonus := 0.0

	// ใช้ calculateTalentBonusFromRecipe แทน getTalentBonus เก่า
	var talentBonus float64
	if caster.Character == nil {
		talentBonus = 0.0
	} else if spell.ElementID <= 4 {
		// T0: ใช้ talent เดียว
		ingredients := map[uint]int{spell.ElementID: 1}
		talentBonus = s.calculateTalentBonusFromRecipe(ingredients, caster.Character)
	} else {
		// T1+: ต้องหา recipe
		recipe, err := s.gameDataRepo.FindRecipeByOutputElementID(spell.ElementID)
		if err != nil || recipe == nil {
			talentBonus = 0.0
		} else {
			ingredientCount := make(map[uint]int)
			for _, ing := range recipe.Ingredients {
				ingredientCount[ing.InputElementID]++
			}
			talentBonus = s.calculateTalentBonusFromRecipe(ingredientCount, caster.Character)
		}
	}

	var targetElementID uint = 0
	if target.EnemyID != nil && target.Enemy != nil {
		targetElementID = target.Enemy.ElementID
	} else if target.CharacterID != nil && target.Character != nil {
		targetElementID = target.Character.PrimaryElementID
	}

	elementalModifier := 1.0
	var err error

	// เช็คว่าเป็น friendly effect หรือไม่
	isFriendlyEffect := (effect.EffectID == 1102 || // SHIELD
		effect.EffectID == 1103 || // HEAL
		effect.EffectID == 2101 || // BUFF_HP_REGEN
		effect.EffectID == 2102 || // BUFF_MP_REGEN
		effect.EffectID == 2201 || // BUFF_EVASION
		effect.EffectID == 2202 || // BUFF_DMG_UP
		effect.EffectID == 2203 || // BUFF_RETALIATION
		effect.EffectID == 2204) // BUFF_DEFENSE_UP

	if !isFriendlyEffect {
		elementalModifier, err = s.getElementalModifier(spell.ElementID, targetElementID)
		if err != nil {
			s.appLogger.Error("Failed to get elemental modifier", err, "spell_element", spell.ElementID, "target_element", targetElementID)
			elementalModifier = 1.0
		}
	}

	buffDebuffModifier := 1.0

	var targetEffects []domain.ActiveEffect
	if target.ActiveEffects != nil {
		json.Unmarshal(target.ActiveEffects, &targetEffects)
	}

	for _, activeEffect := range targetEffects {
		switch activeEffect.EffectID {
		case 2204: // BUFF_DEFENSE_UP
			if !isFriendlyEffect {
				reductionPercent := 0.5
				if activeEffect.Value > 0 {
					reductionPercent = float64(activeEffect.Value) / 100.0
				}
				buffDebuffModifier *= (1.0 - reductionPercent)
			}
		case 4102: // DEBUFF_VULNERABLE
			if !isFriendlyEffect {
				increasePercent := float64(activeEffect.Value) / 100.0
				buffDebuffModifier *= (1.0 + increasePercent)
			}
		}
	}

	var casterEffects []domain.ActiveEffect
	if caster.ActiveEffects != nil {
		json.Unmarshal(caster.ActiveEffects, &casterEffects)
	}
	for _, activeEffect := range casterEffects {
		if activeEffect.EffectID == 2202 && !isFriendlyEffect { // BUFF_DAMAGE_UP
			increasePercent := float64(activeEffect.Value) / 100.0
			buffDebuffModifier *= (1.0 + increasePercent)
		}
	}

	finalValue := (baseValue + masteryBonus + talentBonus) * elementalModifier * buffDebuffModifier * powerModifier
	finalValue = math.Round(finalValue*100) / 100

	return finalValue, nil
}

// ==================== ACTIVE FUNCTIONS ====================

// calculateTalentBonusFromRecipe คือ "ผู้เชี่ยวชาญการคำนวณ" ที่ใช้ "กฎ" ของเรา!
func (s *combatService) calculateTalentBonusFromRecipe(ingredients map[uint]int, character *domain.Character) float64 {
	totalBonus := 0.0
	// ⭐️ ดึงค่าตัวหารจาก Game Config ⭐️
	talentDivisorStr, _ := s.gameDataRepo.GetGameConfigValue("TALENT_DMG_DIVISOR")
	talentDivisor := 10.0 // Default fallback
	if talentDivisorStr != "" {
		if val, err := strconv.ParseFloat(talentDivisorStr, 64); err == nil {
			talentDivisor = val
		}
	}

	// --- Logic สำหรับ T0 (S) และ T1 (S+P) ---
	// (กฎอัจฉริยะของน้องชาย: ยกประโยชน์ให้ 100% ทั้งคู่)
	if len(ingredients) <= 2 {
		for elementID := range ingredients {
			totalBonus += s.getTalentValue(character, elementID) / talentDivisor
		}
	} else {
		// --- Logic สำหรับ T2+ (เช่น S+S+P) ---
		// TODO: Implement Dominant Element logic (S+S+P -> 100% S, 25% P)
		// (สำหรับตอนนี้... เราใช้ Logic เดียวกับ T1 ไปก่อนก็ได้)
		for elementID := range ingredients {
			totalBonus += s.getTalentValue(character, elementID) / talentDivisor
		}
	}

	return totalBonus
}

// getTalentValue ดึงค่า talent จาก character ตาม element ID
func (s *combatService) getTalentValue(character *domain.Character, elementID uint) float64 {
	switch elementID {
	case 1:
		return float64(character.TalentS)
	case 2:
		return float64(character.TalentL)
	case 3:
		return float64(character.TalentG)
	case 4:
		return float64(character.TalentP)
	default:
		return 0.0
	}
}

// getElementalModifier ดึงค่า modifier จากความได้เปรียบด้านธาตุ
func (s *combatService) getElementalModifier(attackerID, defenderID uint) (float64, error) {
	if defenderID == 0 {
		return 1.0, nil
	}

	modifierStr, err := s.gameDataRepo.GetMatchupModifier(attackerID, defenderID)
	if err != nil {
		s.appLogger.Error("Failed to get matchup modifier from repo", err)
		return 1.0, err
	}
	if modifierStr == "" {
		s.appLogger.Warn("Elemental matchup not found, using default 1.0", "attacker", attackerID, "defender", defenderID)
		return 1.0, nil
	}

	modifier, _ := strconv.ParseFloat(modifierStr, 64)
	return modifier, nil
}
