package domain

type DeckSlot struct {
	ID        uint `gorm:"primaryKey" json:"id"`
	DeckID    uint `gorm:"not null" json:"-"`
	SlotNum   int  `gorm:"not null" json:"slotNum"`
	ElementID uint `gorm:"not null" json:"elementId"`
}
