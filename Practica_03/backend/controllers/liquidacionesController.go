// Este archivo agrupa las liquidaciones a distribuidores (Fase 5).
// El admin paga la deuda acumulada; el distribuidor solo consulta lo suyo.
package controllers

import (
	// Importamos el framework Fiber para manejar la petición y la respuesta.
	"github.com/gofiber/fiber/v2"
	// Importamos nuestros modelos para APIError.
	"multicatalogo-backend/models"
	// Importamos el repositorio para crear y leer liquidaciones.
	"multicatalogo-backend/repository"
)

// CrearLiquidacion liquida el saldo pendiente de un distribuidor (solo admin).
// El cuerpo espera {"referido_id": N}; si no queda deuda responde 409.
func CrearLiquidacion(c *fiber.Ctx) error {
	var req struct {
		ReferidoID int `json:"referido_id"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.APIError{
			Status:  fiber.StatusBadRequest,
			Message: "Cuerpo de petición inválido",
		})
	}
	// El nodo raíz de la semilla tiene id 0, por eso solo descartamos negativos.
	if req.ReferidoID < 0 {
		return c.Status(fiber.StatusBadRequest).JSON(models.APIError{
			Status:  fiber.StatusBadRequest,
			Message: "Debe indicar un distribuidor válido (referido_id)",
		})
	}

	liquidacion, err := repository.CrearLiquidacion(c.Context(), req.ReferidoID)
	if err != nil {
		if err == repository.ErrSinComisionesPendientes {
			return c.Status(fiber.StatusConflict).JSON(models.APIError{
				Status:  fiber.StatusConflict,
				Message: "El distribuidor no tiene comisiones pendientes",
				Details: fiber.Map{"referido_id": req.ReferidoID},
			})
		}
		return errorInterno(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(liquidacion)
}

// GetLiquidaciones devuelve el histórico global de pagos (solo admin).
func GetLiquidaciones(c *fiber.Ctx) error {
	pagina, limite := paginaDe(c)
	resultado, err := repository.ObtenerLiquidaciones(c.Context(), pagina, limite)
	if err != nil {
		return errorInterno(c, err)
	}
	return c.JSON(resultado)
}

// GetLiquidacionesMias devuelve los pagos recibidos por el usuario del token.
func GetLiquidacionesMias(c *fiber.Ctx) error {
	usuarioID, ok := usuarioIDDelToken(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(models.APIError{
			Status:  fiber.StatusUnauthorized,
			Message: "Sesión no válida o expirada",
		})
	}

	nodoID, tieneNodo, err := repository.NodoDeRed(c.Context(), usuarioID)
	if err != nil {
		return errorInterno(c, err)
	}
	if !tieneNodo {
		return c.JSON(models.Pagina{Data: []models.Liquidacion{}, Page: 1, Limit: 10, Total: 0})
	}

	pagina, limite := paginaDe(c)
	resultado, err := repository.ObtenerLiquidacionesMias(c.Context(), nodoID, pagina, limite)
	if err != nil {
		return errorInterno(c, err)
	}
	return c.JSON(resultado)
}

// GetDeudas devuelve la cola de deuda por distribuidor (solo admin).
func GetDeudas(c *fiber.Ctx) error {
	deudas, err := repository.ObtenerDeudas(c.Context())
	if err != nil {
		return errorInterno(c, err)
	}
	return c.JSON(deudas)
}
