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
	MatchType       string                 `json:"match_type" validate:"required,oneof=TRAINING STORY PVP"`
	StageID         *uint                  `json:"stage_id,omitempty"`         // Required for STORY
	OpponentID      *uint                  `json:"opponent_id,omitempty"`      // Required for PVP
	TrainingEnemies []TrainingEnemyInput   `json:"training_enemies,omitempty"` // Required for TRAINING
	DeckID          *uint                  `json:"deck_id,omitempty"`
	Deck            []DeckSlotInput        `json:"deck,omitempty" validate:"max=8,dive"`
	Modifiers       *domain.MatchModifiers `json:"modifiers,omitempty"`
}

type PerformActionRequest struct {
	ActionType string  `json:"action_type" validate:"required,oneof=END_TURN CAST_SPELL"`
	CastMode   string  `json:"cast_mode,omitempty" validate:"omitempty,oneof=INSTANT CHARGE OVERCHARGE"`
	SpellID    *uint   `json:"spell_id,omitempty"`  // ⭐️ สำหรับ "CAST_SPELL"
	TargetID   *string `json:"target_id,omitempty"` // ⭐️ สำหรับ "CAST_SPELL"
}

type PerformActionResponse struct {
	UpdatedMatch    *domain.CombatMatch  `json:"updatedMatch"`
	PerformedAction PerformActionRequest `json:"performedAction"`
}

// --- DTO สำหรับ ResolveSpell (Endpoint แยก) ---
type ResolveSpellRequest struct {
	ElementID       uint  `json:"element_id" validate:"required"` // ธาตุที่ต้องการหาเวท
	MasteryID       uint  `json:"mastery_id" validate:"required"` // ศาสตร์ที่ต้องการใช้
	CasterElementID *uint `json:"caster_element_id,omitempty"`    // ธาตุหลักของผู้ที่จะใช้เวท (optional - ถ้าไม่ส่งจะใช้จาก match's player character)
}

type ResolveSpellResponse struct {
	Spell             *domain.Spell `json:"spell"`
	ElementRequested  uint          `json:"element_requested"`
	MasteryRequested  uint          `json:"mastery_requested"`
	CasterElementUsed uint          `json:"caster_element_used"`
	//FallbackDescription string        `json:"fallback_description,omitempty"`
	//FallbackUsed        bool          `json:"fallback_used"`
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
	router.Get("/resolve-spell", h.ResolveSpell) // ⭐️ GET Endpoint สำหรับ ResolveSpell
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

// ✨⭐️ Handler สำหรับ ResolveSpell (GET Endpoint) ⭐️✨
func (h *CombatHandler) ResolveSpell(c *fiber.Ctx) error {
	claims := c.Locals("user_claims").(*appauth.Claims)
	playerID := claims.UserID

	// รับ parameters จาก query string
	elementIDStr := c.Query("element_id")
	masteryIDStr := c.Query("mastery_id")
	casterElementIDStr := c.Query("caster_element_id")

	// Validate required parameters
	if elementIDStr == "" || masteryIDStr == "" {
		return apperrors.New(422, "MISSING_PARAMETERS", "กรุณาระบุ element_id และ mastery_id")
	}

	// Parse element_id
	elementID := c.QueryInt("element_id", 0)
	if elementID <= 0 {
		return apperrors.New(422, "INVALID_ELEMENT_ID", "element_id ต้องเป็นตัวเลขที่มากกว่า 0")
	}

	// Parse mastery_id
	masteryID := c.QueryInt("mastery_id", 0)
	if masteryID <= 0 {
		return apperrors.New(422, "INVALID_MASTERY_ID", "mastery_id ต้องเป็นตัวเลขที่มากกว่า 0")
	}

	// กำหนด casterElementID (ถ้าไม่ส่งมา ต้องหาจาก character ของ player)
	casterElementID := uint(0)
	if casterElementIDStr != "" {
		casterElementIDInt := c.QueryInt("caster_element_id", 0)
		if casterElementIDInt <= 0 {
			return apperrors.New(422, "INVALID_CASTER_ELEMENT_ID", "caster_element_id ต้องเป็นตัวเลขที่มากกว่า 0")
		}
		casterElementID = uint(casterElementIDInt)
		h.appLogger.Info("Using custom caster element ID", "casterElement", casterElementID, "playerID", playerID)
	} else {
		// ดึง character ของ player เพื่อเอา PrimaryElementID
		// NOTE: ต้องมี method ใน service หรือ character repo
		h.appLogger.Info("No caster_element_id provided, need to fetch from player's active character", "playerID", playerID)
		// TODO: ต้องเพิ่ม logic ดึง character's primary element
		return apperrors.New(422, "MISSING_CASTER_ELEMENT", "กรุณาระบุ caster_element_id หรือให้ระบบดึงจาก character")
	}

	// เรียกใช้ ResolveSpell
	spell, err := h.service.ResolveSpell(uint(elementID), uint(masteryID), casterElementID)
	if err != nil {
		return err
	}

	// สร้าง response
	response := ResolveSpellResponse{
		Spell:             spell,
		ElementRequested:  uint(elementID),
		MasteryRequested:  uint(masteryID),
		CasterElementUsed: casterElementID,
		//	FallbackUsed:      false, // TODO: ต้องเพิ่ม flag ใน ResolveSpell เพื่อบอกว่าใช้ fallback หรือไม่
	}

	return appresponse.Success(c, fiber.StatusOK, "Spell resolved successfully", response, nil)
}
