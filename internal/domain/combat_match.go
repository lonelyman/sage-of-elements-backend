// file: internal/domain/combat_match.go
package domain

import (
	"time"

	"github.com/gofrs/uuid"
	"gorm.io/datatypes"
)

type MatchStatus string

const (
	MatchInProgress MatchStatus = "IN_PROGRESS"
	MatchFinished   MatchStatus = "FINISHED"
	MatchAborted    MatchStatus = "ABORTED"
)

// MatchType คือประเภทของการต่อสู้
type MatchType string

const (
	MatchTypeTraining MatchType = "TRAINING" // โหมดฝึกซ้อม - เลือกศัตรูเอง
	MatchTypeStory    MatchType = "STORY"    // โหมดเนื้อเรื่อง - ด่านที่กำหนด
	MatchTypePVP      MatchType = "PVP"      // ต่อสู้กับผู้เล่นอื่น
)

// MatchModifiers คือ struct สำหรับเก็บ "กฎพิเศษ" ในการต่อสู้
type MatchModifiers struct {
	DisableTimer      bool `json:"disable_timer"`
	InfiniteHP        bool `json:"infinite_hp"`
	InfiniteResources bool `json:"infinite_resources"`
}

// CombatMatch แทนการต่อสู้ 1 ครั้ง
type CombatMatch struct {
	ID          uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	MatchType   MatchType      `gorm:"type:varchar(20);not null;default:'TRAINING'" json:"matchType"` // TRAINING, STORY, PVP
	StageID     *uint          `gorm:"index;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"stageID,omitempty"`
	Stage       *Stage         `gorm:"foreignKey:StageID;references:ID" json:"stage,omitempty"`
	Status      MatchStatus    `gorm:"type:varchar(20);not null" json:"status"`
	Modifiers   datatypes.JSON `gorm:"type:jsonb" json:"modifiers"`
	TurnNumber  int            `gorm:"not null;default:1" json:"turnNumber"`
	CurrentTurn uuid.UUID      `gorm:"type:uuid" json:"currentTurn"`
	Combatants  []*Combatant   `gorm:"foreignKey:MatchID;constraint:OnDelete:CASCADE;" json:"combatants"`
	CreatedAt   time.Time      `json:"createdAt"`
	UpdatedAt   time.Time      `gorm:"index" json:"updatedAt"` // ✅ ใช้สำหรับตรวจจับ match ค้าง (GORM auto-update)
	FinishedAt  *time.Time     `json:"finishedAt"`
}
