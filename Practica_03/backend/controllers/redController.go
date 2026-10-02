// Seguimos dentro del paquete controllers, ahora para la sección de la red multinivel.
package controllers

import (
	// Importamos Fiber para manejar la respuesta HTTP.
	"github.com/gofiber/fiber/v2"
	// Importamos el repositorio para consultar el árbol de referidos.
	"multicatalogo-backend/repository"
)

// GetRed devuelve el árbol completo de la red multinivel del usuario.
func GetRed(c *fiber.Ctx) error {
	// Fiber serializa la estructura recursiva de referidos a JSON.
	return c.JSON(repository.ObtenerRed())
}
