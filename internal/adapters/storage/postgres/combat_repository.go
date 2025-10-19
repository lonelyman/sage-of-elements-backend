// file: internal/adapters/storage/postgres/combat_repository.go
package postgres

import (
	"fmt"
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
		fmt.Printf("err%v", err)
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

// FindMatchByID ค้นหา Match ด้วย ID พร้อม Preload ข้อมูลที่จำเป็นทั้งหมด
func (r *combatRepository) FindMatchByID(matchID string) (*domain.CombatMatch, error) {
	var match domain.CombatMatch
	err := r.db.
		Preload("Combatants.Character.PrimaryElement"). // Preload ให้ลึกขึ้นเผื่อใช้
		Preload("Combatants.Enemy.Element").
		Preload("Combatants.Enemy.Abilities").       // <-- ⭐️ สั่งให้โหลดท่าโจมตีของศัตรูมาด้วย!
		Preload("Combatants.Enemy.AI.AbilityToUse"). // <-- ⭐️ สั่งให้โหลดกฎ AI และท่าที่ผูกกับกฎนั้นมาด้วย!
		Preload("Combatants.Deck").
		Where("id = ?", matchID).
		First(&match).Error
	return &match, err
}

// UpdateMatch บันทึกสถานะล่าสุดของ Match และ Combatant ทุกตัว
func (r *combatRepository) UpdateMatch(match *domain.CombatMatch) (*domain.CombatMatch, error) {
	err := r.db.Transaction(func(tx *gorm.DB) error {
		// 1. บันทึก Combatant ทุกตัวก่อน
		for i := range match.Combatants {
			if err := tx.Save(match.Combatants[i]).Error; err != nil {
				return err
			}
		}

		// 2. บันทึก Match หลักทีหลัง
		if err := tx.Save(match).Error; err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	// ดึงข้อมูลล่าสุดทั้งหมดกลับมาอีกครั้ง (เหมือนเดิม)
	return r.FindMatchByID(match.ID.String())
}
