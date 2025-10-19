// file: internal/modules/deck/service.go
package deck

import (
	"fmt"
	"sage-of-elements-backend/internal/domain"
	"sage-of-elements-backend/internal/modules/character"
	"sage-of-elements-backend/pkg/apperrors"
	"sage-of-elements-backend/pkg/applogger"
)

// --- Interface (สัญญา) ---

type DeckService interface {
	CreateDeck(playerID, characterID uint, name string) (*domain.Deck, error)
	GetDecksByCharacterID(playerID, characterID uint) ([]domain.Deck, error)
	UpdateDeck(playerID, deckID uint, req UpdateDeckRequest) (*domain.Deck, error)
	DeleteDeck(playerID, deckID uint) error
}

// --- Implementation (การทำงานจริง) ---

type deckService struct {
	appLogger     applogger.Logger
	deckRepo      DeckRepository
	characterRepo character.CharacterRepository // ⭐️ ต้องใช้ CharacterRepo เพื่อตรวจสอบความเป็นเจ้าของ!
}

// NewDeckService creates a new instance of deckService.
func NewDeckService(
	appLogger applogger.Logger,
	deckRepo DeckRepository,
	characterRepo character.CharacterRepository,
) DeckService {
	return &deckService{
		appLogger:     appLogger,
		deckRepo:      deckRepo,
		characterRepo: characterRepo,
	}
}

func (s *deckService) CreateDeck(playerID, characterID uint, name string) (*domain.Deck, error) {
	// 1. ตรวจสอบความเป็นเจ้าของ: เช็คว่าตัวละครนี้เป็นของผู้เล่นคนนี้จริงหรือไม่
	char, err := s.characterRepo.FindByID(characterID)
	if err != nil {
		return nil, apperrors.SystemError("error checking character ownership")
	}
	if char == nil || char.PlayerID != playerID {
		return nil, apperrors.PermissionDeniedError("you do not own this character")
	}

	// 2. ตรวจสอบเงื่อนไขทางธุรกิจ: สร้าง Deck เกินจำนวนสูงสุดหรือยัง?
	count, err := s.deckRepo.CountByCharacterID(characterID)
	if err != nil {
		return nil, apperrors.SystemError("error counting decks")
	}
	if count >= 8 { // สมมติว่าสร้างได้สูงสุด 8 Decks
		return nil, apperrors.New(422, "MAX_DECKS_REACHED", "Maximum number of decks reached")
	}

	// 3. เตรียมข้อมูล Deck ใหม่
	newDeck := &domain.Deck{
		CharacterID:  characterID,
		Name:         name,
		IsActive:     false,      // Deck ใหม่จะไม่เป็น Deck หลักเสมอ
		DisplayOrder: int(count), // กำหนดลำดับการแสดงผลเป็นลำดับสุดท้าย
	}

	// 4. สั่งให้ Repository บันทึก
	return s.deckRepo.Create(newDeck)
}

func (s *deckService) GetDecksByCharacterID(playerID, characterID uint) ([]domain.Deck, error) {
	// 1. ตรวจสอบความเป็นเจ้าของ
	char, err := s.characterRepo.FindByID(characterID)
	if err != nil {
		return nil, apperrors.SystemError("error checking character ownership")
	}
	if char == nil || char.PlayerID != playerID {
		return nil, apperrors.PermissionDeniedError("you do not own this character")
	}

	// 2. สั่งให้ Repository ไปดึงข้อมูลมา
	return s.deckRepo.FindByCharacterID(characterID)
}

func (s *deckService) UpdateDeck(playerID, deckID uint, req UpdateDeckRequest) (*domain.Deck, error) {
	// 1. ตรวจสอบความเป็นเจ้าของ Deck (เหมือนเดิม)
	deck, err := s.deckRepo.FindByID(deckID)
	if err != nil || deck == nil {
		return nil, apperrors.NotFoundError("deck not found")
	}
	char, _ := s.characterRepo.FindByID(deck.CharacterID)
	if char == nil || char.PlayerID != playerID {
		return nil, apperrors.PermissionDeniedError("you do not have permission to edit this deck")
	}

	// --- ✨⭐️ เพิ่ม "ด่านตรวจสอบ" ใหม่ตรงนี้! ⭐️✨ ---
	// 2. ตรวจสอบว่า Client ส่ง Slot ซ้ำซ้อนมาหรือไม่
	slotTracker := make(map[int]bool) // สร้าง "สมุดจด"
	for _, s := range req.Slots {
		if s.SlotNum < 1 || s.SlotNum > 8 {
			return nil, apperrors.New(400, "INVALID_SLOT_NUMBER", fmt.Sprintf("Slot number %d is out of range (must be 1-8)", s.SlotNum))
		}

		if slotTracker[s.SlotNum] {
			// ถ้า "สมุดจด" บอกว่าเคยเจอช่องนี้แล้ว...
			return nil, apperrors.New(400, "DUPLICATE_SLOT", fmt.Sprintf("Slot number %d is duplicated", s.SlotNum))
		}
		slotTracker[s.SlotNum] = true // "จด" ว่าเจอช่องนี้แล้ว
	}
	// ---------------------------------------------

	// 3. TODO (อนาคต): ตรวจสอบว่าธาตุ T1 ที่ส่งมา... ผู้เล่น "ค้นพบ" แล้วจริงๆ

	// 4. แปลงข้อมูล (เหมือนเดิม)
	var newSlots []*domain.DeckSlot
	for _, s := range req.Slots {
		newSlots = append(newSlots, &domain.DeckSlot{
			SlotNum:   s.SlotNum,
			ElementID: s.ElementID,
		})
	}

	// 5. สั่งให้ Repository ทำการอัปเดต!
	return s.deckRepo.Update(deckID, req.Name, newSlots)
}

func (s *deckService) DeleteDeck(playerID, deckID uint) error {
	// 1. ตรวจสอบความเป็นเจ้าของ Deck
	deck, err := s.deckRepo.FindByID(deckID)
	if err != nil || deck == nil {
		return apperrors.NotFoundError("deck not found")
	}
	// เช็คผ่าน Character ที่ผูกกับ Deck
	char, _ := s.characterRepo.FindByID(deck.CharacterID)
	if char == nil || char.PlayerID != playerID {
		return apperrors.PermissionDeniedError("you do not have permission to delete this deck")
	}

	// 2. สั่งให้ Repository ลบ
	return s.deckRepo.Delete(deckID)
}
