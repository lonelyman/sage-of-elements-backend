package player

import (
	"sage-of-elements-backend/internal/domain"
	"sage-of-elements-backend/pkg/appauth"
	"sage-of-elements-backend/pkg/apperrors"
	"sage-of-elements-backend/pkg/applogger"
	"time"

	"github.com/gofiber/fiber/v2"
)

// PlayerService คือ "สัญญา" สำหรับ Business Logic ที่เกี่ยวกับ Player
type PlayerService interface {
	Register(username, email, password string) (*domain.Player, error)
	Login(username, password string) (accessToken string, refreshToken string, err error)
	GetProfile(playerID uint) (*domain.Player, error)
	RefreshToken(refreshToken string) (newAccessToken string, err error)
}

// playerService คือ struct ที่จะเก็บ Logic การทำงานจริง
type playerService struct {
	appLogger  applogger.Logger
	authSvc    appauth.Service
	playerRepo PlayerRepository
}

// NewPlayerService คือฟังก์ชันสำหรับสร้าง Service ขึ้นมาใช้งาน
func NewPlayerService(appLogger applogger.Logger, authSvc appauth.Service, playerRepo PlayerRepository) PlayerService {
	return &playerService{
		appLogger:  appLogger,
		authSvc:    authSvc,
		playerRepo: playerRepo,
	}
}

// --- เราจะมาเติม Logic ในฟังก์ชันเหล่านี้กันในขั้นตอนต่อไป ---

// Register คือฟังก์ชันสำหรับสมัครสมาชิก
func (s *playerService) Register(username, email, password string) (*domain.Player, error) {
	// 1. Validation: ตรวจสอบ Input พื้นฐาน
	// if len(username) < 4 {
	// 	return nil, errors.New("username must be at least 4 characters long")
	// }
	// if len(password) < 8 {
	// 	return nil, errors.New("password must be at least 8 characters long")
	// }

	existingUser, err := s.playerRepo.FindByUsername(username)
	if err != nil {
		// ถ้ามี Error จาก DB ให้หยุดทำงาน
		s.appLogger.Error("database error on find by username", err)
		return nil, apperrors.SystemError("database error")
	}
	if existingUser != nil {
		// ถ้า "เจอ" ผู้เล่น (ผลลัพธ์ไม่เป็น nil) ถึงจะหมายความว่าชื่อซ้ำ!
		return nil, apperrors.AlreadyExistsError("username already exists", nil)
	}

	// ตรวจสอบ Email
	existingEmail, err := s.playerRepo.FindByEmail(email)
	if err != nil {
		s.appLogger.Error("database error on find by email", err)
		return nil, apperrors.SystemError("database error")
	}
	if existingEmail != nil {
		// ถ้า "เจอ" อีเมล ถึงจะหมายความว่าอีเมลซ้ำ!
		return nil, apperrors.AlreadyExistsError("email already exists", nil)
	}

	// 3. Hash รหัสผ่าน (เรียกใช้ Auth Service)
	hashedPassword, err := s.authSvc.HashPassword(password)
	if err != nil {
		return nil, apperrors.SystemErrorWithDetails("failed to hash password", err)
	}

	// 4. สร้าง Player object (ไม่มี HashedPassword แล้ว!)
	newPlayer := &domain.Player{
		Username: username,
		Email:    email,
		Status:   "active",
	}

	// 5. ✨ สร้าง PlayerAuth object สำหรับการ Login แบบ "local" ✨
	playerAuth := &domain.PlayerAuth{
		Provider: "local",
		Secret:   hashedPassword,
	}

	// 6. ส่งทั้งสอง object ไปให้ Repository จัดการใน Transaction เดียว!
	createdPlayer, err := s.playerRepo.CreateWithAuth(newPlayer, playerAuth)
	if err != nil {
		s.appLogger.Error("failed to create player with auth", err)
		return nil, apperrors.SystemError("failed to create player")
	}

	return createdPlayer, nil
}

func (s *playerService) Login(username, password string) (accessToken string, refreshToken string, err error) {
	// 1. ค้นหาผู้ใช้ด้วย Username
	player, err := s.playerRepo.FindByUsername(username)
	if err != nil || player == nil {
		return "", "", apperrors.UnauthorizedError("invalid username or password")
	}

	// 2. ค้นหา "วิธีการ Login" แบบ "local" (รหัสผ่าน) ของผู้ใช้คนนั้น
	playerAuth, err := s.playerRepo.FindAuthByPlayerIDAndProvider(player.ID, "local")
	if err != nil || playerAuth == nil {
		return "", "", apperrors.UnauthorizedError("password login is not enabled for this account")
	}

	// 3. เปรียบเทียบรหัสผ่าน
	err = s.authSvc.ComparePassword(playerAuth.Secret, password)
	if err != nil {
		return "", "", apperrors.UnauthorizedError("invalid username or password")
	}

	// 4. สร้าง Tokens (Access & Refresh)
	accessToken, refreshToken, err = s.authSvc.GenerateTokens(player.ID, "PLAYER")
	if err != nil {
		return "", "", apperrors.SystemErrorWithDetails("failed to generate tokens", err)
	}

	// 5. บันทึก Refresh Token และ LastLoginAt ลง Database
	now := time.Now()
	player.LastLoginAt = &now
	playerAuth.RefreshToken = &refreshToken

	if _, err := s.playerRepo.Update(player); err != nil {
		s.appLogger.Warn("failed to update last login time", "player_id", player.ID)
	}
	if _, err := s.playerRepo.UpdateAuth(playerAuth); err != nil {
		s.appLogger.Warn("failed to save refresh token", "player_auth_id", playerAuth.ID)
	}

	return accessToken, refreshToken, nil
}

func (s *playerService) GetProfile(playerID uint) (*domain.Player, error) {
	// 1. ตรวจสอบ Input พื้นฐาน
	if playerID == 0 {
		return nil, apperrors.InvalidFormatError("Invalid player ID", nil)
	}

	// 2. เรียกใช้ Repository เพื่อค้นหาผู้เล่นด้วย ID
	player, err := s.playerRepo.FindByID(playerID)
	if err != nil {
		// ถ้าเกิด Error จาก Database
		return nil, apperrors.SystemErrorWithDetails("database error on find by id", err)
	}
	if player == nil {
		// ถ้าหาผู้เล่นไม่เจอ
		return nil, apperrors.NotFoundError("player not found")
	}

	// 3. คืนค่า Player ที่หาเจอ
	return player, nil
}

// RefreshToken คือฟังก์ชันสำหรับขอ Access Token ใบใหม่โดยใช้ Refresh Token
func (s *playerService) RefreshToken(refreshToken string) (newAccessToken string, err error) {
	// 1. ตรวจสอบความถูกต้องและวันหมดอายุของ Refresh Token
	claims, err := s.authSvc.ValidateRefreshToken(refreshToken)
	if err != nil {
		// ถ้า Token ปลอม, หมดอายุ, หรือผิดพลาด
		return "", apperrors.NewWithDetails(fiber.StatusUnauthorized, apperrors.ErrInvalidToken, "Invalid or expired refresh token", err.Error())
	}

	// 2. ✨ หัวใจด้านความปลอดภัย! ✨
	// ค้นหาใน Database ว่ามี "ใคร" เป็นเจ้าของ Refresh Token ใบนี้
	playerAuth, err := s.playerRepo.FindAuthByRefreshToken(refreshToken)
	if err != nil {
		s.appLogger.Error("database error on find by refresh token", err)
		return "", apperrors.SystemError("database error")
	}
	// ถ้าหาไม่เจอ หรือ Token ใน DB ไม่ตรงกับที่ส่งมา (อาจจะถูกขโมยหรือเป็นใบเก่า) -> ไม่อนุญาต
	if playerAuth == nil || playerAuth.PlayerID != claims.UserID {
		return "", apperrors.UnauthorizedError("Invalid refresh token")
	}

	// 3. ถ้าทุกอย่างถูกต้อง... สร้าง Access Token ใบใหม่ให้
	newAccessToken, _, err = s.authSvc.GenerateTokens(playerAuth.PlayerID, claims.Role)
	if err != nil {
		return "", apperrors.SystemErrorWithDetails("failed to generate new access token", err)
	}

	return newAccessToken, nil
}
