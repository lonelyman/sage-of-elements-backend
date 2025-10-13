package character

import (
	"sage-of-elements-backend/pkg/appauth"
	"sage-of-elements-backend/pkg/apperrors"
	"sage-of-elements-backend/pkg/appresponse"
	"sage-of-elements-backend/pkg/appvalidator"
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	// Import jwt เพื่ออ่านข้อมูลจาก token
)

// characterHandler คือ struct ที่จัดการ Request/Response ของ Character
type characterHandler struct {
	validator *validator.Validate
	service   CharacterService
}

// NewCharacterHandler คือฟังก์ชันสำหรับสร้าง Handler
func NewCharacterHandler(validator *validator.Validate, service CharacterService) *characterHandler {
	return &characterHandler{
		validator: validator,
		service:   service,
	}
}

// CreateCharacterRequest คือ Struct สำหรับรับข้อมูล JSON
type CreateCharacterRequest struct {
	Name             string `json:"character_name" validate:"required,min=3"`
	Gender           string `json:"gender" validate:"required,oneof=MALE FEMALE"`
	PrimaryElementID uint   `json:"primary_element_id" validate:"required,min=1,max=4"`
	InitialMasteryID uint   `json:"initial_mastery_id" validate:"required,min=1,max=4"`
}

func (h *characterHandler) RegisterProtectedRoutes(router fiber.Router) {
	router.Post("/", h.CreateCharacter)
	router.Get("/", h.ListCharacters)
	router.Get("/:id", h.GetCharacterByID)
	router.Delete("/:id", h.DeleteCharacter)
	router.Get("/:id/inventory", h.GetInventory)

}

// CreateCharacter คือ Handler Function สำหรับ Endpoint สร้างตัวละคร
func (h *characterHandler) CreateCharacter(c *fiber.Ctx) error {
	// --- ส่วนที่สำคัญที่สุด: ดึงข้อมูลผู้ใช้ที่ Login อยู่ ออกมาจาก Token ---
	// (ในอนาคต เราจะมี Middleware ที่ทำส่วนนี้ให้โดยอัตโนมัติ)
	claims := c.Locals("user_claims").(*appauth.Claims)
	if claims == nil {
		// เพิ่มการตรวจสอบเผื่อไว้ ว่าถ้าเกิดอะไรผิดพลาด Middleware ส่ง nil มา
		return apperrors.UnauthorizedError("Invalid token claims")
	}
	playerID := claims.UserID

	// ----------------------------------------------------------------

	req := new(CreateCharacterRequest)

	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Cannot parse JSON"})
	}

	if validationResult := appvalidator.Validate(h.validator, req); !validationResult.IsValid {
		return c.Status(fiber.StatusBadRequest).JSON(validationResult)
	}

	// เรียกใช้ Service โดยส่ง playerID ที่ได้จาก Token เข้าไปด้วย
	newCharacter, err := h.service.CreateCharacter(
		playerID,
		req.Name,
		req.Gender,
		req.PrimaryElementID,
		req.InitialMasteryID,
	)
	if err != nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": err.Error()})
	}

	// ส่ง Response กลับไป (เราอาจจะอยากสร้าง DTO (Data Transfer Object) ที่สวยงามกว่านี้ในอนาคต)
	return c.Status(fiber.StatusCreated).JSON(newCharacter)
}

// ListCharacters คือ Handler Function สำหรับดึงรายชื่อตัวละครทั้งหมดของผู้ใช้ที่ Login อยู่
func (h *characterHandler) ListCharacters(c *fiber.Ctx) error {
	// 1. ดึงข้อมูล Claims ที่ Middleware ส่งมาให้
	claims := c.Locals("user_claims").(*appauth.Claims)
	if claims == nil {
		return apperrors.UnauthorizedError("Invalid token claims")
	}
	playerID := claims.UserID

	// 2. เรียกใช้ Service เพื่อดึงข้อมูลตัวละครทั้งหมด
	characters, err := h.service.ListCharacters(playerID)
	if err != nil {
		return err // ส่งต่อ Error ให้ Central Error Handler
	}

	// (ในอนาคต เราอาจจะสร้าง DTO สำหรับ List เพื่อซ่อนข้อมูลบางอย่าง)
	// แต่สำหรับตอนนี้ การส่งกลับไปทั้ง Slice เลยก็ถือว่าใช้ได้

	// 3. ส่ง Response กลับไป
	return appresponse.Success(c, fiber.StatusOK, "ok", characters, nil)
}

// GetCharacterByID คือ Handler Function สำหรับดึงข้อมูลตัวละคร 1 ตัวด้วย ID
func (h *characterHandler) GetCharacterByID(c *fiber.Ctx) error {
	// 1. ดึงข้อมูล Claims ที่ Middleware ส่งมาให้ เพื่อเอา PlayerID
	claims := c.Locals("user_claims").(*appauth.Claims)
	playerID := claims.UserID

	// 2. ดึง Character ID มาจาก URL Parameter (/:id)
	charIDStr := c.Params("id")
	charID, err := strconv.ParseUint(charIDStr, 10, 64)
	if err != nil {
		return apperrors.InvalidFormatError("Invalid character ID format", nil)
	}

	// 3. เรียกใช้ Service เพื่อดึงข้อมูลตัวละคร
	// Service ของเราจะทำการ "ตรวจสอบความเป็นเจ้าของ" ให้เอง!
	character, err := h.service.GetCharacterByID(playerID, uint(charID))
	if err != nil {
		return err // ส่งต่อ Error ให้ Central Error Handler
	}

	// 4. ส่ง Response กลับไป
	return appresponse.Success(c, fiber.StatusOK, "ok", character, nil)
}

// DeleteCharacter คือ Handler Function สำหรับลบตัวละคร
func (h *characterHandler) DeleteCharacter(c *fiber.Ctx) error {
	// 1. ดึงข้อมูล Claims ที่ Middleware ส่งมาให้ เพื่อเอา PlayerID
	claims := c.Locals("user_claims").(*appauth.Claims)
	playerID := claims.UserID

	// 2. ดึง Character ID มาจาก URL Parameter (/:id)
	charIDStr := c.Params("id")
	charID, err := strconv.ParseUint(charIDStr, 10, 64)
	if err != nil {
		return apperrors.InvalidFormatError("Invalid character ID format", nil)
	}

	// 3. เรียกใช้ Service เพื่อทำงาน Logic หลัก (ซึ่งมีการตรวจสอบความเป็นเจ้าของอยู่ข้างใน)
	err = h.service.DeleteCharacter(playerID, uint(charID))
	if err != nil {
		return err // ส่งต่อ Error ให้ Central Error Handler
	}

	// 4. ✨ ส่ง Response 204 No Content ✨
	// เป็นมาตรฐานสากลว่าถ้า DELETE สำเร็จ... เราจะส่ง Status 204 กลับไป ซึ่งหมายถึง "สำเร็จ แต่ไม่มีข้อมูลอะไรจะส่งกลับไปให้นะ"
	return appresponse.NoContent(c)
}

// GetInventory คือ Handler Function สำหรับดึงข้อมูลคลังของตัวละคร
func (h *characterHandler) GetInventory(c *fiber.Ctx) error {
	// 1. ดึงข้อมูล Claims ที่ Middleware ส่งมาให้ เพื่อเอา PlayerID
	claims := c.Locals("user_claims").(*appauth.Claims)
	playerID := claims.UserID

	// 2. ดึง Character ID มาจาก URL Parameter (/:id)
	charIDStr := c.Params("id")
	charID, err := strconv.ParseUint(charIDStr, 10, 64)
	if err != nil {
		return apperrors.InvalidFormatError("Invalid character ID format", nil)
	}

	// 3. เรียกใช้ Service เพื่อดึงข้อมูล
	inventoryResponse, err := h.service.GetInventory(playerID, uint(charID))
	if err != nil {
		return err // ส่งต่อ Error ให้ Central Error Handler
	}

	// 4. ส่ง Response กลับไป
	return appresponse.Success(c, fiber.StatusOK, "Inventory retrieved successfully", inventoryResponse, nil)
}

func (h *characterHandler) AdvanceTutorial(c *fiber.Ctx) error {
	// 1. ดึง PlayerID จาก Token และ CharacterID จาก URL
	claims := c.Locals("user_claims").(*appauth.Claims)
	charIDStr := c.Params("id")
	charID, _ := strconv.ParseUint(charIDStr, 10, 32)

	// 2. เรียกใช้ Service
	updatedChar, err := h.service.AdvanceTutorialStep(claims.UserID, uint(charID))
	if err != nil {
		return err
	}

	// 3. ส่งข้อมูลตัวละครที่อัปเดตแล้วกลับไป
	return appresponse.Success(c, fiber.StatusOK, "Tutorial step advanced", updatedChar, nil)
}

func (h *characterHandler) SkipTutorial(c *fiber.Ctx) error {
	// 1. ดึง PlayerID จาก Token และ CharacterID จาก URL
	claims := c.Locals("user_claims").(*appauth.Claims)
	charIDStr := c.Params("id")
	charID, _ := strconv.ParseUint(charIDStr, 10, 32)

	// 2. เรียกใช้ Service
	updatedChar, err := h.service.SkipTutorial(claims.UserID, uint(charID))
	if err != nil {
		return err
	}

	// 3. ส่งข้อมูลตัวละครที่อัปเดตแล้วกลับไป
	return appresponse.Success(c, fiber.StatusOK, "Tutorial skipped", updatedChar, nil)
}
