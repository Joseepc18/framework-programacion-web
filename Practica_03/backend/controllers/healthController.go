// Seguimos dentro del paquete controllers, ahora para la verificación de estado del servidor.
package controllers

import (
	// Importamos Fiber para manejar la respuesta HTTP.
	"github.com/gofiber/fiber/v2"
	// Importamos config para consultar el estado de Redis.
	"multicatalogo-backend/config"
	// Importamos repository para comprobar la base de datos PostgreSQL.
	"multicatalogo-backend/repository"
)

// HealthCheck confirma que el servidor y sus dependencias (PostgreSQL y Redis) están operativas.
func HealthCheck(c *fiber.Ctx) error {
	dbStatus := "up"
	if db := repository.GetDB(); db == nil || db.Ping(c.Context()) != nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"status":   "error",
			"database": "down",
			"message":  "Base de datos PostgreSQL no disponible",
		})
	}

	redisStatus := "up"
	if config.Cache == nil || config.Cache.Ping(c.Context()).Err() != nil {
		redisStatus = "down"
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"status":   "degraded",
			"database": dbStatus,
			"redis":    redisStatus,
			"message":  "Servicio operativo pero con caché degradada",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":   "ok",
		"database": dbStatus,
		"redis":    redisStatus,
		"message":  "Todos los servicios operativos",
	})
}
