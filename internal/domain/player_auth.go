package domain

// PlayerAuth เก็บข้อมูลการยืนยันตัวตนของผู้เล่นแต่ละคน
type PlayerAuth struct {
	ID           uint    `gorm:"primaryKey;comment:ID เฉพาะของวิธีการยืนยันตัวตน (PK)"`
	PlayerID     uint    `gorm:"not null;comment:ID ของผู้เล่นที่เป็นเจ้าของ (FK to players)"`
	Provider     string  `gorm:"size:50;not null;comment:ประเภทของผู้ให้บริการ (local, google, apple)"`
	ProviderID   string  `gorm:"size:255;unique;comment:ID ที่ได้จากผู้ให้บริการ (เช่น Google ID)"`
	Secret       string  `gorm:"comment:เก็บ Hashed Password สำหรับ Provider 'local' เท่านั้น"`
	RefreshToken *string `gorm:"size:512;comment:เก็บ Refresh Token ล่าสุดของผู้ใช้"`
}
