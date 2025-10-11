package domain

// Recipe เก็บสูตรการสร้างธาตุพันธะ
type Recipe struct {
	ID              uint                `gorm:"primaryKey;comment:ID เฉพาะของสูตร (PK)"`
	OutputElementID uint                `gorm:"not null;comment:ID ของธาตุที่สร้างได้"`
	BaseMPCost      int                 `gorm:"not null;comment:ต้นทุน MP พื้นฐานในการหลอมรวม"`
	OutputElement   *Element            `gorm:"foreignKey:OutputElementID;references:ID"`
	Ingredients     []*RecipeIngredient `gorm:"foreignKey:RecipeID"`
}
