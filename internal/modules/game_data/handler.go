package game_data

import (
	"sage-of-elements-backend/pkg/appresponse"

	"github.com/gofiber/fiber/v2"
)

// gameDataHandler คือ struct ที่จัดการ Request/Response ของ Game Data
type gameDataHandler struct {
	service Service
}

// NewHandler คือฟังก์ชันสำหรับสร้าง Handler
func NewGameDataHandler(service Service) *gameDataHandler {
	return &gameDataHandler{
		service: service,
	}
}

// GetMasterData คือ Handler Function ที่จะถูกเรียกโดย Endpoint
func (h *gameDataHandler) GetMasterData(c *fiber.Ctx) error {
	// 1. เรียกใช้ Service เพื่อดึงข้อมูล Master Data ทั้งหมด
	// (Service ของเราจะจัดการเรื่อง Cache ให้เองโดยอัตโนมัติ)
	masterData, err := h.service.GetMasterData()
	if err != nil {
		// ส่งต่อ Error ให้ Central Error Handler
		return err
	}

	// 2. ส่ง Response กลับไป
	return appresponse.Success(c, fiber.StatusOK, "ok", masterData, nil)
}

// RegisterRoutes ลงทะเบียน Endpoint ทั้งหมดของ Game Data Module
func (h *gameDataHandler) RegisterProtectedRoutes(router fiber.Router) {
	router.Get("/master", h.GetMasterData)
}
