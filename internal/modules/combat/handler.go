// file: internal/modules/combat/handler.go
package combat

import (
	"sage-of-elements-backend/internal/domain"
	"sage-of-elements-backend/pkg/appauth"
	"sage-of-elements-backend/pkg/apperrors"
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

type CreateMatchRequest struct {
	CharacterID     uint                   `json:"character_id" validate:"required"`
	MatchType       string                 `json:"match_type" validate:"required,oneof=TRAINING PVE_STAGE"`
	StageID         *uint                  `json:"stage_id,omitempty"`         // Optional, needed for PVE_STAGE
	TrainingEnemies []domain.StageEnemy    `json:"training_enemies,omitempty"` // Optional, needed for TRAINING
	DeckID          *uint                  `json:"deck_id,omitempty"`          // Optional, for using a saved deck
	Deck            []DeckSlotInput        `json:"deck,omitempty"`             // Optional, for sending a custom deck
	Modifiers       *domain.MatchModifiers `json:"modifiers,omitempty"`
}

// --- Handler ---
type CombatHandler struct {
	validator *validator.Validate
	service   CombatService
}

func NewCombatHandler(validator *validator.Validate, service CombatService) *CombatHandler {
	return &CombatHandler{
		validator: validator,
		service:   service,
	}
}

func (h *CombatHandler) RegisterProtectedRoutes(router fiber.Router) {

	router.Post("/", h.CreateMatch)
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
