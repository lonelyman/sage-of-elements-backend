// file: internal/modules/combat/repository.go
package combat

import "sage-of-elements-backend/internal/domain"

type CombatRepository interface {
	CreateMatch(match *domain.CombatMatch) (*domain.CombatMatch, error)
	FindMatchByID(matchID string) (*domain.CombatMatch, error)
	UpdateMatch(match *domain.CombatMatch) (*domain.CombatMatch, error)

	// 🧹 Cleanup Methods - สำหรับจัดการ match ค้าง
	FindStaleMatches(inactiveMinutes int) ([]*domain.CombatMatch, error)       // หา match ที่ไม่มีความเคลื่อนไหวนานเกินกำหนด
	AbortStaleMatches(inactiveMinutes int) (int64, error)                      // Abort match ค้างทั้งหมด (return จำนวนที่ abort)
	FindPlayerActiveMatch(characterID uint) (*domain.CombatMatch, error)       // หา match ที่ผู้เล่นกำลังเล่นอยู่
	AbortMatchByID(matchID string, reason string) (*domain.CombatMatch, error) // Abort match เฉพาะ ID
}
