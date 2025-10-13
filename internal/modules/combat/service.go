// file: internal/modules/combat/service.go
package combat

import (
	"encoding/json"
	"fmt"
	"sage-of-elements-backend/internal/domain"
	"sage-of-elements-backend/internal/modules/character"
	"sage-of-elements-backend/internal/modules/enemy"
	"sage-of-elements-backend/internal/modules/pve"
	"sage-of-elements-backend/pkg/apperrors"
	"sage-of-elements-backend/pkg/applogger"

	"github.com/gofrs/uuid"
	"gorm.io/datatypes"
)

// --- Interface (เหมือนเดิม) ---
type CombatService interface {
	CreateMatch(playerID uint, req CreateMatchRequest) (*domain.CombatMatch, error)
}

// --- Implementation ---
type combatService struct {
	appLogger     applogger.Logger
	combatRepo    CombatRepository
	characterRepo character.CharacterRepository
	enemyRepo     enemy.EnemyRepository
	pveRepo       pve.PveRepository
}

func NewCombatService(
	appLogger applogger.Logger,
	combatRepo CombatRepository,
	characterRepo character.CharacterRepository,
	enemyRepo enemy.EnemyRepository,
	pveRepo pve.PveRepository,
) CombatService {
	return &combatService{
		appLogger:     appLogger,
		combatRepo:    combatRepo,
		characterRepo: characterRepo,
		enemyRepo:     enemyRepo,
		pveRepo:       pveRepo,
	}
}

// CreateMatch คือ Logic การสร้างห้องต่อสู้สำหรับ "โหมดฝึกซ้อม"
func (s *combatService) CreateMatch(playerID uint, req CreateMatchRequest) (*domain.CombatMatch, error) {
	// 1. ตรวจสอบความเป็นเจ้าของตัวละคร
	playerChar, err := s.characterRepo.FindByID(req.CharacterID)
	if err != nil || playerChar == nil || playerChar.PlayerID != playerID {
		return nil, apperrors.PermissionDeniedError("character not found or you are not the owner")
	}

	// 2. สร้าง Combatant ของ "ผู้เล่น"
	playerCombatantID, _ := uuid.NewV7()
	playerCombatant := &domain.Combatant{
		ID:          playerCombatantID,
		CharacterID: playerChar.ID,
		Initiative:  50 + playerChar.TalentG,
		CurrentHP:   900 + (playerChar.TalentS * 30),
		CurrentMP:   playerChar.CurrentMP,
		CurrentAP:   0,
	}

	var combatants []*domain.Combatant
	combatants = append(combatants, playerCombatant)

	// 3. สร้าง Combatant ของ "ศัตรู" (ตอนนี้รองรับแค่ Training Mode)
	if req.MatchType == "TRAINING" {
		for _, enemyInfo := range req.TrainingEnemies {
			enemyData, err := s.enemyRepo.FindByID(enemyInfo.EnemyID)
			if err != nil || enemyData == nil {
				return nil, apperrors.NotFoundError(fmt.Sprintf("enemy with id %d not found", enemyInfo.EnemyID))
			}
			enemyCombatantID, _ := uuid.NewV7()
			enemyCombatant := &domain.Combatant{
				ID:         enemyCombatantID,
				EnemyID:    &enemyData.ID,
				Initiative: enemyData.Initiative,
				CurrentHP:  enemyData.MaxHP,
				CurrentMP:  9999, // ศัตรู MP ไม่จำกัด
				CurrentAP:  0,
			}
			combatants = append(combatants, enemyCombatant)
		}
	} else {
		// TODO: เพิ่ม Logic สำหรับ PVE_STAGE
		return nil, apperrors.InvalidFormatError("unsupported match type", nil)
	}

	// 4. ตัดสินลำดับเทิร์น (หา Initiative สูงสุด)
	var currentTurnID uuid.UUID = playerCombatantID
	maxInitiative := playerCombatant.Initiative
	for _, c := range combatants {
		if c.Initiative > maxInitiative {
			maxInitiative = c.Initiative
			currentTurnID = c.ID
		}
	}

	// 5. ประกอบร่างห้องต่อสู้
	// 5.1 แปลงร่าง Modifiers (struct -> JSON bytes) ก่อน
	var modifiersJSON datatypes.JSON
	if req.Modifiers != nil {
		jsonBytes, err := json.Marshal(req.Modifiers)
		if err != nil {
			s.appLogger.Error("failed to marshal match modifiers", err)
			return nil, apperrors.SystemError("failed to create match due to modifier issue")
		}
		modifiersJSON = jsonBytes
	}

	// 5.2 ประกอบร่าง CombatMatch object
	matchID, _ := uuid.NewV7()
	newMatch := &domain.CombatMatch{
		ID:          matchID,
		Status:      domain.MatchInProgress,
		Modifiers:   modifiersJSON, // <-- ⭐️ ใช้ตัวแปรที่แปลงร่างแล้ว
		TurnNumber:  1,
		CurrentTurn: currentTurnID,
		Combatants:  combatants,
	}
	// --------------------------

	// 6. บันทึกลง Database (เหมือนเดิม)
	return s.combatRepo.CreateMatch(newMatch)
}
