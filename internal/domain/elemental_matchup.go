// file: internal/domain/elemental_matchup.go
package domain

// ElementalMatchup เก็บข้อมูลตัวคูณดาเมจระหว่างธาตุ 2 ชนิด
type ElementalMatchup struct {
	AttackingElementID uint    `gorm:"primaryKey;comment:ID ของธาตุผู้โจมตี"`
	DefendingElementID uint    `gorm:"primaryKey;comment:ID ของธาตุผู้ป้องกัน"`
	Modifier           float64 `gorm:"not null;comment:ตัวคูณดาเมจ (เช่น 1.3, 0.7, 1.0)"`
}
