// Este archivo expone la administración de cuentas (Fase 7), solo para admin.
package controllers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"multicatalogo-backend/models"
	"multicatalogo-backend/repository"
)

// GetUsuarios devuelve una página de cuentas; acepta ?q= (nombre o correo), ?page y ?limit.
func GetUsuarios(c *fiber.Ctx) error {
	pagina, limite := paginaDe(c)
	resultado, err := repository.ObtenerUsuarios(c.Context(), c.Query("q"), pagina, limite)
	if err != nil {
		return errorInterno(c, err)
	}
	return c.JSON(resultado)
}

// ActualizarUsuario cambia el rol y/o activa o desactiva una cuenta (PATCH /api/usuarios/:id).
// Un admin no puede quitarse a sí mismo el rol ni desactivar su propia cuenta:
// evitaría quedarse sin administradores.
func ActualizarUsuario(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil || id < 1 {
		return c.Status(fiber.StatusBadRequest).JSON(models.APIError{
			Status:  fiber.StatusBadRequest,
			Message: "El ID del usuario debe ser un número entero",
			Details: fiber.Map{"id": c.Params("id")},
		})
	}

	var req struct {
		Rol    *string `json:"rol"`
		Activo *bool   `json:"activo"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.APIError{
			Status:  fiber.StatusBadRequest,
			Message: "Cuerpo de petición inválido",
		})
	}
	if req.Rol == nil && req.Activo == nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.APIError{
			Status:  fiber.StatusBadRequest,
			Message: "Indique rol y/o activo para actualizar",
		})
	}
	if req.Rol != nil && *req.Rol != "admin" && *req.Rol != "cliente" {
		return c.Status(fiber.StatusBadRequest).JSON(models.APIError{
			Status:  fiber.StatusBadRequest,
			Message: "El rol debe ser 'admin' o 'cliente'",
		})
	}

	propioID, _ := usuarioIDDelToken(c)
	if propioID == id &&
		((req.Activo != nil && !*req.Activo) || (req.Rol != nil && *req.Rol != "admin")) {
		return c.Status(fiber.StatusConflict).JSON(models.APIError{
			Status:  fiber.StatusConflict,
			Message: "No puede desactivar su cuenta ni quitarse el rol de administrador",
		})
	}

	usuario, existe, err := repository.ActualizarUsuario(c.Context(), id, req.Rol, req.Activo)
	if err != nil {
		return errorInterno(c, err)
	}
	if !existe {
		return c.Status(fiber.StatusNotFound).JSON(models.APIError{
			Status:  fiber.StatusNotFound,
			Message: "Usuario no encontrado",
			Details: fiber.Map{"id": id},
		})
	}
	return c.JSON(usuario)
}
