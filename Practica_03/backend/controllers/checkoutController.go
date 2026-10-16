// Declaramos el paquete controllers para la gestión del endpoint de compras.
package controllers

import (
	"context"
	"strings"

	"github.com/gofiber/fiber/v2"
	"multicatalogo-backend/models"
)

// Checkout maneja la compra transaccional validando stock e identidad obtenida del JWT.
func (ctrl *CheckoutController) Checkout(c *fiber.Ctx) error {
	var req models.CheckoutRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.APIError{
			Status:  fiber.StatusBadRequest,
			Message: "Cuerpo de petición inválido",
		})
	}

	// Identidad completa del comprador (claim sub + email) desde el token validado
	emailAuth, ok := c.Locals("email").(string)
	if !ok || strings.TrimSpace(emailAuth) == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(models.APIError{
			Status:  fiber.StatusUnauthorized,
			Message: "Sesión no válida o expirada",
		})
	}

	// El claim sub (ID de usuario) es el dueño real que se persiste en pedidos.usuario_id
	usuarioID, _ := c.Locals("usuario_id").(int)

	if len(req.Items) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(models.APIError{
			Status:  fiber.StatusBadRequest,
			Message: "El pedido debe contener al menos un producto",
		})
	}

	// Ejecutamos la transacción en el repositorio (FOR UPDATE, recálculo de precios y stock)
	var duenoID *int
	if usuarioID > 0 {
		duenoID = &usuarioID
	}
	res, err := ctrl.repo.CrearPedido(c.Context(), duenoID, emailAuth, req)
	if err != nil {
		if strings.Contains(err.Error(), "Stock insuficiente") ||
			strings.Contains(err.Error(), "Producto desactivado") ||
			strings.Contains(err.Error(), "producto no encontrado") ||
			err == models.ErrStockInsuficiente ||
			err == models.ErrCantidadInvalida ||
			err == models.ErrCarritoVacio {
			return c.Status(fiber.StatusBadRequest).JSON(models.APIError{
				Status:  fiber.StatusBadRequest,
				Message: err.Error(),
			})
		}
		return errorInterno(c, err)
	}

	// Invalidamos la caché en Redis tras la mutación de inventario
	ctrl.cache.Eliminar(context.Background(), "productos:all", "red:tree:"+emailAuth)

	ctrl.mailer.EnviarConfirmacionPedido(c.Context(), emailAuth, res.Numero, res.Total)

	// Disparamos la actualización del upline en una goroutine con contexto propio
	go func(email string, total float64) {
		ctx := context.Background()
		_ = ctrl.repo.ActualizarComisionesUpline(ctx, email, total)
	}(emailAuth, res.Total)

	return c.Status(fiber.StatusCreated).JSON(res)
}

// CheckoutController agrupa los manejadores de su área con el repositorio y la caché inyectados al arrancar.
type CheckoutController struct {
	repo   DBRepository
	cache  CacheRepository
	mailer Mailer
}

// NuevoCheckoutController construye el controlador con el repositorio, la caché y el mailer.
func NuevoCheckoutController(repo DBRepository, cache CacheRepository, mailer Mailer) *CheckoutController {
	return &CheckoutController{repo: repo, cache: cache, mailer: mailer}
}
