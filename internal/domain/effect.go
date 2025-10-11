package domain

import "gorm.io/datatypes"

// Effect คือตาราง Master Data ที่เก็บ "ประเภท" ของเอฟเฟกต์ทั้งหมด
type Effect struct {
	ID           uint              `gorm:"primaryKey"`
	Name         string            `gorm:"size:100;not null;unique;comment:ชื่อเอฟเฟกต์ในระบบ (Eng, UNIQUE)"`
	Type         EffectType        `gorm:"size:50;not null;comment:ประเภทหลัก (BUFF, DEBUFF, DAMAGE, etc.)"`
	DisplayNames datatypes.JSONMap `gorm:"type:jsonb;comment:ชื่อที่แสดงผลในเกม (เช่น เชื่องช้า)"`
	Descriptions datatypes.JSONMap `gorm:"type:jsonb;comment:คำอธิบายการทำงาน (Tooltip)"`
}

type EffectType string

// กำหนดค่าคงที่สำหรับ EffectType ทั้งหมด
const (
	EffectTypeDamage       EffectType = "DAMAGE"         // เอฟเฟกต์สร้างความเสียหายพื้นฐาน (กายภาพ/เวท)
	EffectTypeTrueDamage   EffectType = "TRUE_DAMAGE"    // เอฟเฟกต์สร้างความเสียหายจริง (ทะลุเกราะ)
	EffectTypeShield       EffectType = "SHIELD"         // เอฟเฟกต์สร้างเกราะป้องกัน
	EffectTypeHeal         EffectType = "HEAL"           // เอฟเฟกต์ฟื้นฟู HP
	EffectTypeResource     EffectType = "RESOURCE"       // เอฟเฟกต์ที่เกี่ยวกับทรัพยากร (เพิ่ม/ลด AP/MP)
	EffectTypeUtility      EffectType = "UTILITY"        // เอฟเฟกต์ช่วยเหลืออื่นๆ (เช่น ลบล้างสถานะ)
	EffectTypeBuff         EffectType = "BUFF"           // เอฟเฟกต์เสริมพลังทั่วไป
	EffectTypeDebuff       EffectType = "DEBUFF"         // เอฟเฟกต์ลดความสามารถทั่วไป
	EffectTypeDebuffCC     EffectType = "DEBUFF_CC"      // เอฟเฟกต์ขัดขวางการเคลื่อนไหว (เช่น Slow, Root)
	EffectTypeDebuffHardCC EffectType = "DEBUFF_HARD_CC" // เอฟเฟกต์ขัดขวางการกระทำเด็ดขาด (เช่น Stun)
	EffectTypeSynergyBuff  EffectType = "BUFF_SYNERGY"   // เอฟเฟกต์พิเศษที่ได้จาก Synergy
	EffectTypeDebuffDOT    EffectType = "DEBUFF_DOT"     // เอฟเฟกต์สร้างความเสียหายต่อเนื่อง (เช่น เผาไหม้, พิษ)
)
