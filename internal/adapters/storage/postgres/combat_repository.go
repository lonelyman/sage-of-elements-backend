// file: internal/adapters/storage/postgres/combat_repository.go
package postgres

import (
	"fmt"
	"sage-of-elements-backend/internal/domain"
	"sage-of-elements-backend/internal/modules/combat"

	"gorm.io/gorm"
)

type combatRepository struct {
	db *gorm.DB
}

// NewCombatRepository creates a new instance of combatRepository.
func NewCombatRepository(db *gorm.DB) combat.CombatRepository {
	return &combatRepository{db: db}
}

// CreateMatch บันทึกข้อมูล Match และ Combatant ทั้งหมดลง DB ใน Transaction เดียว
func (r *combatRepository) CreateMatch(match *domain.CombatMatch) (*domain.CombatMatch, error) {
	// GORM จะฉลาดพอที่จะสร้าง Match และ Combatant ที่อยู่ใน Slice พร้อมกัน
	if err := r.db.Create(match).Error; err != nil {
		fmt.Printf("err%v", err)
		return nil, err
	}
	// หลังจากสร้างเสร็จ เราต้อง Preload ข้อมูลกลับมาด้วย
	// เพื่อให้ Response ที่ส่งกลับไปมีข้อมูลของ Character และ Enemy ครบถ้วน
	// เราต้อง Preload แบบซ้อนกันลึกๆ (Nested Preloading)
	if err := r.db.
		Preload("Combatants.Character.PrimaryElement").
		Preload("Combatants.Enemy.Element").
		First(match, match.ID).Error; err != nil {
		return nil, err
	}

	return match, nil
}

// FindMatchByID ค้นหา Match ด้วย ID พร้อม Preload ข้อมูลที่จำเป็นทั้งหมด
func (r *combatRepository) FindMatchByID(matchID string) (*domain.CombatMatch, error) {
	var match domain.CombatMatch
	err := r.db.
		Preload("Combatants.Character.PrimaryElement"). // Preload ให้ลึกขึ้นเผื่อใช้
		Preload("Combatants.Enemy.Element").
		Preload("Combatants.Enemy.Abilities").       // <-- ⭐️ สั่งให้โหลดท่าโจมตีของศัตรูมาด้วย!
		Preload("Combatants.Enemy.AI.AbilityToUse"). // <-- ⭐️ สั่งให้โหลดกฎ AI และท่าที่ผูกกับกฎนั้นมาด้วย!
		Preload("Combatants.Deck").
		Where("id = ?", matchID).
		First(&match).Error
	return &match, err
}

// UpdateMatch บันทึกสถานะล่าสุดของ Match และ Combatant ทุกตัว
func (r *combatRepository) UpdateMatch(match *domain.CombatMatch) (*domain.CombatMatch, error) {
	err := r.db.Transaction(func(tx *gorm.DB) error {
		// 1. บันทึก Combatant ทุกตัวก่อน
		for i := range match.Combatants {
			if err := tx.Save(match.Combatants[i]).Error; err != nil {
				return err
			}
		}

		// 2. บันทึก Match หลักทีหลัง
		if err := tx.Save(match).Error; err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	// ดึงข้อมูลล่าสุดทั้งหมดกลับมาอีกครั้ง (เหมือนเดิม)
	return r.FindMatchByID(match.ID.String())
}

// ==================== Cleanup Methods ====================

// FindStaleMatches หา match ที่ไม่มีความเคลื่อนไหวเกินเวลากำหนด (นาที)
// ใช้สำหรับ monitoring หรือ preview ก่อน abort
func (r *combatRepository) FindStaleMatches(inactiveMinutes int) ([]*domain.CombatMatch, error) {
	var matches []*domain.CombatMatch
	err := r.db.
		Where("status = ?", domain.MatchInProgress).
		Where("updated_at < NOW() - INTERVAL '? minutes'", inactiveMinutes).
		Order("updated_at ASC").
		Find(&matches).Error
	return matches, err
}

// AbortStaleMatches abort match ทั้งหมดที่ไม่มีความเคลื่อนไหวเกินเวลากำหนด
// return จำนวน match ที่ถูก abort
func (r *combatRepository) AbortStaleMatches(inactiveMinutes int) (int64, error) {
	now := gorm.Expr("NOW()")
	result := r.db.Model(&domain.CombatMatch{}).
		Where("status = ?", domain.MatchInProgress).
		Where("updated_at < NOW() - INTERVAL '? minutes'", inactiveMinutes).
		Updates(map[string]interface{}{
			"status":      domain.MatchAborted,
			"finished_at": now,
			"updated_at":  now,
		})
	return result.RowsAffected, result.Error
}

// FindPlayerActiveMatch หา match ที่ผู้เล่นกำลังเล่นอยู่ (IN_PROGRESS)
// ใช้สำหรับตรวจสอบก่อนสร้าง match ใหม่ (ป้องกันการเปิดหลายห้องพร้อมกัน)
func (r *combatRepository) FindPlayerActiveMatch(characterID uint) (*domain.CombatMatch, error) {
	var match domain.CombatMatch
	err := r.db.
		Joins("JOIN combatants ON combatants.match_id = combat_matches.id").
		Where("combat_matches.status = ?", domain.MatchInProgress).
		Where("combatants.character_id = ?", characterID).
		Preload("Combatants.Character.PrimaryElement").
		Preload("Combatants.Enemy.Element").
		First(&match).Error

	if err == gorm.ErrRecordNotFound {
		return nil, nil // ไม่มี active match = OK
	}
	return &match, err
}

// AbortMatchByID abort match เฉพาะ ID (ใช้เมื่อผู้เล่น disconnect หรือ forfeit)
func (r *combatRepository) AbortMatchByID(matchID string, reason string) (*domain.CombatMatch, error) {
	var match domain.CombatMatch

	// หา match ก่อน
	if err := r.db.Where("id = ?", matchID).First(&match).Error; err != nil {
		return nil, err
	}

	// ตรวจสอบว่ายังเล่นอยู่หรือเปล่า
	if match.Status != domain.MatchInProgress {
		return &match, nil // จบแล้ว ไม่ต้องทำอะไร
	}

	// Abort
	now := gorm.Expr("NOW()")
	if err := r.db.Model(&match).Updates(map[string]interface{}{
		"status":      domain.MatchAborted,
		"finished_at": now,
		"updated_at":  now,
	}).Error; err != nil {
		return nil, err
	}

	return r.FindMatchByID(matchID)
}
