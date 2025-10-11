package character

import (
	"sage-of-elements-backend/internal/domain"
	"sage-of-elements-backend/internal/modules/game_data"

	"gorm.io/gorm"
)

type CharacterRepository interface {
	// ฟังก์ชันที่มีอยู่แล้ว
	CheckCharacterExists(name string) (*domain.Character, error)
	Create(character *domain.Character) (*domain.Character, error)
	Save(character *domain.Character) (*domain.Character, error)
	FindAllByPlayerID(playerID uint) ([]domain.Character, error)
	FindByID(id uint) (*domain.Character, error)
	Delete(characterID uint) error
	FindInventoryByCharacterID(characterID uint) ([]*domain.DimensionalSealInventory, error)
	UpdateCharacterInTx(tx *gorm.DB, character *domain.Character) error
	ConsumeAndUpdateInventoryInTx(tx *gorm.DB, characterID uint, itemsToConsume map[uint]int, itemsToAdd map[uint]int) error

	// --- ⭐️ เพิ่มแค่ฟังก์ชันนี้เข้ามา! ⭐️ ---
	// ฟังก์ชันใหม่สำหรับคำนวณและบันทึกค่าพลังที่ฟื้นฟู
	RegenerateStats(character *domain.Character, gameDataRepo game_data.GameDataRepository) (*domain.Character, error)
}
