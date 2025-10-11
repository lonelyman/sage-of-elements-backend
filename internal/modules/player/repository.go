package player

import "sage-of-elements-backend/internal/domain"

// PlayerRepository คือ "สัญญา" สำหรับการติดต่อกับฐานข้อมูลที่เกี่ยวกับ Player
type PlayerRepository interface {
	FindByID(id uint) (*domain.Player, error)
	FindByUsername(username string) (*domain.Player, error)
	FindByEmail(email string) (*domain.Player, error)
	FindAuthByPlayerIDAndProvider(playerID uint, provider string) (*domain.PlayerAuth, error)
	CreateWithAuth(player *domain.Player, auth *domain.PlayerAuth) (*domain.Player, error)
	Save(player *domain.Player) (*domain.Player, error)
	Update(player *domain.Player) (*domain.Player, error)
	UpdateAuth(auth *domain.PlayerAuth) (*domain.PlayerAuth, error)
	FindAuthByRefreshToken(token string) (*domain.PlayerAuth, error)
}
