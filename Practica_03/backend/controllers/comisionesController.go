// Este archivo agrupa el ledger de comisiones del usuario (Fase 4).
// Cada cuenta solo consulta LO SUYO: el nodo de su token decide el resultado.
package controllers

import (
	// Importamos el framework Fiber para manejar la petición y la respuesta.
	"github.com/gofiber/fiber/v2"
	// Importamos nuestros modelos para APIError.
	"multicatalogo-backend/models"
)

// GetComisionesMias devuelve las comisiones que OTORGAN al usuario del token
// (su subárbol de upline), paginadas y con el total pendiente y el pagado.
func (ctrl *RedController) GetComisionesMias(c *fiber.Ctx) error {
	usuarioID, ok := usuarioIDDelToken(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(models.APIError{
			Status:  fiber.StatusUnauthorized,
			Message: "Sesión no válida o expirada",
		})
	}

	// El nodo de la red es la puerta de entrada al ledger del usuario.
	nodoID, tieneNodo, err := ctrl.repo.NodoDeRed(c.Context(), usuarioID)
	if err != nil {
		return errorInterno(c, err)
	}
	if !tieneNodo {
		// Cuenta sin red propia: ledger vacío en lugar de un error.
		return c.JSON(models.PaginaComisiones{
			Data:      []models.Comision{},
			Page:      1,
			Limit:     10,
			Total:     0,
			Pendiente: 0,
			Pagado:    0,
		})
	}

	pagina, limite := paginaDe(c)
	resultado, err := ctrl.repo.ObtenerComisionesMias(c.Context(), nodoID, pagina, limite)
	if err != nil {
		return errorInterno(c, err)
	}
	return c.JSON(resultado)
}

// GetRedMia devuelve el subárbol de referidos que cuelga del nodo del token,
// no el árbol global de la semilla, junto con el código de invitación propio.
func (ctrl *RedController) GetRedMia(c *fiber.Ctx) error {
	usuarioID, ok := usuarioIDDelToken(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(models.APIError{
			Status:  fiber.StatusUnauthorized,
			Message: "Sesión no válida o expirada",
		})
	}

	nodoID, tieneNodo, err := ctrl.repo.NodoDeRed(c.Context(), usuarioID)
	if err != nil {
		return errorInterno(c, err)
	}
	if !tieneNodo {
		// Sin nodo de red devolvemos una respuesta vacía y legible.
		return c.JSON(models.RedPropia{Red: nil, Codigo: "", ReferidoID: 0})
	}

	red, err := ctrl.repo.ObtenerRedPropia(c.Context(), nodoID)
	if err != nil {
		return errorInterno(c, err)
	}
	return c.JSON(red)
}
