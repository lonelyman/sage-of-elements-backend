package pve

import (
	"sage-of-elements-backend/pkg/appresponse"

	"github.com/gofiber/fiber/v2"
)

type PveHandler struct {
	pveService PveService
}

func NewPveHandler(pveService PveService) *PveHandler {
	return &PveHandler{pveService: pveService}
}

// ⭐️ พี่ขอเสนอให้เปลี่ยนเส้นทางเป็น /realms เพื่อให้สื่อความหมายมากขึ้นนะ
func (h *PveHandler) RegisterProtectedRoutes(router fiber.Router) {
	router.Get("/realms", h.GetRealms)
}

func (h *PveHandler) GetRealms(c *fiber.Ctx) error {
	realms, err := h.pveService.GetAllActiveRealms()
	if err != nil {
		return err
	}
	return appresponse.Success(c, fiber.StatusOK, "Realms retrieved successfully", realms, nil)
}
