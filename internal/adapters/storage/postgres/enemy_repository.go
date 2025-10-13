// file: internal/adapters/storage/postgres/enemy_repository.go
package postgres

import (
	"sage-of-elements-backend/internal/domain"
	"sage-of-elements-backend/internal/modules/enemy"

	"gorm.io/gorm"
)

type enemyRepository struct {
	db *gorm.DB
}

func NewEnemyRepository(db *gorm.DB) enemy.EnemyRepository {
	return &enemyRepository{db: db}
}

func (r *enemyRepository) FindByID(id uint) (*domain.Enemy, error) {
	var e domain.Enemy
	// Preload ทุกอย่างที่เกี่ยวกับศัตรูมาให้หมด! (Abilities, AI)
	err := r.db.Preload("Element").Preload("Abilities").Preload("AI").First(&e, id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &e, nil
}

func (r *enemyRepository) FindAll() ([]domain.Enemy, error) {
	var enemies []domain.Enemy
	err := r.db.Preload("Element").Find(&enemies).Error
	return enemies, err
}
