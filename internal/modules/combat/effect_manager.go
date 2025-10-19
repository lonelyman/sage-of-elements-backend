// file: internal/modules/combat/effect_manager.go
package combat

import (
	"encoding/json"
	"sage-of-elements-backend/internal/domain"
)

// --- นี่คือบ้านใหม่ของผู้เชี่ยวชาญด้านสถานะ ---

func (s *combatService) applyEffect(caster *domain.Combatant, target *domain.Combatant, effectData map[string]interface{}, spell *domain.Spell) {
	effectID := uint(effectData["effect_id"].(float64))

	switch effectID {
	case 1: // DAMAGE
		tempSpellEffect := &domain.SpellEffect{
			BaseValue: effectData["value"].(float64),
		}
		calculatedDamage, err := s.calculateEffectValue(caster, target, spell, tempSpellEffect)
		if err != nil {
			s.appLogger.Error("Error calculating damage value", err)
			return
		}
		damageToInt := int(calculatedDamage)
		target.CurrentHP -= damageToInt
		s.appLogger.Info("Applied DAMAGE effect", "caster", caster.ID, "target", target.ID, "damage", damageToInt, "target_hp_remaining", target.CurrentHP)

	case 301: // DEBUFF_SLOW
		value := int(effectData["value"].(float64))
		duration := int(effectData["duration"].(float64))
		var activeEffects []domain.ActiveEffect
		if target.ActiveEffects != nil {
			json.Unmarshal(target.ActiveEffects, &activeEffects)
		}
		newEffect := domain.ActiveEffect{
			EffectID:       effectID,
			Value:          value,
			TurnsRemaining: duration,
			SourceID:       caster.ID,
		}
		activeEffects = append(activeEffects, newEffect)
		newEffectsJSON, _ := json.Marshal(activeEffects)
		target.ActiveEffects = newEffectsJSON
		s.appLogger.Info("Applied DEBUFF_SLOW effect", "caster", caster.ID, "target", target.ID, "duration", duration)

	default:
		s.appLogger.Warn("Attempted to apply an unknown or unimplemented effect", "effect_id", effectID)
	}
}

func (s *combatService) processEndOfTurnEffects(combatant *domain.Combatant) {
	if combatant.ActiveEffects == nil {
		return
	}
	var currentEffects []domain.ActiveEffect
	json.Unmarshal(combatant.ActiveEffects, &currentEffects)
	if len(currentEffects) == 0 {
		return
	}
	var remainingEffects []domain.ActiveEffect
	for _, effect := range currentEffects {
		effect.TurnsRemaining--
		if effect.TurnsRemaining > 0 {
			remainingEffects = append(remainingEffects, effect)
		} else {
			s.appLogger.Info("Effect has expired", "combatant_id", combatant.ID, "effect_id", effect.EffectID)
		}
	}
	newEffectsJSON, _ := json.Marshal(remainingEffects)
	combatant.ActiveEffects = newEffectsJSON
}

func (s *combatService) recalculateStats(combatant *domain.Combatant) {
	var baseInitiative int
	if combatant.CharacterID != nil && combatant.Character != nil { // ⭐️ เพิ่ม nil check
		baseInitiative = 50 + combatant.Character.TalentG
	} else if combatant.EnemyID != nil && combatant.Enemy != nil { // ⭐️ เพิ่ม nil check
		baseInitiative = combatant.Enemy.Initiative
	}

	var activeEffects []domain.ActiveEffect
	if combatant.ActiveEffects != nil {
		json.Unmarshal(combatant.ActiveEffects, &activeEffects)
	}
	modifiedInitiative := baseInitiative

	for _, effect := range activeEffects {
		if effect.EffectID == 301 {
			modifiedInitiative += effect.Value
		}
	}
	combatant.Initiative = modifiedInitiative
	s.appLogger.Info("Stats recalculated for combatant", "id", combatant.ID, "base_init", baseInitiative, "modified_init", modifiedInitiative)
}
