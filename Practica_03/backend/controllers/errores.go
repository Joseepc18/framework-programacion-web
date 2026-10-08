// Seguimos dentro del paquete controllers; este archivo agrupa las respuestas de error compartidas.
package controllers

import (
	// Importamos log para registrar en la terminal el detalle técnico del error.
	"log"

	// Importamos Fiber para manejar la respuesta HTTP.
	"github.com/gofiber/fiber/v2"
	// Importamos nuestro paquete de modelos para usar la estructura APIError.
	"multicatalogo-backend/models"
)

// errorInterno registra el error real en la terminal y responde HTTP 500 con un mensaje genérico,
// para no exponer detalles de la base de datos al cliente.
func errorInterno(c *fiber.Ctx, err error) error {
	log.Printf("error interno en %s %s: %v", c.Method(), c.Path(), err)
	return c.Status(fiber.StatusInternalServerError).JSON(models.APIError{
		Status:  fiber.StatusInternalServerError,
		Message: "Error interno del servidor",
	})
}
