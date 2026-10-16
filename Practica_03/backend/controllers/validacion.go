// Este archivo centraliza la validación de entrada que antes cada controlador repetía a mano.
package controllers

import (
	// Importamos net/mail para validar el formato del correo.
	"net/mail"
	// Importamos strings para limpiar espacios en blanco.
	"strings"

	"github.com/gofiber/fiber/v2"
	"multicatalogo-backend/models"
)

// campo es un par nombre/valor de un dato obligatorio del cuerpo de la petición.
type campo struct {
	nombre string
	valor  string
}

// camposVacios devuelve los nombres de los campos obligatorios que llegaron vacíos.
func camposVacios(campos ...campo) []string {
	vacios := []string{}
	for _, f := range campos {
		if strings.TrimSpace(f.valor) == "" {
			vacios = append(vacios, f.nombre)
		}
	}
	return vacios
}

// emailValido comprueba que el texto sea una dirección de correo bien formada.
func emailValido(texto string) bool {
	direccion, err := mail.ParseAddress(strings.TrimSpace(texto))
	return err == nil && direccion.Address == strings.TrimSpace(texto)
}

// respuestaCamposVacios responde 400 con el contrato estándar listando los campos faltantes.
func respuestaCamposVacios(c *fiber.Ctx, mensaje string, vacios []string) error {
	return c.Status(fiber.StatusBadRequest).JSON(models.APIError{
		Status:  fiber.StatusBadRequest,
		Message: mensaje,
		Details: fiber.Map{"camposVacios": vacios},
	})
}

// respuestaInvalida responde 400 con un mensaje de validación y el campo afectado.
func respuestaInvalida(c *fiber.Ctx, mensaje, campo string) error {
	return c.Status(fiber.StatusBadRequest).JSON(models.APIError{
		Status:  fiber.StatusBadRequest,
		Message: mensaje,
		Details: fiber.Map{"campo": campo},
	})
}
