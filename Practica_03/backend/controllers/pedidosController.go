// Este archivo agrupa la consulta y gestión de pedidos (Fase 3 del back-office).
// El cliente solo ve SU historial; el administrador gestiona el ciclo de vida.
package controllers

import (
	// Importamos strconv para convertir el ID de la URL a número.
	"strconv"
	// Importamos strings para validar los parámetros de consulta.
	"strings"

	// Importamos el framework Fiber para manejar la petición y la respuesta.
	"github.com/gofiber/fiber/v2"
	// Importamos nuestros modelos para Pedido, APIError y Pagina.
	"multicatalogo-backend/models"
)

// estadosPedido lista los valores admitidos por el ciclo de vida del pedido.
var estadosPedido = map[string]bool{
	"pendiente":  true,
	"preparando": true,
	"enviado":    true,
	"entregado":  true,
	"cancelado":  true,
}

// usuarioIDDelToken recupera el claim sub inyectado por RequireAuth.
func usuarioIDDelToken(c *fiber.Ctx) (int, bool) {
	id, ok := c.Locals("usuario_id").(int)
	return id, ok && id > 0
}

// emailDelToken recupera el claim email inyectado por RequireAuth.
func emailDelToken(c *fiber.Ctx) string {
	email, _ := c.Locals("email").(string)
	return strings.TrimSpace(email)
}

// esAdminDelToken indica si el token pertenece a un administrador.
func esAdminDelToken(c *fiber.Ctx) bool {
	rol, _ := c.Locals("rol").(string)
	return rol == "admin"
}

// paginaDe la lee los parámetros ?page y ?limit con valores por defecto.
func paginaDe(c *fiber.Ctx) (int, int) {
	return max(atoiSeguro(c.Query("page"), 1), 1), max(atoiSeguro(c.Query("limit"), 10), 1)
}

// atoiSeguro convierte un texto a entero devolviendo un valor por defecto.
func atoiSeguro(texto string, porDefecto int) int {
	n, err := strconv.Atoi(strings.TrimSpace(texto))
	if err != nil {
		return porDefecto
	}
	return n
}

// GetPedidosMios devuelve el historial de pedidos del usuario del token.
// Un cliente jamás ve pedidos de otro: el filtro lo impone el claim sub.
func (ctrl *PedidoController) GetPedidosMios(c *fiber.Ctx) error {
	usuarioID, ok := usuarioIDDelToken(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(models.APIError{
			Status:  fiber.StatusUnauthorized,
			Message: "Sesión no válida o expirada",
		})
	}

	pagina, limite := paginaDe(c)
	resultado, err := ctrl.repo.ObtenerPedidosMios(c.Context(), usuarioID, pagina, limite)
	if err != nil {
		return errorInterno(c, err)
	}
	return c.JSON(resultado)
}

// GetPedidosAdmin lista todos los pedidos con filtro opcional ?estado=.
func (ctrl *PedidoController) GetPedidosAdmin(c *fiber.Ctx) error {
	estado := strings.TrimSpace(c.Query("estado"))
	if estado != "" && !estadosPedido[estado] {
		return c.Status(fiber.StatusBadRequest).JSON(models.APIError{
			Status:  fiber.StatusBadRequest,
			Message: "Estado de pedido no válido",
			Details: fiber.Map{"estado": estado},
		})
	}

	pagina, limite := paginaDe(c)
	resultado, err := ctrl.repo.ObtenerPedidosAdmin(c.Context(), estado, pagina, limite)
	if err != nil {
		return errorInterno(c, err)
	}
	return c.JSON(resultado)
}

// GetPedidoDetalle devuelve cabecera y líneas de un pedido concreto.
// El dueño (por token) o un administrador pueden consultarlo; nadie más.
func (ctrl *PedidoController) GetPedidoDetalle(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.APIError{
			Status:  fiber.StatusBadRequest,
			Message: "El ID del pedido debe ser un número entero",
			Details: fiber.Map{"id": c.Params("id")},
		})
	}

	usuarioID, ok := usuarioIDDelToken(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(models.APIError{
			Status:  fiber.StatusUnauthorized,
			Message: "Sesión no válida o expirada",
		})
	}

	detalle, err := ctrl.repo.ObtenerPedidoDetalle(c.Context(), id, usuarioID, emailDelToken(c), esAdminDelToken(c))
	if err != nil {
		if err == models.ErrPedidoNoEncontrado {
			return c.Status(fiber.StatusNotFound).JSON(models.APIError{
				Status:  fiber.StatusNotFound,
				Message: "Pedido no encontrado",
				Details: fiber.Map{"id": id},
			})
		}
		return errorInterno(c, err)
	}
	return c.JSON(detalle)
}

// CambiarEstadoPedido aplica una transición del ciclo de vida (solo admin).
// La cancelación revierte stock y comisiones dentro de la misma transacción.
func (ctrl *PedidoController) CambiarEstadoPedido(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.APIError{
			Status:  fiber.StatusBadRequest,
			Message: "El ID del pedido debe ser un número entero",
			Details: fiber.Map{"id": c.Params("id")},
		})
	}

	var req struct {
		Estado string `json:"estado"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.APIError{
			Status:  fiber.StatusBadRequest,
			Message: "Cuerpo de petición inválido",
		})
	}

	estado := strings.TrimSpace(strings.ToLower(req.Estado))
	if !estadosPedido[estado] {
		return c.Status(fiber.StatusBadRequest).JSON(models.APIError{
			Status:  fiber.StatusBadRequest,
			Message: "Estado de pedido no válido",
			Details: fiber.Map{"estado": req.Estado},
		})
	}

	actualizado, err := ctrl.repo.CambiarEstadoPedido(c.Context(), id, estado)
	if err != nil {
		if err == models.ErrPedidoNoEncontrado {
			return c.Status(fiber.StatusNotFound).JSON(models.APIError{
				Status:  fiber.StatusNotFound,
				Message: "Pedido no encontrado",
				Details: fiber.Map{"id": id},
			})
		}
		if err == models.ErrTransicionInvalida {
			return c.Status(fiber.StatusBadRequest).JSON(models.APIError{
				Status:  fiber.StatusBadRequest,
				Message: "Transición de estado inválida para el pedido",
				Details: fiber.Map{"id": id, "estado": estado},
			})
		}
		return errorInterno(c, err)
	}

	// La cancelación devuelve stock al catálogo: invalidamos su caché.
	if estado == "cancelado" {
		ctrl.cache.Eliminar(c.Context(), "productos:all")
	}

	return c.JSON(actualizado)
}

// PedidoController agrupa los manejadores de pedidos con el repositorio y la caché inyectados al arrancar.
type PedidoController struct {
	repo  DBRepository
	cache CacheRepository
}

// NuevoPedidoController construye el controlador con las dependencias que recibirá.
func NuevoPedidoController(repo DBRepository, cache CacheRepository) *PedidoController {
	return &PedidoController{repo: repo, cache: cache}
}
