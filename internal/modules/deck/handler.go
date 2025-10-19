// file: internal/modules/deck/handler.go
package deck

import (
	"sage-of-elements-backend/pkg/appauth"
	"sage-of-elements-backend/pkg/apperrors"
	"sage-of-elements-backend/pkg/appresponse"
	"sage-of-elements-backend/pkg/appvalidator"
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

// --- DTOs (Data Transfer Objects) - เหมือนเดิม ---

type CreateDeckRequest struct {
	CharacterID uint   `json:"character_id" validate:"required"`
	Name        string `json:"name" validate:"required,min=3"`
}

type UpdateDeckRequest struct {
	Name  string            `json:"name" validate:"required,min=3"`
	Slots []DeckSlotRequest `json:"slots" validate:"dive"`
}

type DeckSlotRequest struct {
	SlotNum   int  `json:"slotNum" validate:"required,gte=1,lte=8"`
	ElementID uint `json:"elementId" validate:"required,gte=5"`
}

// --- Handler (เหมือนเดิม) ---

type deckHandler struct {
	validator *validator.Validate
	service   DeckService
}

func NewDeckHandler(validator *validator.Validate, service DeckService) *deckHandler {
	return &deckHandler{
		validator: validator,
		service:   service,
	}
}

func (h *deckHandler) RegisterProtectedRoutes(router fiber.Router) {
	router.Post("/", h.CreateDeck)
	router.Get("/", h.GetDecks)
	router.Put("/:id", h.UpdateDeck)
	router.Delete("/:id", h.DeleteDeck)
}

func (h *deckHandler) CreateDeck(c *fiber.Ctx) error {
	claims := c.Locals("user_claims").(*appauth.Claims)
	req := new(CreateDeckRequest)

	if err := c.BodyParser(req); err != nil {
		return apperrors.InvalidFormatError("Cannot parse JSON", nil)
	}

	if validationResult := appvalidator.Validate(h.validator, req); !validationResult.IsValid {
		return apperrors.ValidationError("Validation failed", validationResult.Errors)
	}

	newDeck, err := h.service.CreateDeck(claims.UserID, req.CharacterID, req.Name)
	if err != nil {
		return err
	}
	return appresponse.Success(c, fiber.StatusCreated, "Deck created successfully", newDeck, nil)
}

func (h *deckHandler) GetDecks(c *fiber.Ctx) error {
	claims := c.Locals("user_claims").(*appauth.Claims)
	charIDStr := c.Query("character_id")
	if charIDStr == "" {
		return apperrors.InvalidFormatError("Missing character_id query parameter", nil)
	}
	charID, err := strconv.ParseUint(charIDStr, 10, 32)
	if err != nil {
		return apperrors.InvalidFormatError("Invalid character_id format", nil)
	}

	decks, err := h.service.GetDecksByCharacterID(claims.UserID, uint(charID))
	if err != nil {
		return err
	}
	return appresponse.Success(c, fiber.StatusOK, "Decks retrieved successfully", decks, nil)
}

func (h *deckHandler) UpdateDeck(c *fiber.Ctx) error {
	claims := c.Locals("user_claims").(*appauth.Claims)
	deckIDStr := c.Params("id")
	deckID, err := strconv.ParseUint(deckIDStr, 10, 32)
	if err != nil {
		return apperrors.InvalidFormatError("Invalid deck ID format", nil)
	}

	req := new(UpdateDeckRequest)
	if err := c.BodyParser(req); err != nil {
		return apperrors.InvalidFormatError("Cannot parse JSON", nil)
	}

	if validationResult := appvalidator.Validate(h.validator, req); !validationResult.IsValid {
		return apperrors.ValidationError("Validation failed", validationResult.Errors)
	}

	updatedDeck, err := h.service.UpdateDeck(claims.UserID, uint(deckID), *req)
	if err != nil {
		return err
	}
	return appresponse.Success(c, fiber.StatusOK, "Deck updated successfully", updatedDeck, nil)
}

func (h *deckHandler) DeleteDeck(c *fiber.Ctx) error {
	// 1. ดึง PlayerID จาก Token
	claims := c.Locals("user_claims").(*appauth.Claims)
	playerID := claims.UserID

	// 2. ดึง DeckID จาก URL Param
	deckIDStr := c.Params("id")
	deckID, err := strconv.ParseUint(deckIDStr, 10, 32)
	if err != nil {
		return apperrors.InvalidFormatError("Invalid deck ID format", nil)
	}

	// 3. เรียกใช้ Service เพื่อลบ
	if err := h.service.DeleteDeck(playerID, uint(deckID)); err != nil {
		return err // ส่งต่อให้ Error Handler กลาง
	}

	// 4. ส่ง Response 204 No Content (มาตรฐานสากลสำหรับการลบสำเร็จ)
	return appresponse.NoContent(c)
}
