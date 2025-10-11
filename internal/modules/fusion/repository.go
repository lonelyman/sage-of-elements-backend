package fusion

import (
	"sage-of-elements-backend/internal/domain"

	"gorm.io/gorm"
)

// Repository คือ "สัญญา" สำหรับการจัดการข้อมูลที่เกี่ยวกับการหลอมรวม
type FusionRepository interface {
	// ค้นหาสูตรจากส่วนประกอบที่กำหนด (ต้องทำใน Transaction)
	FindRecipeByIngredients(tx *gorm.DB, ingredients map[uint]int) (*domain.Recipe, error)

	// ตรวจสอบว่าเคยค้นพบสูตรนี้แล้วหรือยัง
	IsRecipeDiscovered(characterID uint, recipeID uint) (bool, error)

	// บันทึกการค้นพบใหม่ (ต้องทำใน Transaction)
	LogDiscovery(tx *gorm.DB, characterID uint, recipeID uint) error
}
