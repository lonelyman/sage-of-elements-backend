// file: internal/modules/combat/spell_application.go
package combat

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"sage-of-elements-backend/internal/domain"

	"github.com/gofrs/uuid"
)

// AppliedEffect เก็บผลลัพธ์ของ effect ที่ถูก apply แล้ว
type AppliedEffect struct {
	EffectID    uint        `json:"effect_id"`
	EffectType  string      `json:"effect_type"` // "DAMAGE", "HEAL", "BUFF", "DEBUFF", "SHIELD"
	TargetID    uuid.UUID   `json:"target_id"`
	FinalValue  float64     `json:"final_value"`
	Evaded      bool        `json:"evaded"`
	Absorbed    float64     `json:"absorbed"`     // shield absorption
	ActualValue float64     `json:"actual_value"` // ค่าที่เกิดขึ้นจริงหลัง shield/defense
	Details     interface{} `json:"details,omitempty"`
}

// EffectApplicationResult เก็บผลลัพธ์รวมของการ apply effects ทั้งหมด
type EffectApplicationResult struct {
	AppliedEffects []AppliedEffect `json:"applied_effects"`
	TotalDamage    float64         `json:"total_damage"`
	TotalHealing   float64         `json:"total_healing"`
	TotalAbsorbed  float64         `json:"total_absorbed"`
	EffectsApplied int             `json:"effects_applied"`
	EffectsEvaded  int             `json:"effects_evaded"`
	Errors         []string        `json:"errors,omitempty"`
}

// ApplyCalculatedEffects เป็น Step 4 ของการร่ายเวท
// รับผิดชอบ apply effect แต่ละตัวพร้อมบันทึกผลลัพธ์
func (s *combatService) ApplyCalculatedEffects(
	caster *domain.Combatant,
	target *domain.Combatant,
	spell *domain.Spell,
	initialValues EffectValueMap,
	modifierCtx *ModifierContext,
) (*EffectApplicationResult, error) {

	s.appLogger.Info("💥 STEP 4: Applying calculated effects",
		"spell_id", spell.ID,
		"effect_count", len(initialValues),
		"combined_modifier", modifierCtx.CombinedMod,
	)

	result := &EffectApplicationResult{
		AppliedEffects: make([]AppliedEffect, 0),
	}

	// Loop แต่ละ effect
	for effectID, initialValue := range initialValues {
		// หา spellEffect data
		var spellEffect *domain.SpellEffect
		for _, se := range spell.Effects {
			if se.EffectID == effectID {
				spellEffect = se
				break
			}
		}

		if spellEffect == nil {
			errMsg := fmt.Sprintf("SpellEffect not found for ID %d", effectID)
			s.appLogger.Warn(errMsg)
			result.Errors = append(result.Errors, errMsg)
			continue
		}

		// คำนวณค่าสุดท้าย
		finalValue := initialValue * modifierCtx.CombinedMod

		// หา final target (สำหรับ SYNERGY_BUFF ต้องเปลี่ยนเป็น caster)
		finalTarget := s._DetermineEffectTarget(caster, target, spell, effectID)

		// Apply effect
		appliedEffect, err := s._ApplySpecificEffect(
			caster,
			finalTarget,
			spell,
			effectID,
			finalValue,
			spellEffect,
		)

		if err != nil {
			errMsg := fmt.Sprintf("Failed to apply effect %d: %v", effectID, err)
			s.appLogger.Error("Effect application error", err, "effect_id", effectID)
			result.Errors = append(result.Errors, errMsg)
			continue
		}

		// บันทึกผลลัพธ์
		result.AppliedEffects = append(result.AppliedEffects, *appliedEffect)
		result.EffectsApplied++

		// สรุปสถิติ
		if appliedEffect.Evaded {
			result.EffectsEvaded++
		} else {
			switch appliedEffect.EffectType {
			case "DAMAGE":
				result.TotalDamage += appliedEffect.ActualValue
				result.TotalAbsorbed += appliedEffect.Absorbed
			case "HEAL":
				result.TotalHealing += appliedEffect.ActualValue
			}
		}
	}

	s.appLogger.Info("✅ STEP 4 Complete: Effects applied",
		"total_effects", result.EffectsApplied,
		"total_damage", result.TotalDamage,
		"total_healing", result.TotalHealing,
		"evaded", result.EffectsEvaded,
	)

	return result, nil
}

// ==================== Helper Functions ====================

// _DetermineEffectTarget หา target ที่ถูกต้องสำหรับ effect
func (s *combatService) _DetermineEffectTarget(
	caster *domain.Combatant,
	target *domain.Combatant,
	spell *domain.Spell,
	effectID uint,
) *domain.Combatant {

	// ดึงข้อมูล effect
	effectInfo, _ := s.gameDataRepo.FindEffectByID(effectID)
	if effectInfo != nil && effectInfo.Type == domain.EffectTypeSynergyBuff {
		s.appLogger.Info("Synergy Buff detected, redirecting to caster", "effect_id", effectID)
		return caster
	}

	// ใช้ target ตาม spell.TargetType
	switch spell.TargetType {
	case domain.TargetTypeSelf:
		return caster
	case domain.TargetTypeEnemy:
		return target
	case domain.TargetTypeAlly:
		return target
	default:
		return target
	}
}

// _ApplySpecificEffect แยกประเภทของ effect แล้วเรียก sub-function ที่เหมาะสม
func (s *combatService) _ApplySpecificEffect(
	caster *domain.Combatant,
	target *domain.Combatant,
	spell *domain.Spell,
	effectID uint,
	finalValue float64,
	spellEffect *domain.SpellEffect,
) (*AppliedEffect, error) {

	s.appLogger.Debug("Applying specific effect",
		"effect_id", effectID,
		"final_value", finalValue,
		"target_id", target.ID,
	)

	// ดึงข้อมูล effect type
	effectInfo, err := s.gameDataRepo.FindEffectByID(effectID)
	if err != nil || effectInfo == nil {
		return nil, fmt.Errorf("effect info not found for ID %d", effectID)
	}

	// Switch ตาม effect type
	switch effectInfo.Type {
	case domain.EffectTypeDamage:
		return s.__ApplyDamageEffect(caster, target, spell, finalValue)

	case domain.EffectTypeHeal:
		return s.__ApplyHealEffect(target, finalValue)

	case domain.EffectTypeShield:
		baseDuration := int(spellEffect.DurationInTurns)
		duration := s._CalculateDurationBonus(caster, baseDuration)
		return s.__ApplyShieldEffect(target, finalValue, duration)

	case domain.EffectTypeBuff:
		baseDuration := int(spellEffect.DurationInTurns)
		duration := s._CalculateDurationBonus(caster, baseDuration)
		return s.__ApplyBuffEffect(target, effectID, finalValue, duration)

	case domain.EffectTypeDebuff:
		baseDuration := int(spellEffect.DurationInTurns)
		duration := s._CalculateDurationBonus(caster, baseDuration)
		return s.__ApplyDebuffEffect(caster, target, effectID, finalValue, duration, spellEffect)

	case domain.EffectTypeSynergyBuff:
		baseDuration := int(spellEffect.DurationInTurns)
		duration := s._CalculateDurationBonus(caster, baseDuration)
		return s.__ApplySynergyBuffEffect(caster, effectID, duration)

	case domain.EffectTypeResource:
		// MP_DAMAGE (effect ID 1104) เป็น RESOURCE type
		if effectID == 1104 {
			return s.__ApplyMpDamageEffect(caster, target, spell, finalValue)
		}
		return nil, fmt.Errorf("unsupported resource effect ID: %d", effectID)

	default:
		return nil, fmt.Errorf("unknown effect type: %s", effectInfo.Type)
	}
}

// ==================== Effect Type Implementations ====================

// __ApplyDamageEffect ทำ damage พร้อมเช็ค evasion, shield, defense
func (s *combatService) __ApplyDamageEffect(
	caster *domain.Combatant,
	target *domain.Combatant,
	spell *domain.Spell,
	damage float64,
) (*AppliedEffect, error) {

	result := &AppliedEffect{
		EffectID:    1101, // DAMAGE
		EffectType:  "DAMAGE",
		TargetID:    target.ID,
		FinalValue:  damage,
		Evaded:      false,
		Absorbed:    0,
		ActualValue: 0,
	}

	// 1. Check Evasion
	if s._CheckEvasion(target) {
		result.Evaded = true
		s.appLogger.Info("Attack EVADED!", "caster", caster.ID, "target", target.ID)
		return result, nil
	}

	// 2. Check Vulnerable (increase damage)
	damageAfterVulnerable := s._ApplyVulnerableModifier(target, damage)

	// 3. Apply Shield Absorption
	damageAfterShield, absorbed := s._ApplyShieldAbsorption(target, damageAfterVulnerable)
	result.Absorbed = absorbed

	// 4. Apply Defense Reduction
	finalDamage := s._ApplyDefenseReduction(target, damageAfterShield)

	// 5. Deduct HP
	oldHP := target.CurrentHP
	target.CurrentHP -= int(finalDamage)
	if target.CurrentHP < 0 {
		target.CurrentHP = 0
	}

	result.ActualValue = float64(oldHP - target.CurrentHP)

	s.appLogger.Info("Damage applied",
		"target_id", target.ID,
		"raw_damage", damage,
		"after_vulnerable", damageAfterVulnerable,
		"absorbed", absorbed,
		"final_damage", finalDamage,
		"hp_before", oldHP,
		"hp_after", target.CurrentHP,
	)

	return result, nil
}

// __ApplyHealEffect ฟื้นฟู HP
func (s *combatService) __ApplyHealEffect(
	target *domain.Combatant,
	healAmount float64,
) (*AppliedEffect, error) {

	result := &AppliedEffect{
		EffectID:   1103, // HEAL
		EffectType: "HEAL",
		TargetID:   target.ID,
		FinalValue: healAmount,
	}

	maxHP := s.getMaxHP(target)
	oldHP := target.CurrentHP

	target.CurrentHP += int(healAmount)
	if target.CurrentHP > maxHP {
		target.CurrentHP = maxHP
	}

	result.ActualValue = float64(target.CurrentHP - oldHP)

	s.appLogger.Info("Heal applied",
		"target_id", target.ID,
		"heal_amount", healAmount,
		"hp_before", oldHP,
		"hp_after", target.CurrentHP,
		"max_hp", maxHP,
	)

	return result, nil
}

// __ApplyShieldEffect ให้ shield
func (s *combatService) __ApplyShieldEffect(
	target *domain.Combatant,
	shieldValue float64,
	duration int,
) (*AppliedEffect, error) {

	result := &AppliedEffect{
		EffectID:    1102, // SHIELD
		EffectType:  "SHIELD",
		TargetID:    target.ID,
		FinalValue:  shieldValue,
		ActualValue: shieldValue,
	}

	// เพิ่ม shield effect
	newEffect := domain.ActiveEffect{
		EffectID:       1102, // SHIELD (same as effect type)
		Value:          int(shieldValue),
		TurnsRemaining: duration,
		SourceID:       target.ID,
	}

	s._AddActiveEffect(target, newEffect)

	s.appLogger.Info("Shield applied",
		"target_id", target.ID,
		"shield_value", shieldValue,
		"duration", duration,
	)

	return result, nil
}

// __ApplyBuffEffect ให้ buff
func (s *combatService) __ApplyBuffEffect(
	target *domain.Combatant,
	effectID uint,
	value float64,
	duration int,
) (*AppliedEffect, error) {

	result := &AppliedEffect{
		EffectID:    effectID,
		EffectType:  "BUFF",
		TargetID:    target.ID,
		FinalValue:  value,
		ActualValue: value,
	}

	newEffect := domain.ActiveEffect{
		EffectID:       effectID,
		Value:          int(value),
		TurnsRemaining: duration,
		SourceID:       target.ID,
	}

	s._AddActiveEffect(target, newEffect)

	s.appLogger.Info("Buff applied",
		"target_id", target.ID,
		"effect_id", effectID,
		"value", value,
		"duration", duration,
	)

	return result, nil
}

// __ApplyDebuffEffect ให้ debuff
func (s *combatService) __ApplyDebuffEffect(
	caster *domain.Combatant,
	target *domain.Combatant,
	effectID uint,
	value float64,
	duration int,
	spellEffect *domain.SpellEffect,
) (*AppliedEffect, error) {

	result := &AppliedEffect{
		EffectID:    effectID,
		EffectType:  "DEBUFF",
		TargetID:    target.ID,
		FinalValue:  value,
		ActualValue: value,
	}

	// สำหรับ DoT effects (IGNITE) ต้องคำนวณใหม่ด้วย Talent P
	finalValue := value
	if effectID == 4201 { // DEBUFF_IGNITE
		finalValue = s._CalculateDoTValue(caster, spellEffect.BaseValue)
	}

	newEffect := domain.ActiveEffect{
		EffectID:       effectID,
		Value:          int(finalValue),
		TurnsRemaining: duration,
		SourceID:       caster.ID,
	}

	s._AddActiveEffect(target, newEffect)

	s.appLogger.Info("Debuff applied",
		"target_id", target.ID,
		"effect_id", effectID,
		"value", finalValue,
		"duration", duration,
	)

	return result, nil
}

// __ApplySynergyBuffEffect ให้ synergy buff (stance)
func (s *combatService) __ApplySynergyBuffEffect(
	caster *domain.Combatant,
	effectID uint,
	duration int,
) (*AppliedEffect, error) {

	result := &AppliedEffect{
		EffectID:   effectID,
		EffectType: "BUFF",
		TargetID:   caster.ID,
	}

	// Synergy buff value = Talent P
	talentP := 0
	if caster.Character != nil {
		talentP = caster.Character.TalentP
	}

	newEffect := domain.ActiveEffect{
		EffectID:       effectID,
		Value:          talentP,
		TurnsRemaining: duration,
		SourceID:       caster.ID,
	}

	s._AddActiveEffect(caster, newEffect)

	result.FinalValue = float64(talentP)
	result.ActualValue = float64(talentP)

	s.appLogger.Info("Synergy Buff applied",
		"caster_id", caster.ID,
		"effect_id", effectID,
		"talent_p", talentP,
		"duration", duration,
	)

	return result, nil
}

// __ApplyMpDamageEffect ทำ MP damage พร้อมเติม MP ให้ caster
func (s *combatService) __ApplyMpDamageEffect(
	caster *domain.Combatant,
	target *domain.Combatant,
	spell *domain.Spell,
	mpDamage float64,
) (*AppliedEffect, error) {

	result := &AppliedEffect{
		EffectID:   1104, // MP_DAMAGE
		EffectType: "MP_DAMAGE",
		TargetID:   target.ID,
		FinalValue: mpDamage,
	}

	// หัก MP จาก target
	oldMP := target.CurrentMP
	target.CurrentMP -= int(mpDamage)
	if target.CurrentMP < 0 {
		target.CurrentMP = 0
	}

	actualDrained := oldMP - target.CurrentMP

	// เติม MP ให้ caster (50%)
	mpGain := int(float64(actualDrained) * 0.5)
	maxMP := s.getMaxMP(caster)
	caster.CurrentMP += mpGain
	if caster.CurrentMP > maxMP {
		caster.CurrentMP = maxMP
	}

	result.ActualValue = float64(actualDrained)
	result.Details = map[string]interface{}{
		"mp_gained": mpGain,
	}

	s.appLogger.Info("MP Damage applied",
		"target_id", target.ID,
		"mp_damage", mpDamage,
		"mp_drained", actualDrained,
		"caster_mp_gain", mpGain,
	)

	return result, nil
}

// ==================== Sub-Helper Functions ====================

// _CheckEvasion ตรวจสอบว่า target หลบได้หรือไม่
func (s *combatService) _CheckEvasion(target *domain.Combatant) bool {
	if target.ActiveEffects == nil {
		return false
	}

	var activeEffects []domain.ActiveEffect
	err := json.Unmarshal(target.ActiveEffects, &activeEffects)
	if err != nil {
		return false
	}

	for _, effect := range activeEffects {
		if effect.EffectID == 2201 { // BUFF_EVASION
			evasionChance := effect.Value
			if evasionChance > 0 {
				roll := rand.Intn(100)
				s.appLogger.Info("Evasion check", "chance", evasionChance, "roll", roll)
				return roll < evasionChance
			}
		}
	}

	return false
}

// _ApplyVulnerableModifier เพิ่ม damage ถ้า target มี vulnerable debuff
func (s *combatService) _ApplyVulnerableModifier(target *domain.Combatant, damage float64) float64 {
	if target.ActiveEffects == nil {
		return damage
	}

	var activeEffects []domain.ActiveEffect
	err := json.Unmarshal(target.ActiveEffects, &activeEffects)
	if err != nil {
		return damage
	}

	for _, effect := range activeEffects {
		if effect.EffectID == 4102 { // DEBUFF_VULNERABLE
			increasePercent := effect.Value
			if increasePercent > 0 {
				modifier := 1.0 + (float64(increasePercent) / 100.0)
				newDamage := damage * modifier
				s.appLogger.Info("Vulnerable modifier applied",
					"increase_percent", increasePercent,
					"damage_before", damage,
					"damage_after", newDamage,
				)
				return newDamage
			}
		}
	}

	return damage
}

// _ApplyShieldAbsorption ให้ shield ดูด damage
func (s *combatService) _ApplyShieldAbsorption(target *domain.Combatant, damage float64) (float64, float64) {
	if target.ActiveEffects == nil {
		return damage, 0
	}

	var activeEffects []domain.ActiveEffect
	err := json.Unmarshal(target.ActiveEffects, &activeEffects)
	if err != nil {
		return damage, 0
	}

	totalAbsorbed := 0.0
	remainingDamage := damage

	// หา shield effects
	newEffects := make([]domain.ActiveEffect, 0)
	for _, effect := range activeEffects {
		if effect.EffectID == 1102 && remainingDamage > 0 { // SHIELD
			shieldValue := float64(effect.Value)
			if shieldValue > remainingDamage {
				// Shield มากกว่า damage → Shield absorb ทั้งหมด
				totalAbsorbed += remainingDamage
				effect.Value -= int(remainingDamage)
				remainingDamage = 0
				newEffects = append(newEffects, effect)
			} else {
				// Damage มากกว่า shield → Shield หมด
				totalAbsorbed += shieldValue
				remainingDamage -= shieldValue
				// ไม่เอา shield ที่หมดแล้วเข้า list
			}
		} else {
			newEffects = append(newEffects, effect)
		}
	}

	// Update active effects
	if totalAbsorbed > 0 {
		newJSON, _ := json.Marshal(newEffects)
		target.ActiveEffects = newJSON
		s.appLogger.Info("Shield absorption",
			"absorbed", totalAbsorbed,
			"damage_after_shield", remainingDamage,
		)
	}

	return remainingDamage, totalAbsorbed
}

// _ApplyDefenseReduction ลด damage จาก defense buff
func (s *combatService) _ApplyDefenseReduction(target *domain.Combatant, damage float64) float64 {
	if target.ActiveEffects == nil {
		return damage
	}

	var activeEffects []domain.ActiveEffect
	err := json.Unmarshal(target.ActiveEffects, &activeEffects)
	if err != nil {
		return damage
	}

	for _, effect := range activeEffects {
		if effect.EffectID == 2204 { // BUFF_DEFENSE_UP
			reductionPercent := effect.Value
			if reductionPercent > 0 {
				modifier := 1.0 - (float64(reductionPercent) / 100.0)
				if modifier < 0 {
					modifier = 0
				}
				newDamage := damage * modifier
				s.appLogger.Info("Defense reduction applied",
					"reduction_percent", reductionPercent,
					"damage_before", damage,
					"damage_after", newDamage,
				)
				return newDamage
			}
		}
	}

	return damage
}

// _CalculateDoTValue คำนวณ DoT damage ด้วย Talent P
func (s *combatService) _CalculateDoTValue(caster *domain.Combatant, baseValue float64) float64 {
	if caster.Character == nil {
		return baseValue
	}

	talentP := float64(caster.Character.TalentP)
	dotDivisor := 10.0 // สามารถดึงจาก config ได้

	finalValue := baseValue + (talentP / dotDivisor)

	s.appLogger.Debug("DoT value calculated",
		"base_value", baseValue,
		"talent_p", talentP,
		"final_value", finalValue,
	)

	return finalValue
}

// _AddActiveEffect เพิ่ม active effect ให้ combatant
func (s *combatService) _AddActiveEffect(combatant *domain.Combatant, newEffect domain.ActiveEffect) {
	var currentEffects []domain.ActiveEffect

	// Load existing effects
	if combatant.ActiveEffects != nil {
		json.Unmarshal(combatant.ActiveEffects, &currentEffects)
	}

	// Add new effect
	currentEffects = append(currentEffects, newEffect)

	// Save back
	newJSON, _ := json.Marshal(currentEffects)
	combatant.ActiveEffects = newJSON
}
