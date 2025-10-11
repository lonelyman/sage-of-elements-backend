package domain

// SpellEffect คือตารางเชื่อมโยงที่บอกว่า "เวท" ใบนี้มี "เอฟเฟกต์" อะไรบ้าง
type SpellEffect struct {
	ID               uint    `gorm:"primaryKey"`
	SpellID          uint    `gorm:"not null"`
	EffectID         uint    `gorm:"not null"`
	Effect           *Effect `gorm:"foreignKey:EffectID;references:ID"`
	BaseValue        float64 `gorm:"not null"`
	DurationInTurns  int
	ConditionType    ConditionType `gorm:"size:50;default:'NONE'"`
	ConditionDetails string
}

type ConditionType string

// กำหนดค่าคงที่สำหรับ ConditionType ทั้งหมด
const (
	ConditionTypeNone ConditionType = "NONE" // ไม่มีเงื่อนไขพิเศษ

	// --- เงื่อนไขเกี่ยวกับ "ฝ่ายเรา" (Self) ---
	ConditionTypeSelfHasBuff    ConditionType = "SELF_HAS_BUFF"    // ฝ่ายเรามีบัฟที่กำหนด
	ConditionTypeSelfHPBelow    ConditionType = "SELF_HP_BELOW"    // HP ของเราต่ำกว่า % ที่กำหนด
	ConditionTypeSelfHPAbove    ConditionType = "SELF_HP_ABOVE"    // HP ของเราสูงกว่า % ที่กำหนด
	ConditionTypeSelfHasElement ConditionType = "SELF_HAS_ELEMENT" // เรามีธาตุที่กำหนดใน "มิติผนึก"
	ConditionTypeSelfAPIs       ConditionType = "SELF_AP_IS"       // AP ของเรามากกว่าหรือเท่ากับค่าที่กำหนด

	// --- เงื่อนไขเกี่ยวกับ "เป้าหมาย" (Target) ---
	ConditionTypeTargetHasDebuff   ConditionType = "TARGET_HAS_DEBUFF"   // เป้าหมายติดดีบัฟที่กำหนด
	ConditionTypeTargetHPBelow     ConditionType = "TARGET_HP_BELOW"     // HP ของเป้าหมายต่ำกว่า % ที่กำหนด
	ConditionTypeTargetHPAbove     ConditionType = "TARGET_HP_ABOVE"     // HP ของเป้าหมายสูงกว่า % ที่กำหนด
	ConditionTypeTargetIsGuarded   ConditionType = "TARGET_IS_GUARDED"   // เป้าหมายมี Shield อยู่
	ConditionTypeTargetIsStaggered ConditionType = "TARGET_IS_STAGGERED" // เป้าหมายติดสถานะ "เสียหลัก"
	ConditionTypeTargetHasShield   ConditionType = "TARGET_HAS_SHIELD"   // เป้าหมายมีเกราะป้องกัน (Shield) อยู่

	// --- เงื่อนไขอื่นๆ ---
	ConditionTypeElementalResonance ConditionType = "ELEMENTAL_RESONANCE" // ธาตุตรงกับพรสวรรค์
	ConditionTypeTurnCountIs        ConditionType = "TURN_COUNT_IS"       // รอบการต่อสู้เป็นเลขคู่/คี่ หรือตรงกับค่าที่กำหนด
	ConditionTypeLastSpellWas       ConditionType = "LAST_SPELL_WAS"      // เวทที่ใช้ก่อนหน้านี้เป็นเวทที่กำหนด
	ConditionTypeComboCountIs       ConditionType = "COMBO_COUNT_IS"      // ค่าคอมโบตรงกับค่าที่กำหนด

	// --- เงื่อนไขแบบสุ่ม ---
	ConditionTypeRandomChance ConditionType = "RANDOM_CHANCE" // มีโอกาส % ที่จะทำงาน
)
