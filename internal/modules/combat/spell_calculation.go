// file: internal/modules/combat/spell_calculation.go
package combat

import (
	"fmt"
	"sage-of-elements-backend/internal/domain"
)

// EffectValueMap เก็บค่าพื้นฐานของแต่ละ effect (ก่อนคูณ modifier)
type EffectValueMap map[uint]float64 // effectID -> initial value

// ModifierContext เก็บ modifier ทั้งหมดสำหรับการคำนวณ
type ModifierContext struct {
	ElementalMod  float64 // ความได้เปรียบด้านธาตุ
	BuffDebuffMod float64 // Buff/Debuff ที่มีผล
	PowerMod      float64 // Casting mode modifier
	CombinedMod   float64 // ผลคูณรวม
}

// CalculateInitialEffectValues คำนวณค่าพื้นฐานของแต่ละ effect
// (STEP 2) รวม base value + mastery bonus + talent bonus
func (s *combatService) CalculateInitialEffectValues(
	spell *domain.Spell,
	caster *domain.Combatant,
) (EffectValueMap, error) {

	s.appLogger.Info("🧮 STEP 2: Calculating initial effect values",
		"spell_id", spell.ID,
		"effect_count", len(spell.Effects),
	)

	result := make(EffectValueMap)

	for _, spellEffect := range spell.Effects {
		// 2.1 Get Base Value
		baseValue := s._GetBaseValue(spellEffect)

		// 2.2 Calculate Mastery Bonus (multiplicative)
		masteryBonus := s._CalculateMasteryBonus(caster, spell.MasteryID)

		// 2.3 Calculate Talent Bonus (additive)
		talentBonus := s._CalculateTalentBonus(caster, spell, spellEffect.EffectID)

		// Combine: (base * mastery) + talent
		initialValue := (baseValue * masteryBonus) + talentBonus

		result[spellEffect.EffectID] = initialValue

		s.appLogger.Debug("Effect value calculated",
			"effect_id", spellEffect.EffectID,
			"base_value", baseValue,
			"mastery_bonus", masteryBonus,
			"talent_bonus", talentBonus,
			"initial_value", initialValue,
		)
	}

	s.appLogger.Info("✅ STEP 2 Complete: Initial values calculated",
		"total_effects", len(result),
	)

	return result, nil
}

// CalculateCombinedModifiers คำนวณ modifier ทั้งหมดที่จะคูณกับค่าพื้นฐาน
// (STEP 3) รวม elemental + buff/debuff + power modifiers
func (s *combatService) CalculateCombinedModifiers(
	caster *domain.Combatant,
	target *domain.Combatant,
	spell *domain.Spell,
	powerMod float64,
	effectID uint,
) (*ModifierContext, error) {

	s.appLogger.Info("⚡ STEP 3: Calculating combined modifiers",
		"caster_id", caster.ID,
		"target_id", target.ID,
		"effect_id", effectID,
	)

	ctx := &ModifierContext{
		PowerMod: powerMod,
	}

	// 3.1 Get Elemental Modifier
	var err error
	ctx.ElementalMod, err = s._GetElementalModifier(spell, caster, target)
	if err != nil {
		// ถ้าหาไม่เจอให้ใช้ค่า default 1.0
		s.appLogger.Warn("Failed to get elemental modifier, using default 1.0", "error", err)
		ctx.ElementalMod = 1.0
	}

	// 3.2 Get Buff/Debuff Modifier
	ctx.BuffDebuffMod = s._GetBuffDebuffModifier(caster, target, effectID)

	// 3.3 Calculate Combined Modifier
	ctx.CombinedMod = ctx.ElementalMod * ctx.BuffDebuffMod * ctx.PowerMod

	s.appLogger.Info("✅ STEP 3 Complete: Modifiers calculated",
		"elemental_mod", ctx.ElementalMod,
		"buff_debuff_mod", ctx.BuffDebuffMod,
		"power_mod", ctx.PowerMod,
		"combined_mod", ctx.CombinedMod,
	)

	return ctx, nil
}

// ==================== STEP 2 Sub-functions ====================

// _GetBaseValue ดึงค่าพื้นฐานจาก spell effect
func (s *combatService) _GetBaseValue(spellEffect *domain.SpellEffect) float64 {
	return spellEffect.BaseValue
}

// _CalculateMasteryBonus คำนวณโบนัสจาก mastery โดยใช้สูตร Level²
// (ตัวอย่าง: Lv.1=1, Lv.2=4, Lv.5=25, Lv.10=100)
func (s *combatService) _CalculateMasteryBonus(
	caster *domain.Combatant,
	masteryID uint,
) float64 {
	// ถ้าไม่มี character (เช่น Enemy) ไม่มี mastery bonus
	if caster.Character == nil {
		return 0.0
	}

	// ดึง mastery level จาก character
	var masteryLevel int = 1 // Default level 1
	for _, mastery := range caster.Character.Masteries {
		if mastery.MasteryID == masteryID {
			masteryLevel = mastery.Level
			break
		}
	}

	// คำนวณโบนัสด้วยสูตร exponential: Level × Level
	bonus := float64(masteryLevel * masteryLevel)

	s.appLogger.Debug("Mastery bonus calculated",
		"mastery_id", masteryID,
		"level", masteryLevel,
		"bonus", bonus,
	)

	return bonus
}

// _CalculateTalentBonus คำนวณโบนัสจาก talent (additive)
func (s *combatService) _CalculateTalentBonus(
	caster *domain.Combatant,
	spell *domain.Spell,
	effectID uint,
) float64 {
	// ใช้ฟังก์ชันเดิมจาก calculator.go
	// NOTE: ต้องหา recipe ก่อนถ้าเป็น T1+
	if spell.ElementID <= 4 {
		// T0: ใช้ talent เดียว
		ingredients := map[uint]int{spell.ElementID: 1}
		return s.calculateTalentBonusFromRecipe(ingredients, caster.Character)
	}

	// T1+: ต้องหา recipe
	recipe, err := s.gameDataRepo.FindRecipeByOutputElementID(spell.ElementID)
	if err != nil || recipe == nil {
		s.appLogger.Warn("No recipe found for T1+ spell, using 0 talent bonus",
			"spell_id", spell.ID,
			"element_id", spell.ElementID,
		)
		return 0.0
	}

	// สร้าง ingredient map
	ingredientCount := make(map[uint]int)
	for _, ing := range recipe.Ingredients {
		ingredientCount[ing.InputElementID]++
	}

	return s.calculateTalentBonusFromRecipe(ingredientCount, caster.Character)
}

// ==================== STEP 3 Sub-functions ====================

// _GetElementalModifier คำนวณ modifier จากความได้เปรียบด้านธาตุ
func (s *combatService) _GetElementalModifier(
	spell *domain.Spell,
	caster *domain.Combatant,
	target *domain.Combatant,
) (float64, error) {

	// หา element ID ของเป้าหมาย
	var targetElementID uint = 0
	if target.Enemy != nil {
		targetElementID = target.Enemy.ElementID
	} else if target.Character != nil {
		targetElementID = target.Character.PrimaryElementID
	}

	// ถ้าไม่มี element ให้ใช้ค่า neutral (1.0)
	if targetElementID == 0 {
		s.appLogger.Debug("Target has no element, using neutral modifier 1.0")
		return 1.0, nil
	}

	// เรียกใช้ฟังก์ชันเดิมจาก calculator.go
	modifier, err := s.getElementalModifier(spell.ElementID, targetElementID)
	if err != nil {
		return 1.0, err
	}

	return modifier, nil
}

// _GetBuffDebuffModifier คำนวณ modifier จาก buff/debuff
func (s *combatService) _GetBuffDebuffModifier(
	caster *domain.Combatant,
	target *domain.Combatant,
	effectID uint,
) float64 {
	// TODO: Implement buff/debuff checking logic
	// - Check caster's buffs (ATK_UP, etc.)
	// - Check target's debuffs (VULNERABLE, etc.)
	// - Return combined multiplier

	// ตอนนี้ return 1.0 (no modifier)
	// จะ implement จริงในภายหลัง
	return 1.0
}

// ==================== Helper (import from calculator.go) ====================

func (s *combatService) _GetPowerModifier(castingMode string) float64 {
	// Power modifier ถูกคำนวณไว้แล้วใน preparation step
	// ฟังก์ชันนี้เอาไว้ใช้ในกรณีที่ต้องการดึงซ้ำ
	var configKey string
	switch castingMode {
	case "CHARGE":
		configKey = "CAST_MODE_CHARGE_POWER_MOD"
	case "OVERCHARGE":
		configKey = "CAST_MODE_OVERCHARGE_POWER_MOD"
	default:
		return 1.0
	}

	modStr, _ := s.gameDataRepo.GetGameConfigValue(configKey)
	var mod float64
	_, _ = fmt.Sscanf(modStr, "%f", &mod)
	if mod == 0 {
		return 1.0
	}
	return mod
}
