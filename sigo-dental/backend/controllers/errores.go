// Este archivo centraliza las respuestas de error compartidas por los controladores.
package controllers

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"sigo-dental-backend/models"
)

// errorInterno registra el detalle en el log del servidor y responde 500 con un mensaje genérico,
// para no filtrar a la persona usuaria información interna (consultas, rutas, errores del driver).
func errorInterno(c *fiber.Ctx, err error) error {
	log.Printf("error interno en %s %s: %v", c.Method(), c.Path(), err)
	return c.Status(fiber.StatusInternalServerError).JSON(models.APIError{
		Status:  fiber.StatusInternalServerError,
		Message: "Error interno del servidor",
	})
}
