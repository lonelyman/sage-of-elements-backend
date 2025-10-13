package domain

import "time"

// Character คือตารางที่เก็บข้อมูลตัวละครของผู้เล่น
type Character struct {
	ID                 uint                         `gorm:"primaryKey;comment:ID เฉพาะของตัวละคร (PK)" json:"id"`
	PlayerID           uint                         `gorm:"not null;comment:ID ของผู้เล่นที่เป็นเจ้าของ (FK to players)" json:"player_id"`
	CharacterName      string                       `gorm:"size:255;not null;comment:ชื่อตัวละครที่ผู้เล่นตั้ง" json:"character_name"`
	Gender             string                       `gorm:"size:50;not null;comment:เพศของตัวละคร (MALE, FEMALE)" json:"gender"`
	PrimaryElementID   uint                         `gorm:"not null;comment:ID ของธาตุปฐมภูมิที่เลือกตอนสร้างตัว" json:"primary_element_id"`
	Level              int                          `gorm:"default:1;comment:เลเวลปัจจุบัน" json:"level"`
	Exp                int                          `gorm:"default:0;comment:ค่าประสบการณ์ปัจจุบัน" json:"exp"`
	CurrentHP          int                          `gorm:"comment:พลังชีวิตปัจจุบัน" json:"current_hp"`
	CurrentMP          int                          `gorm:"comment:พลังเวทปัจจุบัน" json:"current_mp"`
	TalentS            int                          `gorm:"comment:แต้มพรสวรรค์ S" json:"talent_s"`
	TalentL            int                          `gorm:"comment:แต้มพรสวรรค์ L" json:"talent_l"`
	TalentG            int                          `gorm:"comment:แต้มพรสวรรค์ G" json:"talent_g"`
	TalentP            int                          `gorm:"comment:แต้มพรสวรรค์ P" json:"talent_p"`
	StatsUpdatedAt     time.Time                    `gorm:"autoCreateTime;comment:เวลาที่อัปเดตค่าพลังล่าสุด" json:"statsUpdatedAt"`
	TutorialStep       int                          `gorm:"default:1;comment:ขั้นตอนของบทช่วยสอนที่ทำถึง" json:"tutorial_step"`
	PrimaryElement     *Element                     `gorm:"foreignKey:PrimaryElementID;references:ID"`
	Masteries          []*CharacterMastery          `gorm:"foreignKey:CharacterID;constraint:OnDelete:CASCADE;" json:"masteries"`
	DimensionalSeal    []*DimensionalSealInventory  `gorm:"foreignKey:CharacterID;constraint:OnDelete:CASCADE;" json:"dimensionalSeal"`
	JournalDiscoveries []*CharacterJournalDiscovery `gorm:"foreignKey:CharacterID;constraint:OnDelete:CASCADE;" json:"journalDiscoveries"`
}
