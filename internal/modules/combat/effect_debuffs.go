// file: internal/modules/combat/effect_debuffs.go
package combat

import (
	"encoding/json"
	"math"
	"sage-of-elements-backend/internal/domain"
)

// ============================================================================
// 📌 DEBUFF EFFECTS (4000s Range)
// ============================================================================
// Effect IDs: 4101 (Slow), 4102 (Vulnerable), 4201 (Ignite/DoT)
// ============================================================================

// --- ⭐️ ดีบัฟ Slow ⭐️ ---
func (s *combatService) applyDebuffSlow(caster *domain.Combatant, target *domain.Combatant, effectData map[string]interface{}) {
	value := int(effectData["value"].(float64))
	duration := int(effectData["duration"].(float64))

	var activeEffects []domain.ActiveEffect
	if target.ActiveEffects != nil {
		json.Unmarshal(target.ActiveEffects, &activeEffects)
	}
	newEffect := domain.ActiveEffect{
		EffectID:       4101, // DEBUFF_SLOW
		Value:          value,
		TurnsRemaining: duration,
		SourceID:       caster.ID,
	}
	activeEffects = append(activeEffects, newEffect)
	newEffectsJSON, _ := json.Marshal(activeEffects)
	target.ActiveEffects = newEffectsJSON
	s.appLogger.Info("Applied DEBUFF_SLOW effect", "caster", caster.ID, "target", target.ID, "duration", duration)
}

// --- ⭐️ ดีบัฟ Vulnerable ⭐️ ---
func (s *combatService) applyDebuffVulnerable(caster *domain.Combatant, target *domain.Combatant, effectData map[string]interface{}) {
	// --- การตรวจสอบ Type Assertion ---
	// เวท ID 16 (Analyze) มี BaseValue: 10, Duration: 2
	valueFloat, ok1 := effectData["value"].(float64)       // % Damage ที่เพิ่มขึ้น
	durationFloat, ok2 := effectData["duration"].(float64) // ระยะเวลา
	if !ok1 || !ok2 {
		s.appLogger.Warn("Invalid or missing value or duration in effectData for applyDebuffVulnerable", "data", effectData)
		return
	}
	// -----------------------------

	vulnerabilityPercent := int(math.Round(valueFloat))
	duration := int(durationFloat)
	if vulnerabilityPercent < 0 {
		vulnerabilityPercent = 0
	} // ไม่ควรติดลบ

	var activeEffects []domain.ActiveEffect
	if target.ActiveEffects != nil {
		err := json.Unmarshal(target.ActiveEffects, &activeEffects)
		if err != nil {
			s.appLogger.Error("Failed to unmarshal existing active effects for Vulnerable debuff", err, "target_id", target.ID)
			activeEffects = []domain.ActiveEffect{} // เริ่มใหม่ถ้า unmarshal ไม่ได้
		}
	}

	// --- ⭐️ จัดการ Stack: (อันใหม่ทับอันเก่า) ⭐️ ---
	var tempEffects []domain.ActiveEffect
	for _, effect := range activeEffects {
		if effect.EffectID != 4102 { // DEBUFF_VULNERABLE - เก็บ Effect อื่นๆ
			tempEffects = append(tempEffects, effect)
		} else {
			s.appLogger.Info("Replacing existing Vulnerable debuff", "target_id", target.ID)
		}
	}
	activeEffects = tempEffects // ใช้ list ที่กรองอันเก่าออกแล้ว
	// -----------------------------------------------------------------

	// สร้าง Object Debuff ใหม่
	newEffect := domain.ActiveEffect{
		EffectID:       4102,                 // DEBUFF_VULNERABLE
		Value:          vulnerabilityPercent, // % Damage ที่จะได้รับเพิ่ม
		TurnsRemaining: duration,
		SourceID:       caster.ID,
	}
	activeEffects = append(activeEffects, newEffect)

	// แปลงกลับเป็น JSON
	newEffectsJSON, err := json.Marshal(activeEffects)
	if err != nil {
		s.appLogger.Error("Failed to marshal updated active effects for Vulnerable debuff", err, "target_id", target.ID)
		return
	}
	target.ActiveEffects = newEffectsJSON

	s.appLogger.Info("Applied DEBUFF_VULNERABLE effect", "target", target.ID, "duration", duration, "increase_percent", vulnerabilityPercent)
}

// --- ⭐️ ดีบัฟ Ignite (DoT) ⭐️ ---
func (s *combatService) applyDebuffIgnite(caster *domain.Combatant, target *domain.Combatant, effectData map[string]interface{}) {
	// --- การตรวจสอบ Type Assertion ---
	valueFloat, ok1 := effectData["value"].(float64)       // Damage ต่อเทิร์น
	durationFloat, ok2 := effectData["duration"].(float64) // ระยะเวลา
	if !ok1 || !ok2 {
		s.appLogger.Warn("Invalid or missing value or duration in effectData for applyDebuffIgnite", "data", effectData)
		return
	}
	// -----------------------------

	dotPerTurn := int(math.Round(valueFloat))
	duration := int(durationFloat)
	if dotPerTurn < 0 {
		dotPerTurn = 0
	} // Damage ไม่ควรติดลบ

	var activeEffects []domain.ActiveEffect
	if target.ActiveEffects != nil {
		err := json.Unmarshal(target.ActiveEffects, &activeEffects)
		if err != nil {
			s.appLogger.Error("Failed to unmarshal existing active effects for Ignite debuff", err, "target_id", target.ID)
			activeEffects = []domain.ActiveEffect{} // เริ่มใหม่ถ้า unmarshal ไม่ได้
		}
	}

	// --- ⭐️ จัดการ Stack: (อันใหม่ทับอันเก่า) ⭐️ ---
	var tempEffects []domain.ActiveEffect
	for _, effect := range activeEffects {
		if effect.EffectID != 4201 { // DEBUFF_IGNITE - เก็บ Effect อื่นๆ
			tempEffects = append(tempEffects, effect)
		} else {
			s.appLogger.Info("Replacing existing Ignite debuff", "target_id", target.ID)
		}
	}
	activeEffects = tempEffects // ใช้ list ที่กรอง Ignite อันเก่าออกแล้ว
	// -----------------------------------------------------------------

	// สร้าง Object Debuff ใหม่
	newEffect := domain.ActiveEffect{
		EffectID:       4201,       // DEBUFF_IGNITE
		Value:          dotPerTurn, // Damage ที่จะทำต่อเทิร์น
		TurnsRemaining: duration,
		SourceID:       caster.ID,
	}
	activeEffects = append(activeEffects, newEffect)

	// แปลงกลับเป็น JSON
	newEffectsJSON, err := json.Marshal(activeEffects)
	if err != nil {
		s.appLogger.Error("Failed to marshal updated active effects for Ignite debuff", err, "target_id", target.ID)
		return
	}
	target.ActiveEffects = newEffectsJSON

	s.appLogger.Info("Applied DEBUFF_IGNITE effect", "target", target.ID, "duration", duration, "damage_per_turn", dotPerTurn)
}
