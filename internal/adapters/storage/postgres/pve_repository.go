// file: internal/adapters/storage/postgres/pve_repository.go
package postgres

import (
	"sage-of-elements-backend/internal/domain"
	"sage-of-elements-backend/internal/modules/pve"

	"gorm.io/gorm"
)

type pveRepository struct {
	db *gorm.DB
}

func NewPveRepository(db *gorm.DB) pve.PveRepository {
	return &pveRepository{db: db}
}

// ⭐️ แก้ไข Logic การดึงข้อมูลให้ซับซ้อนและถูกต้องขึ้น!
func (r *pveRepository) FindAllActiveRealms() ([]domain.Realm, error) {
	var realms []domain.Realm
	// เราจะดึงข้อมูลแบบซ้อนกัน (Nested Preload) และเรียงลำดับให้สวยงาม
	err := r.db.
		Order("id asc"). // เรียง Realm ตาม ID
		Where("is_active = ?", true).
		// Preload "Chapters" ที่อยู่ใน Realm และสั่งให้เรียง Chapter ตามลำดับ
		Preload("Chapters", func(db *gorm.DB) *gorm.DB {
			return db.Order("chapters.chapter_number asc")
		}).
		// Preload "Stages" ที่อยู่ใน "Chapters" และสั่งให้เรียง Stage ตามลำดับ
		Preload("Chapters.Stages", func(db *gorm.DB) *gorm.DB {
			return db.Order("stages.stage_number asc")
		}).
		Find(&realms).Error
	return realms, err
}
