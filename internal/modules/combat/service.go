// file: internal/modules/combat/service.go
package combat

import (
	"encoding/json"
	"fmt"
	"sage-of-elements-backend/internal/domain"
	"sage-of-elements-backend/internal/modules/character"
	"sage-of-elements-backend/internal/modules/deck"
	"sage-of-elements-backend/internal/modules/enemy"
	"sage-of-elements-backend/internal/modules/game_data"
	"sage-of-elements-backend/internal/modules/pve"
	"sage-of-elements-backend/pkg/apperrors"
	"sage-of-elements-backend/pkg/applogger"
	"strconv"

	"github.com/gofrs/uuid"
	"gorm.io/datatypes"
)

// --- Interface (เหมือนเดิม) ---
type CombatService interface {
	CreateMatch(playerID uint, req CreateMatchRequest) (*domain.CombatMatch, error)
	PerformAction(playerID uint, matchID string, req PerformActionRequest) (*PerformActionResponse, error)
}

// --- Implementation ---
type combatService struct {
	appLogger     applogger.Logger
	combatRepo    CombatRepository
	characterRepo character.CharacterRepository
	enemyRepo     enemy.EnemyRepository
	pveRepo       pve.PveRepository
	gameDataRepo  game_data.GameDataRepository
	deckRepo      deck.DeckRepository
}

func NewCombatService(
	appLogger applogger.Logger,
	combatRepo CombatRepository,
	characterRepo character.CharacterRepository,
	enemyRepo enemy.EnemyRepository,
	pveRepo pve.PveRepository,
	gameDataRepo game_data.GameDataRepository,
	deckRepo deck.DeckRepository,
) CombatService {
	return &combatService{
		appLogger:     appLogger,
		combatRepo:    combatRepo,
		characterRepo: characterRepo,
		enemyRepo:     enemyRepo,
		pveRepo:       pveRepo,
		gameDataRepo:  gameDataRepo,
		deckRepo:      deckRepo,
	}
}

// CreateMatch คือ Logic การสร้างห้องต่อสู้สำหรับ "โหมดฝึกซ้อม"
func (s *combatService) CreateMatch(playerID uint, req CreateMatchRequest) (*domain.CombatMatch, error) {
	// 1. ตรวจสอบความเป็นเจ้าของตัวละคร
	playerChar, err := s.characterRepo.FindByID(req.CharacterID)
	if err != nil || playerChar == nil || playerChar.PlayerID != playerID {
		return nil, apperrors.PermissionDeniedError("character not found or you are not the owner")
	}

	// 2. ดึง "กฎ" การคำนวณ Stat ทั้งหมดมาจาก Cache!
	hpBaseStr, _ := s.gameDataRepo.GetGameConfigValue("STAT_HP_BASE")
	hpPerTalentStr, _ := s.gameDataRepo.GetGameConfigValue("STAT_HP_PER_TALENT_S")
	initBaseStr, _ := s.gameDataRepo.GetGameConfigValue("STAT_INITIATIVE_BASE")
	initPerTalentStr, _ := s.gameDataRepo.GetGameConfigValue("STAT_INITIATIVE_PER_TALENT_G")

	hpBase, _ := strconv.Atoi(hpBaseStr)
	hpPerTalent, _ := strconv.Atoi(hpPerTalentStr)
	initBase, _ := strconv.Atoi(initBaseStr)
	initPerTalent, _ := strconv.Atoi(initPerTalentStr)

	// 3. สร้าง Combatant ของ "ผู้เล่น" (โดยใช้ "กฎ" ที่ดึงมา)
	playerCombatantID, _ := uuid.NewV7()
	playerCombatant := &domain.Combatant{
		ID:          playerCombatantID,
		CharacterID: &playerChar.ID,
		Initiative:  initBase + (playerChar.TalentG * initPerTalent),
		CurrentHP:   hpBase + (playerChar.TalentS * hpPerTalent),
		CurrentMP:   playerChar.CurrentMP,
		CurrentAP:   0,
	}

	// 4. "โหลดคลังกระสุน" T1
	var combatantDeck []*domain.CombatantDeck
	if req.DeckID != nil {
		deckID := *req.DeckID
		deckData, err := s.deckRepo.FindByID(deckID)
		if err != nil {
			return nil, apperrors.NotFoundError("selected deck not found")
		}
		if deckData.CharacterID != req.CharacterID {
			return nil, apperrors.PermissionDeniedError("selected deck does not belong to this character")
		}

		for _, slot := range deckData.Slots {
			newCharge := &domain.CombatantDeck{
				ID:          uuid.Must(uuid.NewV7()),
				CombatantID: playerCombatantID,
				ElementID:   slot.ElementID,
				IsConsumed:  false,
			}
			combatantDeck = append(combatantDeck, newCharge)
		}
	}
	playerCombatant.Deck = combatantDeck

	// 5. สร้าง Combatant ของ "ศัตรู"
	var combatants []*domain.Combatant
	combatants = append(combatants, playerCombatant)
	if req.MatchType == "TRAINING" {

		fmt.Println("req.TrainingEnemies", req.TrainingEnemies)
		for _, enemyInfo := range req.TrainingEnemies {
			s.appLogger.Info("Attempting to find enemy in repo", "enemy_id_to_find", enemyInfo.EnemyID)

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
				CurrentMP:  9999,
				CurrentAP:  0,
			}
			combatants = append(combatants, enemyCombatant)
		}
	} else {
		return nil, apperrors.InvalidFormatError("unsupported match type", nil)
	}

	// 6. ตัดสินลำดับเทิร์น และแจก AP เริ่มต้น
	apPerTurnStr, _ := s.gameDataRepo.GetGameConfigValue("COMBAT_AP_PER_TURN")
	apPerTurn, _ := strconv.Atoi(apPerTurnStr)
	var firstTurnCombatant *domain.Combatant = playerCombatant
	for _, c := range combatants {
		if c.Initiative > firstTurnCombatant.Initiative {
			firstTurnCombatant = c
		}
	}
	firstTurnCombatant.CurrentAP = apPerTurn
	s.appLogger.Info("Granting starting AP", "combatant_id", firstTurnCombatant.ID, "ap", firstTurnCombatant.CurrentAP)

	// 7. ประกอบร่างห้องต่อสู้
	var modifiersJSON datatypes.JSON
	if req.Modifiers != nil {
		jsonBytes, err := json.Marshal(req.Modifiers)
		if err != nil {
			return nil, apperrors.SystemError("failed to create match due to modifier issue")
		}
		modifiersJSON = jsonBytes
	}
	matchID, _ := uuid.NewV7()
	newMatch := &domain.CombatMatch{
		ID:          matchID,
		Status:      domain.MatchInProgress,
		Modifiers:   modifiersJSON,
		TurnNumber:  1,
		CurrentTurn: firstTurnCombatant.ID,
		Combatants:  combatants,
	}

	// 8. บันทึกลง Database
	return s.combatRepo.CreateMatch(newMatch)
}

func (s *combatService) PerformAction(playerID uint, matchID string, req PerformActionRequest) (*PerformActionResponse, error) {
	// 1. โหลด Match
	match, err := s.combatRepo.FindMatchByID(matchID)
	if err != nil {
		return nil, apperrors.NotFoundError("match not found")
	}
	if match.Status != domain.MatchInProgress {
		return nil, apperrors.New(400, "MATCH_FINISHED", "this match is already finished")
	}

	// 2. หาตัวตนและตรวจสอบเทิร์น
	playerCombatant := s.findPlayerCombatant(match)
	if playerCombatant == nil || playerCombatant.Character.PlayerID != playerID {
		return nil, apperrors.PermissionDeniedError("you are not part of this match")
	}
	if match.CurrentTurn != playerCombatant.ID {
		return nil, apperrors.New(400, "NOT_YOUR_TURN", "it's not your turn")
	}

	// --- ✨⭐️ นี่คือ "หัวใจ" ที่อัปเกรดแล้ว! ⭐️✨ ---
	// 3. ส่งต่องานให้ "ผู้เชี่ยวชาญ"
	var actionErr error
	switch req.ActionType {

	case "EMERGENCY_FUSION": // ⭐️ "ภารกิจใหม่" ของเรา!
		// TODO: สร้างฟังก์ชัน executeEmergencyFusion
		actionErr = apperrors.New(501, "NOT_IMPLEMENTED", "EMERGENCY_FUSION is not yet implemented")

	case "END_TURN":
		s.appLogger.Info("Player ended their turn.", "match_id", matchID)
		match = s.endTurn(match)
		match, actionErr = s.startNewTurn(match)

	case "CAST_SPELL":
		actionErr = s.executeCastSpell(playerCombatant, match, req) // ⭐️ เรียกใช้ "สมอง" ที่อัปเกรดแล้ว!

	default:
		return nil, apperrors.InvalidFormatError("unsupported action type", nil)
	}
	if actionErr != nil {
		return nil, actionErr
	}
	// ---------------------------------------------

	// 4. ปลุกชีพ AI (ถ้าจำเป็น)
	nextCombatant := s.findCombatantByID(match, match.CurrentTurn)
	if nextCombatant != nil && nextCombatant.EnemyID != nil {
		s.appLogger.Info("AI turn begins.", "ai_combatant_id", nextCombatant.ID)
		match, err = s.processAITurn(match, nextCombatant)
		if err != nil {
			return nil, err
		}
	}

	// 5. ตรวจสอบจบเกม
	match = s.checkMatchEndCondition(match)

	// 6. บันทึกและส่งคืน
	updatedMatch, err := s.combatRepo.UpdateMatch(match)
	if err != nil {
		return nil, err
	}
	return &PerformActionResponse{
		UpdatedMatch:    updatedMatch,
		PerformedAction: req,
	}, nil
}
