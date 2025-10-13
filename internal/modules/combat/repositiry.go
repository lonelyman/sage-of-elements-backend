// file: internal/modules/combat/repository.go
package combat

import "sage-of-elements-backend/internal/domain"

type CombatRepository interface {
	CreateMatch(match *domain.CombatMatch) (*domain.CombatMatch, error)
}
