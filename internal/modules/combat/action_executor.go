// file: internal/modules/combat/action_executor.go
package combat

import (
	"sage-of-elements-backend/internal/domain"
	"sage-of-elements-backend/pkg/apperrors"

	"github.com/gofrs/uuid"
)

func (s *combatService) executeCastSpell(caster *domain.Combatant, match *domain.CombatMatch, req PerformActionRequest) error {
	// [ด่านที่ 1: การตรวจสอบ]
	if req.SpellID == nil || req.TargetID == nil {
		return apperrors.InvalidFormatError("spell_id and target_id are required", nil)
	}
	spellID, targetUUIDStr := *req.SpellID, *req.TargetID
	targetUUID, _ := uuid.FromString(targetUUIDStr)
	spell, _ := s.gameDataRepo.FindSpellByID(spellID)
	if spell == nil {
		return apperrors.NotFoundError("spell not found")
	}

	targetIndex := -1
	for i, c := range match.Combatants {
		if c.ID == targetUUID {
			targetIndex = i
			break
		}
	}
	if targetIndex == -1 {
		return apperrors.NotFoundError("target not found")
	}
	target := match.Combatants[targetIndex]

	if caster.CurrentAP < spell.APCost {
		return apperrors.New(422, "INSUFFICIENT_AP", "not enough AP")
	}
	if caster.CurrentMP < spell.MPCost {
		return apperrors.New(422, "INSUFFICIENT_MP", "not enough MP")
	}

	// --- ✨⭐️ นี่คือ "กฎ Hotbar 12 ช่อง" ที่แท้จริง! ⭐️✨ ---
	// [ด่านที่ 1.5: ตรวจสอบ "วัตถุดิบ"]
	requiredElementID := spell.ElementID
	if requiredElementID <= 4 {
		// --- นี่คือเวท T0 (เช่น EarthSlam) ---
		// กฎคือ: ใช้ได้เลย! ไม่ต้องเช็ค "มือ" หรือ "Deck"!
		s.appLogger.Info("Casting a T0 spell. No element consumed.", "spell", spell.Name)
	} else {
		// --- นี่คือเวท T1 (เช่น MoltenMeteor) ---
		s.appLogger.Info("Casting a T1 spell. Checking charges in Deck...", "spell", spell.Name)

		// ไปที่ "คลังกระสุน" (Deck) ของผู้เล่น
		chargeIndex := -1
		for i, charge := range caster.Deck {
			// ⭐️ เช็คว่า 1.ธาตุตรง 2.ยังไม่ถูกใช้
			if charge.ElementID == requiredElementID && !charge.IsConsumed {
				chargeIndex = i // เจอ "กระสุน" นัดที่ยังไม่ได้ใช้!
				break
			}
		}

		if chargeIndex == -1 {
			// ถ้าหาไม่เจอ...
			return apperrors.New(422, "INSUFFICIENT_CHARGES", "no unconsumed charges of this T1 element left in deck")
		}

		// "ลั่นไก!" (ใช้กระสุนนัดนั้นไป)
		caster.Deck[chargeIndex].IsConsumed = true
		s.appLogger.Info("Consumed a T1 charge", "element_id", requiredElementID, "charge_index", chargeIndex)
	}
	// ------------------------------------

	// [ด่านที่ 2 & 3: การคำนวณและลงมือ] (เหมือนเดิม)
	s.appLogger.Info("Player is casting spell", "spell", spell.Name, "target_id", target.ID)
	for _, spellEffect := range spell.Effects {
		effectData := map[string]interface{}{
			"effect_id": float64(spellEffect.Effect.ID),
			"value":     float64(spellEffect.BaseValue),
			"duration":  float64(spellEffect.DurationInTurns),
		}
		s.applyEffect(caster, target, effectData, spell)
	}

	// [ด่านที่ 4: หักทรัพยากรผู้ร่าย]
	caster.CurrentAP -= spell.APCost
	caster.CurrentMP -= spell.MPCost

	s.appLogger.Info("Player resources updated", "ap", caster.CurrentAP, "mp", caster.CurrentMP)
	return nil
}
