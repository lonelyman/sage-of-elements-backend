package fusion

import (
	"gorm.io/gorm"

	"sage-of-elements-backend/internal/domain"
	"sage-of-elements-backend/internal/modules/character"
	"sage-of-elements-backend/internal/modules/game_data"
	"sage-of-elements-backend/pkg/apperrors"
	"sage-of-elements-backend/pkg/applogger"
)

// --- DTOs ---
type IngredientInput struct {
	ElementID uint `json:"element_id" validate:"required"`
	Quantity  int  `json:"quantity" validate:"required,gte=1"`
}
type CraftResult struct {
	NewElement       *domain.Element `json:"newElement"`
	IsFirstDiscovery bool            `json:"isFirstDiscovery"`
}

// --- Service Interface (ใช้ชื่อ FusionService) ---
type FusionService interface {
	CraftElement(playerID, characterID uint, ingredients []IngredientInput) (*CraftResult, error)
}

// --- Service Implementation ---
type fusionService struct {
	appLogger     applogger.Logger
	db            *gorm.DB
	fusionRepo    FusionRepository
	characterRepo character.CharacterRepository
	gameDataRepo  game_data.GameDataRepository
}

// (ใช้ชื่อ NewFusionService และคืนค่าเป็น FusionService)
func NewFusionService(
	appLogger applogger.Logger,
	db *gorm.DB,
	fusionRepo FusionRepository,
	characterRepo character.CharacterRepository,
	gameDataRepo game_data.GameDataRepository,
) FusionService {
	return &fusionService{
		appLogger:     appLogger,
		db:            db,
		fusionRepo:    fusionRepo,
		characterRepo: characterRepo,
		gameDataRepo:  gameDataRepo,
	}
}

// CraftElement คือ Logic หลักของการหลอมรวมธาตุ (ฉบับสมบูรณ์)
func (s *fusionService) CraftElement(playerID, characterID uint, ingredients []IngredientInput) (*CraftResult, error) {
	// --- ส่วนที่ 1: ตรวจสอบและค้นหา (ส่วนนี้ยังเหมือนเดิม) ---
	// 1.1 ตรวจสอบสิทธิ์ความเป็นเจ้าของตัวละคร
	char, err := s.characterRepo.FindByID(characterID)
	if err != nil || char == nil {
		return nil, apperrors.NotFoundError("character not found")
	}
	if char.PlayerID != playerID {
		return nil, apperrors.PermissionDeniedError("you are not the owner of this character")
	}

	char, err = s.characterRepo.RegenerateStats(char, s.gameDataRepo)
	if err != nil {
		s.appLogger.Error("failed to regenerate stats for character", err)
		return nil, apperrors.SystemError("failed to regenerate character stats")
	}

	// 1.2 แปลง Input และค้นหาสูตร
	ingredientMap := make(map[uint]int)
	for _, ing := range ingredients {
		ingredientMap[ing.ElementID] += ing.Quantity
	}
	recipe, err := s.fusionRepo.FindRecipeByIngredients(s.db, ingredientMap)
	if err != nil {
		s.appLogger.Error("error finding recipe", err)
		return nil, apperrors.SystemError("error finding recipe")
	}
	if recipe == nil {
		return nil, apperrors.NotFoundError("recipe not found or ingredients are incorrect")
	}
	s.appLogger.Info("DEBUG: Checking MP", "character_mp", char.CurrentMP, "recipe_cost", recipe.BaseMPCost)

	// --- ส่วนที่ 2: ✨ LOGIC ที่ยกเครื่องใหม่ทั้งหมด! ✨ ---
	// 2.1 ตรวจสอบทรัพยากรที่แท้จริง: MP เท่านั้น!
	if char.CurrentMP < recipe.BaseMPCost {
		return nil, apperrors.New(422, "INSUFFICIENT_MP", "Insufficient MP")
	}
	// เราจะไม่เช็ค Inventory สำหรับวัตถุดิบ T0 อีกต่อไปแล้ว!

	// --- ส่วนที่ 3: ✨ Transaction ที่ทำงานตามกฎใหม่! ✨ ---
	var finalResult *CraftResult
	txErr := s.db.Transaction(func(tx *gorm.DB) error {
		// 3.1 "จ่าย" ทรัพยากร -> หัก MP ของตัวละคร
		char.CurrentMP -= recipe.BaseMPCost
		if err := s.characterRepo.UpdateCharacterInTx(tx, char); err != nil {
			return err // ถ้าอัปเดต MP ไม่ได้ ก็ Rollback
		}

		// 3.2 ตรวจสอบและ "บันทึกการค้นพบ" ลงในสมุดบันทึก
		isDiscovered, err := s.fusionRepo.IsRecipeDiscovered(characterID, recipe.ID)
		if err != nil {
			return err // ถ้าเช็คไม่ได้ ก็ Rollback
		}

		// ถ้ายังไม่เคยค้นพบ ก็ให้บันทึกซะ
		if !isDiscovered {
			if err := s.fusionRepo.LogDiscovery(tx, characterID, recipe.ID); err != nil {
				return err // ถ้าบันทึกไม่ได้ ก็ Rollback
			}
		}

		// ⭐️ ข้อแตกต่างที่สำคัญ: เราจะไม่ยุ่งกับ Inventory อีกต่อไปแล้วใน Logic นี้! ⭐️

		// 3.3 เตรียมผลลัพธ์ที่จะส่งกลับ
		finalResult = &CraftResult{
			NewElement:       recipe.OutputElement, // ส่งข้อมูลธาตุที่ค้นพบกลับไป
			IsFirstDiscovery: !isDiscovered,        // บอก Client ด้วยว่านี่คือการค้นพบครั้งแรกหรือไม่
		}

		return nil // ทุกอย่างเรียบร้อย, Commit Transaction!
	})

	// --- ส่วนที่ 4: จัดการ Error ของ Transaction (ส่วนนี้เหมือนเดิม) ---
	if txErr != nil {
		s.appLogger.Error("failed to execute crafting transaction", txErr)
		if appErr, ok := txErr.(*apperrors.AppError); ok {
			return nil, appErr
		}
		return nil, apperrors.SystemError("an unexpected error occurred during transaction")
	}

	return finalResult, nil
}
