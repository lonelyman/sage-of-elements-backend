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
	ResolveSpell(elementID uint, masteryID uint, casterMainElementID uint) (*domain.Spell, error)

	// 🧹 Cleanup Methods
	CleanupStaleMatches(inactiveMinutes int) (int64, error)             // ทำความสะอาด match ค้าง (สำหรับ cron job)
	AbortMatch(matchID string, reason string) error                     // Abort match เฉพาะ (สำหรับ forfeit/disconnect)
	GetPlayerActiveMatch(characterID uint) (*domain.CombatMatch, error) // ตรวจสอบว่าผู้เล่นกำลังเล่นอยู่หรือเปล่า
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

	// 2. ✅ ตรวจสอบว่าผู้เล่นมี active match อยู่หรือเปล่า
	activeMatch, err := s.combatRepo.FindPlayerActiveMatch(req.CharacterID)
	if err != nil {
		s.appLogger.Error("Failed to check active match", err,
			"character_id", req.CharacterID,
		)
		return nil, apperrors.SystemError("failed to check active match")
	}
	if activeMatch != nil {
		return nil, apperrors.New(409, "MATCH_ALREADY_ACTIVE",
			fmt.Sprintf("character already has an active match: %s", activeMatch.ID.String()))
	}

	// 3. ดึง "กฎ" การคำนวณ Stat ทั้งหมดมาจาก Cache!
	hpBaseStr, _ := s.gameDataRepo.GetGameConfigValue("STAT_HP_BASE")
	hpPerTalentStr, _ := s.gameDataRepo.GetGameConfigValue("STAT_HP_PER_TALENT_S")
	initBaseStr, _ := s.gameDataRepo.GetGameConfigValue("STAT_INITIATIVE_BASE")
	initPerTalentStr, _ := s.gameDataRepo.GetGameConfigValue("STAT_INITIATIVE_PER_TALENT_G")

	hpBase, _ := strconv.Atoi(hpBaseStr)
	hpPerTalent, _ := strconv.Atoi(hpPerTalentStr)
	initBase, _ := strconv.Atoi(initBaseStr)
	initPerTalent, _ := strconv.Atoi(initPerTalentStr)

	// 4. สร้าง Combatant ของ "ผู้เล่น" (โดยใช้ "กฎ" ที่ดึงมา)
	playerCombatantID, _ := uuid.NewV7()
	playerCombatant := &domain.Combatant{
		ID:          playerCombatantID,
		CharacterID: &playerChar.ID,
		Initiative:  initBase + (playerChar.TalentG * initPerTalent),
		CurrentHP:   hpBase + (playerChar.TalentS * hpPerTalent),
		CurrentMP:   playerChar.CurrentMP,
		CurrentAP:   0,
	}

	// 5. "โหลดคลังกระสุน" T1
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

	// 6. สร้าง Combatant ของ "ศัตรู" ตามประเภทการต่อสู้
	var combatants []*domain.Combatant
	combatants = append(combatants, playerCombatant)

	switch req.MatchType {
	case "TRAINING":
		// โหมดฝึกซ้อม - เลือกศัตรูเอง
		if len(req.TrainingEnemies) == 0 {
			return nil, apperrors.InvalidFormatError("training_enemies is required for TRAINING mode", nil)
		}

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
				CurrentMP:  9999,
				CurrentAP:  0,
			}
			combatants = append(combatants, enemyCombatant)
		}

	case "STORY":
		// โหมดเนื้อเรื่อง - โหลดศัตรูจากด่าน
		if req.StageID == nil {
			return nil, apperrors.InvalidFormatError("stage_id is required for STORY mode", nil)
		}

		// TODO: ต้องเพิ่ม method ใน PveRepository ก่อน
		// stageData, err := s.pveRepo.FindStageByID(*req.StageID)
		// stageEnemies, err := s.pveRepo.FindStageEnemiesByStageID(*req.StageID)
		// for _, stageEnemy := range stageEnemies {
		//     enemyData, err := s.enemyRepo.FindByID(stageEnemy.EnemyID)
		//     enemyCombatantID, _ := uuid.NewV7()
		//     enemyCombatant := &domain.Combatant{
		//         ID:         enemyCombatantID,
		//         EnemyID:    &enemyData.ID,
		//         Initiative: enemyData.Initiative,
		//         CurrentHP:  enemyData.MaxHP,
		//         CurrentMP:  9999,
		//         CurrentAP:  0,
		//     }
		//     combatants = append(combatants, enemyCombatant)
		// }

		s.appLogger.Warn("STORY mode not fully implemented yet", "stage_id", *req.StageID)
		return nil, apperrors.New(501, "NOT_IMPLEMENTED", "STORY mode is not fully implemented yet")

	case "PVP":
		// PvP - ต่อสู้กับผู้เล่นอื่น
		if req.OpponentID == nil {
			return nil, apperrors.InvalidFormatError("opponent_id is required for PVP mode", nil)
		}

		opponentChar, err := s.characterRepo.FindByID(*req.OpponentID)
		if err != nil || opponentChar == nil {
			return nil, apperrors.NotFoundError(fmt.Sprintf("opponent character with id %d not found", *req.OpponentID))
		}

		// สร้าง Combatant ของฝ่ายตรงข้าม
		opponentCombatantID, _ := uuid.NewV7()
		opponentCombatant := &domain.Combatant{
			ID:          opponentCombatantID,
			CharacterID: &opponentChar.ID,
			Initiative:  initBase + (opponentChar.TalentG * initPerTalent),
			CurrentHP:   hpBase + (opponentChar.TalentS * hpPerTalent),
			CurrentMP:   opponentChar.CurrentMP,
			CurrentAP:   0,
		}

		// TODO: โหลด deck ของฝ่ายตรงข้าม
		// if opponentChar.SelectedDeckID != nil { ... }

		combatants = append(combatants, opponentCombatant)

		s.appLogger.Info("PVP match created",
			"player_char_id", req.CharacterID,
			"opponent_char_id", *req.OpponentID,
		)

	default:
		return nil, apperrors.InvalidFormatError("unsupported match type", nil)
	}

	// 7. ตัดสินลำดับเทิร์น และแจก AP เริ่มต้น
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

	// 8. ประกอบร่างห้องต่อสู้
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
		MatchType:   domain.MatchType(req.MatchType),
		StageID:     req.StageID,
		Status:      domain.MatchInProgress,
		Modifiers:   modifiersJSON,
		TurnNumber:  1,
		CurrentTurn: firstTurnCombatant.ID,
		Combatants:  combatants,
	}

	// 9. บันทึกลง Database
	return s.combatRepo.CreateMatch(newMatch)
}

// PerformAction - ฟังก์ชันหลักในการประมวลผลการกระทำของผู้เล่นในการต่อสู้
//
// 📋 ภาพรวม 6 ขั้นตอนหลัก:
// ┌─────────────────────────────────────────────────────────────────┐
// │ 1. [VALIDATION] โหลดและตรวจสอบ Match                            │
// │    ↓ ตรวจว่า match มีอยู่จริง และยังไม่จบ                       │
// ├─────────────────────────────────────────────────────────────────┤
// │ 2. [AUTHORIZATION] ตรวจสอบสิทธิ์และเทิร์น                       │
// │    ↓ ยืนยันว่าผู้เล่นเป็นเจ้าของตัวละครและถึงเทิร์นของตัวเอง   │
// ├─────────────────────────────────────────────────────────────────┤
// │ 3. [ACTION EXECUTION] ประมวลผลการกระทำของผู้เล่น                │
// │    ↓ แยกเป็น 2 ประเภท:                                          │
// │      • END_TURN: จบเทิร์นและเริ่มเทิร์นใหม่                      │
// │      • CAST_SPELL: ร่ายเวทโจมตี/ฟื้นฟู                          │
// ├─────────────────────────────────────────────────────────────────┤
// │ 4. [EARLY EXIT CHECK] ตรวจสอบจบเกม (เฉพาะ CAST_SPELL)          │
// │    ↓ ถ้ามีคนตายจนเกมจบ → บันทึกและ return ทันที                │
// ├─────────────────────────────────────────────────────────────────┤
// │ 5. [AI PROCESSING] ให้ AI ทุกตัวเล่นต่อเนื่อง                   │
// │    ↓ วน loop จนกว่าจะกลับมาเป็นเทิร์นผู้เล่น                   │
// ├─────────────────────────────────────────────────────────────────┤
// │ 6. [PERSISTENCE] บันทึกผลลัพธ์และส่งคืน                         │
// │    ↓ UpdateMatch → ส่ง response กลับไปหา client                 │
// └─────────────────────────────────────────────────────────────────┘
//
// 🔄 Flow ตัวอย่าง:
//
//	Player Cast Spell → Enemy Dies → Check End → Game Over? Yes → Return
//	Player Cast Spell → Enemy Survives → AI Turn 1 → AI Turn 2 → Back to Player → Return
//	Player End Turn → AI Turn 1 → Check End → AI Turn 2 → Back to Player → Return
func (s *combatService) PerformAction(playerID uint, matchID string, req PerformActionRequest) (*PerformActionResponse, error) {

	// ════════════════════════════════════════════════════════════════
	// ขั้นตอนที่ 1: VALIDATION - โหลดและตรวจสอบ Match
	// ════════════════════════════════════════════════════════════════
	// หน้าที่: ดึงข้อมูลการต่อสู้จาก database และตรวจสอบว่ายังเล่นอยู่
	// Input:   matchID (string) - UUID ของการต่อสู้
	// Output:  match (*domain.CombatMatch) - ข้อมูลการต่อสู้พร้อม combatants
	// Error:   - "match not found" ถ้าไม่มี match นี้ใน DB
	//          - "MATCH_FINISHED" ถ้าเกมจบแล้ว (status != IN_PROGRESS)
	// ────────────────────────────────────────────────────────────────
	match, err := s.combatRepo.FindMatchByID(matchID)
	if err != nil {
		return nil, apperrors.NotFoundError("match not found")
	}
	if match.Status != domain.MatchInProgress {
		return nil, apperrors.New(400, "MATCH_FINISHED", "this match is already finished")
	}

	// ════════════════════════════════════════════════════════════════
	// ขั้นตอนที่ 2: AUTHORIZATION - ตรวจสอบสิทธิ์และเทิร์น
	// ════════════════════════════════════════════════════════════════
	// หน้าที่: ยืนยันตัวตนและตรวจสอบว่าถึงเทิร์นของผู้เล่นหรือยัง
	// Process:
	//   1. หา combatant ของผู้เล่นจาก match.Combatants (CharacterID != nil)
	//   2. ตรวจสอบว่า Character.PlayerID ตรงกับ playerID ที่ส่งมา
	//   3. ตรวจสอบว่า match.CurrentTurn เป็น ID ของ playerCombatant
	// Error:   - "you are not part of this match" ถ้าไม่ใช่เจ้าของ character
	//          - "NOT_YOUR_TURN" ถ้ายังไม่ถึงเทิร์น
	// Note:    ป้องกันการ cheat โดยการส่ง request แทนคนอื่น
	// ────────────────────────────────────────────────────────────────
	playerCombatant := s.findPlayerCombatant(match)
	if playerCombatant == nil || playerCombatant.Character.PlayerID != playerID {
		return nil, apperrors.PermissionDeniedError("you are not part of this match")
	}
	if match.CurrentTurn != playerCombatant.ID {
		return nil, apperrors.New(400, "NOT_YOUR_TURN", "it's not your turn")
	}

	// ════════════════════════════════════════════════════════════════
	// ขั้นตอนที่ 3: ACTION EXECUTION - ประมวลผลการกระทำของผู้เล่น
	// ════════════════════════════════════════════════════════════════
	// หน้าที่: ประมวลผล action ที่ผู้เล่นเลือก (switch-case ตาม ActionType)
	//
	// Note: UpdatedAt จะถูกอัปเดตอัตโนมัติโดย GORM เมื่อเรียก UpdateMatch()
	//       ใช้ UpdatedAt ในการตรวจจับ match ค้าง (stale detection)
	//
	// 3.1) ActionType = "END_TURN"
	//      ├─ endTurn(): เคลียร์ AP ปัจจุบัน, เลื่อน CurrentTurn ไปคนถัดไป
	//      └─ startNewTurn(): แจก AP ใหม่, ลด duration ของ effects, regen resources
	//
	// 3.2) ActionType = "CAST_SPELL"
	//      ├─ executeCastSpellV2(): ร่ายเวท (ตรวจสอบ MP/AP, หา target, คำนวณดาเมจ, apply effects)
	//      ├─ checkMatchEndCondition(): ตรวจว่ามีทีมไหนตายหมดหรือยัง
	//      └─ [EARLY EXIT] ถ้าเกมจบ → UpdateMatch → return ทันที (ไม่ต้องให้ AI เล่นต่อ)
	//
	// Output:  match ที่ถูกแก้ไขแล้ว (CurrentTurn, TurnNumber, Combatants stats)
	// Error:   - actionErr จาก spell casting (ไม่พอ MP, target ไม่ถูกต้อง, etc.)
	//          - "unsupported action type" ถ้าส่ง action ที่ไม่รู้จัก
	// ────────────────────────────────────────────────────────────────
	var actionErr error
	switch req.ActionType {

	case "END_TURN":
		// จบเทิร์นและเริ่มเทิร์นใหม่
		s.appLogger.Info("Player ended their turn.", "match_id", matchID)
		match = s.endTurn(match)                 // เลื่อน CurrentTurn ไปคนถัดไป
		match, actionErr = s.startNewTurn(match) // แจก AP, ลด effect duration, regen

	case "CAST_SPELL":
		// ร่ายเวทโจมตีหรือฟื้นฟู
		actionErr = s.executeCastSpellV2(playerCombatant, match, req)

		// ✅ ตรวจสอบจบเกมทันทีหลังร่ายเวท (เพราะอาจมีคนตายจนเกมจบ)
		match = s.checkMatchEndCondition(match)
		if match.Status != domain.MatchInProgress {
			// เกมจบแล้ว (PLAYER_WIN หรือ PLAYER_LOSE)
			// → บันทึกและ return ทันที (ข้ามขั้นตอน AI และ final save)
			updatedMatch, err := s.combatRepo.UpdateMatch(match)
			if err != nil {
				return nil, err
			}
			return &PerformActionResponse{
				UpdatedMatch:    updatedMatch,
				PerformedAction: req,
			}, nil
		}

	default:
		return nil, apperrors.InvalidFormatError("unsupported action type", nil)
	}

	// ตรวจสอบ error จากการทำ action
	if actionErr != nil {
		return nil, actionErr
	}

	// ════════════════════════════════════════════════════════════════
	// ขั้นตอนที่ 4: AI PROCESSING - ให้ AI ทุกตัวเล่นต่อเนื่อง
	// ════════════════════════════════════════════════════════════════
	// หน้าที่: ประมวลผลเทิร์นของ AI ทุกตัวจนกว่าจะกลับมาเป็นเทิร์นผู้เล่น
	// Process:
	//   1. Loop ตรวจสอบ match.CurrentTurn
	//   2. ถ้า CurrentTurn.EnemyID != nil → เรียก processAITurn() → หมุนเทิร์น
	//   3. ถ้า CurrentTurn.CharacterID != nil → หยุด loop (เทิร์นผู้เล่น)
	//   4. ตรวจสอบจบเกมหลังจาก AI แต่ละตัวเล่นเสร็จ
	//   5. Safety: จำกัดสูงสุด 20 เทิร์นต่อเนื่อง (ป้องกัน infinite loop)
	//
	// ตัวอย่าง Flow:
	//   - Match มี 3 combatants: [Player, Enemy1, Enemy2]
	//   - Player กด END_TURN → CurrentTurn = Enemy1
	//   - processAllAITurns():
	//     • Enemy1 เล่น → หมุนเทิร์น → CurrentTurn = Enemy2
	//     • Enemy2 เล่น → หมุนเทิร์น → CurrentTurn = Player
	//     • เจอ Player → หยุด loop
	//
	// Output:  match ที่ AI เล่นเสร็จแล้ว (CurrentTurn กลับมาที่ผู้เล่น หรือเกมจบ)
	// Error:   error จาก AI turn processing (spell casting, resource issue)
	// ────────────────────────────────────────────────────────────────
	match, err = s.processAllAITurns(match)
	if err != nil {
		return nil, err
	}

	// ════════════════════════════════════════════════════════════════
	// ขั้นตอนที่ 5: PERSISTENCE - บันทึกผลลัพธ์และส่งคืน
	// ════════════════════════════════════════════════════════════════
	// หน้าที่: บันทึกสถานะล่าสุดของ match ลง database และส่งกลับหา client
	// Process:
	//   1. UpdateMatch() → บันทึก match, combatants, effects ทั้งหมด
	//   2. สร้าง PerformActionResponse ที่มี:
	//      - UpdatedMatch: match ที่อัปเดตแล้ว (รวมข้อมูล AI turns)
	//      - PerformedAction: request ที่ผู้เล่นส่งมา (เพื่อให้ client ตรวจสอบ)
	//   3. Return response กลับไป
	//
	// Output:  PerformActionResponse - ข้อมูลการต่อสู้หลังประมวลผลเสร็จ
	// Error:   error จาก database (connection, constraint violation, etc.)
	// ────────────────────────────────────────────────────────────────
	updatedMatch, err := s.combatRepo.UpdateMatch(match)
	if err != nil {
		return nil, err
	}
	return &PerformActionResponse{
		UpdatedMatch:    updatedMatch,
		PerformedAction: req,
	}, nil
}

// ==================== Spell Casting (New Refactored Version) ====================

// executeCastSpellV2 เป็น wrapper สำหรับเรียกใช้ระบบร่ายเวทใหม่
// แทนที่ executeCastSpell เดิม (ที่จะถูก deprecate)
func (s *combatService) executeCastSpellV2(
	caster *domain.Combatant,
	match *domain.CombatMatch,
	req PerformActionRequest,
) error {
	// Validate request
	if req.SpellID == nil || req.TargetID == nil {
		return apperrors.InvalidFormatError("spell_id and target_id are required", nil)
	}

	spellID := *req.SpellID
	targetUUIDStr := *req.TargetID

	// Parse target UUID
	targetUUID, err := uuid.FromString(targetUUIDStr)
	if err != nil {
		return apperrors.InvalidFormatError("invalid target_id format", nil)
	}

	// Get casting mode (default to INSTANT)
	castingMode := "INSTANT"
	if req.CastMode != "" {
		castingMode = req.CastMode
	}

	// เรียกใช้ระบบใหม่
	return s.ExecuteSpellCast(match, caster, targetUUID, spellID, castingMode)
}

// ==================== Cleanup & Management ====================

// CleanupStaleMatches ทำความสะอาด match ที่ค้างเกินเวลากำหนด
// ควรเรียกจาก Cron Job ทุก 5-10 นาที
func (s *combatService) CleanupStaleMatches(inactiveMinutes int) (int64, error) {
	s.appLogger.Info("🧹 Starting stale match cleanup",
		"inactive_minutes", inactiveMinutes,
	)

	affectedRows, err := s.combatRepo.AbortStaleMatches(inactiveMinutes)
	if err != nil {
		s.appLogger.Error("Failed to cleanup stale matches", err)
		return 0, err
	}

	if affectedRows > 0 {
		s.appLogger.Warn("Aborted stale matches",
			"count", affectedRows,
			"inactive_minutes", inactiveMinutes,
		)
	} else {
		s.appLogger.Debug("No stale matches found")
	}

	return affectedRows, nil
}

// AbortMatch ยกเลิก match เฉพาะ ID (ใช้เมื่อผู้เล่น forfeit/disconnect)
func (s *combatService) AbortMatch(matchID string, reason string) error {
	s.appLogger.Info("Aborting match",
		"match_id", matchID,
		"reason", reason,
	)

	match, err := s.combatRepo.AbortMatchByID(matchID, reason)
	if err != nil {
		s.appLogger.Error("Failed to abort match", err,
			"match_id", matchID,
		)
		return err
	}

	if match.Status == domain.MatchAborted {
		s.appLogger.Info("Match aborted successfully",
			"match_id", matchID,
			"status", match.Status,
		)
	}

	return nil
}

// GetPlayerActiveMatch ตรวจสอบว่าผู้เล่นมี match ที่กำลังเล่นอยู่หรือไม่
// ใช้ก่อนสร้าง match ใหม่ เพื่อป้องกันการเปิดหลายห้องพร้อมกัน
func (s *combatService) GetPlayerActiveMatch(characterID uint) (*domain.CombatMatch, error) {
	return s.combatRepo.FindPlayerActiveMatch(characterID)
}
