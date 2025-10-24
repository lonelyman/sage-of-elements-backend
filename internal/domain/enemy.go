// file: internal/domain/enemy.go
package domain

import "gorm.io/datatypes"

// Enemy คือตาราง Master Data ที่เก็บข้อมูลศัตรูทุกชนิดในเกม
type Enemy struct {
	ID           uint           `gorm:"primaryKey;comment:ID เฉพาะของศัตรู (PK)"`
	Name         string         `gorm:"size:100;not null;unique;comment:ชื่อในระบบ (Eng)"`
	DisplayNames datatypes.JSON `gorm:"type:jsonb;comment:ชื่อที่แสดงผลในแต่ละภาษา"`
	Descriptions datatypes.JSON `gorm:"type:jsonb;comment:คำอธิบายเชิง Lore"`
	Level        int            `gorm:"not null;comment:เลเวลพื้นฐานของศัตรู"`
	MaxHP        int            `gorm:"not null;comment:พลังชีวิตสูงสุดพื้นฐาน"`
	Initiative   int            `gorm:"not null;comment:ค่าความเร็วพื้นฐาน"`
	MaxEndurance int            `gorm:"not null;comment:ค่าความอดทนสูงสุดพื้นฐาน"`
	ElementID    uint           `gorm:"comment:ID ของธาตุประจำตัวศัตรู (FK to elements)"`
	Element      *Element       `gorm:"foreignKey:ElementID;references:ID"`

	// --- ⭐️ เพิ่ม 3 บรรทัดนี้เข้ามาเพื่อสร้าง "ความสัมพันธ์"! ⭐️ ---
	// นี่คือ "สะพาน" ที่บอก GORM ว่า Enemy มี Abilities, AI, และ Loots ได้หลายอัน
	Abilities []*EnemyAbility `gorm:"foreignKey:EnemyID;constraint:OnDelete:CASCADE;"`
	AI        []*EnemyAI      `gorm:"foreignKey:EnemyID;constraint:OnDelete:CASCADE;"`
	Loots     []*EnemyLoot    `gorm:"foreignKey:EnemyID;constraint:OnDelete:CASCADE;"`
}

// EnemyAbility คือท่าโจมตีหรือความสามารถ 1 อย่างของศัตรู
type EnemyAbility struct {
	ID           uint           `gorm:"primaryKey;comment:ID เฉพาะของความสามารถ (PK)"`
	EnemyID      uint           `gorm:"not null;comment:ID ของศัตรูที่เป็นเจ้าของท่านี้ (FK to enemies)"`
	Name         string         `gorm:"size:100;not null;unique;comment:ชื่อท่าในระบบ (Eng)"`
	DisplayNames datatypes.JSON `gorm:"type:jsonb;comment:ชื่อท่าที่แสดงในเกม"`
	Descriptions datatypes.JSON `gorm:"type:jsonb;comment:คำอธิบายการทำงานของท่า"`
	APCost       int            `gorm:"not null;comment:ต้นทุน AP ที่ศัตรูต้องใช้"`
	MPCost       int            `gorm:"default:0"`
	EffectsJSON  datatypes.JSON `gorm:"type:jsonb;comment:ผลลัพธ์ของท่าโจมตีในรูปแบบ JSON"`
}

// EnemyLoot คือตารางที่บอกว่าศัตรู 1 ชนิดมีโอกาสดรอปไอเทมอะไรบ้าง
type EnemyLoot struct {
	ID         uint    `gorm:"primaryKey"`
	EnemyID    uint    `gorm:"not null;comment:ID ของศัตรู (FK to enemies)"`
	ElementID  uint    `gorm:"not null;comment:ID ของธาตุที่จะดรอป (FK to elements)"`
	DropChance float64 `gorm:"not null;comment:โอกาสดรอป (0.0 ถึง 1.0)"`
	MinAmount  int     `gorm:"not null;default:1;comment:จำนวนดรอปขั้นต่ำ"`
	MaxAmount  int     `gorm:"not null;default:1;comment:จำนวนดรอปสูงสุด"`
}

// EnemyAI คือ "กฎ" การตัดสินใจ 1 ข้อของศัตรู
type EnemyAI struct {
	ID             uint          `gorm:"primaryKey;comment:ID เฉพาะของกฎ (PK)"`
	EnemyID        uint          `gorm:"not null;comment:ID ของศัตรูที่เป็นเจ้าของกฎนี้ (FK to enemies)"`
	Priority       int           `gorm:"not null;comment:ลำดับความสำคัญ (ยิ่งน้อยยิ่งทำก่อน)"`
	Condition      AICondition   `gorm:"size:100;not null;comment:เงื่อนไขในการทำงาน (เช่น IF_SELF_HP_BELOW)"`
	ConditionValue float64       `gorm:"comment:ค่าของเงื่อนไข (เช่น 0.5 สำหรับ 50%)"`
	Action         AIAction      `gorm:"size:100;not null;comment:การกระทำที่จะทำ (เช่น USE_ABILITY)"`
	Target         AITarget      `gorm:"size:50;not null;comment:เป้าหมายของการกระทำ (SELF, PLAYER)"`
	AbilityToUseID *uint         `gorm:"comment:ID ของท่าที่จะใช้ (ถ้า Action คือ USE_ABILITY)"`
	AbilityToUse   *EnemyAbility `gorm:"foreignKey:AbilityToUseID;references:ID"`
}

// --- Constants / Enums (พร้อมคำอธิบายภาษาไทย) ---

// AICondition คือเงื่อนไขในการทำงานของ AI
type AICondition string

const (
	AIConditionAlways            AICondition = "ALWAYS"              // ทำเสมอ (สำหรับท่า Default ที่ไม่มีเงื่อนไข)
	AIConditionSelfHPBelow       AICondition = "SELF_HP_BELOW"       // HP ของตัวเองต่ำกว่า % ที่กำหนด
	AIConditionSelfHPAbove       AICondition = "SELF_HP_ABOVE"       // HP ของตัวเองสูงกว่า % ที่กำหนด
	AIConditionTargetHPBelow     AICondition = "TARGET_HP_BELOW"     // HP ของเป้าหมายต่ำกว่า % ที่กำหนด
	AIConditionTargetHPAbove     AICondition = "TARGET_HP_ABOVE"     // HP ของเป้าหมายสูงกว่า % ที่กำหนด
	AIConditionSelfHasBuff       AICondition = "SELF_HAS_BUFF"       // ตัวเองมีบัฟ (อาจระบุชื่อบัฟใน ConditionValue)
	AIConditionSelfHasDebuff     AICondition = "SELF_HAS_DEBUFF"     // ตัวเองมีดีบัฟ
	AIConditionTargetHasBuff     AICondition = "TARGET_HAS_BUFF"     // เป้าหมายมีบัฟ
	AIConditionTargetHasDebuff   AICondition = "TARGET_HAS_DEBUFF"   // เป้าหมายมีดีบัฟ
	AIConditionTargetHasShield   AICondition = "TARGET_HAS_SHIELD"   // เป้าหมายมี Shield อยู่
	AIConditionTargetIsStaggered AICondition = "TARGET_IS_STAGGERED" // เป้าหมายติดสถานะ "เสียหลัก"
	AIConditionTurnIs            AICondition = "TURN_IS"             // เทิร์นที่... (เช่น 1, 2, 3)
	AIConditionTurnIsEven        AICondition = "TURN_IS_EVEN"        // เป็นเทิร์นเลขคู่
	AIConditionTurnIsOdd         AICondition = "TURN_IS_ODD"         // เป็นเทิร์นเลขคี่
)

// AIAction คือการกระทำที่ AI สามารถทำได้
type AIAction string

const (
	AIActionUseAbility AIAction = "USE_ABILITY" // ใช้ความสามารถพิเศษ (ตาม AbilityToUseID)
	AIActionDefend     AIAction = "DEFEND"      // ตั้งการ์ดป้องกัน (Logic พิเศษใน Combat Service)
	AIActionPass       AIAction = "PASS"        // อยู่เฉยๆ / ข้ามเทิร์น
	AIActionFlee       AIAction = "FLEE"        // พยายามหลบหนี
)

// AITarget คือเป้าหมายที่เป็นไปได้ของ AI
type AITarget string

const (
	AITargetSelf           AITarget = "SELF"             // ตัวเอง
	AITargetPlayer         AITarget = "PLAYER"           // ผู้เล่น (เป้าหมายหลัก)
	AITargetPlayerLowestHP AITarget = "PLAYER_LOWEST_HP" // ผู้เล่นที่ HP ต่ำที่สุด
	AITargetAlly           AITarget = "ALLY"             // เพื่อนร่วมทีมของศัตรู
	AITargetAllyLowestHP   AITarget = "ALLY_LOWEST_HP"   // เพื่อนร่วมทีมที่ HP ต่ำที่สุด
)
