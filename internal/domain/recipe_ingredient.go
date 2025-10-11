package domain

// RecipeIngredient เก็บส่วนประกอบและปริมาณของแต่ละสูตร
type RecipeIngredient struct {
	ID             uint     `gorm:"primaryKey;comment:ID ของแถวข้อมูล"`
	RecipeID       uint     `gorm:"not null;comment:ID ของสูตรที่เป็นเจ้าของ (FK to recipes)"`
	InputElementID uint     `gorm:"not null;comment:ID ของธาตุที่เป็นส่วนประกอบ"`
	Quantity       int      `gorm:"not null;default:1;comment:ปริมาณที่ต้องใช้"`
	InputElement   *Element `gorm:"foreignKey:InputElementID;references:ID"`
}
