// file: internal/modules/combat/spell_resolver.go
package combat

import (
	"fmt"
	"sage-of-elements-backend/internal/domain"
	"sage-of-elements-backend/pkg/apperrors"
)

// ResolveSpell หาเวทมนตร์ที่เหมาะสม โดยมี Fallback Algorithm ถ้าหาไม่เจอ
// Parameters:
//   - elementID: ธาตุที่ต้องการหาเวท
//   - masteryID: ศาสตร์ที่ต้องการใช้
//   - casterMainElementID: ธาตุหลักของผู้ใช้เวท (สำหรับ Fallback Step 2A)
//
// Returns: spell, error
func (s *combatService) ResolveSpell(elementID uint, masteryID uint, casterMainElementID uint) (*domain.Spell, error) {
	// ลองหาตรงๆ ก่อน
	spell, err := s.gameDataRepo.FindSpellByElementAndMastery(elementID, masteryID)
	if err == nil && spell != nil {
		s.appLogger.Info("Spell found directly", "element", elementID, "mastery", masteryID, "spell", spell.Name)
		return spell, nil // เจอ! จบเลย
	}

	// ถ้าไม่เจอ -> เริ่ม Fallback Algorithm
	//s.appLogger.Warn("Spell not found, starting Fallback Algorithm", "element", elementID, "mastery", masteryID)
	return s.findFallbackSpell(elementID, masteryID, casterMainElementID)
}

// findFallbackSpell ใช้อัลกอริทึม Fallback ตามเงื่อนไขที่กำหนด
func (s *combatService) findFallbackSpell(failedElementID uint, failedMasteryID uint, casterMainElementID uint) (*domain.Spell, error) {
	// --- ขั้นที่ 1: ตรวจสอบเสียงข้างมากในสูตรผสม ---
	recipe, err := s.gameDataRepo.FindRecipeByOutputElementID(failedElementID)

	// ถ้าเป็นธาตุผสม (Tier 1+) และมีสูตร
	if err == nil && recipe != nil && len(recipe.Ingredients) > 0 {
		s.appLogger.Info("Fallback Step 1: Found recipe for element", "element", failedElementID, "ingredients", len(recipe.Ingredients))

		majorityElementID, hasMajority := s.calculateMajorityElement(recipe)

		if hasMajority {
			// **** กรณี 1.1: มีเสียงข้างมาก ****
			s.appLogger.Info("Fallback Step 1.1: Majority element found", "majorityElement", majorityElementID)
			fallbackSpell, err := s.gameDataRepo.FindSpellByElementAndMastery(majorityElementID, failedMasteryID)
			if err == nil && fallbackSpell != nil {
				s.appLogger.Info("Fallback success from majority element", "spell", fallbackSpell.Name)
				return fallbackSpell, nil
			}
			s.appLogger.Warn("Majority element also lacks the spell, proceeding to Step 2")
			// ถ้าธาตุเสียงข้างมากก็ไม่มี -> ไปต่อ Step 2
		} else {
			// **** กรณี 1.2: ไม่มีเสียงข้างมาก (เสมอ) ****
			s.appLogger.Info("Fallback Step 1.2: No majority found, proceeding to Step 2")
		}
	} else {
		// ถ้าเป็น T0 (ไม่มีสูตร) หรือหาสูตรไม่เจอ
		s.appLogger.Info("Fallback Step 1: No recipe found (likely T0 element), proceeding to Step 2")
	}

	// --- ขั้นที่ 2: ตรวจสอบธาตุหลักของผู้ใช้ (หรือสู้กันภายใน) ---
	recipeIngredients := s.getRecipeIngredientIDs(recipe)
	isCasterIngredient := s.isElementInList(casterMainElementID, recipeIngredients)

	if isCasterIngredient {
		// **** กรณี 2A: ผู้ใช้เป็น 1 ในธาตุแม่ ****
		s.appLogger.Info("Fallback Step 2A: Caster is an ingredient", "casterElement", casterMainElementID)
		fallbackSpell, err := s.gameDataRepo.FindSpellByElementAndMastery(casterMainElementID, failedMasteryID)
		if err == nil && fallbackSpell != nil {
			s.appLogger.Info("Fallback success from caster's primary element", "spell", fallbackSpell.Name)
			return fallbackSpell, nil
		}
		s.appLogger.Warn("Caster's primary element lacks the spell, proceeding to Step 2B")
		// ถ้าธาตุหลักผู้ใช้ก็ไม่มี -> ไปต่อ 2B
	}

	// **** กรณี 2B: ผู้ใช้เป็น 'คนนอก' หรือ ธาตุหลักตัวเองไม่มีเวทนั้น ****
	s.appLogger.Info("Fallback Step 2B: Caster is outsider or primary lacks spell, performing internal fight")

	if len(recipeIngredients) > 0 {
		winnerElementID, isTie := s.determineInternalWinner(recipeIngredients)

		if !isTie {
			// **** กรณี 2B.1: มีผู้ชนะจากการสู้กัน ****
			s.appLogger.Info("Fallback Step 2B.1: Internal fight winner", "winner", winnerElementID)
			fallbackSpell, err := s.gameDataRepo.FindSpellByElementAndMastery(winnerElementID, failedMasteryID)
			if err == nil && fallbackSpell != nil {
				s.appLogger.Info("Fallback success from internal fight winner", "spell", fallbackSpell.Name)
				return fallbackSpell, nil
			}
			s.appLogger.Warn("Internal fight winner lacks the spell, proceeding to final fallback")
		}
	}

	// **** กรณี 2B.2: สู้กันแล้ว 'เสมอ' (Tie) หรือ หาเวทจากผู้ชนะไม่ได้ ****
	s.appLogger.Info("Fallback Step 2B.2: Tie or winner lacks spell, using caster advantage logic")

	if len(recipeIngredients) > 0 {
		// ให้ caster สู้กับธาตุแต่ละตัว แล้วเลือกธาตุที่ชนะ caster มากที่สุด (คะแนนสูงที่สุด)
		strongestElement, highestScore := s.findStrongestAgainstCaster(casterMainElementID, recipeIngredients)
		s.appLogger.Info("Strongest element against caster selected", "casterElement", casterMainElementID, "strongestElement", strongestElement, "score", highestScore)

		fallbackSpell, err := s.gameDataRepo.FindSpellByElementAndMastery(strongestElement, failedMasteryID)
		if err == nil && fallbackSpell != nil {
			s.appLogger.Info("Fallback success from strongest element", "spell", fallbackSpell.Name)
			return fallbackSpell, nil
		}
	}

	// ถ้าถึงขั้นนี้แล้วยังไม่เจอ = ไม่มีเวทนี้จริงๆ
	errMsg := fmt.Sprintf("ไม่พบเวทมนตร์ที่สามารถใช้งานได้สำหรับธาตุ %d และศาสตร์ %d", failedElementID, failedMasteryID)
	err = apperrors.New(404, "SPELL_NOT_FOUND", errMsg)
	s.appLogger.Error("No fallback spell found", err, "element", failedElementID, "mastery", failedMasteryID)
	return nil, err
}

// calculateMajorityElement หาเสียงข้างมากจาก recipe
// Returns: (majorityElementID, hasMajority)
func (s *combatService) calculateMajorityElement(recipe *domain.Recipe) (uint, bool) {
	if recipe == nil || len(recipe.Ingredients) == 0 {
		return 0, false
	}

	// นับจำนวนของแต่ละธาตุ
	elementCount := make(map[uint]int)
	for _, ing := range recipe.Ingredients {
		elementCount[ing.InputElementID] += int(ing.Quantity)
	}

	// หาธาตุที่มีจำนวนมากที่สุด
	var maxElement uint
	var maxCount int
	totalCount := 0

	for elemID, count := range elementCount {
		totalCount += count
		if count > maxCount {
			maxCount = count
			maxElement = elemID
		}
	}

	// ตรวจสอบว่าเป็นเสียงข้างมากจริงๆ (มากกว่าครึ่ง)
	hasMajority := maxCount > totalCount/2

	s.appLogger.Info("Majority calculation", "maxElement", maxElement, "maxCount", maxCount, "totalCount", totalCount, "hasMajority", hasMajority)
	return maxElement, hasMajority
}

// getRecipeIngredientIDs แปลง recipe ingredients เป็น array ของ element IDs
func (s *combatService) getRecipeIngredientIDs(recipe *domain.Recipe) []uint {
	if recipe == nil || len(recipe.Ingredients) == 0 {
		return []uint{}
	}

	// ใช้ map เพื่อเก็บธาตุที่ไม่ซ้ำ
	elementSet := make(map[uint]bool)
	for _, ing := range recipe.Ingredients {
		elementSet[ing.InputElementID] = true
	}

	// แปลงเป็น slice
	result := make([]uint, 0, len(elementSet))
	for elemID := range elementSet {
		result = append(result, elemID)
	}

	return result
}

// isElementInList ตรวจสอบว่า elementID อยู่ใน list หรือไม่
func (s *combatService) isElementInList(elementID uint, list []uint) bool {
	for _, id := range list {
		if id == elementID {
			return true
		}
	}
	return false
}

// determineInternalWinner ใช้ตารางแพ้ทางเพื่อหาผู้ชนะจากการสู้กันภายในของธาตุต่างๆ
// Returns: (winnerElementID, isTie)
func (s *combatService) determineInternalWinner(elementIDs []uint) (uint, bool) {
	if len(elementIDs) == 0 {
		return 0, true
	}
	if len(elementIDs) == 1 {
		return elementIDs[0], false
	}

	// นับคะแนนชนะของแต่ละธาตุ
	scores := make(map[uint]int)
	for _, elemID := range elementIDs {
		scores[elemID] = 0
	}

	// ทำการแข่งขันแบบ Round-robin (ทุกคนชนทุกคน)
	for i, attackerID := range elementIDs {
		for j, defenderID := range elementIDs {
			if i == j {
				continue // ไม่ชนตัวเอง
			}

			// ดึงค่า modifier จาก matchup table
			modifierStr, err := s.gameDataRepo.GetMatchupModifier(attackerID, defenderID)
			if err != nil {
				s.appLogger.Warn("Error getting matchup modifier", "attacker", attackerID, "defender", defenderID, "error", err)
				continue
			}

			// ถ้า modifier > 1.0 = attacker ได้เปรียบ = ได้ 1 คะแนน
			if modifierStr > "1.0" {
				scores[attackerID]++
			}
		}
	}

	// หาธาตุที่มีคะแนนสูงสุด
	var maxScore int
	var winner uint
	var tieCount int

	for elemID, score := range scores {
		if score > maxScore {
			maxScore = score
			winner = elemID
			tieCount = 1
		} else if score == maxScore {
			tieCount++
		}
	}

	isTie := tieCount > 1
	s.appLogger.Info("Internal fight result", "winner", winner, "maxScore", maxScore, "isTie", isTie, "scores", scores)

	return winner, isTie
}

// findStrongestAgainstCaster หาธาตุที่ชนะ caster มากที่สุด (มีคะแนนสูงที่สุดเมื่อสู้กับ caster)
// ใช้เมื่อต้องการเลือกธาตุที่แข็งแกร่งกว่า caster
// Returns: (strongestElementID, highestScore)
func (s *combatService) findStrongestAgainstCaster(casterElementID uint, candidateElements []uint) (uint, int) {
	if len(candidateElements) == 0 {
		return 0, 0
	}

	// คำนวณคะแนนของแต่ละธาตุเมื่อสู้กับ caster
	// ถ้า modifier > 1.0 = candidate มีเปรียบ caster = คะแนนสูง
	// ถ้า modifier < 1.0 = candidate เสียเปรียบ caster = คะแนนต่ำ
	scores := make(map[uint]int)

	for _, candidateID := range candidateElements {
		scores[candidateID] = 0

		// ดูว่า candidate vs caster เป็นยังไง
		modifierStr, err := s.gameDataRepo.GetMatchupModifier(candidateID, casterElementID)
		if err != nil {
			s.appLogger.Warn("Error getting matchup modifier", "candidate", candidateID, "caster", casterElementID, "error", err)
			continue
		}

		// ถ้า modifier > 1.0 = candidate ได้เปรียบ = คะแนน +1 (นี่คือที่เราต้องการ)
		// ถ้า modifier < 1.0 = candidate เสียเปรียบ = คะแนน -1
		// ถ้า modifier = 1.0 = เสมอ = คะแนน 0
		if modifierStr > "1.0" {
			scores[candidateID] = 1 // candidate แข็งแกร่งกว่า caster (นี่คือที่เราต้องการ)
		} else if modifierStr < "1.0" {
			scores[candidateID] = -1 // candidate อ่อนแอกว่า caster
		}
	}

	// หาธาตุที่มีคะแนนสูงที่สุด (ชนะ caster มากที่สุด)
	var strongestElement uint
	highestScore := -999

	for elemID, score := range scores {
		if strongestElement == 0 || score > highestScore {
			highestScore = score
			strongestElement = elemID
		}
	}

	// ถ้าไม่มีธาตุที่ชนะ caster ให้เลือกตัวแรก
	if strongestElement == 0 && len(candidateElements) > 0 {
		strongestElement = candidateElements[0]
		highestScore = 0
	}

	s.appLogger.Info("Strongest element against caster found", "caster", casterElementID, "strongest", strongestElement, "score", highestScore, "allScores", scores)

	return strongestElement, highestScore
}
