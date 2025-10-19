// file: internal/modules/combat/calculator.go
package combat

import (
	"fmt"
	"sage-of-elements-backend/internal/domain"
	"strconv"
)

// --- นี่คือบ้านใหม่ของผู้เชี่ยวชาญด้านการคำนวณ ---

func (s *combatService) calculateEffectValue(caster *domain.Combatant, target *domain.Combatant, spell *domain.Spell, effect *domain.SpellEffect) (float64, error) {
	baseValue := effect.BaseValue
	masteryBonus := 0.0
	talentBonus := s.getTalentBonus(caster, spell)

	var targetElementID uint = 0
	if target.EnemyID != nil && target.Enemy != nil { // ⭐️ เพิ่ม nil check
		targetElementID = target.Enemy.ElementID
	}

	elementalModifier, err := s.getElementalModifier(spell.ElementID, targetElementID)
	if err != nil {
		return 0, err
	}

	buffDebuffModifier := 1.0 // TODO: Implement buff/debuff logic
	finalValue := (baseValue + masteryBonus + talentBonus) * elementalModifier * buffDebuffModifier

	s.appLogger.Info("Damage Calculation", "base", baseValue, "talent", talentBonus, "elemental", elementalModifier, "final", finalValue)
	return finalValue, nil
}

// getTalentBonus จะไป "ค้นสูตร" แล้วส่งต่อให้ "ผู้เชี่ยวชาญการคำนวณ"
func (s *combatService) getTalentBonus(caster *domain.Combatant, spell *domain.Spell) float64 {
	if caster.Character == nil {
		return 0.0
	}

	// 1. ตรวจสอบก่อนว่าเวทนี้เป็น T0 หรือไม่ (ธาตุพื้นฐาน 1-4)
	if spell.ElementID <= 4 {
		// ถ้าเป็น T0... ก็ใช้ Logic "กฎธาตุเด่น" 100%
		return s.calculateTalentBonusFromRecipe(map[uint]int{spell.ElementID: 1}, caster.Character)
	}

	// 2. ถ้าเป็น T1 หรือสูงกว่า... ให้ไป "ค้นหาสูตร"
	recipe, err := s.gameDataRepo.FindRecipeByOutputElementID(spell.ElementID)
	if err != nil {
		// --- ⭐️ แก้ไขตรงนี้! ⭐️ ---
		// "ยัด" ข้อมูล spell.ID เข้าไปใน "ข้อความ" ด้วย fmt.Sprintf
		// แล้วส่ง "ตัว err" เป็น "ของชิ้นที่ 2"
		s.appLogger.Error(fmt.Sprintf("Failed to find recipe for T1 spell (spell_id: %d)", spell.ID), err)
		return 0.0 // ⭐️ คืนค่า 0.0 (ไม่มีโบนัส)
	}
	if recipe == nil {
		s.appLogger.Warn("No recipe found for T1+ spell", "spell_id", spell.ID, "element_id", spell.ElementID)
		return 0.0
	}

	// 3. เราได้ "สูตรผสม" (Ingredients) มาแล้ว!
	// แปลงสูตรให้อยู่ในรูปแบบที่นับง่าย (เช่น map[S]1, map[P]1)
	ingredientCount := make(map[uint]int)
	for _, ing := range recipe.Ingredients {
		ingredientCount[ing.InputElementID]++
	}

	// 4. ส่ง "สูตรที่นับแล้ว" ไปให้ "ผู้เชี่ยวชาญ" คำนวณ!
	return s.calculateTalentBonusFromRecipe(ingredientCount, caster.Character)
}

// --- 📝 เพิ่ม "ผู้ช่วย" คนใหม่นี้เข้าไป! 📝 ---
// calculateTalentBonusFromRecipe คือ "ผู้เชี่ยวชาญการคำนวณ" ที่ใช้ "กฎ" ของเรา!
func (s *combatService) calculateTalentBonusFromRecipe(ingredients map[uint]int, character *domain.Character) float64 {
	totalBonus := 0.0
	talentDivisor := 10.0 // ⭐️ TODO: ดึง "10.0" มาจาก Game Config ("TALENT_DMG_DIVISOR")

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

// (ฟังก์ชันลูกตัวจิ๋ว ที่ช่วยให้โค้ดสะอาด - อาจจะมีอยู่แล้ว)
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
