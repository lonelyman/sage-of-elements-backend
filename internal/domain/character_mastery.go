package domain

// CharacterMastery คือตารางที่เก็บว่าผู้เล่นมี Level/MXP เท่าไหร่ในแต่ละศาสตร์
type CharacterMastery struct {
	CharacterID uint     `gorm:"primaryKey;autoIncrement:false;comment:ID ของตัวละคร (Composite PK)" json:"-"`
	MasteryID   uint     `gorm:"primaryKey;autoIncrement:false;comment:ID ของศาสตร์ (Composite PK)" json:"mastery_id"`
	Level       int      `gorm:"default:1;comment:ระดับความชำนาญ" json:"level"`
	Mxp         int      `gorm:"default:0;comment:ค่าประสบการณ์ความชำนาญ" json:"mxp"`
	Mastery     *Mastery `gorm:"foreignKey:MasteryID" json:"mastery"`
}
