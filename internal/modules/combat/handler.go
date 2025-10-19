// file: internal/modules/combat/handler.go
package combat

import (
	"sage-of-elements-backend/internal/domain"
	"sage-of-elements-backend/pkg/appauth"
	"sage-of-elements-backend/pkg/apperrors"
	"sage-of-elements-backend/pkg/applogger"
	"sage-of-elements-backend/pkg/appresponse"
	"sage-of-elements-backend/pkg/appvalidator"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

// --- DTOs (Data Transfer Objects) ---
type DeckSlotInput struct {
	SlotNum   int  `json:"slotNum" validate:"required,gte=1,lte=8"`
	ElementID uint `json:"elementId" validate:"required,gte=5"`
}

type TrainingEnemyInput struct {
	EnemyID uint `json:"enemy_id" validate:"required"`
}
type CreateMatchRequest struct {
	CharacterID     uint                   `json:"character_id" validate:"required"`
	MatchType       string                 `json:"match_type" validate:"required,oneof=TRAINING PVE_STAGE"`
	StageID         *uint                  `json:"stage_id,omitempty"`
	TrainingEnemies []TrainingEnemyInput   `json:"training_enemies,omitempty" validate:"dive"`
	DeckID          *uint                  `json:"deck_id,omitempty"`
	Deck            []DeckSlotInput        `json:"deck,omitempty" validate:"max=8,dive"`
	Modifiers       *domain.MatchModifiers `json:"modifiers,omitempty"`
}

type PerformActionRequest struct {
	ActionType string  `json:"action_type" validate:"required,oneof=END_TURN DRAW_CHARGE CAST_SPELL EMERGENCY_FUSION"`
	CastMode   string  `json:"cast_mode,omitempty" validate:"omitempty,oneof=INSTANT CHARGE OVERCHARGE"`
	ElementID  *uint   `json:"element_id,omitempty"` // ⭐️ สำหรับ "DRAW_CHARGE" หรือ "EMERGENCY_FUSION"
	SpellID    *uint   `json:"spell_id,omitempty"`   // ⭐️ สำหรับ "CAST_SPELL"
	TargetID   *string `json:"target_id,omitempty"`  // ⭐️ สำหรับ "CAST_SPELL"
}

type PerformActionResponse struct {
	UpdatedMatch    *domain.CombatMatch  `json:"updatedMatch"`
	PerformedAction PerformActionRequest `json:"performedAction"`
}

// --- Handler ---
type CombatHandler struct {
	appLogger applogger.Logger
	validator *validator.Validate
	service   CombatService
}

func NewCombatHandler(appLogger applogger.Logger, validator *validator.Validate, service CombatService) *CombatHandler {
	return &CombatHandler{
		appLogger: appLogger,
		validator: validator,
		service:   service,
	}
}

func (h *CombatHandler) RegisterProtectedRoutes(router fiber.Router) {
	router.Post("/", h.CreateMatch)
	router.Post("/:id/actions", h.PerformAction)
}

// --- Handler Functions ---
func (h *CombatHandler) CreateMatch(c *fiber.Ctx) error {
	// 1. ดึง PlayerID จาก Token
	claims := c.Locals("user_claims").(*appauth.Claims)
	playerID := claims.UserID

	// 2. Parse & Validate Body
	req := new(CreateMatchRequest)
	if err := c.BodyParser(req); err != nil {
		return apperrors.InvalidFormatError("Cannot parse JSON", nil)
	}
	if validationResult := appvalidator.Validate(h.validator, req); !validationResult.IsValid {
		return c.Status(fiber.StatusBadRequest).JSON(validationResult)
	}

	// 3. เรียกใช้ Service เพื่อสร้างห้องต่อสู้
	newMatch, err := h.service.CreateMatch(playerID, *req)
	if err != nil {
		return err // ส่งต่อให้ Error Handler กลาง
	}

	// 4. ส่งสถานะของห้องที่เพิ่งสร้างเสร็จกลับไป
	return appresponse.Success(c, fiber.StatusCreated, "Match created successfully", newMatch, nil)
}

// ✨⭐️ สร้าง Handler Function ใหม่สำหรับรับ Action! ⭐️✨
func (h *CombatHandler) PerformAction(c *fiber.Ctx) error {
	claims := c.Locals("user_claims").(*appauth.Claims)
	matchID := c.Params("id")

	req := new(PerformActionRequest) // <-- ใช้ DTO จาก service.go
	if err := c.BodyParser(req); err != nil {
		return apperrors.InvalidFormatError("Cannot parse JSON", nil)
	}
	if validationResult := appvalidator.Validate(h.validator, req); !validationResult.IsValid {
		return apperrors.ValidationError("Validation failed", validationResult.Errors)
	}

	actionResponse, err := h.service.PerformAction(claims.UserID, matchID, *req)
	if err != nil {
		return err
	}

	return appresponse.Success(c, fiber.StatusOK, "Action performed successfully", actionResponse, nil)
}
