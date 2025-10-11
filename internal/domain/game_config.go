package domain

// GameConfig คือตารางที่เก็บ "กฎ" และ "ค่าคงที่" ทั้งหมดของเกม
type GameConfig struct {
	Key         string `gorm:"primaryKey;size:100"`
	Value       string `gorm:"not null"`
	Description string `gorm:"comment:คำอธิบายของค่าคงที่นี้ (เพื่อให้นักพัฒนาคนอื่นเข้าใจ)"`
}

// TableName คือฟังก์ชันที่บอก GORM ว่าให้สร้างตารางชื่อ "game_configs"
func (GameConfig) TableName() string {
	return "game_configs"
}
