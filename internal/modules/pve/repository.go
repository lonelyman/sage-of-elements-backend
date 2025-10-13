// file: internal/modules/pve/repository.go
package pve

import "sage-of-elements-backend/internal/domain"

type PveRepository interface {
	FindAllActiveRealms() ([]domain.Realm, error)
}
