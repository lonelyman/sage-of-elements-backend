// file: internal/adapters/storage/postgres/deck_repository.go
package postgres

import (
	"sage-of-elements-backend/internal/domain"
	"sage-of-elements-backend/internal/modules/deck"

	"gorm.io/gorm"
)

type deckRepository struct {
	db *gorm.DB
}

// NewDeckRepository creates a new instance of deckRepository.
func NewDeckRepository(db *gorm.DB) deck.DeckRepository {
	return &deckRepository{
		db: db,
	}
}

func (r *deckRepository) Create(d *domain.Deck) (*domain.Deck, error) {
	if err := r.db.Create(d).Error; err != nil {
		return nil, err
	}
	return d, nil
}

func (r *deckRepository) FindByCharacterID(characterID uint) ([]domain.Deck, error) {
	var decks []domain.Deck
	// ใช้ .Preload("Slots") เพื่อให้ GORM ดึงข้อมูลในตาราง deck_slots ที่ผูกกันมาให้ด้วย
	// และ .Order("display_order asc") เพื่อจัดเรียงตามที่ผู้เล่นตั้งค่าไว้
	err := r.db.Preload("Slots").Order("display_order asc").Where("character_id = ?", characterID).Find(&decks).Error
	return decks, err
}

func (r *deckRepository) FindByID(deckID uint) (*domain.Deck, error) {
	var d domain.Deck
	// Preload "Slots" มาด้วยเสมอ เพื่อให้ได้ข้อมูลที่ครบถ้วน
	err := r.db.Preload("Slots").First(&d, deckID).Error
	return &d, err
}

func (r *deckRepository) CountByCharacterID(characterID uint) (int64, error) {
	var count int64
	err := r.db.Model(&domain.Deck{}).Where("character_id = ?", characterID).Count(&count).Error
	return count, err
}

func (r *deckRepository) Update(deckID uint, name string, slots []*domain.DeckSlot) (*domain.Deck, error) {
	// เราจะใช้ Transaction ที่นี่ เพื่อให้แน่ใจว่าการอัปเดตทั้งหมดสำเร็จหรือล้มเหลวไปพร้อมกัน
	err := r.db.Transaction(func(tx *gorm.DB) error {
		// 1. อัปเดตข้อมูลหลักของ Deck (เช่น ชื่อ)
		if err := tx.Model(&domain.Deck{}).Where("id = ?", deckID).Update("name", name).Error; err != nil {
			return err
		}

		// 2. ลบ Slots เก่าทั้งหมดที่ผูกกับ Deck นี้ทิ้งไป
		if err := tx.Where("deck_id = ?", deckID).Delete(&domain.DeckSlot{}).Error; err != nil {
			return err
		}

		// 3. ถ้ามี Slots ใหม่ส่งมา... ก็สร้างมันขึ้นมาใหม่ทั้งหมด
		if len(slots) > 0 {
			// กำหนด DeckID ให้กับทุก Slot ใหม่
			for _, slot := range slots {
				slot.DeckID = deckID
			}
			if err := tx.Create(&slots).Error; err != nil {
				return err
			}
		}

		return nil // Commit transaction
	})

	if err != nil {
		return nil, err
	}

	// ดึงข้อมูล Deck ที่อัปเดตแล้วทั้งหมดกลับไปให้ Service
	return r.FindByID(deckID)
}

func (r *deckRepository) Delete(deckID uint) error {
	// GORM จะจัดการลบ Deck และ Slots ที่ผูกกัน (constraint:OnDelete:CASCADE) ให้โดยอัตโนมัติ
	return r.db.Delete(&domain.Deck{}, deckID).Error
}
