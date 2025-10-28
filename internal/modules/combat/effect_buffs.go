// file: internal/modules/combat/effect_buffs.go
package combat

import (
	"encoding/json"
	"math"
	"sage-of-elements-backend/internal/domain"
)

// ============================================================================
// 📌 BUFF EFFECTS (2000s Range)
// ============================================================================
// Effect IDs: 2101 (HP Regen), 2102 (MP Regen), 2201 (Evasion),
//             2202 (Damage Up), 2203 (Retaliation), 2204 (Defense Up)
// ============================================================================

// --- ⭐️ บัฟ HP Regen ⭐️ ---
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
		EffectID:       2101,        // BUFF_HP_REGEN
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

// --- ⭐️ บัฟ MP Regen ⭐️ ---
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
		EffectID:       2102,         // BUFF_MP_REGEN
		Value:          regenPerTurn, // ค่า MP ที่จะฟื้นต่อเทิร์น
		TurnsRemaining: duration,
		SourceID:       caster.ID,
	}
	activeEffects = append(activeEffects, newEffect)
	newEffectsJSON, _ := json.Marshal(activeEffects)
	target.ActiveEffects = newEffectsJSON

	s.appLogger.Info("Applied BUFF_MP_REGEN effect", "target", target.ID, "duration", duration, "regen_per_turn", regenPerTurn)
}

// --- ⭐️ บัฟ Evasion ⭐️ ---
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
		if effect.EffectID != 2201 { // BUFF_EVASION - เก็บ Effect อื่นๆ ไว้
			tempEffects = append(tempEffects, effect)
		} else {
			s.appLogger.Info("Replacing existing Evasion buff", "target_id", target.ID)
		}
	}
	activeEffects = tempEffects // ใช้ list ที่กรองแล้ว
	// ------------------------------------

	// สร้าง Object บัฟใหม่
	newEffect := domain.ActiveEffect{
		EffectID:       2201,           // BUFF_EVASION
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

// --- ⭐️ บัฟ Damage Up ⭐️ ---
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
		if effect.EffectID != 2202 { // BUFF_DMG_UP
			tempEffects = append(tempEffects, effect)
		} else {
			s.appLogger.Info("Replacing existing Damage Up buff", "target_id", target.ID)
		}
	}
	activeEffects = tempEffects
	// ------------------------------------

	// สร้าง Object บัฟใหม่
	newEffect := domain.ActiveEffect{
		EffectID:       2202,                  // BUFF_DMG_UP
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

// --- ⭐️ บัฟ Retaliation ⭐️ ---
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
		if effect.EffectID != 2203 { // BUFF_RETALIATION - เก็บ Effect อื่นๆ
			tempEffects = append(tempEffects, effect)
		} else {
			s.appLogger.Info("Replacing existing Retaliation buff", "target_id", target.ID)
		}
	}
	activeEffects = tempEffects
	// -----------------------------------------------------------------

	// สร้าง Object Buff ใหม่
	newEffect := domain.ActiveEffect{
		EffectID:       2203,              // BUFF_RETALIATION
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

// --- ⭐️ บัฟ Defense Up ⭐️ ---
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
		EffectID:       2204, // BUFF_DEFENSE_UP
		Value:          value,
		TurnsRemaining: duration,
		SourceID:       caster.ID,
	}
	activeEffects = append(activeEffects, newEffect)
	newEffectsJSON, _ := json.Marshal(activeEffects)
	target.ActiveEffects = newEffectsJSON
	s.appLogger.Info("Applied BUFF_DEFENSE_UP effect", "target", target.ID, "duration", duration)
}
