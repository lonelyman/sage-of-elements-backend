package domain

import "gorm.io/datatypes"

// Realm คือตาราง Master Data ที่เก็บ "โหมด PvE" ทั้งหมด
type Realm struct {
	ID              uint           `gorm:"primaryKey;comment:ID เฉพาะของโหมด (PK)"`
	Name            string         `gorm:"size:100;not null;unique;comment:ชื่อในระบบ (Eng)"`
	DisplayNames    datatypes.JSON `gorm:"type:jsonb;comment:ชื่อโหมดที่แสดงในเกม"`
	Descriptions    datatypes.JSON `gorm:"type:jsonb;comment:คำอธิบายโหมด"`
	IsActive        bool           `gorm:"not null;default:true;comment:โหมดนี้เปิดให้เล่นอยู่หรือไม่"`
	UnlockCondition datatypes.JSON `gorm:"type:jsonb;comment:เงื่อนไขในการปลดล็อก (JSON)"`
}

// Chapter คือบทของเนื้อเรื่อง หรือกลุ่มของด่านที่อยู่ใน Realm เดียวกัน
type Chapter struct {
	ID            uint           `gorm:"primaryKey;comment:ID เฉพาะของบท (PK)"`
	RealmID       uint           `gorm:"not null;comment:ID ของ Realm ที่บทนี้สังกัด (FK to realms)"`
	ChapterNumber int            `gorm:"not null;comment:ลำดับของบท (1, 2, 3...)"`
	Name          string         `gorm:"size:100;not null;unique;comment:ชื่อในระบบ (Eng)"`
	DisplayNames  datatypes.JSON `gorm:"type:jsonb;comment:ชื่อบทที่แสดงในเกม"`
	Descriptions  datatypes.JSON `gorm:"type:jsonb;comment:คำอธิบายเรื่องย่อของบท"`
	Realm         *Realm         `gorm:"foreignKey:RealmID;references:ID"` // GORM Preload
}

// Stage คือ "ด่าน" 1 ด่านที่ผู้เล่นจะต้องเข้าไปต่อสู้
type Stage struct {
	ID                uint           `gorm:"primaryKey;comment:ID เฉพาะของด่าน (PK)"`
	ChapterID         uint           `gorm:"not null;comment:ID ของ Chapter ที่ด่านนี้สังกัด (FK to chapters)"`
	StageNumber       int            `gorm:"not null;comment:ลำดับของด่านในบท (1, 2, 3...)"`
	Name              string         `gorm:"size:100;not null;unique;comment:ชื่อในระบบ (Eng)"`
	DisplayNames      datatypes.JSON `gorm:"type:jsonb;comment:ชื่อด่านที่แสดงในเกม"`
	Descriptions      datatypes.JSON `gorm:"type:jsonb;comment:คำอธิบายเรื่องย่อของด่าน"`
	StageType         StageType      `gorm:"size:50;not null;comment:ประเภทของด่าน (STORY, ELITE, BOSS)"`
	FirstClearRewards datatypes.JSON `gorm:"type:jsonb;comment:รางวัลสำหรับการเคลียร์ครั้งแรก (JSON)"`
	Chapter           *Chapter       `gorm:"foreignKey:ChapterID;references:ID"` // GORM Preload
}

// --- Constants / Enums ---
// StageType คือประเภทของด่าน
type StageType string

const (
	StageTypeStory StageType = "STORY" // ด่านเนื้อเรื่องปกติ
	StageTypeElite StageType = "ELITE" // ด่านระดับสูง (มอนสเตอร์เก่งกว่าปกติ)
	StageTypeBoss  StageType = "BOSS"  // ด่านบอส
)

// StageEnemy คือข้อมูลการปรากฏตัวของศัตรู 1 ตัวในด่าน 1 ด่าน
type StageEnemy struct {
	ID       uint `gorm:"primaryKey;comment:ID เฉพาะของแถวข้อมูล (PK)"`
	StageID  uint `gorm:"not null;comment:ID ของด่าน (FK to stages)"`
	EnemyID  uint `gorm:"not null;comment:ID ของศัตรู (FK to enemies)"`
	Position int  `gorm:"not null;comment:ตำแหน่งในสนามรบ (เช่น 1=หน้า, 2=หลัง)"`

	// --- ✨ ฟิลด์พิเศษเพื่อความยืดหยุ่น! ✨ ---
	LevelOverride *int `gorm:"comment:ใช้ Level นี้แทนค่า Default ในตาราง enemies (ถ้าไม่ NULL)"`
}
