package postgres

import (
	"sage-of-elements-backend/internal/domain"
	"sage-of-elements-backend/internal/modules/fusion"

	"gorm.io/gorm"
)

type fusionRepository struct {
	db *gorm.DB
}

// NewFusionRepository คือฟังก์ชันสำหรับสร้าง Repository
func NewFusionRepository(db *gorm.DB) fusion.FusionRepository {
	return &fusionRepository{
		db: db,
	}
}

// FindRecipeByIngredients คือ Logic ที่ซับซ้อนที่สุดของเรา!
// ทำหน้าที่ค้นหาสูตรที่ตรงกับส่วนประกอบที่ส่งมาเป๊ะๆ
// FindRecipeByIngredients คือ Logic ที่ถูก "ยกเครื่อง" ใหม่ทั้งหมด
func (r *fusionRepository) FindRecipeByIngredients(tx *gorm.DB, ingredients map[uint]int) (*domain.Recipe, error) {
	var recipeID uint

	// 1. เริ่มสร้าง Query บนตาราง recipe_ingredients
	query := tx.Model(&domain.RecipeIngredient{}).
		Group("recipe_id").
		// 2. กฎข้อแรก: ต้องมี "จำนวน" ส่วนประกอบเท่ากับที่เราส่งมาเป๊ะๆ
		Having("COUNT(*) = ?", len(ingredients))

	// 3. กฎข้อที่สอง: ต้องมี "ส่วนประกอบแต่ละชิ้น" ครบตามที่เราส่งมา
	for elementID, quantity := range ingredients {
		query = query.Having(
			"SUM(CASE WHEN input_element_id = ? AND quantity = ? THEN 1 ELSE 0 END) = 1",
			elementID,
			quantity,
		)
	}

	// 4. ดึง recipe_id ที่ผ่านเงื่อนไขทั้งหมดออกมา
	err := query.Pluck("recipe_id", &recipeID).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil // ไม่เจอสูตร, ถือว่าปกติ
		}
		return nil, err
	}

	// 5. ถ้าเจอ ID... ก็ไปดึงข้อมูล Recipe ทั้งหมดพร้อม Preload
	var recipe domain.Recipe
	err = tx.
		Preload("OutputElement").
		Preload("Ingredients.InputElement").
		First(&recipe, recipeID).Error

	if err != nil {
		return nil, err
	}

	return &recipe, nil
}

// IsRecipeDiscovered ตรวจสอบว่าเคยค้นพบสูตรนี้แล้วหรือยัง
func (r *fusionRepository) IsRecipeDiscovered(characterID uint, recipeID uint) (bool, error) {
	var count int64
	err := r.db.Model(&domain.CharacterJournalDiscovery{}).
		Where("character_id = ? AND recipe_id = ?", characterID, recipeID).
		Count(&count).Error

	return count > 0, err
}

// LogDiscovery บันทึกการค้นพบใหม่
func (r *fusionRepository) LogDiscovery(tx *gorm.DB, characterID uint, recipeID uint) error {
	discovery := &domain.CharacterJournalDiscovery{
		CharacterID: characterID,
		RecipeID:    recipeID,
	}
	// ใช้ FirstOrCreate เพื่อป้องกันการสร้างข้อมูลซ้ำซ้อน (เผื่อไว้)
	return tx.FirstOrCreate(discovery).Error
}
