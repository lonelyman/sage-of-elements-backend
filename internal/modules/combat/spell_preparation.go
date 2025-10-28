// file: internal/modules/combat/spell_preparation.go
package combat

import (
	"fmt"
	"sage-of-elements-backend/internal/domain"
	"sage-of-elements-backend/pkg/apperrors"
	"strconv"

	"github.com/gofrs/uuid"
)

// SpellPreparationResult เก็บผลลัพธ์จากการเตรียมการร่ายเวท
type SpellPreparationResult struct {
	Spell           *domain.Spell
	Caster          *domain.Combatant
	Target          *domain.Combatant
	FinalAPCost     int
	FinalMPCost     int
	PowerModifier   float64
	ConsumedCharges []uint // Element IDs ที่ถูก consume จาก deck
}

// PrepareAndValidateCast เป็น Step 1 ของการร่ายเวท
// รับผิดชอบ: ดึงข้อมูล, ตรวจสอบ, คำนวณ cost, หักทรัพยากร
func (s *combatService) PrepareAndValidateCast(
	match *domain.CombatMatch,
	caster *domain.Combatant,
	targetID uuid.UUID,
	spellID uint,
	castingMode string,
) (*SpellPreparationResult, error) {

	s.appLogger.Info("🎯 STEP 1: Preparing spell cast",
		"caster_id", caster.ID,
		"target_id", targetID,
		"spell_id", spellID,
		"casting_mode", castingMode,
	)

	// 1.1 Fetch Spell Data
	spell, err := s._FetchSpellData(spellID)
	if err != nil {
		return nil, err
	}

	// 1.2 Find Target
	target := s._FindTarget(match, targetID)
	if target == nil {
		return nil, apperrors.NotFoundError("target not found in this match")
	}

	// 1.3 Validate Targeting Rules
	if err := s._ValidateTargeting(spell, caster.ID, targetID, target); err != nil {
		return nil, err
	}

	// 1.4 Calculate Final Cost & Power Modifier
	finalAP, finalMP, powerMod, err := s._CalculateFinalCost(spell.APCost, spell.MPCost, castingMode)
	if err != nil {
		return nil, err
	}

	// 1.5 Validate & Deduct Resources (AP/MP)
	if err := s._ValidateAndDeductResources(caster, finalAP, finalMP); err != nil {
		return nil, err
	}

	// 1.6 Validate & Consume Element Charges (for T1+ spells)
	consumedCharges, err := s._ValidateAndConsumeCharges(caster, spell)
	if err != nil {
		return nil, err
	}

	s.appLogger.Info("✅ STEP 1 Complete: Spell preparation successful",
		"spell_name", spell.Name,
		"final_ap_cost", finalAP,
		"final_mp_cost", finalMP,
		"power_modifier", powerMod,
		"consumed_charges", consumedCharges,
	)

	return &SpellPreparationResult{
		Spell:           spell,
		Caster:          caster,
		Target:          target,
		FinalAPCost:     finalAP,
		FinalMPCost:     finalMP,
		PowerModifier:   powerMod,
		ConsumedCharges: consumedCharges,
	}, nil
}

// ==================== Sub-functions ====================

// _FetchSpellData ดึงข้อมูล spell จาก database
// NOTE: Frontend จะเรียก /resolve-spell ก่อนแล้วส่ง spell_id มาให้
func (s *combatService) _FetchSpellData(spellID uint) (*domain.Spell, error) {
	spell, err := s.gameDataRepo.FindSpellByID(spellID)
	if err != nil {
		s.appLogger.Error("Failed to find spell by ID", err, "spell_id", spellID)
		return nil, apperrors.SystemError("failed to retrieve spell data")
	}
	if spell == nil {
		return nil, apperrors.NotFoundError("spell not found")
	}
	return spell, nil
}

// _FindTarget หา combatant ตาม UUID
func (s *combatService) _FindTarget(match *domain.CombatMatch, targetID uuid.UUID) *domain.Combatant {
	return s.findCombatantByID(match, targetID)
}

// _ValidateTargeting ตรวจสอบว่า target ตรงกับ spell.TargetType หรือไม่
func (s *combatService) _ValidateTargeting(
	spell *domain.Spell,
	casterID uuid.UUID,
	targetID uuid.UUID,
	target *domain.Combatant,
) error {
	isValidTarget := true
	casterIDStr := casterID.String()
	targetIDStr := targetID.String()

	switch spell.TargetType {
	case domain.TargetTypeSelf:
		if targetIDStr != casterIDStr {
			isValidTarget = false
			s.appLogger.Warn("Invalid target for SELF spell",
				"spell", spell.Name,
				"caster", casterIDStr,
				"target", targetIDStr,
			)
		}

	case domain.TargetTypeEnemy:
		if targetIDStr == casterIDStr {
			isValidTarget = false
			s.appLogger.Warn("Invalid target for ENEMY spell (cannot target self)",
				"spell", spell.Name,
				"caster", casterIDStr,
				"target", targetIDStr,
			)
		}
		// Optional: เช็คว่า target เป็นศัตรูจริงๆ
		// if target.CharacterID != nil { ... }

	case domain.TargetTypeAlly:
		// Target ต้องเป็น self หรือพันธมิตร
		if targetIDStr != casterIDStr {
			isActuallyAlly := target.CharacterID != nil
			if !isActuallyAlly {
				isValidTarget = false
				s.appLogger.Warn("Invalid target for ALLY spell (target is not an ally)",
					"spell", spell.Name,
					"caster", casterIDStr,
					"target", targetIDStr,
				)
			}
		}

	default:
		s.appLogger.Warn("Unknown or unhandled TargetType in validation",
			"spell", spell.Name,
			"target_type", spell.TargetType,
		)
	}

	if !isValidTarget {
		displaySpellName := spell.Name
		if nameTH, ok := spell.DisplayNames["th"].(string); ok && nameTH != "" {
			displaySpellName = nameTH
		} else if nameEN, ok := spell.DisplayNames["en"].(string); ok && nameEN != "" {
			displaySpellName = nameEN
		}
		return apperrors.New(422, "INVALID_TARGET",
			fmt.Sprintf("เลือกเป้าหมายสำหรับเวท '%s' ไม่ถูกต้อง", displaySpellName))
	}

	return nil
}

// _CalculateFinalCost คำนวณ AP/MP cost ตาม casting mode
func (s *combatService) _CalculateFinalCost(
	baseAP int,
	baseMP int,
	castingMode string,
) (finalAP int, finalMP int, powerMod float64, err error) {

	// Default values
	finalAP = baseAP
	finalMP = baseMP
	powerMod = 1.0

	// ถ้าเป็น INSTANT ให้คืนค่า default
	if castingMode == "" || castingMode == "INSTANT" {
		return finalAP, finalMP, powerMod, nil
	}

	// ดึง config values
	var apAddStr, mpAddStr, powerModStr string

	switch castingMode {
	case "CHARGE":
		apAddStr, _ = s.gameDataRepo.GetGameConfigValue("CAST_MODE_CHARGE_AP_ADD")
		mpAddStr, _ = s.gameDataRepo.GetGameConfigValue("CAST_MODE_CHARGE_MP_ADD")
		powerModStr, _ = s.gameDataRepo.GetGameConfigValue("CAST_MODE_CHARGE_POWER_MOD")

	case "OVERCHARGE":
		apAddStr, _ = s.gameDataRepo.GetGameConfigValue("CAST_MODE_OVERCHARGE_AP_ADD")
		mpAddStr, _ = s.gameDataRepo.GetGameConfigValue("CAST_MODE_OVERCHARGE_MP_ADD")
		powerModStr, _ = s.gameDataRepo.GetGameConfigValue("CAST_MODE_OVERCHARGE_POWER_MOD")

	default:
		s.appLogger.Warn("Unknown CastMode received, defaulting to INSTANT", "received_mode", castingMode)
		return finalAP, finalMP, powerMod, nil
	}

	// Parse values
	apAdd, _ := strconv.Atoi(apAddStr)
	mpAdd, _ := strconv.Atoi(mpAddStr)
	powerMod, _ = strconv.ParseFloat(powerModStr, 64)

	// Apply additive cost
	finalAP = baseAP + apAdd
	finalMP = baseMP + mpAdd

	s.appLogger.Info("Casting mode cost calculated",
		"mode", castingMode,
		"base_ap", baseAP,
		"base_mp", baseMP,
		"final_ap", finalAP,
		"final_mp", finalMP,
		"power_mod", powerMod,
	)

	return finalAP, finalMP, powerMod, nil
}

// _ValidateAndDeductResources ตรวจสอบและหัก AP/MP
func (s *combatService) _ValidateAndDeductResources(
	caster *domain.Combatant,
	apCost int,
	mpCost int,
) error {
	// Validate AP
	if caster.CurrentAP < apCost {
		return apperrors.New(422, "INSUFFICIENT_AP",
			fmt.Sprintf("AP ไม่พอ (ต้องการ: %d, มี: %d)", apCost, caster.CurrentAP))
	}

	// Validate MP
	if caster.CurrentMP < mpCost {
		return apperrors.New(422, "INSUFFICIENT_MP",
			fmt.Sprintf("MP ไม่พอ (ต้องการ: %d, มี: %d)", mpCost, caster.CurrentMP))
	}

	// Deduct resources
	mpBeforeCast := caster.CurrentMP // เก็บไว้ log
	caster.CurrentAP -= apCost
	caster.CurrentMP -= mpCost

	s.appLogger.Info("Resources deducted from caster",
		"caster_id", caster.ID,
		"ap_cost", apCost,
		"mp_cost", mpCost,
		"ap_remaining", caster.CurrentAP,
		"mp_before", mpBeforeCast,
		"mp_remaining", caster.CurrentMP,
	)

	return nil
}

// _ValidateAndConsumeCharges ตรวจสอบและ consume element charge สำหรับ T1+ spells
func (s *combatService) _ValidateAndConsumeCharges(
	caster *domain.Combatant,
	spell *domain.Spell,
) ([]uint, error) {

	// T0 spells (ID 1-4) ไม่ต้อง consume charge
	if spell.ElementID <= 4 {
		s.appLogger.Info("T0 spell detected, no charge consumption needed",
			"spell_id", spell.ID,
			"spell_name", spell.Name,
		)
		return []uint{}, nil
	}

	// T1+ spells ต้องมี charge
	s.appLogger.Info("T1+ spell detected, checking for available charges",
		"spell_id", spell.ID,
		"required_element", spell.ElementID,
	)

	// ค้นหา charge ที่ตรงกับ element และยังไม่ได้ consume
	var foundCharge *domain.CombatantDeck
	for i := range caster.Deck {
		charge := caster.Deck[i]
		if charge.ElementID == spell.ElementID && !charge.IsConsumed {
			foundCharge = charge
			break
		}
	}

	if foundCharge == nil {
		return nil, apperrors.New(422, "NO_ELEMENT_CHARGE",
			fmt.Sprintf("ไม่มี Element Charge สำหรับธาตุ ID %d", spell.ElementID))
	}

	// Mark as consumed
	foundCharge.IsConsumed = true

	s.appLogger.Info("Element charge consumed",
		"charge_id", foundCharge.ID,
		"element_id", foundCharge.ElementID,
	)

	return []uint{foundCharge.ElementID}, nil
}
