// file: internal/modules/combat/service.go
package combat

import (
	"encoding/json"
	"fmt"
	"sage-of-elements-backend/internal/domain"
	"sage-of-elements-backend/internal/modules/character"
	"sage-of-elements-backend/internal/modules/enemy"
	"sage-of-elements-backend/internal/modules/game_data"
	"sage-of-elements-backend/internal/modules/pve"
	"sage-of-elements-backend/pkg/apperrors"
	"sage-of-elements-backend/pkg/applogger"
	"sort"
	"strconv"

	"github.com/gofrs/uuid"
	"gorm.io/datatypes"
)

// --- Interface (เหมือนเดิม) ---
type CombatService interface {
	CreateMatch(playerID uint, req CreateMatchRequest) (*domain.CombatMatch, error)
	PerformAction(playerID uint, matchID string, req PerformActionRequest) (*domain.CombatMatch, error)
}

// --- Implementation ---
type combatService struct {
	appLogger     applogger.Logger
	combatRepo    CombatRepository
	characterRepo character.CharacterRepository
	enemyRepo     enemy.EnemyRepository
	pveRepo       pve.PveRepository
	gameDataRepo  game_data.GameDataRepository
}

func NewCombatService(
	appLogger applogger.Logger,
	combatRepo CombatRepository,
	characterRepo character.CharacterRepository,
	enemyRepo enemy.EnemyRepository,
	pveRepo pve.PveRepository,
	gameDataRepo game_data.GameDataRepository,
) CombatService {
	return &combatService{
		appLogger:     appLogger,
		combatRepo:    combatRepo,
		characterRepo: characterRepo,
		enemyRepo:     enemyRepo,
		pveRepo:       pveRepo,
		gameDataRepo:  gameDataRepo,
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
		CharacterID: &playerChar.ID,
		Initiative:  50 + playerChar.TalentG,
		CurrentHP:   900 + (playerChar.TalentS * 30),
		CurrentMP:   playerChar.CurrentMP,
		CurrentAP:   0,
	}

	var combatants []*domain.Combatant
	combatants = append(combatants, playerCombatant)

	// 3. สร้าง Combatant ของ "ศัตรู" (ตอนนี้รองรับแค่ Training Mode)
	if req.MatchType == "TRAINING" {

		fmt.Println("req.TrainingEnemies", req.TrainingEnemies)
		for _, enemyInfo := range req.TrainingEnemies {
			s.appLogger.Info("Attempting to find enemy in repo", "enemy_id_to_find", enemyInfo.EnemyID)

			enemyData, err := s.enemyRepo.FindByID(enemyInfo.EnemyID)
			s.appLogger.Dump(enemyData)
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

	// 4. ตัดสินลำดับเทิร์น และแจก AP เริ่มต้นในครั้งเดียว!

	// 4.1 ดึงกฎของเกมมาก่อน
	apPerTurnStr, _ := s.gameDataRepo.GetGameConfigValue("COMBAT_AP_PER_TURN")
	apPerTurn, _ := strconv.Atoi(apPerTurnStr)

	// 4.2 หา "ตัว" ผู้เล่นที่จะได้เริ่มก่อน (เริ่มต้นโดยให้เป็นผู้เล่น)
	var firstTurnCombatant *domain.Combatant = playerCombatant
	for _, c := range combatants {
		if c.Initiative > firstTurnCombatant.Initiative {
			firstTurnCombatant = c // ถ้าเจอคนเร็วกว่า ก็เปลี่ยน "ตัว" ที่จะเริ่มก่อนเลย
		}
	}

	// 4.3 มอบ AP เริ่มต้นให้กับคนที่ได้เล่นก่อน!
	firstTurnCombatant.CurrentAP = apPerTurn
	s.appLogger.Info("Granting starting AP", "combatant_id", firstTurnCombatant.ID, "ap", firstTurnCombatant.CurrentAP)
	// -----------------------------------------------------------

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
		CurrentTurn: firstTurnCombatant.ID,
		Combatants:  combatants,
	}
	// --------------------------

	// 6. บันทึกลง Database (เหมือนเดิม)
	return s.combatRepo.CreateMatch(newMatch)
}

func (s *combatService) PerformAction(playerID uint, matchID string, req PerformActionRequest) (*domain.CombatMatch, error) {
	// 1. โหลดสถานะ Match ปัจจุบันขึ้นมาจาก DB
	match, err := s.combatRepo.FindMatchByID(matchID)
	if err != nil {
		return nil, apperrors.NotFoundError("match not found")
	}

	// 2. ตรวจสอบเงื่อนไขพื้นฐาน
	if match.Status != domain.MatchInProgress {
		return nil, apperrors.New(400, "MATCH_FINISHED", "this match is already finished")
	}

	// 3. ค้นหา "ตัวตน" ของผู้เล่นในสนามรบ (Player's Combatant)
	var playerCombatant *domain.Combatant
	for _, c := range match.Combatants {
		// เช็คว่า Combatant คนนี้คือผู้เล่นที่ส่งคำขอมาหรือไม่
		if c.Character != nil && c.Character.PlayerID == playerID {
			playerCombatant = c
			break
		}
	}
	if playerCombatant == nil {
		return nil, apperrors.PermissionDeniedError("you are not part of this match")
	}

	// 4. ตรวจสอบว่า "ใช่เทิร์นของคุณหรือไม่?"
	if match.CurrentTurn != playerCombatant.ID {
		return nil, apperrors.New(400, "NOT_YOUR_TURN", "it's not your turn to perform an action")
	}

	// 5. ประมวลผล Action ตามประเภท
	switch req.ActionType {
	case "DRAW_ELEMENT":
		// --- ✨⭐️ เพิ่ม Logic ใหม่ของเราที่นี่! ⭐️✨ ---
		if req.ElementID == nil {
			return nil, apperrors.InvalidFormatError("element_id is required for DRAW_ELEMENT action", nil)
		}
		elementID := *req.ElementID

		// กฎ: การจั่ว T0 ใช้ 1 AP (ตอนนี้ Hardcode ไปก่อน)
		apCost := 1
		if playerCombatant.CurrentAP < apCost {
			return nil, apperrors.New(422, "INSUFFICIENT_AP", "not enough AP to draw element")
		}

		// กฎ: อนุญาตให้จั่วได้แค่ธาตุ T0 (ID 1-4) เท่านั้น
		if elementID < 1 || elementID > 4 {
			return nil, apperrors.New(400, "INVALID_ACTION", "can only draw T0 elements with this action")
		}

		// ลงมือ "จั่ว"!
		var hand []uint
		if playerCombatant.Hand != nil {
			json.Unmarshal(playerCombatant.Hand, &hand)
		}
		hand = append(hand, elementID)
		newHandJSON, _ := json.Marshal(hand)

		playerCombatant.Hand = newHandJSON
		playerCombatant.CurrentAP -= apCost

		s.appLogger.Info("Player drew an element", "element_id", elementID, "ap_remaining", playerCombatant.CurrentAP)

	case "END_TURN":
		s.appLogger.Info("Player ended their turn.", "match_id", matchID)
		// จบเทิร์นของผู้เล่น และส่งต่อให้คนถัดไป
		match = s.endTurn(match)
		// เตรียมความพร้อมสำหรับเทิร์นของ "คนถัดไป" (ซึ่งอาจจะเป็น AI)
		match, err = s.startNewTurn(match)
		if err != nil {
			return nil, err
		}

	case "CAST_SPELL":
		// --- ✨⭐️ นี่คือ Logic ใหม่ทั้งหมดของเรา! ⭐️✨ ---
		s.appLogger.Info("Player attempts to cast a spell.", "match_id", matchID)

		// --- [ด่านที่ 1: การตรวจสอบ (The Gatekeeper)] ---
		// 1.1 ตรวจสอบ Input
		if req.SpellID == nil || req.TargetID == nil {
			return nil, apperrors.InvalidFormatError("spell_id and target_id are required for CAST_SPELL", nil)
		}
		spellID := *req.SpellID
		targetUUID, err := uuid.FromString(*req.TargetID)
		if err != nil {
			return nil, apperrors.InvalidFormatError("invalid target_id format", nil)
		}

		// 1.2 หาข้อมูลเวทและเป้าหมาย
		spell, err := s.gameDataRepo.FindSpellByID(spellID)
		if err != nil || spell == nil {
			return nil, apperrors.NotFoundError("spell not found")
		}
		target := s.findCombatantByID(match, targetUUID)
		if target == nil {
			return nil, apperrors.NotFoundError("target not found in this match")
		}

		// 1.3 ตรวจสอบทรัพยากร (AP & MP)
		if playerCombatant.CurrentAP < spell.APCost {
			return nil, apperrors.New(422, "INSUFFICIENT_AP", "not enough AP to cast spell")
		}
		if playerCombatant.CurrentMP < spell.MPCost {
			return nil, apperrors.New(422, "INSUFFICIENT_MP", "not enough MP to cast spell")
		}

		// 1.4 ตรวจสอบ "วัตถุดิบ" (ธาตุในมือ)
		// (Logic ง่ายๆ ตอนนี้: เวท T0 ต้องการธาตุของตัวเอง 1 ใบ)
		requiredElementID := spell.ElementID
		var hand []uint
		if playerCombatant.Hand != nil {
			json.Unmarshal(playerCombatant.Hand, &hand)
		}
		elementFound := false
		elementIndex := -1
		for i, elementInHand := range hand {
			if elementInHand == requiredElementID {
				elementFound = true
				elementIndex = i
				break
			}
		}
		if !elementFound {
			return nil, apperrors.New(422, "INSUFFICIENT_ELEMENTS", "required element not in hand")
		}

		// --- [ด่านที่ 2: การคำนวณ (The Calculator) - เวอร์ชันง่าย] ---
		// TODO: เราจะกลับมาใส่ "สูตรคำนวณดาเมจเต็ม" ที่นี่ในอนาคต!
		var totalDamage float64 = 0.0
		for _, effect := range spell.Effects {
			if effect.Effect.Name == "DAMAGE" { // ใช้ Name เพื่อความชัดเจน
				totalDamage += effect.BaseValue
			}
		}

		// --- [ด่านที่ 3: การลงมือ (The Executioner)] ---
		// 3.1 หัก HP เป้าหมาย
		target.CurrentHP -= int(totalDamage)
		s.appLogger.Info("Player cast spell", "spell", spell.Name, "target", target.ID, "damage", totalDamage)

		// 3.2 หักทรัพยากรผู้ร่าย
		playerCombatant.CurrentAP -= spell.APCost
		playerCombatant.CurrentMP -= spell.MPCost

		// "ใช้" ธาตุออกจากมือ
		hand = append(hand[:elementIndex], hand[elementIndex+1:]...)
		newHandJSON, _ := json.Marshal(hand)
		playerCombatant.Hand = newHandJSON

		s.appLogger.Info("Player resources updated", "ap", playerCombatant.CurrentAP, "mp", playerCombatant.CurrentMP)

	default:
		return nil, apperrors.InvalidFormatError("unsupported action type", nil)
	}

	// --- ✨⭐️ ส่วนที่เพิ่มเข้ามา: "ปลุกชีพ AI"! ⭐️✨ ---
	// 6. ตรวจสอบว่าเทิร์นต่อไปเป็นของ AI หรือไม่
	nextCombatant := s.findCombatantByID(match, match.CurrentTurn)
	if nextCombatant != nil && nextCombatant.EnemyID != nil {
		s.appLogger.Info("AI turn begins.", "ai_combatant_id", nextCombatant.ID)
		// ถ้าใช่... ให้ประมวลผลเทิร์นของ AI ทันที!
		match, err = s.processAITurn(match, nextCombatant)
		if err != nil {
			return nil, err // ถ้า AI ทำงานผิดพลาด
		}
	}
	// ----------------------------------------------------

	// 7. บันทึกสถานะ Match ที่อัปเดตแล้วกลับลง DB
	return s.combatRepo.UpdateMatch(match)
}

// endTurn คือฟังก์ชันสำหรับ "จบเทิร์น" และหาคนเล่นถัดไป
func (s *combatService) endTurn(match *domain.CombatMatch) *domain.CombatMatch {
	var currentTurnIndex int
	for i, c := range match.Combatants {
		if c.ID == match.CurrentTurn {
			currentTurnIndex = i
			break
		}
	}
	nextTurnIndex := (currentTurnIndex + 1) % len(match.Combatants)
	nextCombatant := match.Combatants[nextTurnIndex]

	match.CurrentTurn = nextCombatant.ID
	if nextTurnIndex == 0 {
		match.TurnNumber++
	}
	return match
}

// findCombatantByID ค้นหา Combatant ใน Match ด้วย ID
func (s *combatService) findCombatantByID(match *domain.CombatMatch, id uuid.UUID) *domain.Combatant {
	for _, c := range match.Combatants {
		if c.ID == id {
			return c
		}
	}
	return nil
}

// processAITurn คือ "สมอง" ของศัตรู
func (s *combatService) processAITurn(match *domain.CombatMatch, aiCombatant *domain.Combatant) (*domain.CombatMatch, error) {
	// 1. ดึงข้อมูล AI Rules ของศัตรูตัวนี้
	// (เราต้อง Preload "AI" มาตั้งแต่ตอน FindByID ใน enemyRepo)
	aiRules := aiCombatant.Enemy.AI
	// เรียงกฎตาม Priority (สำคัญน้อยไปมาก -> สำคัญมาก่อน)
	sort.Slice(aiRules, func(i, j int) bool {
		return aiRules[i].Priority < aiRules[j].Priority
	})

	// 2. หา "เป้าหมาย" (ตอนนี้มีแค่ผู้เล่นคนเดียว)
	var target *domain.Combatant
	for _, c := range match.Combatants {
		if c.CharacterID != nil {
			target = c
			break
		}
	}
	if target == nil {
		return nil, apperrors.SystemError("player target not found in match")
	}

	// 3. วน Loop เช็ค "กฎ" แต่ละข้อ เพื่อหา Action ที่จะทำ
	var actionToPerform *domain.EnemyAbility
	for _, rule := range aiRules {
		// (ตอนนี้เรารองรับแค่ Condition: ALWAYS และ TURN_IS ก่อนนะ)
		conditionMet := false
		switch rule.Condition {
		case domain.AIConditionAlways:
			conditionMet = true
		case domain.AIConditionTurnIs:
			if float64(match.TurnNumber) == rule.ConditionValue {
				conditionMet = true
			}
		}

		if conditionMet {
			// เจอกฎที่ตรงเงื่อนไขแล้ว!
			actionToPerform = rule.AbilityToUse
			s.appLogger.Info("AI decided to use ability", "ability_name", actionToPerform.Name, "rule_priority", rule.Priority)
			break // หยุดหาทันที
		}
	}

	// 4. ถ้ามี Action ที่จะทำ... ก็ลงมือเลย!
	if actionToPerform != nil {
		aiCombatant.CurrentAP -= actionToPerform.APCost
		// อ่านค่าดาเมจจาก JSON
		var effects []map[string]interface{}
		json.Unmarshal(actionToPerform.EffectsJSON, &effects)

		for _, effect := range effects {
			if effect["effect_id"] == float64(1) { // 1 = DAMAGE
				damage := int(effect["value"].(float64))
				target.CurrentHP -= damage
				s.appLogger.Info("AI dealt damage to player", "damage", damage, "player_hp_remaining", target.CurrentHP)
			}
		}
	}

	// 5. จบเทิร์นของตัวเองโดยอัตโนมัติ!
	s.appLogger.Info("AI ended its turn.", "match_id", match.ID)
	match = s.endTurn(match)

	// เตรียมความพร้อมสำหรับเทิร์นของ "คนถัดไป" (ซึ่งก็คือผู้เล่น)
	match, err := s.startNewTurn(match)
	if err != nil {
		return nil, err
	}

	return match, nil
}

// startNewTurn คือฟังก์ชันที่เตรียมความพร้อมเมื่อเทิร์นใหม่เริ่มต้น (เช่น เพิ่ม AP)
func (s *combatService) startNewTurn(match *domain.CombatMatch) (*domain.CombatMatch, error) {
	// 1. หาว่าเทิร์นนี้เป็นของใคร
	currentCombatant := s.findCombatantByID(match, match.CurrentTurn)
	if currentCombatant == nil {
		return nil, apperrors.SystemError("failed to find current combatant")
	}

	apPerTurnStr, _ := s.gameDataRepo.GetGameConfigValue("COMBAT_AP_PER_TURN")
	apPerTurn, _ := strconv.Atoi(apPerTurnStr)

	baseApCapStr, _ := s.gameDataRepo.GetGameConfigValue("COMBAT_BASE_AP_CAP")
	baseApCap, _ := strconv.Atoi(baseApCapStr)
	// 3. (อนาคต) ตรวจสอบบัฟที่ส่งผลต่อ AP
	apCapBonus := 0 // TODO: Check for buffs
	currentMaxAP := baseApCap + apCapBonus

	// 4. เพิ่ม AP ให้กับเจ้าของเทิร์น
	currentCombatant.CurrentAP += apPerTurn

	// 5. ตรวจสอบกับขีดจำกัดสูงสุด
	if currentCombatant.CurrentAP > currentMaxAP {
		currentCombatant.CurrentAP = currentMaxAP
	}

	s.appLogger.Info("New turn started", "combatant_id", currentCombatant.ID, "new_ap", currentCombatant.CurrentAP)
	return match, nil
}
