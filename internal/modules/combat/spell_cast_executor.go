// file: internal/modules/combat/spell_cast_executor.go
package combat

import (
	"sage-of-elements-backend/internal/domain"

	"github.com/gofrs/uuid"
)

// ExecuteSpellCast เป็น Main Entry Point สำหรับการร่ายเวท
// รับผิดชอบ orchestrate ทั้งกระบวนการร่ายเวทตั้งแต่ต้นจนจบ
//
// Flow:
// 1. PrepareAndValidateCast → ตรวจสอบ + หักทรัพยากร
// 2. CalculateInitialEffectValues → คำนวณค่าพื้นฐาน
// 3. CalculateCombinedModifiers → คำนวณ modifier
// 4. ApplyCalculatedEffects → ประยุกต์ effect
// 5. SaveCombatState → บันทึกสถานะ (auto-saved ใน match update)
func (s *combatService) ExecuteSpellCast(
	match *domain.CombatMatch,
	caster *domain.Combatant,
	targetID uuid.UUID,
	spellID uint,
	castingMode string,
) error {

	s.appLogger.Info("🚀 BEGIN: ExecuteSpellCast",
		"match_id", match.ID,
		"caster_id", caster.ID,
		"target_id", targetID,
		"spell_id", spellID,
		"casting_mode", castingMode,
	)

	// ==================== STEP 1: Preparation ====================
	prepResult, err := s.PrepareAndValidateCast(match, caster, targetID, spellID, castingMode)
	if err != nil {
		s.appLogger.Error("STEP 1 failed: Preparation error", err)
		return err
	}

	// ==================== STEP 2: Calculate Initial Values ====================
	initialValues, err := s.CalculateInitialEffectValues(prepResult.Spell, prepResult.Caster)
	if err != nil {
		s.appLogger.Error("STEP 2 failed: Initial value calculation error", err)
		return err
	}

	// ==================== STEP 3: Calculate Modifiers ====================
	// NOTE: คำนวณ modifier ครั้งเดียวสำหรับทุก effect (เพื่อความเร็ว)
	// ถ้ามี effect ที่ต้องการ modifier แยก ให้ปรับใน Step 4
	modifierCtx, err := s.CalculateCombinedModifiers(
		prepResult.Caster,
		prepResult.Target,
		prepResult.Spell,
		prepResult.PowerModifier,
		0, // effect ID 0 = ใช้ modifier ทั่วไป
	)
	if err != nil {
		s.appLogger.Error("STEP 3 failed: Modifier calculation error", err)
		return err
	}

	// ==================== STEP 4: Apply Effects ====================
	applicationResult, err := s.ApplyCalculatedEffects(
		prepResult.Caster,
		prepResult.Target,
		prepResult.Spell,
		initialValues,
		modifierCtx,
	)
	if err != nil {
		s.appLogger.Error("STEP 4 failed: Effect application error", err)
		return err
	}

	// Log summary
	if len(applicationResult.Errors) > 0 {
		s.appLogger.Warn("Some effects failed to apply",
			"error_count", len(applicationResult.Errors),
			"errors", applicationResult.Errors,
		)
	}

	// ==================== STEP 5: Save State ====================
	// NOTE: State จะถูกบันทึกโดย PerformAction เมื่อ return
	// ไม่ต้อง save ที่นี่

	s.appLogger.Info("✅ SUCCESS: ExecuteSpellCast completed",
		"spell_name", prepResult.Spell.Name,
		"effects_applied", applicationResult.EffectsApplied,
		"total_damage", applicationResult.TotalDamage,
		"total_healing", applicationResult.TotalHealing,
		"evaded", applicationResult.EffectsEvaded,
	)

	return nil
}
