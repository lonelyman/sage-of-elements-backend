// file: internal/modules/combat/effect_synergy.go
package combat

import (
	"encoding/json"
	"sage-of-elements-backend/internal/domain"
)

// ============================================================================
// 📌 SYNERGY EFFECTS (3000s Range)
// ============================================================================
// Effect IDs: 3101 (Stance S), 3102 (Stance L), 3103 (Stance G), 3104 (Stance P)
// ============================================================================
// Note: Stances are mutually exclusive - applying a new stance replaces the old one
// ============================================================================

// --- ⭐️ Stance S (Strength) ⭐️ ---
func (s *combatService) applySynergyGrantStanceS(caster *domain.Combatant, target *domain.Combatant, effectData map[string]interface{}) {
	// --- ⭐️ 1. Get Value (เหมือนเดิม) ⭐️ ---
	value := 0
	if v, ok := effectData["value"]; ok {
		value = int(v.(float64))
	}

	// --- ⭐️ 2. Get Duration (เพิ่ม Error Check) ⭐️ ---
	durationFloat, ok := effectData["duration"].(float64)
	if !ok {
		s.appLogger.Warn("Invalid or missing duration in effectData for applySynergyGrantStanceS", "data", effectData)
		return // Exit if duration is missing
	}
	duration := int(durationFloat)

	// --- ⭐️ 3. Unmarshal Existing Effects (เพิ่ม Error Check) ⭐️ ---
	var activeEffects []domain.ActiveEffect
	if target.ActiveEffects != nil {
		err := json.Unmarshal(target.ActiveEffects, &activeEffects)
		if err != nil {
			s.appLogger.Error("Failed to unmarshal existing active effects for Stance S buff", err, "target_id", target.ID)
			activeEffects = []domain.ActiveEffect{} // Start with an empty list if JSON is bad
		}
	}

	// --- ⭐️ 4. Add "Replace" Logic - Remove ALL existing Stances ⭐️ ---
	var tempEffects []domain.ActiveEffect
	for _, effect := range activeEffects {
		// Keep effects that are NOT any Stance (3101-3104)
		if effect.EffectID < 3101 || effect.EffectID > 3104 {
			tempEffects = append(tempEffects, effect)
		} else {
			s.appLogger.Info("Replacing existing Stance", "target_id", target.ID, "old_stance_id", effect.EffectID)
		}
	}
	activeEffects = tempEffects // Use the filtered list (without any old Stance)

	// --- ⭐️ 5. Create and Append New Effect (เหมือนเดิม) ⭐️ ---
	newEffect := domain.ActiveEffect{
		EffectID:       3101, // STANCE_S
		Value:          value,
		TurnsRemaining: duration,
		SourceID:       caster.ID,
	}
	activeEffects = append(activeEffects, newEffect) // Add the new Stance S

	// --- ⭐️ 6. Marshal and Save (เพิ่ม Error Check) ⭐️ ---
	newEffectsJSON, err := json.Marshal(activeEffects)
	if err != nil {
		s.appLogger.Error("Failed to marshal updated active effects for Stance S buff", err, "target_id", target.ID)
		return // Don't save if marshaling fails
	}
	target.ActiveEffects = newEffectsJSON

	// --- ⭐️ 7. Log (เหมือนเดิม) ⭐️ ---
	s.appLogger.Info("Applied SYNERGY_GRANT_STANCE_S effect", "target", target.ID, "duration", duration)
}

// --- ⭐️ Stance L (Life) ⭐️ ---
func (s *combatService) applySynergyGrantStanceL(caster *domain.Combatant, target *domain.Combatant, effectData map[string]interface{}) {
	// --- 1. Get Value (เหมือนเดิม) ---
	value := 0
	if v, ok := effectData["value"]; ok {
		value = int(v.(float64))
	}

	// --- 2. Get Duration (เหมือนเดิม, already has error check) ---
	durationFloat, ok := effectData["duration"].(float64)
	if !ok {
		s.appLogger.Warn("Invalid or missing duration in effectData for applySynergyGrantStanceL", "data", effectData)
		return
	}
	duration := int(durationFloat)

	// --- 3. Unmarshal Existing Effects (เหมือนเดิม, already has error check) ---
	var activeEffects []domain.ActiveEffect
	if target.ActiveEffects != nil {
		err := json.Unmarshal(target.ActiveEffects, &activeEffects)
		if err != nil {
			s.appLogger.Error("Failed to unmarshal existing active effects for Stance L buff", err, "target_id", target.ID)
			activeEffects = []domain.ActiveEffect{}
		}
	}

	// --- ⭐️ 4. Add "Replace" Logic - Remove ALL existing Stances ⭐️ ---
	var tempEffects []domain.ActiveEffect
	for _, effect := range activeEffects {
		if effect.EffectID < 3101 || effect.EffectID > 3104 {
			tempEffects = append(tempEffects, effect)
		} else {
			s.appLogger.Info("Replacing existing Stance", "target_id", target.ID, "old_stance_id", effect.EffectID)
		}
	}
	activeEffects = tempEffects // Use the filtered list

	// --- 5. Create and Append New Effect (เหมือนเดิม) ---
	newEffect := domain.ActiveEffect{
		EffectID:       3102, // STANCE_L
		Value:          value,
		TurnsRemaining: duration,
		SourceID:       caster.ID,
	}
	activeEffects = append(activeEffects, newEffect)

	// --- ⭐️ 6. Marshal and Save (เพิ่ม Error Check) ⭐️ ---
	newEffectsJSON, err := json.Marshal(activeEffects)
	if err != nil {
		s.appLogger.Error("Failed to marshal updated active effects for Stance L buff", err, "target_id", target.ID)
		return
	}
	target.ActiveEffects = newEffectsJSON

	// --- 7. Log (เหมือนเดิม) ---
	s.appLogger.Info("Applied SYNERGY_GRANT_STANCE_L effect", "target", target.ID, "duration", duration)
}

// --- ⭐️ Stance G (Gale) ⭐️ ---
func (s *combatService) applySynergyGrantStanceG(caster *domain.Combatant, target *domain.Combatant, effectData map[string]interface{}) {
	// --- 1. Get Value (เหมือนเดิม) ---
	value := 0
	if v, ok := effectData["value"]; ok {
		value = int(v.(float64))
	}

	// --- 2. Get Duration (เหมือนเดิม, already has error check) ---
	durationFloat, ok := effectData["duration"].(float64)
	if !ok {
		s.appLogger.Warn("Invalid or missing duration in effectData for applySynergyGrantStanceG", "data", effectData)
		return
	}
	duration := int(durationFloat)

	// --- 3. Unmarshal Existing Effects (เหมือนเดิม, already has error check) ---
	var activeEffects []domain.ActiveEffect
	if target.ActiveEffects != nil {
		err := json.Unmarshal(target.ActiveEffects, &activeEffects)
		if err != nil {
			s.appLogger.Error("Failed to unmarshal existing active effects for Stance G buff", err, "target_id", target.ID)
			activeEffects = []domain.ActiveEffect{}
		}
	}

	// --- ⭐️ 4. Add "Replace" Logic - Remove ALL existing Stances ⭐️ ---
	var tempEffects []domain.ActiveEffect
	for _, effect := range activeEffects {
		if effect.EffectID < 3101 || effect.EffectID > 3104 {
			tempEffects = append(tempEffects, effect)
		} else {
			s.appLogger.Info("Replacing existing Stance", "target_id", target.ID, "old_stance_id", effect.EffectID)
		}
	}
	activeEffects = tempEffects // Use the filtered list

	// --- 5. Create and Append New Effect (เหมือนเดิม) ---
	newEffect := domain.ActiveEffect{
		EffectID:       3103, // STANCE_G
		Value:          value,
		TurnsRemaining: duration,
		SourceID:       caster.ID,
	}
	activeEffects = append(activeEffects, newEffect)

	// --- ⭐️ 6. Marshal and Save (เพิ่ม Error Check) ⭐️ ---
	newEffectsJSON, err := json.Marshal(activeEffects)
	if err != nil {
		s.appLogger.Error("Failed to marshal updated active effects for Stance G buff", err, "target_id", target.ID)
		return
	}
	target.ActiveEffects = newEffectsJSON

	// --- 7. Log (เหมือนเดิม) ---
	s.appLogger.Info("Applied SYNERGY_GRANT_STANCE_G effect", "target", target.ID, "duration", duration)
}

// --- ⭐️ Stance P (Precision) ⭐️ ---
func (s *combatService) applySynergyGrantStanceP(caster *domain.Combatant, target *domain.Combatant, effectData map[string]interface{}) {
	// --- 1. Get Value (เหมือนเดิม) ---
	value := 0
	if v, ok := effectData["value"]; ok {
		value = int(v.(float64))
	}

	// --- 2. Get Duration (เหมือนเดิม, already has error check) ---
	durationFloat, ok := effectData["duration"].(float64)
	if !ok {
		s.appLogger.Warn("Invalid or missing duration in effectData for applySynergyGrantStanceP", "data", effectData)
		return
	}
	duration := int(durationFloat)

	// --- 3. Unmarshal Existing Effects (เหมือนเดิม, already has error check) ---
	var activeEffects []domain.ActiveEffect
	if target.ActiveEffects != nil {
		err := json.Unmarshal(target.ActiveEffects, &activeEffects)
		if err != nil {
			s.appLogger.Error("Failed to unmarshal existing active effects for Stance P buff", err, "target_id", target.ID)
			activeEffects = []domain.ActiveEffect{}
		}
	}

	// --- ⭐️ 4. Add "Replace" Logic - Remove ALL existing Stances ⭐️ ---
	var tempEffects []domain.ActiveEffect
	for _, effect := range activeEffects {
		if effect.EffectID < 3101 || effect.EffectID > 3104 {
			tempEffects = append(tempEffects, effect)
		} else {
			s.appLogger.Info("Replacing existing Stance", "target_id", target.ID, "old_stance_id", effect.EffectID)
		}
	}
	activeEffects = tempEffects // Use the filtered list

	// --- 5. Create and Append New Effect (เหมือนเดิม) ---
	newEffect := domain.ActiveEffect{
		EffectID:       3104, // STANCE_P
		Value:          value,
		TurnsRemaining: duration,
		SourceID:       caster.ID,
	}
	activeEffects = append(activeEffects, newEffect)

	// --- ⭐️ 6. Marshal and Save (เพิ่ม Error Check) ⭐️ ---
	newEffectsJSON, err := json.Marshal(activeEffects)
	if err != nil {
		s.appLogger.Error("Failed to marshal updated active effects for Stance P buff", err, "target_id", target.ID)
		return
	}
	target.ActiveEffects = newEffectsJSON

	// --- 7. Log (เหมือนเดิม) ---
	s.appLogger.Info("Applied SYNERGY_GRANT_STANCE_P effect", "target", target.ID, "duration", duration)
}
