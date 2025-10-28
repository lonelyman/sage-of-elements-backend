// file: internal/modules/combat/effect_manager.go
package combat

import (
	"encoding/json"
	"sage-of-elements-backend/internal/domain"
)

// ============================================================================
// 📌 EFFECT MANAGER - CORE FUNCTIONS
// ============================================================================
// This file contains:
// - Effect tick processing (HP/MP Regen, DoT)
// - Effect expiry handling
// - Stat recalculation (Initiative, Defense, etc.)
// - Effect dispatcher (applyEffect - routes to specialist functions)
// - Helper functions (getMaxHP, getMaxMP)
//
// Specialist effect application functions are in separate files:
// - effect_direct.go   : Damage, Heal, Shield, MP Damage (1000s)
// - effect_buffs.go    : Buffs (2000s)
// - effect_debuffs.go  : Debuffs (4000s)
// - effect_synergy.go  : Stance effects (3000s)
// ============================================================================

func (s *combatService) processEffectTicksAndExpiry(combatant *domain.Combatant) {
	// 1. ตรวจสอบว่ามี ActiveEffects หรือไม่
	if combatant.ActiveEffects == nil {
		s.appLogger.Debug("No active effects to process", "combatant_id", combatant.ID)
		return // ไม่มีอะไรให้ทำ ออกเลย
	}

	// 2. ลอง Unmarshal JSON ของ ActiveEffects
	var currentEffects []domain.ActiveEffect
	err := json.Unmarshal(combatant.ActiveEffects, &currentEffects)
	if err != nil || len(currentEffects) == 0 {
		if err != nil {
			s.appLogger.Error("Failed to unmarshal active effects in processEffectTicksAndExpiry", err, "combatant_id", combatant.ID, "raw_json", string(combatant.ActiveEffects))
		} else {
			s.appLogger.Debug("Active effects list is empty", "combatant_id", combatant.ID)
		}
		// ถ้า JSON เสีย หรือ List ว่างเปล่า ก็ไม่มีอะไรให้ทำต่อ
		// อาจจะเคลียร์ combatant.ActiveEffects ทิ้งถ้า JSON เสีย?
		// combatant.ActiveEffects = nil // Optional: Clear invalid JSON
		return
	}

	// 3. ดึงค่า Max HP/MP มาเตรียมไว้
	maxHP := s.getMaxHP(combatant)
	maxMP := s.getMaxMP(combatant)

	// 4. เตรียม List ใหม่สำหรับเก็บ Effect ที่ยังคงอยู่
	var remainingEffects []domain.ActiveEffect
	somethingChanged := false // Flag ตรวจสอบว่ามีการเปลี่ยนแปลงเกิดขึ้นหรือไม่

	s.appLogger.Debug("Processing effect ticks and expiry", "combatant_id", combatant.ID, "effect_count_before", len(currentEffects))

	// 5. วนลูป Effect แต่ละอัน
	for _, effect := range currentEffects {
		currentEffect := effect // สร้าง copy เพื่อทำงาน จะได้ไม่กระทบค่าใน loop เดิม

		// --- 5.1 ทำ Effect ที่ทำงานตามเวลา (Tick Effects) ---
		switch currentEffect.EffectID {
		case 2101: // BUFF_HP_REGEN
			healAmount := currentEffect.Value // ค่า Heal ต่อเทิร์นจาก Value ของบัฟ
			if healAmount > 0 {               // ทำงานเฉพาะเมื่อมีค่า Heal
				newHP := combatant.CurrentHP + healAmount
				if newHP > maxHP {
					newHP = maxHP
				} // กันเลือดเกิน
				if newHP != combatant.CurrentHP { // อัปเดตและ Log เฉพาะเมื่อมีการเปลี่ยนแปลง
					combatant.CurrentHP = newHP
					s.appLogger.Info("Applied HP_REGEN tick", "combatant_id", combatant.ID, "heal", healAmount, "new_hp", combatant.CurrentHP)
					somethingChanged = true
				}
			}
		case 2102: // BUFF_MP_REGEN
			regenAmount := currentEffect.Value // ค่า Regen ต่อเทิร์นจาก Value ของบัฟ
			if regenAmount > 0 {               // ทำงานเฉพาะเมื่อมีค่า Regen
				newMP := combatant.CurrentMP + regenAmount
				if newMP > maxMP {
					newMP = maxMP
				} // กัน MP เกิน
				if newMP != combatant.CurrentMP { // อัปเดตและ Log เฉพาะเมื่อมีการเปลี่ยนแปลง
					combatant.CurrentMP = newMP
					s.appLogger.Info("Applied MP_REGEN tick", "combatant_id", combatant.ID, "regen", regenAmount, "new_mp", combatant.CurrentMP)
					somethingChanged = true
				}
			}
		case 4201: // DEBUFF_IGNITE
			dotAmount := currentEffect.Value // Damage ต่อเทิร์นจาก Value ของดีบัฟ
			if dotAmount > 0 {               // ทำงานเฉพาะเมื่อมีค่า Damage
				newHP := combatant.CurrentHP - dotAmount
				if newHP < 0 {
					newHP = 0
				} // กันเลือดติดลบ
				if newHP != combatant.CurrentHP { // อัปเดตและ Log เฉพาะเมื่อมีการเปลี่ยนแปลง
					combatant.CurrentHP = newHP
					s.appLogger.Info("Applied IGNITE DoT tick", "combatant_id", combatant.ID, "damage", dotAmount, "new_hp", combatant.CurrentHP) // ⭐️ แก้ Log!
					somethingChanged = true
				}
			}
		}
		// --- สิ้นสุด Tick Effects ---

		// --- 5.2 ลด Duration และเช็คหมดอายุ ---
		previousDuration := currentEffect.TurnsRemaining // เก็บค่าเดิมไว้เช็ค
		currentEffect.TurnsRemaining--                   // ลดเวลา 1 เทิร์น

		if currentEffect.TurnsRemaining > 0 {
			// ยังไม่หมดอายุ เก็บ Effect (ที่อาจจะมีการเปลี่ยนแปลงค่า Value จาก Tick) ไว้ใน List ใหม่
			remainingEffects = append(remainingEffects, currentEffect)
			// ถ้า Duration ลดลงจริง ถือว่ามีการเปลี่ยนแปลง
			if currentEffect.TurnsRemaining != previousDuration {
				somethingChanged = true
			}
		} else {
			// Effect หมดอายุแล้ว Log บอก และตั้ง Flag ว่ามีการเปลี่ยนแปลง
			s.appLogger.Info("Effect has expired", "combatant_id", combatant.ID, "effect_id", currentEffect.EffectID, "value_at_expiry", currentEffect.Value)
			somethingChanged = true
		}
		// --- สิ้นสุดเช็คหมดอายุ ---

	} // จบ Loop for effect

	// 6. อัปเดต ActiveEffects ใน Combatant เฉพาะเมื่อมีการเปลี่ยนแปลงเกิดขึ้น
	if somethingChanged {
		newEffectsJSON, marshalErr := json.Marshal(remainingEffects)
		if marshalErr != nil {
			s.appLogger.Error("Failed to marshal remaining effects after processing", marshalErr, "combatant_id", combatant.ID)
			// ถ้า Marshal ไม่ได้ อาจจะปล่อย JSON เดิมไว้ หรือเคลียร์ทิ้ง? ตอนนี้ปล่อยไว้ก่อน
		} else {
			combatant.ActiveEffects = newEffectsJSON // อัปเดต JSON ใหม่
			s.appLogger.Info("Updated active effects after processing ticks/expiry", "combatant_id", combatant.ID, "remaining_count", len(remainingEffects))
		}
	} else {
		// ถ้าไม่มีอะไรเปลี่ยนแปลงเลย ก็ไม่ต้องทำอะไร ไม่ต้อง Marshal ใหม่
		s.appLogger.Debug("No changes to active effects after processing ticks/expiry", "combatant_id", combatant.ID)
	}
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
		if effect.EffectID == 4101 { // DEBUFF_SLOW
			modifiedInitiative += effect.Value
		}
	}
	combatant.Initiative = modifiedInitiative
	s.appLogger.Info("Stats recalculated for combatant", "id", combatant.ID, "base_init", baseInitiative, "modified_init", modifiedInitiative)
}

func (s *combatService) getMaxHP(combatant *domain.Combatant) int {
	if combatant.CharacterID != nil && combatant.Character != nil {
		// เราใช้ current_hp จาก character data เป็น MaxHP (ตามที่เราทำใน turn_manager)
		return combatant.Character.CurrentHP
	} else if combatant.EnemyID != nil && combatant.Enemy != nil {
		return combatant.Enemy.MaxHP
	}
	s.appLogger.Warn("Could not determine MaxHP for combatant", "id", combatant.ID)
	return 0
}

// --- ⭐️ เพิ่ม Helper Function: getMaxMP ⭐️ ---
func (s *combatService) getMaxMP(combatant *domain.Combatant) int {
	if combatant.CharacterID != nil && combatant.Character != nil {
		// เราใช้ current_mp จาก character data เป็น MaxMP (ตามที่เราทำใน turn_manager)
		return combatant.Character.CurrentMP
	} else if combatant.EnemyID != nil && combatant.Enemy != nil {
		// Enemy บางตัวอาจจะมี MaxMP (ถ้าออกแบบไว้) หรือใช้ค่าคงที่เยอะๆ
		// return combatant.Enemy.MaxMP // ถ้ามีฟิลด์นี้
		return 9999 // หรือใช้ค่า Default ไปก่อน
	}
	s.appLogger.Warn("Could not determine MaxMP for combatant", "id", combatant.ID)
	return 0
}

// ============================================================================
// 📌 EFFECT DISPATCHER
// ============================================================================
// applyEffect - Routes effect application to specialist functions based on Effect ID
// This is the central hub that delegates work to specific effect handlers
// ============================================================================

func (s *combatService) applyEffect(caster *domain.Combatant, target *domain.Combatant, effectData map[string]interface{}, spell *domain.Spell) {
	effectID := uint(effectData["effect_id"].(float64))

	// Route to appropriate specialist based on new 1000-based ID structure
	switch effectID {
	// --- Direct Effects (1000s) ---
	case 1101: // DAMAGE
		s.applyDamage(caster, target, effectData, spell)
	case 1102: // SHIELD
		s.applyShield(caster, target, effectData, spell)
	case 1103: // HEAL
		s.applyHeal(caster, target, effectData, spell)
	case 1104: // MP_DAMAGE
		s.applyMpDamage(caster, target, effectData, spell)

	// --- Buffs (2000s) ---
	case 2101: // BUFF_HP_REGEN
		s.applyBuffHpRegen(caster, target, effectData)
	case 2102: // BUFF_MP_REGEN
		s.applyBuffMpRegen(caster, target, effectData)
	case 2201: // BUFF_EVASION
		s.applyBuffEvasion(caster, target, effectData)
	case 2202: // BUFF_DAMAGE_UP
		s.applyBuffDamageUp(caster, target, effectData)
	case 2203: // BUFF_RETALIATION
		s.applyBuffRetaliation(caster, target, effectData)
	case 2204: // BUFF_DEFENSE_UP
		s.applyBuffDefenseUp(caster, target, effectData)

	// --- Synergy Effects (3000s) ---
	case 3101: // STANCE_S
		s.applySynergyGrantStanceS(caster, target, effectData)
	case 3102: // STANCE_L
		s.applySynergyGrantStanceL(caster, target, effectData)
	case 3103: // STANCE_G
		s.applySynergyGrantStanceG(caster, target, effectData)
	case 3104: // STANCE_P
		s.applySynergyGrantStanceP(caster, target, effectData)

	// --- Debuffs (4000s) ---
	case 4101: // DEBUFF_SLOW
		s.applyDebuffSlow(caster, target, effectData)
	case 4102: // DEBUFF_VULNERABLE
		s.applyDebuffVulnerable(caster, target, effectData)
	case 4201: // DEBUFF_IGNITE
		s.applyDebuffIgnite(caster, target, effectData)

	default:
		s.appLogger.Warn("Attempted to apply an unknown or unimplemented effect", "effect_id", effectID)
	}
}
