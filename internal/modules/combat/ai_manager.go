// file: internal/modules/combat/ai_manager.go
package combat

import (
	"encoding/json"
	"sage-of-elements-backend/internal/domain"
	"sage-of-elements-backend/pkg/apperrors"
	"sort"
)

// --- นี่คือบ้านใหม่ของผู้เชี่ยวชาญด้าน AI ---

func (s *combatService) processAITurn(match *domain.CombatMatch, aiCombatant *domain.Combatant) (*domain.CombatMatch, error) {
	if aiCombatant.Enemy == nil {
		s.appLogger.Warn("AI Combatant has no Enemy data loaded, skipping turn.", "ai_combatant_id", aiCombatant.ID)
		match = s.endTurn(match)
		match, err := s.startNewTurn(match)
		if err != nil {
			return nil, err
		}
		return match, nil
	}

	aiRules := aiCombatant.Enemy.AI
	sort.Slice(aiRules, func(i, j int) bool {
		return aiRules[i].Priority < aiRules[j].Priority
	})

	var target *domain.Combatant
	for _, c := range match.Combatants {
		if c.CharacterID != nil {
			target = c
			break
		}
	}
	if target == nil {
		return nil, apperrors.SystemError("player target not found in match")
	}

	var actionToPerform *domain.EnemyAbility
	for _, rule := range aiRules {
		conditionMet := false
		switch rule.Condition {
		case domain.AIConditionAlways:
			conditionMet = true
		case domain.AIConditionTurnIs:
			if float64(match.TurnNumber) == rule.ConditionValue {
				conditionMet = true
			}
			// TODO: เพิ่ม case domain.AIConditionSelfHPBelow
		}

		if conditionMet {
			actionToPerform = rule.AbilityToUse
			s.appLogger.Info("AI decided to use ability", "ability_name", actionToPerform.Name, "rule_priority", rule.Priority)
			break
		}
	}

	if actionToPerform != nil {
		if aiCombatant.CurrentAP < actionToPerform.APCost {
			s.appLogger.Warn("AI has insufficient AP for chosen action, passing turn.", "ability", actionToPerform.Name)
		} else {
			aiCombatant.CurrentAP -= actionToPerform.APCost
			dummySpell := &domain.Spell{
				ElementID: aiCombatant.Enemy.ElementID,
			}
			var effects []map[string]interface{}
			json.Unmarshal(actionToPerform.EffectsJSON, &effects)
			for _, effectData := range effects {
				s.applyEffect(aiCombatant, target, effectData, dummySpell)
			}
		}
	}

	s.appLogger.Info("AI ended its turn.", "match_id", match.ID)
	match = s.endTurn(match)

	match, err := s.startNewTurn(match)
	if err != nil {
		return nil, err
	}

	return match, nil
}
