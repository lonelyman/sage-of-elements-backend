package pve

import "sage-of-elements-backend/internal/domain"

type PveService interface {
	GetAllActiveRealms() ([]domain.Realm, error)
}

type pveService struct {
	pveRepo PveRepository
}

func NewPveService(pveRepo PveRepository) PveService {
	return &pveService{pveRepo: pveRepo}
}

func (s *pveService) GetAllActiveRealms() ([]domain.Realm, error) {
	return s.pveRepo.FindAllActiveRealms()
}
