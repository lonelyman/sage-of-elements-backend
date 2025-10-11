package player

import (
	"time"

	"sage-of-elements-backend/pkg/appauth"
	"sage-of-elements-backend/pkg/apperrors"
	"sage-of-elements-backend/pkg/appresponse"
	"sage-of-elements-backend/pkg/appvalidator"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

// playerHandler (เหมือนเดิม)
type playerHandler struct {
	validator *validator.Validate
	service   PlayerService
}

// NewPlayerHandler (เหมือนเดิม)
func NewPlayerHandler(validator *validator.Validate, service PlayerService) *playerHandler {
	return &playerHandler{
		validator: validator,
		service:   service,
	}
}

// --- DTOs (Data Transfer Objects) ---

// RegisterRequest (เพิ่ม Email!)
type RegisterRequest struct {
	Username string `json:"username" validate:"required,min=4"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

// RegisterResponse (สร้างใหม่!)
type RegisterResponse struct {
	ID        uint      `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

// LoginRequest (เหมือนเดิม)
type LoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

// LoginResponse (สร้างใหม่!)
type LoginResponse struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}

// ProfileResponse (สร้างใหม่เพื่อให้ชัดเจน)
type ProfileResponse struct {
	ID        uint      `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refreshToken" validate:"required"`
}

// --- Route Registration ---

func (h *playerHandler) RegisterPublicRoutes(router fiber.Router) {
	router.Post("/register", h.Register)
	router.Post("/login", h.Login)
	router.Post("/refresh-token", h.RefreshToken)
}

func (h *playerHandler) RegisterProtectedRoutes(router fiber.Router) {
	router.Get("/me", h.GetProfile)
}

// --- Handlers (ฉบับ Refactor!) ---

func (h *playerHandler) Register(c *fiber.Ctx) error {
	req := new(RegisterRequest)
	if err := c.BodyParser(req); err != nil {
		return apperrors.New(fiber.StatusBadRequest, "INVALID_REQUEST", "Cannot parse JSON")
	}
	if validationResult := appvalidator.Validate(h.validator, req); !validationResult.IsValid {
		return apperrors.ValidationError("Validation failed", validationResult.Errors)
	}

	// เรียกใช้ Service เวอร์ชันใหม่ (ส่ง Email ไปด้วย!)
	newPlayer, err := h.service.Register(req.Username, req.Email, req.Password)
	if err != nil {
		return err // ส่งต่อ Error ให้ Central Error Handler
	}

	// สร้าง Response DTO ที่ปลอดภัย
	responsePayload := &RegisterResponse{
		ID:        newPlayer.ID,
		Username:  newPlayer.Username,
		Email:     newPlayer.Email,
		CreatedAt: newPlayer.CreatedAt,
	}

	return appresponse.Success(c, fiber.StatusCreated, "User registered successfully", responsePayload, nil)
}

func (h *playerHandler) Login(c *fiber.Ctx) error {
	req := new(LoginRequest)
	if err := c.BodyParser(req); err != nil {
		return apperrors.New(fiber.StatusBadRequest, "INVALID_REQUEST", "Cannot parse JSON")
	}
	if validationResult := appvalidator.Validate(h.validator, req); !validationResult.IsValid {
		return apperrors.ValidationError("Validation failed", validationResult.Errors)
	}

	// รับ Token มา 2 ตัว!
	accessToken, refreshToken, err := h.service.Login(req.Username, req.Password)
	if err != nil {
		return err
	}

	// สร้าง Response DTO สำหรับ Login
	responsePayload := &LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}

	return appresponse.Success(c, fiber.StatusOK, "Login successful", responsePayload, nil)
}

func (h *playerHandler) GetProfile(c *fiber.Ctx) error {
	claims := c.Locals("user_claims").(*appauth.Claims)
	playerID := claims.UserID

	player, err := h.service.GetProfile(playerID)
	if err != nil {
		return err
	}

	// สร้าง Response DTO ที่ปลอดภัย
	responsePayload := &ProfileResponse{
		ID:        player.ID,
		Username:  player.Username,
		Email:     player.Email,
		CreatedAt: player.CreatedAt,
	}

	return appresponse.Success(c, fiber.StatusOK, "User profile retrieved successfully", responsePayload, nil)
}

// RefreshToken คือ Handler Function สำหรับขอ Access Token ใหม่
func (h *playerHandler) RefreshToken(c *fiber.Ctx) error {
	// 1. Parse Request Body
	req := new(RefreshTokenRequest)
	if err := c.BodyParser(req); err != nil {
		return apperrors.New(fiber.StatusBadRequest, "INVALID_REQUEST", "Cannot parse JSON")
	}
	if validationResult := appvalidator.Validate(h.validator, req); !validationResult.IsValid {
		return apperrors.ValidationError("Validation failed", validationResult.Errors)
	}

	// 2. เรียกใช้ Service เพื่อทำงาน Logic หลัก
	newAccessToken, err := h.service.RefreshToken(req.RefreshToken)
	if err != nil {
		return err // ส่งต่อ Error ให้ Central Error Handler
	}

	// 3. ส่ง Access Token ใบใหม่กลับไป
	return appresponse.Success(c, fiber.StatusOK, "Token refreshed successfully", fiber.Map{
		"accessToken": newAccessToken,
	}, nil)
}
