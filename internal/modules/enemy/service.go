package enemy

import "sage-of-elements-backend/internal/domain"

type EnemyService interface {
	GetAllEnemies() ([]domain.Enemy, error)
}

type enemyService struct {
	enemyRepo EnemyRepository
}

func NewEnemyService(enemyRepo EnemyRepository) EnemyService {
	return &enemyService{enemyRepo: enemyRepo}
}

func (s *enemyService) GetAllEnemies() ([]domain.Enemy, error) {
	return s.enemyRepo.FindAll()
}
