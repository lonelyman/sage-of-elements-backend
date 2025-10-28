// file: internal/modules/combat/effect_direct.go
package combat

import (
	"encoding/json"
	"math"
	"math/rand"
	"sage-of-elements-backend/internal/domain"
	"sort"
)

// ============================================================================
// 📌 DIRECT EFFECTS (1000s Range)
// ============================================================================
// Effect IDs: 1101 (Damage), 1102 (Shield), 1103 (Heal), 1104 (MP Damage)
// ============================================================================

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
				if effect.EffectID == 2201 { // BUFF_EVASION
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

	// --- ⭐️ ขั้นตอนที่ 4: Logic เช็ค Vulnerable (ID 4102) บนเป้าหมาย ⭐️ ---
	// (เพิ่ม Damage ที่จะได้รับ)
	var targetActiveEffectsForVulnerable []domain.ActiveEffect
	damageIncreasePercent := 0 // % Damage ที่จะเพิ่มขึ้น (Default = 0)
	hasVulnerableDebuff := false

	if target.ActiveEffects != nil {
		// ⭐️ ใช้ตัวแปรใหม่ (targetActiveEffectsForVulnerable)
		err := json.Unmarshal(target.ActiveEffects, &targetActiveEffectsForVulnerable)
		if err == nil {
			for _, effect := range targetActiveEffectsForVulnerable {
				if effect.EffectID == 4102 { // DEBUFF_VULNERABLE
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

	// --- ⭐️ ขั้นตอนที่ 6: Logic จัดการ Shield (ID 1102) บนเป้าหมาย ⭐️ ---
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
				isIShield := tempEffects[i].EffectID == 1102 // SHIELD
				isJShield := tempEffects[j].EffectID == 1102
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
				if tempEffects[i].EffectID == 1102 && tempEffects[i].Value > 0 && remainingDamage > 0 {
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
				if !(effect.EffectID == 1102 && effect.Value <= 0) {
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

	// --- ⭐️ ขั้นตอนที่ 8: (ใหม่!) Logic เช็ค Retaliation (ID 2203) บนเป้าหมาย ⭐️ ---
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
				if effect.EffectID == 2203 && effect.Value > 0 { // BUFF_RETALIATION
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

// --- ⭐️ เพิ่ม ผู้เชี่ยวชาญ Shield! ⭐️ ---
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
			if effect.EffectID >= 3101 && effect.EffectID <= 3104 && effect.DurationInTurns > 0 { // STANCE_S/L/G/P
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
		if effect.EffectID != 1102 { // SHIELD - เก็บ Effect อื่นๆ
			tempEffects = append(tempEffects, effect)
		} else {
			s.appLogger.Info("Replacing existing Shield buff", "target_id", target.ID, "old_shield_value", effect.Value)
		}
	}
	activeEffects = tempEffects // ใช้ list ที่กรอง Shield อันเก่าออกแล้ว
	// --- ⭐️ สิ้นสุด Logic Replace ⭐️ ---

	// สร้าง Object Shield Effect ใหม่
	newEffect := domain.ActiveEffect{
		EffectID:       1102, // SHIELD
		Value:          shieldHP,
		TurnsRemaining: shieldDuration,
		SourceID:       caster.ID,
	}
	activeEffects = append(activeEffects, newEffect) // เพิ่ม Shield "อันใหม่" เข้าไป
	newEffectsJSON, _ := json.Marshal(activeEffects)
	target.ActiveEffects = newEffectsJSON

	s.appLogger.Info("Applied SHIELD effect", "target", target.ID, "shield_hp", shieldHP, "duration", shieldDuration)
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
	if uint(effectIDFloat) == 1104 { // MP_DAMAGE
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
		s.appLogger.Info("DEBUG: Condition if uint(effectIDFloat) == 1104 is FALSE", "effectIDFloat", effectIDFloat) // ⭐️ Log 5 (เผื่อเช็ค ID ผิด) ⭐️
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
