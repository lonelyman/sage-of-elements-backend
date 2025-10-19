package enemy

import (
	"sage-of-elements-backend/pkg/appresponse"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

type EnemyHandler struct {
	validator *validator.Validate
	service   EnemyService
}

func NewEnemyHandler(validator *validator.Validate, service EnemyService) *EnemyHandler {
	return &EnemyHandler{validator: validator, service: service}
}

func (h *EnemyHandler) RegisterProtectedRoutes(router fiber.Router) {
	router.Get("/", h.GetAllEnemies)
}

func (h *EnemyHandler) GetAllEnemies(c *fiber.Ctx) error {
	enemies, err := h.service.GetAllEnemies()
	if err != nil {
		return err
	}
	return appresponse.Success(c, fiber.StatusOK, "Enemies retrieved successfully", enemies, nil)
}
