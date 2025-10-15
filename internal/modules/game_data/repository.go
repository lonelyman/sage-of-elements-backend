package game_data

import "sage-of-elements-backend/internal/domain"

// Repository คือ "สัญญา" สำหรับการดึงข้อมูลแกนกลางทั้งหมด
type GameDataRepository interface {
	FindAllElements() ([]domain.Element, error)
	FindAllMasteries() ([]domain.Mastery, error)
	FindAllRecipes() ([]domain.Recipe, error)
	FindAllSpells() ([]domain.Spell, error)

	GetGameConfigValue(key string) (string, error)
	FindAllGameConfigs() ([]domain.GameConfig, error)
	FindSpellByID(id uint) (*domain.Spell, error)
}
