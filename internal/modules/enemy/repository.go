// file: internal/modules/enemy/repository.go
package enemy

import "sage-of-elements-backend/internal/domain"

type EnemyRepository interface {
	FindByID(id uint) (*domain.Enemy, error)
	FindAll() ([]domain.Enemy, error) // <--- เพิ่มฟังก์ชันนี้เผื่อสำหรับ API GET /enemies
}
