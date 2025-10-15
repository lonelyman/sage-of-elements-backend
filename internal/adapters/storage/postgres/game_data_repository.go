package postgres

import (
	"log"
	"sage-of-elements-backend/internal/adapters/cache/redis"
	"sage-of-elements-backend/internal/domain"
	"sage-of-elements-backend/internal/modules/game_data"

	"gorm.io/gorm"
)

// gameDataRepository คือ struct ที่ implement game_data.Repository interface
type gameDataRepository struct {
	db          *gorm.DB
	configCache *redis.GameConfigCacheRepository
}

// NewGameDataRepository คือฟังก์ชันสำหรับสร้าง Repository
func NewGameDataRepository(db *gorm.DB, configCache *redis.GameConfigCacheRepository) game_data.GameDataRepository {
	return &gameDataRepository{
		db:          db,
		configCache: configCache,
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

// GetGameConfigValue จะถาม Cache ก่อนเสมอ!
func (r *gameDataRepository) GetGameConfigValue(key string) (string, error) {
	// 1. ถาม Cache ที่ถูกต้อง!
	cachedValue, err := r.configCache.GetConfig(key)
	if err != nil {
		log.Printf("WARN: Redis error on GetConfig: %v. Falling back to DB.", err)
		return r.getConfigFromDB(key)
	}
	if cachedValue != "" {
		return cachedValue, nil // Cache Hit!
	}

	// 2. ถ้า Cache Miss...
	log.Printf("WARN: Cache miss for key '%s'. Falling back to DB.", key)
	return r.getConfigFromDB(key)
}

// getConfigFromDB คือฟังก์ชันลูกที่ใช้ถาม Database โดยตรง (โค้ดเดิมของน้องชาย)
func (r *gameDataRepository) getConfigFromDB(key string) (string, error) {
	var config domain.GameConfig
	if err := r.db.Where("key = ?", key).First(&config).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return "", nil
		}
		return "", err
	}
	return config.Value, nil
}
func (r *gameDataRepository) FindAllGameConfigs() ([]domain.GameConfig, error) {
	var configs []domain.GameConfig
	err := r.db.Find(&configs).Error
	return configs, err
}

func (r *gameDataRepository) FindSpellByID(id uint) (*domain.Spell, error) {
	var spell domain.Spell
	// Preload "Effects.Effect" เพื่อให้เราดึงข้อมูล Effect ที่ซ้อนอยู่ข้างในมาด้วย
	err := r.db.Preload("Effects.Effect").First(&spell, id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &spell, nil
}
