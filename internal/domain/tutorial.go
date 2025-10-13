// file: internal/domain/tutorial.go
package domain

// เราจะนิยาม "ชื่อเรียก" ของแต่ละขั้นตอน Tutorial ไว้ที่นี่
const (
	TutorialStepNone      = 0  // สำหรับผู้เล่นที่ Skip หรือยังไม่ได้เริ่ม
	TutorialStepStart     = 1  // เพิ่งสร้างตัวละครเสร็จ, รอเข้าห้องทดลอง
	TutorialStepCrafted   = 2  // คราฟต์ของครั้งแรกสำเร็จแล้ว, รอจัด Deck
	TutorialStepDeckBuilt = 3  // จัด Deck ครั้งแรกสำเร็จแล้ว, รอเข้าสู้
	TutorialStepCompleted = 99 // จบ Tutorial โดยสมบูรณ์
)
