// file: internal/adapters/storage/postgres/combat_repository.go
package postgres

import (
	"sage-of-elements-backend/internal/domain"
	"sage-of-elements-backend/internal/modules/combat"

	"gorm.io/gorm"
)

type combatRepository struct {
	db *gorm.DB
}

// NewCombatRepository creates a new instance of combatRepository.
func NewCombatRepository(db *gorm.DB) combat.CombatRepository {
	return &combatRepository{db: db}
}

// CreateMatch บันทึกข้อมูล Match และ Combatant ทั้งหมดลง DB ใน Transaction เดียว
func (r *combatRepository) CreateMatch(match *domain.CombatMatch) (*domain.CombatMatch, error) {
	// GORM จะฉลาดพอที่จะสร้าง Match และ Combatant ที่อยู่ใน Slice พร้อมกัน
	if err := r.db.Create(match).Error; err != nil {
		return nil, err
	}

	// หลังจากสร้างเสร็จ เราต้อง Preload ข้อมูลกลับมาด้วย
	// เพื่อให้ Response ที่ส่งกลับไปมีข้อมูลของ Character และ Enemy ครบถ้วน
	// เราต้อง Preload แบบซ้อนกันลึกๆ (Nested Preloading)
	if err := r.db.
		Preload("Combatants.Character.PrimaryElement").
		Preload("Combatants.Enemy.Element").
		First(match, match.ID).Error; err != nil {
		return nil, err
	}

	return match, nil
}
