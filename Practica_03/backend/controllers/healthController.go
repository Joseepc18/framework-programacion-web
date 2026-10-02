// Seguimos dentro del paquete controllers, ahora para la verificación de estado del servidor.
package controllers

import (
	// Importamos Fiber para manejar la respuesta HTTP.
	"github.com/gofiber/fiber/v2"
)

// HealthCheck confirma que el servidor está levantado y respondiendo peticiones.
func HealthCheck(c *fiber.Ctx) error {
	// Retornamos HTTP 200 con un JSON fijo que indica que el servidor está operativo.
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  "ok",
		"message": "Servidor Go/Fiber operativo",
	})
}
