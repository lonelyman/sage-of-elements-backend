package postgres

import (
	"sage-of-elements-backend/internal/domain"
	"sage-of-elements-backend/internal/modules/game_data"

	"gorm.io/gorm"
)

// gameDataRepository คือ struct ที่ implement game_data.Repository interface
type gameDataRepository struct {
	db *gorm.DB
}

// NewGameDataRepository คือฟังก์ชันสำหรับสร้าง Repository
func NewGameDataRepository(db *gorm.DB) game_data.GameDataRepository {
	return &gameDataRepository{
		db: db,
	}
}

// FindAllElements ดึงข้อมูลธาตุทั้งหมด
func (r *gameDataRepository) FindAllElements() ([]domain.Element, error) {
	var elements []domain.Element
	if err := r.db.Find(&elements).Error; err != nil {
		return nil, err
	}
	return elements, nil
}

// FindAllMasteries ดึงข้อมูลศาสตร์ทั้งหมด
func (r *gameDataRepository) FindAllMasteries() ([]domain.Mastery, error) {
	var masteries []domain.Mastery
	if err := r.db.Find(&masteries).Error; err != nil {
		return nil, err
	}
	return masteries, nil
}

// FindAllRecipes ดึงข้อมูลสูตรผสมทั้งหมด (พร้อมส่วนประกอบ)
func (r *gameDataRepository) FindAllRecipes() ([]domain.Recipe, error) {
	var recipes []domain.Recipe
	// เราต้อง Preload "Ingredients.InputElement" เพื่อให้ข้อมูลส่วนประกอบมาครบ
	err := r.db.Preload("Ingredients.InputElement").Find(&recipes).Error
	if err != nil {
		return nil, err
	}
	return recipes, nil
}

// FindAllSpells ดึงข้อมูลเวทมนตร์ทั้งหมด (พร้อมเอฟเฟกต์)
func (r *gameDataRepository) FindAllSpells() ([]domain.Spell, error) {
	var spells []domain.Spell
	// เราต้อง Preload "Effects.Effect" เพื่อให้ข้อมูลเอฟเฟกต์มาครบ
	err := r.db.Preload("Effects.Effect").Find(&spells).Error
	if err != nil {
		return nil, err
	}
	return spells, nil
}

// GetGameConfigValue ดึงค่า Config 1 ตัวจากตาราง game_configs ด้วย key
func (r *gameDataRepository) GetGameConfigValue(key string) (string, error) {
	var config domain.GameConfig
	// ใช้ .Where("key = ?", key) เพื่อค้นหาแถวที่มี key ตรงกัน
	// และใช้ .First(&config) เพื่อดึงข้อมูลแค่แถวเดียว
	if err := r.db.Where("key = ?", key).First(&config).Error; err != nil {
		// ถ้า GORM คืนค่า ErrRecordNotFound แปลว่าไม่เจอ key นั้น
		if err == gorm.ErrRecordNotFound {
			// เราจะคืนค่า string ว่างๆ กลับไป และไม่ถือว่าเป็น Error
			return "", nil
		}
		// แต่ถ้าเป็น Error อื่นๆ (เช่น DB down) ก็ต้องส่ง Error กลับไป
		return "", err
	}
	// ถ้าเจอ ก็คืนค่า Value กลับไป
	return config.Value, nil
}
