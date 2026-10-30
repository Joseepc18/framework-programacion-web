package controllers

import "github.com/gofiber/fiber/v2"

// HealthController comprueba el estado de PostgreSQL y de Redis a través de sus interfaces.
type HealthController struct {
	repo  DBRepository
	cache CacheRepository
}

// NuevoHealthController construye el controlador con las dependencias que revisará.
func NuevoHealthController(repo DBRepository, cache CacheRepository) *HealthController {
	return &HealthController{repo: repo, cache: cache}
}

// HealthCheck confirma que el servidor y sus dependencias (PostgreSQL y Redis) están operativas.
// Sin PostgreSQL responde 503; sin Redis responde 200 "degraded" porque la API sigue sirviendo.
func (ctrl *HealthController) HealthCheck(c *fiber.Ctx) error {
	if ctrl.repo.Ping(c.Context()) != nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"status":   "error",
			"database": "down",
			"message":  "Base de datos PostgreSQL no disponible",
		})
	}

	if ctrl.cache.Ping(c.Context()) != nil {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"status":   "degraded",
			"database": "up",
			"redis":    "down",
			"message":  "Servicio operativo pero con caché degradada",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":   "ok",
		"database": "up",
		"redis":    "up",
		"message":  "Todos los servicios operativos",
	})
}
