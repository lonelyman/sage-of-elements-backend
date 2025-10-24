// file: internal/modules/combat/calculator.go
package combat

import (
	"encoding/json"
	"fmt"
	"math"
	"sage-of-elements-backend/internal/domain"
	"strconv"
)

// --- นี่คือบ้านใหม่ของผู้เชี่ยวชาญด้านการคำนวณ ---
// ⭐️ เพิ่ม argument: powerModifier float64 ⭐️
func (s *combatService) calculateEffectValue(caster *domain.Combatant, target *domain.Combatant, spell *domain.Spell, effect *domain.SpellEffect, powerModifier float64) (float64, error) {
	baseValue := effect.BaseValue
	masteryBonus := 0.0                                             // TODO: Implement mastery bonus
	talentBonus := s.getTalentBonus(caster, spell, effect.EffectID) // ⭐️ ส่ง EffectID ไปด้วย! ⭐️

	var targetElementID uint = 0
	if target.EnemyID != nil && target.Enemy != nil {
		targetElementID = target.Enemy.ElementID
	} else if target.CharacterID != nil && target.Character != nil {
		targetElementID = target.Character.PrimaryElementID // สมมติว่า Player ก็มีธาตุ
	}

	elementalModifier := 1.0 // ค่าเริ่มต้น
	var err error
	// --- ⭐️ เพิ่มเช็ค: Heal ไม่ควรสนธาตุ! ⭐️ ---
	isHealEffect := (effect.EffectID == 3 || effect.EffectID == 100) // ID 3 = Heal, 100 = HoT
	if !isHealEffect {
		elementalModifier, err = s.getElementalModifier(spell.ElementID, targetElementID)
		if err != nil {
			// อาจจะแค่ Log Warning แล้วใช้ 1.0 แทนที่จะ Return Error?
			s.appLogger.Error("Failed to get elemental modifier", err, "spell_element", spell.ElementID, "target_element", targetElementID)
			elementalModifier = 1.0
		}
	} else {
		s.appLogger.Info("Skipping elemental modifier for Heal effect", "effect_id", effect.EffectID)
	}
	// ------------------------------------

	buffDebuffModifier := 1.0 // ค่าเริ่มต้น

	var targetEffects []domain.ActiveEffect
	if target.ActiveEffects != nil {
		json.Unmarshal(target.ActiveEffects, &targetEffects)
	}

	for _, activeEffect := range targetEffects {
		// --- ⭐️ ปรับปรุง Logic Buff/Debuff ⭐️ ---
		switch activeEffect.EffectID {
		case 110: // BUFF_DEFENSE_UP (กายาเหล็ก/ผิวศิลา)
			if !isHealEffect { // บัฟป้องกัน ไม่ควรลด Heal
				// สมมติว่า BaseValue ของ Effect 110 คือ % ลดทอน (เช่น 0 = ไม่ลด, 50 = ลด 50%)
				// หรืออาจจะเป็นค่า Def คงที่ที่ต้องเอาไปคำนวณร่วมกับ Atk? -> ตอนนี้ทำแบบ % ไปก่อน
				reductionPercent := 0.5 // Default ลด 50% ถ้า BaseValue = 0 (จาก Harden)
				if activeEffect.Value > 0 {
					reductionPercent = float64(activeEffect.Value) / 100.0 // ถ้ามี Value ให้ใช้ค่านั้น
				}
				buffDebuffModifier *= (1.0 - reductionPercent) // ลด Damage ตาม %
				s.appLogger.Info("Applying DEFENSE_UP modifier", "target_id", target.ID, "reduction", reductionPercent)
			}
		case 302: // DEBUFF_VULNERABLE (เปิดจุดอ่อน/Analyze)
			if !isHealEffect { // Vulnerable ไม่ควรเพิ่ม Heal
				// สมมติว่า BaseValue ของ Effect 302 คือ % ที่โดนแรงขึ้น
				increasePercent := float64(activeEffect.Value) / 100.0
				buffDebuffModifier *= (1.0 + increasePercent) // เพิ่ม Damage ตาม %
				s.appLogger.Info("Applying VULNERABLE modifier", "target_id", target.ID, "increase", increasePercent)
			}
			// TODO: เพิ่ม case 103 (BUFF_DAMAGE_UP ของ Caster) -> อันนี้ต้องเช็คที่ Caster ไม่ใช่ Target
		}
		// ------------------------------------
	}

	// --- ✨⭐️ เช็ค Buff ที่ Caster! ⭐️✨ ---
	var casterEffects []domain.ActiveEffect
	if caster.ActiveEffects != nil {
		err := json.Unmarshal(caster.ActiveEffects, &casterEffects)
		if err != nil {
			s.appLogger.Error("Failed to unmarshal caster active effects", err, "caster_id", caster.ID)
		}
	}
	for _, activeEffect := range casterEffects {
		switch activeEffect.EffectID {
		case 103: // BUFF_DAMAGE_UP (Caster)
			if !isHealEffect { // Damage Up ไม่ควรเพิ่ม Heal
				increasePercent := float64(activeEffect.Value) / 100.0
				buffDebuffModifier *= (1.0 + increasePercent) // เพิ่ม Damage ตาม % จากบัฟผู้ร่าย
				s.appLogger.Info("Applying DAMAGE_UP modifier (from Caster)", "caster_id", caster.ID, "increase", increasePercent)
			}
		}
	}
	// --- ✨⭐️ สิ้นสุดการเช็ค Buff ที่ Caster ⭐️✨ ---

	// ⭐️ คำนวณ Final Value โดยรวมทุกอย่าง (รวม powerModifier!) ⭐️
	finalValue := (baseValue + masteryBonus + talentBonus) * elementalModifier * buffDebuffModifier * powerModifier

	// ปัดเศษทศนิยมเหลือ 2 ตำแหน่ง (เผื่อ Debug ง่ายขึ้น)
	finalValue = math.Round(finalValue*100) / 100

	logMessage := "Effect Value Calculation"
	if isHealEffect {
		logMessage = "Heal Value Calculation"
	}

	s.appLogger.Info(logMessage,
		"base", baseValue,
		"talent", talentBonus,
		"elemental", elementalModifier,
		"buff_debuff", buffDebuffModifier,
		"power_mod", powerModifier,
		"final", finalValue,
		"spell_id", spell.ID,
		"effect_id", effect.EffectID, // เพิ่ม EffectID เข้าไปใน Log
	)
	return finalValue, nil
}

// getTalentBonus จะไป "ค้นสูตร" แล้วส่งต่อให้ "ผู้เชี่ยวชาญการคำนวณ"
func (s *combatService) getTalentBonus(caster *domain.Combatant, spell *domain.Spell, effectID uint) float64 {
	// ถ้าไม่มีข้อมูลตัวละคร (เช่น เป็น Enemy ร่าย) ก็ไม่มีโบนัส Talent
	if caster.Character == nil {
		return 0.0
	}

	// --- ⭐️ เพิ่ม Logic พิเศษสำหรับ Heal! ⭐️ ---
	// เช็คว่า Effect ที่กำลังคำนวณ เป็น Heal หรือ HoT หรือไม่
	isHealEffect := (effectID == 3 || effectID == 100)
	if isHealEffect {
		s.appLogger.Info("Calculating Talent Bonus for Heal using TalentL", "effect_id", effectID)
		// ⭐️ TODO: ดึงค่าตัวหาร (10.0) มาจาก Game Config สำหรับ Heal โดยเฉพาะ ⭐️
		talentDivisor := 10.0
		// บังคับให้ใช้ Talent L (ID 2) เสมอสำหรับ Heal
		return s.getTalentValue(caster.Character, 2) / talentDivisor
	}
	// --- ⭐️ สิ้นสุด Logic พิเศษสำหรับ Heal ⭐️ ---

	// --- Logic เดิมสำหรับ Effect อื่นๆ (Damage, Debuff, etc.) ---
	// ตรวจสอบว่าเป็นเวท T0 (ธาตุพื้นฐาน 1-4) หรือไม่
	if spell.ElementID <= 4 {
		// ถ้าเป็น T0... ใช้ Talent ของธาตุนั้น 100%
		return s.calculateTalentBonusFromRecipe(map[uint]int{spell.ElementID: 1}, caster.Character)
	}

	// ถ้าเป็น T1+... ไปค้นหาสูตรผสมธาตุ
	recipe, err := s.gameDataRepo.FindRecipeByOutputElementID(spell.ElementID)
	if err != nil {
		s.appLogger.Error(fmt.Sprintf("Failed to find recipe for T1+ spell (spell_id: %d)", spell.ID), err)
		return 0.0 // คืนค่า 0.0 ถ้าหาสูตรไม่เจอ
	}
	if recipe == nil {
		s.appLogger.Warn("No recipe found for T1+ spell", "spell_id", spell.ID, "element_id", spell.ElementID)
		return 0.0
	}

	// แปลงสูตรเป็น map เพื่อนับจำนวนธาตุตั้งต้น
	ingredientCount := make(map[uint]int)
	for _, ing := range recipe.Ingredients {
		ingredientCount[ing.InputElementID]++
	}

	// ส่ง map สูตรไปคำนวณโบนัส Talent ตามกฎ (เช่น กฎธาตุเด่น)
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
