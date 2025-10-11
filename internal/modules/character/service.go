package character

import (
	"errors"
	"time"

	"sage-of-elements-backend/internal/domain"
	"sage-of-elements-backend/internal/modules/game_data"
	"sage-of-elements-backend/pkg/apperrors"
	"sage-of-elements-backend/pkg/applogger"
	"strconv"
)

// CharacterService คือ "สัญญา" สำหรับ Business Logic ที่เกี่ยวกับตัวละคร
type CharacterService interface {
	CreateCharacter(playerID uint, name string, gender string, elementID uint, masteryID uint) (*domain.Character, error)
	ListCharacters(playerID uint) ([]domain.Character, error)
	GetCharacterByID(playerID, characterID uint) (*domain.Character, error)
	DeleteCharacter(playerID, characterID uint) error
	GetInventory(playerID, characterID uint) (*InventoryResponse, error)
}

// characterService คือ struct ที่จะเก็บ Logic การทำงานจริง
type characterService struct {
	appLogger     applogger.Logger
	repoCharacter CharacterRepository
	repoGameData  game_data.GameDataRepository
	// อนาคตเราอาจจะต้องใช้ repo อื่นๆ เช่น elementRepo เพื่อตรวจสอบข้อมูล
}

type InventoryResponse struct {
	CharacterID uint                               `json:"characterId"`
	CurrentMP   int                                `json:"currentMp"`
	Inventory   []*domain.DimensionalSealInventory `json:"inventory"`
}

// NewCharacterService คือฟังก์ชันสำหรับสร้าง Service ขึ้นมาใช้งาน
func NewCharacterService(appLogger applogger.Logger, repoCharacter CharacterRepository, repoGameData game_data.GameDataRepository) CharacterService {
	return &characterService{
		appLogger:     appLogger,
		repoCharacter: repoCharacter,
		repoGameData:  repoGameData,
	}
}

// CreateCharacter คือฟังก์ชันสำหรับสร้างตัวละครใหม่
func (s *characterService) CreateCharacter(playerID uint, name string, gender string, elementID uint, masteryID uint) (*domain.Character, error) {
	// 1. Validation และ 2. Duplicate Check (เหมือนเดิม)
	if len(name) < 3 {
		return nil, errors.New("character name must be at least 3 characters")
	}
	existingChar, _ := s.repoCharacter.CheckCharacterExists(name)
	if existingChar != nil {
		return nil, errors.New("character name already exists")
	}

	// 3. คำนวณค่า Talent เริ่มต้น!
	talents := calculateInitialTalents(gender, elementID)

	// 4. สร้าง Struct Character หลัก
	newCharacter := &domain.Character{
		PlayerID:         playerID,
		CharacterName:    name,
		Gender:           gender,
		PrimaryElementID: elementID,
		Level:            1,
		Exp:              0,
		TalentS:          talents["S"],
		TalentL:          talents["L"],
		TalentG:          talents["G"],
		TalentP:          talents["P"],
	}

	// --- ⭐️ ส่วนที่เพิ่มเข้ามา ⭐️ ---
	// 5. คำนวณ MaxHP/MP และตั้งค่า HP/MP เริ่มต้น
	baseHpStr, _ := s.repoGameData.GetGameConfigValue("STAT_HP_BASE")
	hpPerTalentSStr, _ := s.repoGameData.GetGameConfigValue("STAT_HP_PER_TALENT_S")
	baseMpStr, _ := s.repoGameData.GetGameConfigValue("STAT_MP_BASE")
	mpPerTalentLStr, _ := s.repoGameData.GetGameConfigValue("STAT_MP_PER_TALENT_L")

	baseHp, _ := strconv.Atoi(baseHpStr)
	hpPerTalentS, _ := strconv.Atoi(hpPerTalentSStr)
	baseMp, _ := strconv.Atoi(baseMpStr)
	mpPerTalentL, _ := strconv.Atoi(mpPerTalentLStr)

	maxHP := baseHp + (newCharacter.TalentS * hpPerTalentS)
	maxMP := baseMp + (newCharacter.TalentL * mpPerTalentL)

	newCharacter.CurrentHP = maxHP
	newCharacter.CurrentMP = maxMP
	// ---------------------------------

	// 6. สร้างข้อมูล Mastery เริ่มต้น
	masteries := []*domain.CharacterMastery{
		{MasteryID: 1, Level: 1, Mxp: 0},
		{MasteryID: 2, Level: 1, Mxp: 0},
		{MasteryID: 3, Level: 1, Mxp: 0},
		{MasteryID: 4, Level: 1, Mxp: 0},
	}
	newCharacter.Masteries = masteries

	// 7. บันทึกลง Database ผ่าน Repository
	savedCharacter, err := s.repoCharacter.Save(newCharacter)
	if err != nil {
		s.appLogger.Error("failed to save character in repository", err)
		return nil, errors.New("failed to save character to database")
	}

	return savedCharacter, nil
}

// calculateInitialTalents เป็น helper function ที่ใช้คำนวณ Talent เริ่มต้น
func calculateInitialTalents(gender string, elementID uint) map[string]int {
	talents := map[string]int{"S": 3, "L": 3, "G": 3, "P": 3}

	// Affinity Map (กฎความสัมพันธ์ของธาตุ)
	// สมมติว่า ID 1=S, 2=L, 3=G, 4=P
	affinityMap := map[uint]uint{1: 4, 2: 3, 3: 2, 4: 1} // S->P, L->G, G->L, P->S

	elementMap := map[uint]string{1: "S", 2: "L", 3: "G", 4: "P"}

	// 1. ใส่ 90 แต้มให้ธาตุหลัก
	primaryElementString := elementMap[elementID]
	talents[primaryElementString] = 90

	// 2. ใส่ 4 แต้มให้ธาตุคู่หู
	affinityElementID := affinityMap[elementID]
	affinityElementString := elementMap[affinityElementID]
	talents[affinityElementString] = 4

	// 3. ใส่โบนัสเพศ +5
	switch gender {
	case "MALE":
		talents["P"] += 5
	case "FEMALE":
		talents["L"] += 5
	}

	return talents
}

// ListCharacters คือฟังก์ชันสำหรับดึงรายชื่อตัวละครทั้งหมดของผู้เล่น
func (s *characterService) ListCharacters(playerID uint) ([]domain.Character, error) {
	// 1. ตรวจสอบ Input พื้นฐาน (เผื่อไว้)
	if playerID == 0 {
		return nil, apperrors.InvalidFormatError("Invalid player ID", nil)
	}

	// 2. เรียกใช้ Repository เพื่อค้นหาตัวละครทั้งหมดด้วย Player ID
	characters, err := s.repoCharacter.FindAllByPlayerID(playerID)
	if err != nil {
		// ถ้าเกิด Error จาก Database
		s.appLogger.Error("failed to find characters by player id in repository", err)
		return nil, apperrors.SystemError("failed to retrieve characters")
	}

	// 3. คืนค่า Slice ของ Characters ที่หาเจอ (อาจจะเป็น Slice ว่างๆ ถ้ายังไม่มีตัวละคร)
	return characters, nil
}

// GetCharacterByID คือฟังก์ชันสำหรับดึงข้อมูลของตัวละครที่ระบุ ID
func (s *characterService) GetCharacterByID(playerID, characterID uint) (*domain.Character, error) {
	// 1. ตรวจสอบ Input พื้นฐาน
	if characterID == 0 {
		return nil, apperrors.InvalidFormatError("Invalid character ID", nil)
	}

	// 2. เรียกใช้ Repository เพื่อค้นหาตัวละครด้วย ID
	character, err := s.repoCharacter.FindByID(characterID)
	if err != nil {
		s.appLogger.Error("failed to find character by id in repository", err)
		return nil, apperrors.SystemError("failed to retrieve character")
	}
	if character == nil {
		return nil, apperrors.NotFoundError("character not found")
	}

	// 3. ✨ หัวใจด้านความปลอดภัย! ✨
	// ตรวจสอบให้แน่ใจว่าผู้เล่นที่กำลังขอข้อมูล (playerID) เป็น "เจ้าของ" ตัวละครตัวนั้นจริงๆ
	if character.PlayerID != playerID {
		return nil, apperrors.PermissionDeniedError("you are not the owner of this character")
	}

	// 4. คืนค่า Character ที่หาเจอ
	return character, nil
}

// DeleteCharacter คือฟังก์ชันสำหรับลบตัวละคร
func (s *characterService) DeleteCharacter(playerID, characterID uint) error {
	// 1. ตรวจสอบ Input พื้นฐาน
	if characterID == 0 {
		return apperrors.InvalidFormatError("Invalid character ID", nil)
	}

	// 2. ดึงข้อมูลตัวละครขึ้นมาก่อน เพื่อ "ตรวจสอบความเป็นเจ้าของ"
	character, err := s.repoCharacter.FindByID(characterID)
	if err != nil {
		s.appLogger.Error("failed to find character by id in repository for deletion", err)
		return apperrors.SystemError("failed to retrieve character for deletion")
	}
	if character == nil {
		return apperrors.NotFoundError("character not found")
	}

	// 3. ✨ หัวใจด้านความปลอดภัย! ✨
	// ตรวจสอบว่าผู้เล่นที่ส่งคำขอ (playerID) เป็น "เจ้าของ" ตัวละครตัวนั้นจริงๆ หรือไม่
	if character.PlayerID != playerID {
		return apperrors.PermissionDeniedError("you do not have permission to delete this character")
	}

	// 4. ถ้าทุกอย่างถูกต้อง... สั่งให้ Repository ทำการลบ
	s.appLogger.Info("deleting character", "character_id", characterID, "player_id", playerID)
	err = s.repoCharacter.Delete(characterID)
	if err != nil {
		s.appLogger.Error("failed to delete character in repository", err)
		return apperrors.SystemError("failed to delete character")
	}

	return nil // คืนค่า nil แปลว่าลบสำเร็จ
}

// GetInventory คือฟังก์ชันสำหรับดึงข้อมูลคลังและสถานะปัจจุบันของตัวละคร
func (s *characterService) GetInventory(playerID, characterID uint) (*InventoryResponse, error) {
	// 1. ดึงข้อมูลตัวละครหลักขึ้นมาก่อน เพื่อตรวจสอบความเป็นเจ้าของและเอาค่า MP
	character, err := s.repoCharacter.FindByID(characterID)
	if err != nil {
		s.appLogger.Error("failed to find character by id for inventory check", err)
		return nil, apperrors.SystemError("failed to retrieve character")
	}
	if character == nil {
		return nil, apperrors.NotFoundError("character not found")
	}

	// 2. ตรวจสอบความเป็นเจ้าของ
	if character.PlayerID != playerID {
		return nil, apperrors.PermissionDeniedError("you are not the owner of this character")
	}

	// 3. ดึงข้อมูลไอเทมในคลังทั้งหมด
	inventoryItems, err := s.repoCharacter.FindInventoryByCharacterID(characterID)
	if err != nil {
		s.appLogger.Error("failed to find inventory for character", err)
		return nil, apperrors.SystemError("failed to retrieve inventory")
	}

	// 4. ประกอบร่างเป็น Response DTO
	response := &InventoryResponse{
		CharacterID: characterID,
		CurrentMP:   character.CurrentMP,
		Inventory:   inventoryItems,
	}

	return response, nil
}

// ✨ สร้างฟังก์ชันใหม่สำหรับคำนวณการฟื้นฟู ✨
func (s *characterService) RegenerateStats(character *domain.Character) (*domain.Character, error) {
	// 1. คำนวณเวลาที่ผ่านไปตั้งแต่การอัปเดตครั้งล่าสุด
	timePassed := time.Since(character.StatsUpdatedAt)
	minutesPassed := int(timePassed.Minutes())

	if minutesPassed <= 0 {
		return character, nil // ถ้ายังไม่ถึงนาที ก็ไม่ต้องทำอะไร
	}

	// 2. คำนวณค่า MaxMP (เราอาจจะสร้าง helper function แยกในอนาคต)
	baseMpStr, _ := s.repoGameData.GetGameConfigValue("STAT_MP_BASE")
	mpPerTalentLStr, _ := s.repoGameData.GetGameConfigValue("STAT_MP_PER_TALENT_L")
	baseMp, _ := strconv.Atoi(baseMpStr)
	mpPerTalentL, _ := strconv.Atoi(mpPerTalentLStr)
	maxMP := baseMp + (character.TalentL * mpPerTalentL)

	// 3. คำนวณ MP ที่ควรจะฟื้นฟู (สมมติว่าฟื้น 5 MP/นาที)
	mpToRegen := minutesPassed * 5 // เราสามารถดึง "5" มาจาก Game Config ได้ในอนาคต

	// 4. อัปเดตค่า MP ปัจจุบัน
	newMP := character.CurrentMP + mpToRegen
	if newMP > maxMP {
		newMP = maxMP // ไม่ให้ฟื้นเกิน MaxMP
	}

	if newMP == character.CurrentMP {
		return character, nil // ถ้าไม่มีอะไรเปลี่ยนแปลง ก็ไม่ต้องอัปเดต DB
	}

	// 5. อัปเดตค่าและ "นาฬิกา"
	character.CurrentMP = newMP
	character.StatsUpdatedAt = time.Now()

	// 6. บันทึกข้อมูลใหม่ลง Database
	updatedChar, err := s.repoCharacter.Create(character)
	if err != nil {
		return nil, err
	}

	s.appLogger.Info("Character stats regenerated", "char_id", character.ID, "mp_regened", mpToRegen)
	return updatedChar, nil
}
