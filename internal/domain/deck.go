package domain

type Deck struct {
	ID           uint        `gorm:"primaryKey" json:"id"`
	CharacterID  uint        `gorm:"not null;index" json:"-"`
	Name         string      `gorm:"size:100;not null" json:"name"`
	IsActive     bool        `gorm:"not null;default:false" json:"isActive"`
	DisplayOrder int         `gorm:"not null;default:0" json:"displayOrder"`
	Slots        []*DeckSlot `gorm:"foreignKey:DeckID;constraint:OnDelete:CASCADE;" json:"slots"`
}
