package fusion

import (
	"sage-of-elements-backend/pkg/appauth"
	"sage-of-elements-backend/pkg/apperrors"
	"sage-of-elements-backend/pkg/appresponse"
	"sage-of-elements-backend/pkg/appvalidator"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

// fusionHandler คือ struct ที่จัดการ Request/Response
type fusionHandler struct {
	validator *validator.Validate
	service   FusionService
}

// NewFusionHandler คือฟังก์ชันสำหรับสร้าง Handler
func NewFusionHandler(validator *validator.Validate, service FusionService) *fusionHandler {
	return &fusionHandler{
		validator: validator,
		service:   service,
	}
}

// CraftRequest คือ DTO สำหรับรับข้อมูล JSON
type CraftRequest struct {
	CharacterID uint              `json:"character_id" validate:"required"`
	Ingredients []IngredientInput `json:"ingredients" validate:"required,min=1"`
}

// RegisterRoutes ลงทะเบียน Endpoint ทั้งหมดของ Fusion Module
func (h *fusionHandler) RegisterProtectedRoutes(router fiber.Router) {
	router.Post("/craft", h.CraftElement)
}

// CraftElement คือ Handler Function สำหรับ Endpoint หลอมรวมธาตุ
func (h *fusionHandler) CraftElement(c *fiber.Ctx) error {
	// 1. ดึงข้อมูล Claims ที่ Middleware ส่งมาให้ เพื่อเอา PlayerID
	claims := c.Locals("user_claims").(*appauth.Claims)
	playerID := claims.UserID

	// 2. Parse Request Body
	req := new(CraftRequest)
	if err := c.BodyParser(req); err != nil {
		return apperrors.New(fiber.StatusBadRequest, "INVALID_REQUEST", "Cannot parse JSON")
	}

	// 3. Validate Request Body
	if validationResult := appvalidator.Validate(h.validator, req); !validationResult.IsValid {
		return apperrors.ValidationError("Validation failed", validationResult.Errors)
	}

	// 4. เรียกใช้ Service เพื่อทำงาน Logic หลัก
	result, err := h.service.CraftElement(playerID, req.CharacterID, req.Ingredients)
	if err != nil {
		return err // ส่งต่อ Error ให้ Central Error Handler
	}

	// 5. ส่ง Response กลับไป
	return appresponse.Success(c, fiber.StatusCreated, "Crafting successful!", result, nil)
}
