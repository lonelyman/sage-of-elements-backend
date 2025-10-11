package domain

import "time"

// Player คือตารางที่เก็บข้อมูลผู้เล่น
type Player struct {
	ID                              uint          `gorm:"primaryKey;comment:ID เฉพาะของผู้เล่น (PK)" json:"id"`
	Username                        string        `gorm:"size:255;not null;unique;comment:ชื่อที่แสดงในเกม (Unique)" json:"username"`
	Email                           string        `gorm:"size:255;not null;unique;comment:อีเมลหลักสำหรับยืนยันตัวตน"`
	Status                          string        `gorm:"size:50;not null;default:'active';comment:สถานะบัญชี (active, banned)" json:"status"`
	IsEmailVerified                 bool          `gorm:"not null;default:false;comment:สถานะการยืนยันอีเมล" json:"is_email_verified"`
	EmailVerificationToken          *string       `gorm:"unique;comment:Token ลับสำหรับยืนยันอีเมล"`
	EmailVerificationTokenExpiresAt *time.Time    `gorm:"comment:เวลาหมดอายุของ Token ยืนยันอีเมล"`
	PasswordResetToken              *string       `gorm:"unique;comment:Token ลับสำหรับรีเซ็ตรหัสผ่าน"`
	PasswordResetTokenExpiresAt     *time.Time    `gorm:"comment:เวลาหมดอายุของ Token รีเซ็ตรหัสผ่าน"`
	CreatedAt                       time.Time     `gorm:"comment:วันที่สมัคร" json:"created_at"`
	LastLoginAt                     *time.Time    `gorm:"comment:เวลาที่เข้าสู่ระบบครั้งล่าสุด" json:"last_login_at"`
	PlayerAuth                      []*PlayerAuth `gorm:"foreignKey:PlayerID" json:"-"`
	Characters                      []*Character  `gorm:"foreignKey:PlayerID" json:"characters"`
}
