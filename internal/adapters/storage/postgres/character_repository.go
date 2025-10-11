package postgres

import (
	"fmt"
	"sage-of-elements-backend/internal/domain"
	"sage-of-elements-backend/internal/modules/character" // Import interface จาก module
	"sage-of-elements-backend/internal/modules/game_data"
	"strconv"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// characterRepository คือ struct ที่ implement CharacterRepository interface
type characterRepository struct {
	db *gorm.DB
}

// NewCharacterRepository คือฟังก์ชันสำหรับสร้าง Repository ขึ้นมาใช้งาน
func NewCharacterRepository(db *gorm.DB) character.CharacterRepository {
	return &characterRepository{
		db: db,
	}
}

// CheckCharacterExists ตรวจสอบว่ามีชื่อตัวละครนี้อยู่แล้วหรือไม่
func (r *characterRepository) CheckCharacterExists(name string) (*domain.Character, error) {
	var char domain.Character
	result := r.db.Where("character_name = ?", name).First(&char)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil // ไม่พบชื่อซ้ำ, ถือว่าปกติ
		}
		return nil, result.Error // Error อื่นๆ
	}

	return &char, nil
}

// Create บันทึกตัวละครใหม่ (พร้อมข้อมูล Mastery) ลงในฐานข้อมูล
func (r *characterRepository) Create(character *domain.Character) (*domain.Character, error) {
	// GORM ฉลาดพอที่จะรู้ว่าเมื่อเรา Create Character ที่มี Field Masteries (ที่เป็น Slice) อยู่
	// มันจะทำการสร้างข้อมูลในตาราง character_mastery ให้โดยอัตโนมัติ!
	result := r.db.Create(character)
	if result.Error != nil {
		return nil, result.Error
	}

	return character, nil
}

// Save คือฟังก์ชัน Save ตัวจริง! ใช้ได้ทั้งสร้างและอัปเดต
// GORM จะเช็คเองว่าถ้า ID มีค่า ก็จะ UPDATE, ถ้า ID เป็นค่าว่าง ก็จะ INSERT
func (r *characterRepository) Save(character *domain.Character) (*domain.Character, error) {
	if err := r.db.Save(character).Error; err != nil {
		return nil, err
	}
	return character, nil
}

// FindAllByPlayerID ค้นหาตัวละครทั้งหมดที่เป็นของ Player ID ที่กำหนด
func (r *characterRepository) FindAllByPlayerID(playerID uint) ([]domain.Character, error) {
	var characters []domain.Character

	// ใช้ .Where เพื่อกรองหา "player_id" ที่ตรงกัน
	// และใช้ .Find เพื่อดึงข้อมูลทั้งหมดที่เจอ (แทน .First ที่ดึงแค่ตัวเดียว)
	// พี่เพิ่ม .Preload("PrimaryElement") และ .Preload("Masteries.Mastery") เข้าไปด้วย...
	// เพื่อให้ GORM ดึงข้อมูลที่เกี่ยวข้อง (ธาตุหลัก, ชื่อศาสตร์) มาให้เราใน Query เดียวเลย!
	result := r.db.
		Preload("PrimaryElement").
		Preload("Masteries.Mastery").
		Where("player_id = ?", playerID).
		Find(&characters)

	if result.Error != nil {
		// ถ้าเกิด Error ใดๆ ก็ตาม ให้ return error กลับไป
		return nil, result.Error
	}

	return characters, nil
}

// FindByID ค้นหาตัวละครด้วย ID เพียงตัวเดียว
func (r *characterRepository) FindByID(id uint) (*domain.Character, error) {
	var character domain.Character

	// ใช้ .First เพื่อดึงข้อมูลตัวเดียวด้วย Primary Key
	// และ Preload ข้อมูลที่เกี่ยวข้องมาให้ครบถ้วนเหมือนเดิม
	result := r.db.
		Preload("PrimaryElement").
		Preload("Masteries.Mastery").
		First(&character, id)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil // ไม่พบข้อมูล
		}
		return nil, result.Error // Error อื่นๆ
	}

	return &character, nil
}

// Delete ทำหน้าที่ลบตัวละครด้วย ID
func (r *characterRepository) Delete(characterID uint) error {
	// ใช้คำสั่ง .Delete ของ GORM เพื่อลบข้อมูล
	// GORM จะสร้าง SQL "DELETE FROM characters WHERE id = ?" ให้โดยอัตโนมัติ
	result := r.db.Delete(&domain.Character{}, characterID)

	// // ตรวจสอบว่ามีแถวข้อมูลถูกลบไปจริงหรือไม่
	// if result.RowsAffected == 0 {
	// 	return gorm.ErrRecordNotFound // คืนค่า Error มาตรฐานของ GORM ว่า "หาไม่เจอ"
	// }

	return result.Error
}

// FindInventoryByCharacterID ค้นหาไอเทมทั้งหมดในคลังของตัวละคร
func (r *characterRepository) FindInventoryByCharacterID(characterID uint) ([]*domain.DimensionalSealInventory, error) {
	var inventory []*domain.DimensionalSealInventory
	err := r.db.Where("character_id = ?", characterID).Find(&inventory).Error
	if err != nil {
		return nil, err
	}
	return inventory, nil
}

// UpdateCharacterInTx อัปเดตข้อมูลตัวละครภายใน Transaction
func (r *characterRepository) UpdateCharacterInTx(tx *gorm.DB, character *domain.Character) error {
	return tx.Save(character).Error
}

// ConsumeAndUpdateInventoryInTx คือ Logic ที่ซับซ้อนที่สุดของเรา
// ทำหน้าที่ "เช็คของ", "หักของ", และ "เพิ่มของ" ทั้งหมดใน Transaction เดียว
func (r *characterRepository) ConsumeAndUpdateInventoryInTx(tx *gorm.DB, characterID uint, itemsToConsume map[uint]int, itemsToAdd map[uint]int) error {
	// 1. ดึงข้อมูลคลังปัจจุบันทั้งหมดขึ้นมา และ Lock แถวข้อมูลไว้
	var currentInventory []*domain.DimensionalSealInventory
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("character_id = ?", characterID).Find(&currentInventory).Error
	if err != nil {
		return err
	}

	inventoryMap := make(map[uint]*domain.DimensionalSealInventory)
	for _, item := range currentInventory {
		inventoryMap[item.ElementID] = item
	}

	// 2. ตรวจสอบและหัก "ของที่ต้องใช้" (Consume)
	for elementID, quantityNeeded := range itemsToConsume {
		item, ok := inventoryMap[elementID]
		if !ok || item.Quantity < quantityNeeded {
			return fmt.Errorf("insufficient ingredient: element_id %d", elementID)
		}
		item.Quantity -= quantityNeeded
	}

	// 3. เพิ่ม "ของที่ได้" (Add)
	for elementID, quantityGained := range itemsToAdd {
		if item, ok := inventoryMap[elementID]; ok {
			item.Quantity += quantityGained
		} else {
			// ถ้าเป็นไอเทมใหม่ที่ยังไม่มีในคลัง ก็สร้างขึ้นมาใหม่
			newItem := &domain.DimensionalSealInventory{
				CharacterID: characterID,
				ElementID:   elementID,
				Quantity:    quantityGained,
				ItemType:    domain.ItemTypeNormal,
			}
			currentInventory = append(currentInventory, newItem)
			inventoryMap[elementID] = newItem
		}
	}

	// 4. บันทึกการเปลี่ยนแปลงทั้งหมดกลับลง Database
	for _, item := range currentInventory {
		if item.Quantity > 0 {
			if err := tx.Save(item).Error; err != nil {
				return err
			}
		} else {
			// ถ้าจำนวนเหลือ 0 ก็ลบทิ้งไปเลย
			if err := tx.Delete(item).Error; err != nil {
				return err
			}
		}
	}

	return nil
}

// RegenerateStats คือกลไกการฟื้นฟู MP ของเรา
func (r *characterRepository) RegenerateStats(character *domain.Character, gameDataRepo game_data.GameDataRepository) (*domain.Character, error) {
	timePassed := time.Since(character.StatsUpdatedAt)
	minutesPassed := int(timePassed.Minutes())

	if minutesPassed <= 0 {
		return character, nil
	}

	baseMpStr, _ := gameDataRepo.GetGameConfigValue("STAT_MP_BASE")
	mpPerTalentLStr, _ := gameDataRepo.GetGameConfigValue("STAT_MP_PER_TALENT_L")
	baseMp, _ := strconv.Atoi(baseMpStr)
	mpPerTalentL, _ := strconv.Atoi(mpPerTalentLStr)

	maxMP := baseMp + (character.TalentL * mpPerTalentL)

	mpToRegen := minutesPassed * 5 // สมมติว่า 5 MP/นาที

	newMP := character.CurrentMP + mpToRegen
	if newMP > maxMP {
		newMP = maxMP
	}

	if newMP == character.CurrentMP {
		return character, nil
	}

	character.CurrentMP = newMP
	character.StatsUpdatedAt = time.Now()

	// เรียกใช้ Save (ตัวใหม่) เพื่อบันทึกข้อมูล
	return r.Save(character)
}
