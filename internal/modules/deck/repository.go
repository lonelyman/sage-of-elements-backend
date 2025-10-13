// file: internal/modules/deck/repository.go
package deck

import (
	"sage-of-elements-backend/internal/domain"
)

// DeckRepository defines the contract for deck database operations.
type DeckRepository interface {
	// สร้าง Deck ใหม่ (พร้อม Slot ว่างๆ)
	Create(deck *domain.Deck) (*domain.Deck, error)

	// ค้นหา Deck ทั้งหมดของตัวละคร
	FindByCharacterID(characterID uint) ([]domain.Deck, error)

	// ค้นหา Deck ด้วย ID
	FindByID(deckID uint) (*domain.Deck, error)

	// อัปเดต Deck ( Logic จะซับซ้อนหน่อย เพราะต้องลบ Slot เก่า แล้วสร้าง Slot ใหม่)
	Update(deckID uint, name string, slots []*domain.DeckSlot) (*domain.Deck, error)

	// นับจำนวน Deck ของตัวละคร
	CountByCharacterID(characterID uint) (int64, error)
}
