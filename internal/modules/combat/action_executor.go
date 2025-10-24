// file: internal/modules/combat/action_executor.go
package combat

import (
	"fmt"
	"math"
	"sage-of-elements-backend/internal/domain"
	"sage-of-elements-backend/pkg/apperrors"
	"strconv"
)

// --- "ผู้เชี่ยวชาญ" ของเรามีแค่คนเดียวในไฟล์นี้! ---

// executeCastSpell คือ "สมอง" ที่ทำงานตาม "กฎ Hotbar 12 ช่อง"
// executeCastSpell คือ "สมอง" ที่ทำงานตาม "กฎ Hotbar 12 ช่อง"
func (s *combatService) executeCastSpell(caster *domain.Combatant, match *domain.CombatMatch, req PerformActionRequest) error {
	// [ด่านที่ 1: การตรวจสอบ]
	if req.SpellID == nil || req.TargetID == nil {
		return apperrors.InvalidFormatError("spell_id and target_id are required", nil)
	}
	spellID, targetUUIDStr := *req.SpellID, *req.TargetID
	spell, err := s.gameDataRepo.FindSpellByID(spellID)
	if err != nil {
		s.appLogger.Error("Failed to find spell by ID", err, "spell_id", spellID)
		return apperrors.SystemError("failed to retrieve spell data")
	}
	if spell == nil {
		return apperrors.NotFoundError("spell not found")
	}

	targetIndex := -1
	for i, c := range match.Combatants {
		if c.ID.String() == targetUUIDStr {
			targetIndex = i
			break
		}
	}
	if targetIndex == -1 {
		return apperrors.NotFoundError("target not found")
	}
	target := match.Combatants[targetIndex]

	// --- ✨⭐️ ตรวจสอบ Targeting! ⭐️✨ ---
	isValidTarget := true
	casterIDStr := caster.ID.String()

	switch spell.TargetType {
	case domain.TargetTypeSelf:
		if targetUUIDStr != casterIDStr {
			isValidTarget = false
			s.appLogger.Warn("Invalid target for SELF spell", "spell", spell.Name, "caster", casterIDStr, "target", targetUUIDStr)
		}
	case domain.TargetTypeEnemy:
		if targetUUIDStr == casterIDStr {
			isValidTarget = false
			s.appLogger.Warn("Invalid target for ENEMY spell (cannot target self)", "spell", spell.Name, "caster", casterIDStr, "target", targetUUIDStr)
		} // else { ... (optional enemy check) ... }
	case domain.TargetTypeAlly:
		if targetUUIDStr != casterIDStr {
			isActuallyAlly := target.CharacterID != nil
			if !isActuallyAlly {
				isValidTarget = false
				s.appLogger.Warn("Invalid target for ALLY spell (target is not an ally)", "spell", spell.Name, "caster", casterIDStr, "target", targetUUIDStr)
			}
		}
	default:
		s.appLogger.Warn("Unknown or unhandled TargetType in validation", "spell", spell.Name, "target_type", spell.TargetType)
	}

	if !isValidTarget {
		displaySpellName := spell.Name // ... (logicหา display name เหมือนเดิม) ...
		if nameTH, ok := spell.DisplayNames["th"].(string); ok && nameTH != "" {
			displaySpellName = nameTH
		} else if nameEN, ok := spell.DisplayNames["en"].(string); ok && nameEN != "" {
			displaySpellName = nameEN
		}
		return apperrors.New(422, "INVALID_TARGET", fmt.Sprintf("เลือกเป้าหมายสำหรับเวท '%s' ไม่ถูกต้อง", displaySpellName))
	}
	// --- ✨⭐️ สิ้นสุด Targeting ⭐️✨ ---

	// --- [ด่านที่ 2: ดึง Config Cast Mode] ---
	// ... (ดึง ocApMod, ocMpMod etc. เหมือนเดิม) ...
	ocApModStr, _ := s.gameDataRepo.GetGameConfigValue("CAST_MODE_OVERCHARGE_AP_MOD")
	ocMpModStr, _ := s.gameDataRepo.GetGameConfigValue("CAST_MODE_OVERCHARGE_MP_MOD")
	ocPowerModStr, _ := s.gameDataRepo.GetGameConfigValue("CAST_MODE_OVERCHARGE_POWER_MOD")
	chApModStr, _ := s.gameDataRepo.GetGameConfigValue("CAST_MODE_CHARGE_AP_MOD")
	chMpModStr, _ := s.gameDataRepo.GetGameConfigValue("CAST_MODE_CHARGE_MP_MOD")
	chPowerModStr, _ := s.gameDataRepo.GetGameConfigValue("CAST_MODE_CHARGE_POWER_MOD")
	ocApMod, _ := strconv.ParseFloat(ocApModStr, 64)
	ocMpMod, _ := strconv.ParseFloat(ocMpModStr, 64)
	ocPowerMod, _ := strconv.ParseFloat(ocPowerModStr, 64)
	chApMod, _ := strconv.ParseFloat(chApModStr, 64)
	chMpMod, _ := strconv.ParseFloat(chMpModStr, 64)
	chPowerMod, _ := strconv.ParseFloat(chPowerModStr, 64)

	// [ด่านที่ 3: คำนวณ Cost & Power Modifier]
	// ... (คำนวณ actualAPCost, actualMPCost, powerModifier เหมือนเดิม) ...
	actualAPCost := spell.APCost
	actualMPCost := spell.MPCost
	powerModifier := 1.0
	castMode := "INSTANT"
	if req.CastMode != "" {
		castMode = req.CastMode
	}
	switch castMode {
	case "OVERCHARGE":
		actualAPCost = int(math.Ceil(float64(spell.APCost) * ocApMod))
		actualMPCost = int(math.Ceil(float64(spell.MPCost) * ocMpMod))
		powerModifier = ocPowerMod
	case "CHARGE":
		actualAPCost = int(math.Ceil(float64(spell.APCost) * chApMod))
		actualMPCost = int(math.Ceil(float64(spell.MPCost) * chMpMod))
		powerModifier = chPowerMod
	case "INSTANT": // default
	default:
		s.appLogger.Warn("Unknown CastMode received, defaulting to INSTANT", "received_mode", castMode)
		castMode = "INSTANT"
	}
	s.appLogger.Info("Cast Mode calculated", "mode", castMode, "final_ap", actualAPCost, "final_mp", actualMPCost, "power_mod", powerModifier)

	// [ด่านที่ 4: ตรวจสอบทรัพยากร]
	if caster.CurrentAP < actualAPCost { /* return error */
		return apperrors.New(422, "INSUFFICIENT_AP", fmt.Sprintf("AP ไม่พอ (ต้องการ: %d, มี: %d)", actualAPCost, caster.CurrentAP))
	}
	if caster.CurrentMP < actualMPCost { /* return error */
		return apperrors.New(422, "INSUFFICIENT_MP", fmt.Sprintf("MP ไม่พอ (ต้องการ: %d, มี: %d)", actualMPCost, caster.CurrentMP))
	}

	// --- ✨⭐️ [ด่านที่ 4.5: หักทรัพยากรผู้ร่าย *ก่อน*!] ⭐️✨ ---
	casterMpBeforeCast := caster.CurrentMP // เก็บ MP ก่อนหัก ไว้เผื่อ Debug
	caster.CurrentAP -= actualAPCost
	caster.CurrentMP -= actualMPCost
	s.appLogger.Info("Player resources deducted for cast", "caster_id", caster.ID, "ap_cost", actualAPCost, "mp_cost", actualMPCost, "ap_left", caster.CurrentAP, "mp_left_after_cost", caster.CurrentMP)
	// ---------------------------------------------

	// [ด่านที่ 5: ตรวจสอบ Element Charges (ถ้ามี)]
	// ... (Logic เช็ค T0/T1+ เหมือนเดิม) ...
	requiredElementID := spell.ElementID
	if requiredElementID <= 4 {
		s.appLogger.Info("Casting a T0 spell. No element charge consumed.", "spell", spell.Name)
	} else {
		s.appLogger.Warn("T1+ Spell charge consumption logic not implemented yet", "spell_id", spellID)
		// TODO: Implement charge consumption
	}

	// [ด่านที่ 6: ลงมือ!]
	s.appLogger.Info("Player is casting spell", "spell", spell.Name, "mode", castMode, "caster", caster.ID, "target", target.ID)

	for _, spellEffect := range spell.Effects {
		// ... (โค้ดในลูปเหมือนเดิม: ดึง effectInfo, สร้าง effectData, หา finalTarget) ...
		effectInfo, _ := s.gameDataRepo.FindEffectByID(spellEffect.EffectID)
		if effectInfo == nil {
			s.appLogger.Warn("Could not find effect info for EffectID", "effect_id", spellEffect.EffectID)
			continue
		}
		effectData := map[string]interface{}{
			"effect_id":      float64(spellEffect.EffectID),
			"value":          spellEffect.BaseValue,
			"duration":       float64(spellEffect.DurationInTurns),
			"power_modifier": powerModifier,
			"cast_mode":      castMode,
		}
		var finalTarget *domain.Combatant
		switch spell.TargetType {
		case domain.TargetTypeSelf:
			finalTarget = caster
		case domain.TargetTypeEnemy:
			finalTarget = target
		case domain.TargetTypeAlly:
			finalTarget = target // Assume target chosen is correct for now
		default:
			finalTarget = target
		}
		if effectInfo.Type == domain.EffectTypeSynergyBuff {
			finalTarget = caster
			s.appLogger.Info("Synergy Buff detected, overriding target to caster", "effect_id", spellEffect.EffectID)
		}

		// เรียก applyEffect (ซึ่งจะไปเรียก applyMpDamage และอาจจะเพิ่ม MP ให้ caster)
		s.applyEffect(caster, finalTarget, effectData, spell)
	}

	// [ด่านที่ 7: หักทรัพยากรผู้ร่าย] <-- ไม่มีแล้ว!

	// ⭐️ Log สุดท้าย - เอา casterMpBeforeCast มาใช้! ⭐️
	s.appLogger.Info("Player resources updated after effects", "caster_id", caster.ID, "ap_left", caster.CurrentAP, "mp_before_cast", casterMpBeforeCast, "mp_cost", actualMPCost, "final_mp", caster.CurrentMP) // <-- เพิ่ม mp_before_cast, mp_cost
	return nil
}
