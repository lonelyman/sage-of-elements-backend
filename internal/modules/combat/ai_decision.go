// file: internal/modules/combat/ai_decision.go
package combat

import (
	"sage-of-elements-backend/internal/domain"
)

// ==================== AI Decision Making ====================
// ไฟล์นี้รวบรวมฟังก์ชันที่เกี่ยวกับการตัดสินใจของ AI
// - การเช็คเงื่อนไข (condition checking)
// - การเลือก action (action selection)
// - การตรวจสอบทรัพยากร (resource validation)

// AIDecisionContext เก็บข้อมูลที่ AI ต้องการในการตัดสินใจ
type AIDecisionContext struct {
	AICombatant    *domain.Combatant
	PlayerTarget   *domain.Combatant
	Match          *domain.CombatMatch
	AvailableRules []*domain.EnemyAI
}

// AISelectedAction เก็บผลลัพธ์ของการตัดสินใจ
type AISelectedAction struct {
	Ability *domain.EnemyAbility
	Target  *domain.Combatant
	Rule    *domain.EnemyAI
}

// SelectNextAction เป็นฟังก์ชันหลักในการเลือก action ที่ AI จะทำ
// จะวน loop ตามลำดับ priority ของ rules จนกว่าจะเจอ action ที่ทำได้
func (s *combatService) SelectNextAction(ctx *AIDecisionContext) *AISelectedAction {
	s.appLogger.Debug("🤖 AI starting action selection",
		"ai_id", ctx.AICombatant.ID,
		"current_ap", ctx.AICombatant.CurrentAP,
		"current_mp", ctx.AICombatant.CurrentMP,
		"rules_count", len(ctx.AvailableRules),
	)

	// วน loop ตาม priority (rules ถูกเรียงจากน้อยไปมากแล้ว)
	for _, rule := range ctx.AvailableRules {
		// 1. เช็คเงื่อนไข
		if !s._EvaluateCondition(ctx, rule) {
			s.appLogger.Debug("AI rule condition not met, skipping",
				"priority", rule.Priority,
				"condition", rule.Condition,
			)
			continue
		}

		// 2. เช็คว่ามี ability ให้ใช้หรือไม่
		if rule.AbilityToUse == nil {
			s.appLogger.Warn("AI rule has no ability, skipping",
				"priority", rule.Priority,
			)
			continue
		}

		// 3. เช็คทรัพยากร (AP, MP)
		if !s._CanAffordAbility(ctx.AICombatant, rule.AbilityToUse) {
			s.appLogger.Debug("AI cannot afford ability, checking next rule",
				"priority", rule.Priority,
				"ability", rule.AbilityToUse.Name,
				"need_ap", rule.AbilityToUse.APCost,
				"need_mp", rule.AbilityToUse.MPCost,
				"have_ap", ctx.AICombatant.CurrentAP,
				"have_mp", ctx.AICombatant.CurrentMP,
			)
			continue
		}

		// 4. กำหนดเป้าหมาย
		target := s._DetermineTarget(ctx, rule)
		if target == nil {
			s.appLogger.Warn("AI could not determine target, skipping",
				"priority", rule.Priority,
			)
			continue
		}

		// 5. เจอ action ที่ใช้ได้แล้ว!
		s.appLogger.Info("✅ AI selected action",
			"ability", rule.AbilityToUse.Name,
			"priority", rule.Priority,
			"target", target.ID,
			"ap_cost", rule.AbilityToUse.APCost,
			"mp_cost", rule.AbilityToUse.MPCost,
		)

		return &AISelectedAction{
			Ability: rule.AbilityToUse,
			Target:  target,
			Rule:    rule,
		}
	}

	// ไม่เจอ action ไหนที่ทำได้เลย
	s.appLogger.Info("❌ AI found no valid action",
		"ai_id", ctx.AICombatant.ID,
		"ap_remaining", ctx.AICombatant.CurrentAP,
	)
	return nil
}

// ==================== Condition Evaluation ====================

// _EvaluateCondition เช็คว่าเงื่อนไขของ rule นี้เป็นจริงหรือไม่
func (s *combatService) _EvaluateCondition(ctx *AIDecisionContext, rule *domain.EnemyAI) bool {
	switch rule.Condition {
	case domain.AIConditionAlways:
		return s._CheckAlways()

	case domain.AIConditionTurnIs:
		return s._CheckTurnIs(ctx.Match, rule.ConditionValue)

	case domain.AIConditionSelfHPBelow:
		return s._CheckSelfHPBelow(ctx.AICombatant, rule.ConditionValue)

	default:
		s.appLogger.Warn("Unknown AI condition type",
			"condition", rule.Condition,
		)
		return false
	}
}

// _CheckAlways - เงื่อนไขที่เป็นจริงเสมอ
func (s *combatService) _CheckAlways() bool {
	return true
}

// _CheckTurnIs - เช็คว่าเป็นเทิร์นที่กำหนดหรือไม่
func (s *combatService) _CheckTurnIs(match *domain.CombatMatch, targetTurn float64) bool {
	return float64(match.TurnNumber) == targetTurn
}

// _CheckSelfHPBelow - เช็คว่า HP ของตัวเองต่ำกว่าเปอร์เซ็นต์ที่กำหนดหรือไม่
func (s *combatService) _CheckSelfHPBelow(combatant *domain.Combatant, threshold float64) bool {
	if combatant.Enemy == nil || combatant.Enemy.MaxHP == 0 {
		return false
	}

	currentHPRatio := float64(combatant.CurrentHP) / float64(combatant.Enemy.MaxHP)
	return currentHPRatio <= threshold
}

// ==================== Resource Validation ====================

// _CanAffordAbility เช็คว่า AI มีทรัพยากรพอที่จะใช้ ability นี้หรือไม่
func (s *combatService) _CanAffordAbility(combatant *domain.Combatant, ability *domain.EnemyAbility) bool {
	hasEnoughAP := combatant.CurrentAP >= ability.APCost
	hasEnoughMP := combatant.CurrentMP >= ability.MPCost

	return hasEnoughAP && hasEnoughMP
}

// ==================== Target Selection ====================

// _DetermineTarget กำหนดเป้าหมายของ action ตาม rule
func (s *combatService) _DetermineTarget(ctx *AIDecisionContext, rule *domain.EnemyAI) *domain.Combatant {
	switch rule.Target {
	case domain.AITargetPlayer:
		return ctx.PlayerTarget

	case domain.AITargetSelf:
		return ctx.AICombatant

	default:
		// Default ให้ target เป็น player
		s.appLogger.Debug("Unknown AI target type, defaulting to player",
			"target_type", rule.Target,
		)
		return ctx.PlayerTarget
	}
}
