// file: internal/modules/combat/effect_manager.go
package combat

import (
	"encoding/json"
	"math"
	"math/rand"
	"sage-of-elements-backend/internal/domain"
	"sort"
)

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
		case 100: // BUFF_HP_REGEN
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
		case 101: // BUFF_MP_REGEN
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
		case 306: // DEBUFF_IGNITE
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
		if effect.EffectID == 301 {
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

// --- "ผู้จัดการ" (The Manager) ---
// applyEffect ตอนนี้จะทำหน้าที่แค่ "ส่งต่องาน" ให้ผู้เชี่ยวชาญที่ถูกต้อง!
func (s *combatService) applyEffect(caster *domain.Combatant, target *domain.Combatant, effectData map[string]interface{}, spell *domain.Spell) {
	effectID := uint(effectData["effect_id"].(float64))

	// "ผู้จัดการ" จะดู ID แล้วเลือก "ผู้เชี่ยวชาญ"
	switch effectID {
	case 1: // DAMAGE
		s.applyDamage(caster, target, effectData, spell)
	case 2: // SHIELD
		s.applyShield(caster, target, effectData, spell)
	case 3: // HEAL
		s.applyHeal(caster, target, effectData, spell)
	case 5: // MP_DAMAGE (หรือ MP_DRAIN?)
		s.applyMpDamage(caster, target, effectData, spell)
	case 100: // BUFF_HP_REGEN
		s.applyBuffHpRegen(caster, target, effectData)
	case 101: // BUFF_MP_REGEN
		s.applyBuffMpRegen(caster, target, effectData)
	case 102: // BUFF_EVASION
		s.applyBuffEvasion(caster, target, effectData)
	case 103: // BUFF_DAMAGE_UP
		s.applyBuffDamageUp(caster, target, effectData)
	case 104: // BUFF_RETALIATION
		s.applyBuffRetaliation(caster, target, effectData)
	case 110: // BUFF_DEFENSE_UP
		s.applyBuffDefenseUp(caster, target, effectData)
	case 200: // SYNERGY_GRANT_STANCE_S
		s.applySynergyGrantStanceS(caster, target, effectData)
	case 201: // SYNERGY_GRANT_STANCE_L
		s.applySynergyGrantStanceL(caster, target, effectData)
	case 202: // SYNERGY_GRANT_STANCE_G
		s.applySynergyGrantStanceG(caster, target, effectData)
	case 203: // SYNERGY_GRANT_STANCE_P
		s.applySynergyGrantStanceP(caster, target, effectData)
	case 301: // DEBUFF_SLOW
		s.applyDebuffSlow(caster, target, effectData)
	case 302: // DEBUFF_VULNERABLE
		s.applyDebuffVulnerable(caster, target, effectData)
	case 306: // DEBUFF_IGNITE
		s.applyDebuffIgnite(caster, target, effectData)

	default:
		s.appLogger.Warn("Attempted to apply an unknown or unimplemented effect", "effect_id", effectID)
	}
}

// --- ⭐️ ผู้เชี่ยวชาญด้านการทำ Damage (เวอร์ชันอัปเกรดเต็มรูปแบบ) ⭐️ ---
func (s *combatService) applyDamage(caster *domain.Combatant, target *domain.Combatant, effectData map[string]interface{}, spell *domain.Spell) {

	// --- ⭐️ ขั้นตอนที่ 1: Logic เช็ค Evasion (หลบหลีก) ⭐️ ---
	// ต้องเช็คก่อน! ถ้าหลบได้ คือจบเลย ไม่ต้องคำนวณอะไรต่อ
	var targetActiveEffectsForEvasion []domain.ActiveEffect // ⭐️ ใช้ตัวแปรเฉพาะส่วน
	evasionChance := 0                                      // % หลบหลีกเริ่มต้น
	hasEvasionBuff := false

	if target.ActiveEffects != nil {
		// ⭐️ ใช้ตัวแปรใหม่ (targetActiveEffectsForEvasion)
		err := json.Unmarshal(target.ActiveEffects, &targetActiveEffectsForEvasion)
		if err == nil {
			for _, effect := range targetActiveEffectsForEvasion {
				if effect.EffectID == 102 { // เจอ Buff Evasion!
					evasionChance = effect.Value // ดึง % หลบหลีกมาจาก Value ของบัฟ
					hasEvasionBuff = true
					s.appLogger.Info("Target has Evasion buff", "target_id", target.ID, "chance", evasionChance)
					break // เจออันเดียวพอ (สมมติว่า Evasion ไม่ Stack?)
				}
			}
		}
	}

	if hasEvasionBuff && evasionChance > 0 {
		// สุ่มเลข 0-99
		roll := rand.Intn(100) // ผลลัพธ์คือ 0 ถึง 99
		s.appLogger.Info("Performing Evasion check", "target_id", target.ID, "chance", evasionChance, "roll", roll)
		if roll < evasionChance { // ถ้าเลขสุ่ม < โอกาสหลบ
			s.appLogger.Info("Attack EVADED!", "caster", caster.ID, "target", target.ID, "spell_id", spell.ID)
			// TODO: อาจจะส่ง Event บอก Client ว่า "MISS!"
			return // ⭐️ จบการทำงาน! ไม่ต้องคำนวณ Damage หรือ Shield ต่อ!
		} else {
			s.appLogger.Info("Evasion check failed, attack proceeds", "target_id", target.ID)
		}
	}
	// --- ⭐️ สิ้นสุด Logic Evasion ⭐️ ---

	// --- ⭐️ ขั้นตอนที่ 2: ตรวจสอบข้อมูล (Validation) และเตรียมการ ⭐️ ---
	effectIDFloat, ok1 := effectData["effect_id"].(float64)
	baseValueFloat, ok2 := effectData["value"].(float64)
	powerModifierFloat, ok3 := effectData["power_modifier"].(float64)
	if !ok1 || !ok2 {
		s.appLogger.Warn("Invalid or missing effect_id or value in effectData for applyDamage", "data", effectData)
		return // ไม่ทำ Damage ถ้าข้อมูลพื้นฐานผิดพลาด
	}
	if !ok3 {
		powerModifierFloat = 1.0 // ถ้าไม่มี power_modifier (เช่น AI โจมตี) ให้ใช้ค่า Default 1.0
	}
	// ------------------------------------

	tempSpellEffect := &domain.SpellEffect{
		EffectID:  uint(effectIDFloat),
		BaseValue: baseValueFloat,
	}
	powerModifier := powerModifierFloat // ใช้ค่าที่ตรวจสอบแล้ว

	// --- ⭐️ ขั้นตอนที่ 3: คำนวณ Damage พื้นฐาน (Base Calculation) ⭐️ ---
	// (เรียก calculateEffectValue ที่อาจจะรวม Talent, บัฟ Caster, ธาตุ ฯลฯ)
	calculatedDamage, err := s.calculateEffectValue(caster, target, spell, tempSpellEffect, powerModifier)
	if err != nil {
		s.appLogger.Error("Error calculating damage value", err)
		return
	}

	// --- ⭐️ ขั้นตอนที่ 4: Logic เช็ค Vulnerable (ID 302) บนเป้าหมาย ⭐️ ---
	// (เพิ่ม Damage ที่จะได้รับ)
	var targetActiveEffectsForVulnerable []domain.ActiveEffect
	damageIncreasePercent := 0 // % Damage ที่จะเพิ่มขึ้น (Default = 0)
	hasVulnerableDebuff := false

	if target.ActiveEffects != nil {
		// ⭐️ ใช้ตัวแปรใหม่ (targetActiveEffectsForVulnerable)
		err := json.Unmarshal(target.ActiveEffects, &targetActiveEffectsForVulnerable)
		if err == nil {
			for _, effect := range targetActiveEffectsForVulnerable {
				if effect.EffectID == 302 { // เจอดีบัฟ Vulnerable!
					damageIncreasePercent = effect.Value // ดึง % มาจาก Value
					hasVulnerableDebuff = true
					s.appLogger.Info("Target has Vulnerable debuff", "target_id", target.ID, "increase_percent", damageIncreasePercent)
					break // เจออันเดียวพอ (สมมติว่าไม่ Stack)
				}
			}
		} else {
			s.appLogger.Error("Failed to unmarshal active effects for Vulnerable check", err, "target_id", target.ID)
		}
	}

	if hasVulnerableDebuff && damageIncreasePercent > 0 {
		// เพิ่ม Damage (คำนวณแบบ % ของ float64)
		multiplier := 1.0 + (float64(damageIncreasePercent) / 100.0) // เช่น 10% -> 1.1
		originalCalculatedDamage := calculatedDamage                 // เก็บไว้ดูใน Log
		calculatedDamage = calculatedDamage * multiplier
		s.appLogger.Info("Applied Vulnerable damage increase", "target_id", target.ID, "original_damage", originalCalculatedDamage, "multiplier", multiplier, "final_damage", calculatedDamage)
	}
	// --- ⭐️ สิ้นสุด Logic Vulnerable ⭐️ ---

	// --- ⭐️ ขั้นตอนที่ 5: แปลง Damage เป็น int (หลังคำนวณ % แล้ว) ⭐️ ---
	damageDealt := int(math.Round(calculatedDamage)) // ปัดเศษ Damage
	if damageDealt < 0 {
		damageDealt = 0
	} // Damage ไม่ควรติดลบ

	// --- ⭐️ ขั้นตอนที่ 6: Logic จัดการ Shield (ID 2) บนเป้าหมาย ⭐️ ---
	// (ลด Damage ด้วย Shield ก่อน)
	remainingDamage := damageDealt           // Damage ที่เหลือหลังจากหัก Shield
	var activeEffects []domain.ActiveEffect  // List บัฟ/ดีบัฟ ปัจจุบันของเป้าหมาย
	var updatedEffects []domain.ActiveEffect // List บัฟ/ดีบัฟ ที่จะเหลืออยู่หลังจบ Logic นี้
	hasShieldEffect := false                 // Flag ว่าเป้าหมายมี Shield หรือไม่
	shieldAbsorbedTotal := 0                 // เก็บว่า Shield ดูดซับไปเท่าไหร่

	if target.ActiveEffects != nil {
		err := json.Unmarshal(target.ActiveEffects, &activeEffects)
		if err == nil { // ถ้า Unmarshal สำเร็จ
			tempEffects := make([]domain.ActiveEffect, len(activeEffects))
			copy(tempEffects, activeEffects)

			// เรียงลำดับ Shield ตาม TurnsRemaining น้อยไปมาก (จะได้ลดอันใกล้หมดอายุก่อน)
			sort.SliceStable(tempEffects, func(i, j int) bool {
				isIShield := tempEffects[i].EffectID == 2
				isJShield := tempEffects[j].EffectID == 2
				if isIShield && isJShield {
					return tempEffects[i].TurnsRemaining < tempEffects[j].TurnsRemaining
				}
				if isIShield {
					return true
				}
				if isJShield {
					return false
				}
				return i < j
			})

			// วนลูปเช็ค Effect แต่ละอันใน Slice ชั่วคราว
			for i := range tempEffects {
				if tempEffects[i].EffectID == 2 && tempEffects[i].Value > 0 && remainingDamage > 0 {
					hasShieldEffect = true
					shieldHP := tempEffects[i].Value
					absorbedDamage := 0

					if remainingDamage >= shieldHP {
						absorbedDamage = shieldHP
						remainingDamage -= shieldHP
						tempEffects[i].Value = 0 // Shield แตก!
						s.appLogger.Info("Shield broke!", "target_id", target.ID, "shield_index", i, "absorbed", absorbedDamage)
					} else {
						absorbedDamage = remainingDamage
						tempEffects[i].Value -= remainingDamage
						remainingDamage = 0 // Damage โดนดูดหมดแล้ว
						s.appLogger.Info("Shield absorbed damage", "target_id", target.ID, "shield_index", i, "absorbed", absorbedDamage, "shield_hp_left", tempEffects[i].Value)
					}
					shieldAbsorbedTotal += absorbedDamage
				}
			} // จบ Loop Shield

			// กรองเอา Shield ที่แตก (Value <= 0) ออกไป
			for _, effect := range tempEffects {
				if !(effect.EffectID == 2 && effect.Value <= 0) {
					updatedEffects = append(updatedEffects, effect)
				} else {
					s.appLogger.Info("Removing broken shield from active effects", "target_id", target.ID, "effect_id", effect.EffectID)
				}
			}

			// Marshal list ที่กรองแล้ว กลับเป็น JSON เพื่ออัปเดต
			newEffectsJSON, marshalErr := json.Marshal(updatedEffects)
			if marshalErr == nil {
				target.ActiveEffects = newEffectsJSON // อัปเดต ActiveEffects ของเป้าหมาย
			} else {
				s.appLogger.Error("Failed to marshal updated active effects after shield processing", marshalErr, "target_id", target.ID)
			}

		} else { // ถ้า Unmarshal ไม่สำเร็จ
			s.appLogger.Error("Failed to unmarshal active effects for shield check", err, "target_id", target.ID)
			remainingDamage = damageDealt // ถือว่าไม่มี Shield
		}
	} else { // ถ้าไม่มี ActiveEffects เลย
		remainingDamage = damageDealt // ก็ไม่มี Shield
	}
	// --- ⭐️ สิ้นสุด Logic Shield ⭐️ ---

	// --- ⭐️ ขั้นตอนที่ 7: ลด HP เป้าหมาย (ส่วนที่ทะลุ Shield) ⭐️ ---
	hpBefore := target.CurrentHP // เก็บ HP ก่อนโดน Damage (ส่วนที่ทะลุ Shield)
	var hpDamageDealt int = 0    // เก็บว่า HP โดนลดไปเท่าไหร่จริงๆ
	if remainingDamage > 0 {     // ถ้ามี Damage เหลือหลังจากหัก Shield
		hpDamageDealt = remainingDamage
		target.CurrentHP -= remainingDamage
		if target.CurrentHP < 0 {
			target.CurrentHP = 0
		} // ป้องกันเลือดติดลบ
	}
	hpAfter := target.CurrentHP // HP สุดท้าย

	// --- ⭐️ ขั้นตอนที่ 8: (ใหม่!) Logic เช็ค Retaliation (ID 104) บนเป้าหมาย ⭐️ ---
	// (สะท้อน Damage กลับไปหา Caster)
	var targetActiveEffectsForRetaliation []domain.ActiveEffect
	retaliationDamage := 0 // Damage ที่จะสะท้อน (Default = 0)
	hasRetaliationBuff := false

	// เงื่อนไข: 1. การโจมตีต้อง "โดน" (โดน HP หรือ โดน Shield)
	//          2. ต้องไม่ใช่การโจมตีใส่ตัวเอง
	//          3. เป้าหมายต้องมี ActiveEffects
	if (hpDamageDealt > 0 || shieldAbsorbedTotal > 0) && (caster.ID != target.ID) && target.ActiveEffects != nil {
		// ⭐️ ใช้ตัวแปรใหม่ (targetActiveEffectsForRetaliation)
		err := json.Unmarshal(target.ActiveEffects, &targetActiveEffectsForRetaliation)
		if err == nil {
			for _, effect := range targetActiveEffectsForRetaliation {
				if effect.EffectID == 104 && effect.Value > 0 { // เจอบัฟ Retaliation และมีค่า Damage
					retaliationDamage = effect.Value // ดึง Damage สะท้อนมาจาก Value
					hasRetaliationBuff = true
					s.appLogger.Info("Target has Retaliation buff", "target_id", target.ID, "retaliation_damage", retaliationDamage)
					break // เจออันเดียวพอ (สมมติว่าไม่ Stack)
				}
			}
		} else {
			s.appLogger.Error("Failed to unmarshal active effects for Retaliation check", err, "target_id", target.ID)
		}
	}

	if hasRetaliationBuff && retaliationDamage > 0 {
		// สะท้อน Damage กลับไปหา Caster!
		casterHpBefore := caster.CurrentHP
		caster.CurrentHP -= retaliationDamage
		if caster.CurrentHP < 0 {
			caster.CurrentHP = 0
		} // กัน Caster เลือดติดลบ
		s.appLogger.Info("Applied Retaliation damage to caster", "caster_id", caster.ID, "damage_taken", retaliationDamage, "caster_hp_before", casterHpBefore, "caster_hp_after", caster.CurrentHP)
	}
	// --- ⭐️ สิ้นสุด Logic Retaliation ⭐️ ---

	// --- ⭐️ ขั้นตอนที่ 9: Log ผลลัพธ์สุดท้าย ⭐️ ---
	s.appLogger.Info("Applied DAMAGE effect",
		"caster", caster.ID,
		"target", target.ID,
		"initial_damage", damageDealt, // Damage ที่คำนวณได้ตอนแรก (รวม Vulnerable แล้ว)
		"absorbed_by_shield", shieldAbsorbedTotal, // Damage ที่ Shield ดูดซับไปทั้งหมด
		"hp_damage", hpDamageDealt, // Damage ที่ลง HP จริงๆ
		"target_hp_before", hpBefore, // HP ก่อนโดนส่วนที่ทะลุ Shield
		"target_hp_after", hpAfter, // HP สุดท้าย
		"shield_active", hasShieldEffect, // มี Shield ทำงานหรือไม่
	)
}

// --- ปรับปรุงผู้เชี่ยวชาญ Heal ---
func (s *combatService) applyHeal(caster *domain.Combatant, target *domain.Combatant, effectData map[string]interface{}, spell *domain.Spell) {
	// --- ⭐️ เพิ่มการตรวจสอบ Type Assertion ⭐️ ---
	effectIDFloat, ok1 := effectData["effect_id"].(float64)
	baseValueFloat, ok2 := effectData["value"].(float64)
	powerModifierFloat, ok3 := effectData["power_modifier"].(float64)
	if !ok1 || !ok2 {
		s.appLogger.Warn("Invalid or missing effect_id or value in effectData for applyHeal", effectData)
		return // ไม่ Heal ถ้าข้อมูลพื้นฐานผิดพลาด
	}
	if !ok3 {
		powerModifierFloat = 1.0 // ⭐️ ถ้าไม่มี power_modifier ให้ใช้ค่า Default 1.0 ⭐️
	}
	// --- ⭐️ สิ้นสุดการตรวจสอบ ⭐️ ---
	tempSpellEffect := &domain.SpellEffect{
		EffectID:  uint(effectIDFloat),
		BaseValue: baseValueFloat,
	}
	powerModifier := powerModifierFloat // ใช้ค่าที่ตรวจสอบแล้ว
	// --- ⭐️ เพิ่มเช็ค: Heal ควร scale กับ powerModifier หรือไม่? (ตอนนี้ให้ scale ไปก่อน) ⭐️ ---
	// ถ้าไม่อยากให้ Heal แรงขึ้นตาม Charge/Overcharge ให้ส่ง 1.0 แทน powerModifier ตรงนี้
	s.appLogger.Info("Heal scaling with powerModifier", "modifier", powerModifier)
	// ------------------------------------------------------------------------------------

	// ⭐️ ส่ง powerModifier ให้ calculateEffectValue! ⭐️
	calculatedHeal, err := s.calculateEffectValue(caster, target, spell, tempSpellEffect, powerModifier)
	if err != nil {
		s.appLogger.Error("Error calculating heal value", err)
		return
	}

	healAmount := int(math.Round(calculatedHeal)) // ⭐️ ปัดเศษ Heal ก่อนแปลงเป็น int ⭐️
	if healAmount < 0 {
		healAmount = 0
	} // Heal ไม่ควรติดลบ

	maxHP := s.getMaxHP(target)

	// ⭐️ ตรวจสอบ HP ก่อนเพิ่ม ⭐️
	hpBefore := target.CurrentHP
	target.CurrentHP += healAmount
	if target.CurrentHP > maxHP {
		target.CurrentHP = maxHP
	}
	hpAfter := target.CurrentHP

	s.appLogger.Info("Applied HEAL_HP effect", "caster", caster.ID, "target", target.ID, "heal", healAmount, "target_hp_before", hpBefore, "target_hp_after", hpAfter)
}

// --- ⭐️ ผู้เชี่ยวชาญด้านการมอบสถานะ Slow ⭐️ ---
func (s *combatService) applyDebuffSlow(caster *domain.Combatant, target *domain.Combatant, effectData map[string]interface{}) {
	value := int(effectData["value"].(float64))
	duration := int(effectData["duration"].(float64))

	var activeEffects []domain.ActiveEffect
	if target.ActiveEffects != nil {
		json.Unmarshal(target.ActiveEffects, &activeEffects)
	}
	newEffect := domain.ActiveEffect{
		EffectID:       301, // DEBUFF_SLOW
		Value:          value,
		TurnsRemaining: duration,
		SourceID:       caster.ID,
	}
	activeEffects = append(activeEffects, newEffect)
	newEffectsJSON, _ := json.Marshal(activeEffects)
	target.ActiveEffects = newEffectsJSON
	s.appLogger.Info("Applied DEBUFF_SLOW effect", "caster", caster.ID, "target", target.ID, "duration", duration)
}

// --- ⭐️ ผู้เชี่ยวชาญด้านการเพิ่มพลังป้องกัน ⭐️ ---
func (s *combatService) applyBuffDefenseUp(caster *domain.Combatant, target *domain.Combatant, effectData map[string]interface{}) {
	// (ท่า S_HARDEN ไม่ได้ส่ง "value" มา, เราจะสมมติว่า value = 0)
	value := 0
	if v, ok := effectData["value"]; ok {
		value = int(v.(float64))
	}
	duration := int(effectData["duration"].(float64))

	var activeEffects []domain.ActiveEffect
	if target.ActiveEffects != nil {
		json.Unmarshal(target.ActiveEffects, &activeEffects)
	}
	newEffect := domain.ActiveEffect{
		EffectID:       110, // BUFF_DEFENSE_UP
		Value:          value,
		TurnsRemaining: duration,
		SourceID:       caster.ID,
	}
	activeEffects = append(activeEffects, newEffect)
	newEffectsJSON, _ := json.Marshal(activeEffects)
	target.ActiveEffects = newEffectsJSON
	s.appLogger.Info("Applied BUFF_DEFENSE_UP effect", "target", target.ID, "duration", duration)
}

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

	// --- ⭐️ 4. Add "Replace" Logic ⭐️ ---
	var tempEffects []domain.ActiveEffect
	for _, effect := range activeEffects {
		if effect.EffectID != 200 { // Keep effects that are NOT Stance S (ID 200)
			tempEffects = append(tempEffects, effect)
		} else {
			s.appLogger.Info("Replacing existing Stance S buff", "target_id", target.ID)
		}
	}
	activeEffects = tempEffects // Use the filtered list (without old Stance S)

	// --- ⭐️ 5. Create and Append New Effect (เหมือนเดิม) ⭐️ ---
	newEffect := domain.ActiveEffect{
		EffectID:       200, // SYNERGY_GRANT_STANCE_S
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

// --- ⭐️ เพิ่ม ผู้เชี่ยวชาญ MP Damage/Drain! ⭐️ ---
func (s *combatService) applyMpDamage(caster *domain.Combatant, target *domain.Combatant, effectData map[string]interface{}, spell *domain.Spell) {
	// --- การตรวจสอบ Type Assertion ---
	effectIDFloat, ok1 := effectData["effect_id"].(float64)
	baseValueFloat, ok2 := effectData["value"].(float64)
	powerModifierFloat, ok3 := effectData["power_modifier"].(float64)
	if !ok1 || !ok2 {
		// แก้ไข: เปลี่ยน Error เป็น Warn และใส่ effectData เข้าไปใน Log
		s.appLogger.Warn("Invalid or missing effect_id or value in effectData for applyMpDamage", "data", effectData)
		return
	}
	if !ok3 {
		powerModifierFloat = 1.0
	}
	// -----------------------------

	tempSpellEffect := &domain.SpellEffect{
		EffectID:  uint(effectIDFloat),
		BaseValue: baseValueFloat,
	}
	powerModifier := powerModifierFloat

	// --- คำนวณ MP Damage (Option 2: Scale กับ Talent) ---
	calculatedMpDamage, err := s.calculateEffectValue(caster, target, spell, tempSpellEffect, powerModifier)
	if err != nil {
		s.appLogger.Error("Error calculating MP damage value, falling back to base calculation", err)
		calculatedMpDamage = tempSpellEffect.BaseValue * powerModifier // Fallback
	}
	// ----------------------------------------------------

	mpDamageDealt := int(math.Round(calculatedMpDamage)) // ปัดเศษ
	if mpDamageDealt < 0 {
		mpDamageDealt = 0
	} // ไม่ควรติดลบ

	// --- ลด MP เป้าหมาย ---
	mpBefore := target.CurrentMP
	actualMpLost := 0 // เก็บว่าเป้าหมายเสีย MP ไปเท่าไหร่จริงๆ

	// ⭐️ แก้ไข Logic การลด MP ให้ถูกต้อง ⭐️
	if target.CurrentMP >= mpDamageDealt {
		target.CurrentMP -= mpDamageDealt
		actualMpLost = mpDamageDealt // เสียไปตามที่คำนวณได้
	} else { // ถ้า MP เป้าหมายไม่พอให้ลด
		actualMpLost = target.CurrentMP // เสียไปเท่าที่มี
		target.CurrentMP = 0
	}
	// --------------------------

	mpAfter := target.CurrentMP

	// --- ✨⭐️ เพิ่ม Logic การ Drain MP! ⭐️✨ ---
	if uint(effectIDFloat) == 5 {
		s.appLogger.Info("DEBUG: Entering MP Drain logic block", "actual_mp_lost", actualMpLost, "caster_current_mp", caster.CurrentMP) // ⭐️ Log 1 ⭐️
		// ⭐️ TODO: ดึงค่าประสิทธิภาพการดูด (Drain Efficiency) มาจาก Game Config ⭐️
		drainEfficiency := 0.5                                               // สมมติว่าดูดได้ 50% ของ MP ที่เป้าหมายเสียไปจริง
		mpGained := int(math.Round(float64(actualMpLost) * drainEfficiency)) // คำนวณ MP ที่ Caster จะได้

		s.appLogger.Info("DEBUG: Calculated MP Gained", "mpGained", mpGained) // ⭐️ Log 2 ⭐️

		if mpGained > 0 { // ถ้าคำนวณแล้วได้ MP คืน
			s.appLogger.Info("DEBUG: mpGained > 0, attempting to add MP to caster") // ⭐️ Log 3 ⭐️
			casterMaxMP := s.getMaxMP(caster)                                       // หา MaxMP ของ Caster
			casterMpBeforeDrain := caster.CurrentMP
			caster.CurrentMP += mpGained        // เพิ่ม MP ให้ Caster
			if caster.CurrentMP > casterMaxMP { // กัน MP ล้น
				caster.CurrentMP = casterMaxMP
			}
			s.appLogger.Info("Caster drained MP", "caster", caster.ID, "gained", mpGained, "caster_mp_before", casterMpBeforeDrain, "caster_mp_after", caster.CurrentMP)
		} else {
			s.appLogger.Info("DEBUG: mpGained is not > 0", "mpGained", mpGained) // ⭐️ Log 4 (เผื่อคำนวณผิด) ⭐️
		}
	} else {
		s.appLogger.Info("DEBUG: Condition if uint(effectIDFloat) == 5 is FALSE", "effectIDFloat", effectIDFloat) // ⭐️ Log 5 (เผื่อเช็ค ID ผิด) ⭐️
	}
	// --- ✨⭐️ สิ้นสุด Logic Drain ⭐️✨ ---

	// ปรับ Log นิดหน่อยให้แสดงค่าที่ถูกต้อง
	s.appLogger.Info("Applied MP_DAMAGE effect",
		"caster", caster.ID,
		"target", target.ID,
		"mp_damage_calculated", mpDamageDealt, // ค่าที่คำนวณได้
		"target_mp_lost", actualMpLost, // ค่าที่เป้าหมายเสียไปจริง
		"target_mp_before", mpBefore,
		"target_mp_after", mpAfter,
	)
}

// ----------------------------------------------------

// --- ⭐️ เพิ่ม ผู้เชี่ยวชาญ บัฟ MP Regen! ⭐️ ---
func (s *combatService) applyBuffMpRegen(caster *domain.Combatant, target *domain.Combatant, effectData map[string]interface{}) {
	// --- การตรวจสอบ Type Assertion ---
	// Effect นี้มาจาก Meditate ซึ่ง BaseValue=10 ใน Seed (เวอร์ชันปรับปรุง)
	valueFloat, ok1 := effectData["value"].(float64)
	durationFloat, ok2 := effectData["duration"].(float64)
	if !ok1 || !ok2 {
		s.appLogger.Warn("Invalid or missing value or duration in effectData for applyBuffMpRegen", effectData)
		return
	}
	// -----------------------------

	// --- ⭐️ คิดว่าจะให้ค่า Regen ต่อเทิร์น scale กับ Talent ไหม? ⭐️ ---
	// Option 1: ไม่ scale ใช้ค่า BaseValue ตรงๆ (ง่ายสุด)
	regenPerTurn := int(math.Round(valueFloat))
	// Option 2: Scale กับ Talent L (เหมือน Heal) -> ต้องเรียก calculateEffectValue
	// ต้องส่ง spell เข้ามาด้วยถ้าจะ scale... ตอนนี้เอาแบบไม่ scale ก่อน
	// -----------------------------------------------------------------

	duration := int(durationFloat)
	if regenPerTurn < 0 {
		regenPerTurn = 0
	} // Regen ไม่ควรติดลบ

	var activeEffects []domain.ActiveEffect
	if target.ActiveEffects != nil {
		json.Unmarshal(target.ActiveEffects, &activeEffects)
	}

	newEffect := domain.ActiveEffect{
		EffectID:       101,          // BUFF_MP_REGEN
		Value:          regenPerTurn, // ค่า MP ที่จะฟื้นต่อเทิร์น
		TurnsRemaining: duration,
		SourceID:       caster.ID,
	}
	activeEffects = append(activeEffects, newEffect)
	newEffectsJSON, _ := json.Marshal(activeEffects)
	target.ActiveEffects = newEffectsJSON

	s.appLogger.Info("Applied BUFF_MP_REGEN effect", "target", target.ID, "duration", duration, "regen_per_turn", regenPerTurn)
}

func (s *combatService) applyBuffHpRegen(caster *domain.Combatant, target *domain.Combatant, effectData map[string]interface{}) {
	// --- การตรวจสอบ Type Assertion ---
	valueFloat, ok1 := effectData["value"].(float64)       // ค่าฮีลต่อเทิร์น
	durationFloat, ok2 := effectData["duration"].(float64) // ระยะเวลา
	if !ok1 || !ok2 {
		s.appLogger.Warn("Invalid or missing value or duration in effectData for applyBuffHpRegen", effectData)
		return // ไม่แปะบัฟถ้าข้อมูลผิดพลาด
	}
	// -----------------------------

	// --- ⭐️ คิดว่าจะให้ค่า Heal ต่อเทิร์น scale กับ Talent ไหม? ⭐️ ---
	// Option 1: ไม่ scale ใช้ค่า BaseValue ตรงๆ (ง่ายสุด - ปัจจุบันใช้อันนี้)
	healPerTurn := int(math.Round(valueFloat))
	// Option 2: Scale กับ Talent L (เหมือน Heal ตรง) -> ต้องเรียก calculateEffectValue
	// ถ้าจะทำ Option 2 ต้องส่ง spell เข้ามาให้ function นี้ด้วย
	// -----------------------------------------------------------------

	duration := int(durationFloat)
	if healPerTurn < 0 {
		healPerTurn = 0
	} // Heal ไม่ควรติดลบ

	var activeEffects []domain.ActiveEffect
	if target.ActiveEffects != nil {
		err := json.Unmarshal(target.ActiveEffects, &activeEffects)
		// เพิ่ม error handling ตอน unmarshal เผื่อ JSON เสีย
		if err != nil {
			s.appLogger.Error("Failed to unmarshal existing active effects for HP Regen buff", err, "target_id", target.ID)
			// อาจจะ return หรือจะลองแปะบัฟทับไปเลยก็ได้ ขึ้นอยู่กับการออกแบบ
			activeEffects = []domain.ActiveEffect{} // เริ่มต้นใหม่ถ้า unmarshal ไม่ได้
		}
	}

	// สร้าง Object บัฟใหม่
	newEffect := domain.ActiveEffect{
		EffectID:       100,         // BUFF_HP_REGEN
		Value:          healPerTurn, // ค่า HP ที่จะฟื้นต่อเทิร์น
		TurnsRemaining: duration,
		SourceID:       caster.ID,
	}

	// เพิ่มบัฟใหม่เข้าไปใน list
	activeEffects = append(activeEffects, newEffect)

	// แปลงกลับเป็น JSON เพื่อเก็บลง DB
	newEffectsJSON, err := json.Marshal(activeEffects)
	if err != nil {
		s.appLogger.Error("Failed to marshal updated active effects for HP Regen buff", err, "target_id", target.ID)
		return // ไม่บันทึกถ้า marshal ไม่ได้
	}
	target.ActiveEffects = newEffectsJSON

	s.appLogger.Info("Applied BUFF_HP_REGEN effect", "target", target.ID, "duration", duration, "heal_per_turn", healPerTurn)
}

func (s *combatService) applySynergyGrantStanceL(caster *domain.Combatant, target *domain.Combatant, effectData map[string]interface{}) {
	// --- 1. Get Value (เหมือนเดิม) ---
	value := 0
	if v, ok := effectData["value"]; ok {
		value = int(v.(float64))
	}

	// --- 2. Get Duration (เหมือนเดิม, already has error check) ---
	durationFloat, ok := effectData["duration"].(float64)
	if !ok {
		s.appLogger.Warn("Invalid or missing duration in effectData for applySynergyGrantStanceL", "data", effectData) // Corrected log data
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

	// --- ⭐️ 4. Add "Replace" Logic ⭐️ ---
	var tempEffects []domain.ActiveEffect
	for _, effect := range activeEffects {
		if effect.EffectID != 201 { // 👈 Change ID check to 201
			tempEffects = append(tempEffects, effect)
		} else {
			s.appLogger.Info("Replacing existing Stance L buff", "target_id", target.ID) // 👈 Update log message
		}
	}
	activeEffects = tempEffects // Use the filtered list

	// --- 5. Create and Append New Effect (เหมือนเดิม) ---
	newEffect := domain.ActiveEffect{
		EffectID:       201, // Correct ID for Stance L
		Value:          value,
		TurnsRemaining: duration,
		SourceID:       caster.ID,
	}
	activeEffects = append(activeEffects, newEffect)

	// --- ⭐️ 6. Marshal and Save (เพิ่ม Error Check) ⭐️ ---
	newEffectsJSON, err := json.Marshal(activeEffects) // 👈 Add error check variable 'err'
	if err != nil {
		s.appLogger.Error("Failed to marshal updated active effects for Stance L buff", err, "target_id", target.ID) // 👈 Add error handling
		return
	}
	target.ActiveEffects = newEffectsJSON

	// --- 7. Log (เหมือนเดิม) ---
	s.appLogger.Info("Applied SYNERGY_GRANT_STANCE_L effect", "target", target.ID, "duration", duration)
}

// --- ⭐️ เพิ่ม ผู้เชี่ยวชาญ Stance G! (เวอร์ชันแก้ไข Bug ซ้อนทับ) ⭐️ ---
func (s *combatService) applySynergyGrantStanceG(caster *domain.Combatant, target *domain.Combatant, effectData map[string]interface{}) {
	// --- 1. Get Value (เหมือนเดิม) ---
	value := 0
	if v, ok := effectData["value"]; ok {
		value = int(v.(float64))
	}

	// --- 2. Get Duration (เหมือนเดิม, already has error check) ---
	durationFloat, ok := effectData["duration"].(float64)
	if !ok {
		s.appLogger.Warn("Invalid or missing duration in effectData for applySynergyGrantStanceG", "data", effectData) // Corrected log data
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

	// --- ⭐️ 4. Add "Replace" Logic ⭐️ ---
	var tempEffects []domain.ActiveEffect
	for _, effect := range activeEffects {
		if effect.EffectID != 202 { // 👈 Change ID check to 202
			tempEffects = append(tempEffects, effect)
		} else {
			s.appLogger.Info("Replacing existing Stance G buff", "target_id", target.ID) // 👈 Update log message
		}
	}
	activeEffects = tempEffects // Use the filtered list

	// --- 5. Create and Append New Effect (เหมือนเดิม) ---
	newEffect := domain.ActiveEffect{
		EffectID:       202, // Correct ID for Stance G
		Value:          value,
		TurnsRemaining: duration,
		SourceID:       caster.ID,
	}
	activeEffects = append(activeEffects, newEffect)

	// --- ⭐️ 6. Marshal and Save (เพิ่ม Error Check) ⭐️ ---
	newEffectsJSON, err := json.Marshal(activeEffects) // 👈 Add error check variable 'err'
	if err != nil {
		s.appLogger.Error("Failed to marshal updated active effects for Stance G buff", err, "target_id", target.ID) // 👈 Add error handling
		return
	}
	target.ActiveEffects = newEffectsJSON

	// --- 7. Log (เหมือนเดิม) ---
	s.appLogger.Info("Applied SYNERGY_GRANT_STANCE_G effect", "target", target.ID, "duration", duration)
}

// --- ⭐️ เพิ่ม ผู้เชี่ยวชาญ Stance P! (เวอร์ชันแก้ไข Bug ซ้อนทับ) ⭐️ ---
func (s *combatService) applySynergyGrantStanceP(caster *domain.Combatant, target *domain.Combatant, effectData map[string]interface{}) {
	// --- 1. Get Value (เหมือนเดิม) ---
	value := 0
	if v, ok := effectData["value"]; ok {
		value = int(v.(float64))
	}

	// --- 2. Get Duration (เหมือนเดิม, already has error check) ---
	durationFloat, ok := effectData["duration"].(float64)
	if !ok {
		s.appLogger.Warn("Invalid or missing duration in effectData for applySynergyGrantStanceP", "data", effectData) // Corrected log data
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

	// --- ⭐️ 4. Add "Replace" Logic ⭐️ ---
	var tempEffects []domain.ActiveEffect
	for _, effect := range activeEffects {
		if effect.EffectID != 203 { // 👈 Change ID check to 203
			tempEffects = append(tempEffects, effect)
		} else {
			s.appLogger.Info("Replacing existing Stance P buff", "target_id", target.ID) // 👈 Update log message
		}
	}
	activeEffects = tempEffects // Use the filtered list

	// --- 5. Create and Append New Effect (เหมือนเดิม) ---
	newEffect := domain.ActiveEffect{
		EffectID:       203, // Correct ID for Stance P
		Value:          value,
		TurnsRemaining: duration,
		SourceID:       caster.ID,
	}
	activeEffects = append(activeEffects, newEffect)

	// --- ⭐️ 6. Marshal and Save (เพิ่ม Error Check) ⭐️ ---
	newEffectsJSON, err := json.Marshal(activeEffects) // 👈 Add error check variable 'err'
	if err != nil {
		s.appLogger.Error("Failed to marshal updated active effects for Stance P buff", err, "target_id", target.ID) // 👈 Add error handling
		return
	}
	target.ActiveEffects = newEffectsJSON

	// --- 7. Log (เหมือนเดิม) ---
	s.appLogger.Info("Applied SYNERGY_GRANT_STANCE_P effect", "target", target.ID, "duration", duration)
}

func (s *combatService) applyShield(caster *domain.Combatant, target *domain.Combatant, effectData map[string]interface{}, spell *domain.Spell) {
	// --- การตรวจสอบ Type Assertion ---
	effectIDFloat, ok1 := effectData["effect_id"].(float64)
	baseValueFloat, ok2 := effectData["value"].(float64)
	durationFloat, ok3 := effectData["duration"].(float64) // ⭐️ 1. อ่าน Duration
	powerModifierFloat, ok4 := effectData["power_modifier"].(float64)

	if !ok1 || !ok2 {
		s.appLogger.Warn("Invalid or missing effect_id or value in effectData for applyShield", effectData)
		return
	}
	if !ok4 {
		powerModifierFloat = 1.0
	}
	// -----------------------------

	tempSpellEffect := &domain.SpellEffect{
		EffectID:  uint(effectIDFloat),
		BaseValue: baseValueFloat,
	}
	powerModifier := powerModifierFloat

	// --- ⭐️ 2. Logic การหา Duration ที่เราแก้ไขแล้ว (สมบูรณ์) ⭐️ ---
	var shieldDuration int = 0
	if ok3 && durationFloat > 0 {
		shieldDuration = int(durationFloat)
		s.appLogger.Info("Using Shield's own duration from effectData", "duration", shieldDuration)
	}
	if shieldDuration == 0 {
		foundStanceDuration := false
		for _, effect := range spell.Effects {
			if effect.EffectID >= 200 && effect.EffectID <= 203 && effect.DurationInTurns > 0 {
				shieldDuration = effect.DurationInTurns
				foundStanceDuration = true
				s.appLogger.Info("Using Stance duration for Shield", "stance_effect_id", effect.EffectID, "duration", shieldDuration)
				break
			}
		}
		if !foundStanceDuration {
			s.appLogger.Warn("No Stance effect found for Shield spell, attempting default", "spell_id", spell.ID)
		}
	}
	if shieldDuration == 0 {
		shieldDuration = 1
		s.appLogger.Warn("No duration found in effectData or Stance, defaulting duration", "spell_id", spell.ID, "default_duration", shieldDuration)
	}
	// --- ⭐️ สิ้นสุด Logic Duration ⭐️ ---

	// --- ⭐️ Logic คำนวณ Shield HP (เหมือนเดิม) ⭐️ ---
	calculatedShieldHP, err := s.calculateEffectValue(caster, target, spell, tempSpellEffect, powerModifier)
	if err != nil {
		s.appLogger.Error("Error calculating Shield HP value", err)
		calculatedShieldHP = tempSpellEffect.BaseValue * powerModifier // Fallback
	}
	shieldHP := int(math.Round(calculatedShieldHP))
	if shieldHP < 0 {
		shieldHP = 0
	}
	// --- ⭐️ สิ้นสุด Logic Shield HP ⭐️ ---

	var activeEffects []domain.ActiveEffect
	if target.ActiveEffects != nil {
		err := json.Unmarshal(target.ActiveEffects, &activeEffects)
		if err != nil {
			s.appLogger.Error("Failed to unmarshal existing active effects for Shield buff", err, "target_id", target.ID)
			activeEffects = []domain.ActiveEffect{}
		}
	}

	// --- ⭐️ ณัชชาเพิ่มตรงนี้ 3: Logic "อันใหม่ทับอันเก่า" (Replace) ⭐️ ---
	var tempEffects []domain.ActiveEffect
	for _, effect := range activeEffects {
		if effect.EffectID != 2 { // เก็บ Effect อื่นๆ ที่ไม่ใช่ Shield (ID 2)
			tempEffects = append(tempEffects, effect)
		} else {
			s.appLogger.Info("Replacing existing Shield buff", "target_id", target.ID, "old_shield_value", effect.Value)
		}
	}
	activeEffects = tempEffects // ใช้ list ที่กรอง Shield อันเก่าออกแล้ว
	// --- ⭐️ สิ้นสุด Logic Replace ⭐️ ---

	// สร้าง Object Shield Effect ใหม่
	newEffect := domain.ActiveEffect{
		EffectID:       2,
		Value:          shieldHP,
		TurnsRemaining: shieldDuration,
		SourceID:       caster.ID,
	}
	activeEffects = append(activeEffects, newEffect) // เพิ่ม Shield "อันใหม่" เข้าไป
	newEffectsJSON, _ := json.Marshal(activeEffects)
	target.ActiveEffects = newEffectsJSON

	s.appLogger.Info("Applied SHIELD effect", "target", target.ID, "shield_hp", shieldHP, "duration", shieldDuration)
}

// --- ⭐️ เพิ่ม ผู้เชี่ยวชาญ บัฟ Evasion! ⭐️ ---
func (s *combatService) applyBuffEvasion(caster *domain.Combatant, target *domain.Combatant, effectData map[string]interface{}) {
	// --- การตรวจสอบ Type Assertion ---
	// Blur (ID 10) มี BaseValue: 20, Duration: 1
	valueFloat, ok1 := effectData["value"].(float64)       // % หลบหลีก
	durationFloat, ok2 := effectData["duration"].(float64) // ระยะเวลา
	if !ok1 || !ok2 {
		s.appLogger.Warn("Invalid or missing value or duration in effectData for applyBuffEvasion", "data", effectData)
		return
	}
	// -----------------------------

	// --- ⭐️ Evasion ควร Scale กับ Talent ไหม? (เช่น Talent G?) ⭐️ ---
	// Option 1: ไม่ scale ใช้ BaseValue ตรงๆ (ง่ายสุด - ตอนนี้ใช้อันนี้)
	evasionPercent := int(math.Round(valueFloat))
	// Option 2: Scale กับ Talent G -> ต้องเรียก calculateEffectValue และส่ง spell
	// -----------------------------------------------------------------

	duration := int(durationFloat)
	// กันค่า % แปลกๆ (อยู่ระหว่าง 0-100)
	if evasionPercent < 0 {
		evasionPercent = 0
	}
	if evasionPercent > 100 {
		evasionPercent = 100
	} // อาจจะ Cap ที่ 95%?

	var activeEffects []domain.ActiveEffect
	if target.ActiveEffects != nil {
		err := json.Unmarshal(target.ActiveEffects, &activeEffects)
		if err != nil {
			s.appLogger.Error("Failed to unmarshal existing active effects for Evasion buff", err, "target_id", target.ID)
			activeEffects = []domain.ActiveEffect{}
		}
	}

	// --- ⭐️ จัดการ Evasion Stack ยังไง? ⭐️ ---
	// Option A: ไม่ Stack, อันใหม่ทับอันเก่า (ง่ายสุด - ตอนนี้ทำแบบนี้)
	// Option B: Stack แต่มี Cap (ซับซ้อนขึ้น)
	// Option C: Refresh Duration อันเก่า (เหมือนที่เราคุยกันเรื่อง Stance)
	// ตอนนี้ทำแบบ A: ลบอันเก่า (ถ้ามี) ก่อนเพิ่มอันใหม่
	var tempEffects []domain.ActiveEffect
	for _, effect := range activeEffects {
		if effect.EffectID != 102 { // เก็บ Effect อื่นๆ ไว้
			tempEffects = append(tempEffects, effect)
		} else {
			s.appLogger.Info("Replacing existing Evasion buff", "target_id", target.ID)
		}
	}
	activeEffects = tempEffects // ใช้ list ที่กรองแล้ว
	// ------------------------------------

	// สร้าง Object บัฟใหม่
	newEffect := domain.ActiveEffect{
		EffectID:       102,            // BUFF_EVASION
		Value:          evasionPercent, // % หลบหลีก
		TurnsRemaining: duration,
		SourceID:       caster.ID,
	}
	activeEffects = append(activeEffects, newEffect)

	// แปลงกลับเป็น JSON
	newEffectsJSON, _ := json.Marshal(activeEffects) // Error handling?
	target.ActiveEffects = newEffectsJSON

	s.appLogger.Info("Applied BUFF_EVASION effect", "target", target.ID, "duration", duration, "evasion_percent", evasionPercent)
}

// --- ⭐️ เพิ่ม ผู้เชี่ยวชาญ บัฟ Damage Up! ⭐️ ---
func (s *combatService) applyBuffDamageUp(caster *domain.Combatant, target *domain.Combatant, effectData map[string]interface{}) {
	// --- การตรวจสอบ Type Assertion ---
	// Empower (ID 15) มี BaseValue: 15, Duration: 1
	valueFloat, ok1 := effectData["value"].(float64)       // % เพิ่ม Damage
	durationFloat, ok2 := effectData["duration"].(float64) // ระยะเวลา
	if !ok1 || !ok2 {
		s.appLogger.Warn("Invalid or missing value or duration in effectData for applyBuffDamageUp", "data", effectData)
		return
	}
	// -----------------------------

	// --- ⭐️ Damage Buff ควร Scale กับ Talent ไหม? (เช่น Talent P?) ⭐️ ---
	// Option 1: ไม่ scale ใช้ BaseValue ตรงๆ (ง่ายสุด - ตอนนี้ใช้อันนี้)
	damageIncreasePercent := int(math.Round(valueFloat))
	// Option 2: Scale กับ Talent P -> ต้องเรียก calculateEffectValue และส่ง spell
	// -----------------------------------------------------------------

	duration := int(durationFloat)
	if damageIncreasePercent < 0 {
		damageIncreasePercent = 0
	} // ไม่ควรติดลบ

	// เป้าหมายของ Damage Buff คือ Caster เสมอ (ตาม Logic ใน executeCastSpell ที่ override target สำหรับ Buff)
	var activeEffects []domain.ActiveEffect
	if target.ActiveEffects != nil { // target ในที่นี้คือ caster
		err := json.Unmarshal(target.ActiveEffects, &activeEffects)
		if err != nil {
			s.appLogger.Error("Failed to unmarshal existing active effects for Damage Up buff", err, "target_id", target.ID)
			activeEffects = []domain.ActiveEffect{}
		}
	}

	// --- ⭐️ จัดการ Stack ยังไง? (เหมือน Evasion: อันใหม่ทับอันเก่า) ⭐️ ---
	var tempEffects []domain.ActiveEffect
	for _, effect := range activeEffects {
		if effect.EffectID != 103 {
			tempEffects = append(tempEffects, effect)
		} else {
			s.appLogger.Info("Replacing existing Damage Up buff", "target_id", target.ID)
		}
	}
	activeEffects = tempEffects
	// ------------------------------------

	// สร้าง Object บัฟใหม่
	newEffect := domain.ActiveEffect{
		EffectID:       103,                   // BUFF_DAMAGE_UP
		Value:          damageIncreasePercent, // % เพิ่ม Damage
		TurnsRemaining: duration,
		SourceID:       caster.ID, // Source คือ caster คนเดิม
	}
	activeEffects = append(activeEffects, newEffect)

	// แปลงกลับเป็น JSON
	newEffectsJSON, _ := json.Marshal(activeEffects)
	target.ActiveEffects = newEffectsJSON // target คือ caster

	s.appLogger.Info("Applied BUFF_DAMAGE_UP effect", "target", target.ID, "duration", duration, "damage_increase_percent", damageIncreasePercent)
}

// --- ⭐️ เพิ่ม "ผู้เชี่ยวชาญ" คนใหม่: แปะสถานะ Ignite! ⭐️ ---
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
		if effect.EffectID != 306 { // เก็บ Effect อื่นๆ ที่ไม่ใช่ Ignite
			tempEffects = append(tempEffects, effect)
		} else {
			s.appLogger.Info("Replacing existing Ignite debuff", "target_id", target.ID)
		}
	}
	activeEffects = tempEffects // ใช้ list ที่กรอง Ignite อันเก่าออกแล้ว
	// -----------------------------------------------------------------

	// สร้าง Object Debuff ใหม่
	newEffect := domain.ActiveEffect{
		EffectID:       306,        // DEBUFF_IGNITE
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

// --- ⭐️ เพิ่ม "ผู้เชี่ยวชาญ" คนใหม่: แปะสถานะ Vulnerable! ⭐️ ---
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
		if effect.EffectID != 302 { // เก็บ Effect อื่นๆ ที่ไม่ใช่ Vulnerable
			tempEffects = append(tempEffects, effect)
		} else {
			s.appLogger.Info("Replacing existing Vulnerable debuff", "target_id", target.ID)
		}
	}
	activeEffects = tempEffects // ใช้ list ที่กรองอันเก่าออกแล้ว
	// -----------------------------------------------------------------

	// สร้าง Object Debuff ใหม่
	newEffect := domain.ActiveEffect{
		EffectID:       302,                  // DEBUFF_VULNERABLE
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

// --- ⭐️ เพิ่ม "ผู้เชี่ยวชาญ" คนใหม่: แปะสถานะ Retaliation! ⭐️ ---
func (s *combatService) applyBuffRetaliation(caster *domain.Combatant, target *domain.Combatant, effectData map[string]interface{}) {
	// --- การตรวจสอบ Type Assertion ---
	// เวท ID 14 (StaticField) มี BaseValue: 10, Duration: 2 (หลังจากเราแก้ Seeder)
	valueFloat, ok1 := effectData["value"].(float64)       // Damage ที่จะสะท้อน
	durationFloat, ok2 := effectData["duration"].(float64) // ระยะเวลา
	if !ok1 || !ok2 {
		s.appLogger.Warn("Invalid or missing value or duration in effectData for applyBuffRetaliation", "data", effectData)
		return
	}
	// -----------------------------

	retaliationDamage := int(math.Round(valueFloat))
	duration := int(durationFloat)
	if retaliationDamage < 0 {
		retaliationDamage = 0
	}

	var activeEffects []domain.ActiveEffect
	if target.ActiveEffects != nil {
		err := json.Unmarshal(target.ActiveEffects, &activeEffects)
		if err != nil {
			s.appLogger.Error("Failed to unmarshal existing active effects for Retaliation buff", err, "target_id", target.ID)
			activeEffects = []domain.ActiveEffect{}
		}
	}

	// --- ⭐️ จัดการ Stack: (อันใหม่ทับอันเก่า) ⭐️ ---
	var tempEffects []domain.ActiveEffect
	for _, effect := range activeEffects {
		if effect.EffectID != 104 { // เก็บ Effect อื่นๆ ที่ไม่ใช่ Retaliation
			tempEffects = append(tempEffects, effect)
		} else {
			s.appLogger.Info("Replacing existing Retaliation buff", "target_id", target.ID)
		}
	}
	activeEffects = tempEffects
	// -----------------------------------------------------------------

	// สร้าง Object Buff ใหม่
	newEffect := domain.ActiveEffect{
		EffectID:       104,               // BUFF_RETALIATION
		Value:          retaliationDamage, // Damage ที่จะสะท้อน
		TurnsRemaining: duration,
		SourceID:       caster.ID,
	}
	activeEffects = append(activeEffects, newEffect)

	// แปลงกลับเป็น JSON
	newEffectsJSON, err := json.Marshal(activeEffects)
	if err != nil {
		s.appLogger.Error("Failed to marshal updated active effects for Retaliation buff", err, "target_id", target.ID)
		return
	}
	target.ActiveEffects = newEffectsJSON

	s.appLogger.Info("Applied BUFF_RETALIATION effect", "target", target.ID, "duration", duration, "retaliation_damage", retaliationDamage)
}
