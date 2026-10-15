// Seguimos dentro del paquete controllers; este archivo implementa el microservicio de checkout.
package controllers

import (
	// Importamos context para el contexto propio de la Goroutine.
	"context"
	// Importamos log para registrar el resultado del trabajo en segundo plano.
	"log"
	// Importamos strings para limpiar espacios en blanco de los campos recibidos.
	"strings"

	// Importamos Fiber para manejar la respuesta HTTP.
	"github.com/gofiber/fiber/v2"
	// Importamos nuestro paquete de modelos para usar CheckoutRequest y APIError.
	"multicatalogo-backend/models"
)

// CheckoutController procesa la compra que envía el frontend. Recibe inyectados el repositorio
// (transacciones ACID en PostgreSQL) y la caché (invalidación de la red del usuario).
type CheckoutController struct {
	// repo es el acceso a PostgreSQL.
	repo DBRepository
	// cache es el almacenamiento en caché (Redis).
	cache CacheRepository
}

// NuevoCheckoutController construye el controlador con las dependencias que recibirá.
func NuevoCheckoutController(repo DBRepository, cache CacheRepository) *CheckoutController {
	return &CheckoutController{repo: repo, cache: cache}
}

// Procesar atiende POST /api/checkout: valida el pedido, lo guarda en una transacción ACID
// y lanza las comisiones de la red en segundo plano.
func (ctrl *CheckoutController) Procesar(c *fiber.Ctx) error {
	// Parseamos el JSON que envía Checkout.tsx (cliente, método de pago e items del carrito).
	var req models.CheckoutRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.APIError{
			Status:  fiber.StatusBadRequest,
			Message: "Cuerpo de petición inválido",
		})
	}

	// Validamos los campos obligatorios antes de tocar la base de datos.
	faltantes := []string{}
	if strings.TrimSpace(req.Cliente.Nombre) == "" {
		faltantes = append(faltantes, "cliente.nombre")
	}
	if strings.TrimSpace(req.Cliente.Email) == "" {
		faltantes = append(faltantes, "cliente.email")
	}
	if strings.TrimSpace(req.MetodoPago) == "" {
		faltantes = append(faltantes, "metodoPago")
	}
	if len(req.Items) == 0 {
		faltantes = append(faltantes, "items")
	}

	if len(faltantes) > 0 {
		return c.Status(fiber.StatusBadRequest).JSON(models.APIError{
			Status:  fiber.StatusBadRequest,
			Message: "Faltan datos obligatorios del pedido",
			Details: fiber.Map{"camposVacios": faltantes},
		})
	}

	// Cada item debe traer una cantidad positiva: no se compran cero unidades.
	for _, item := range req.Items {
		if item.Cantidad <= 0 {
			return c.Status(fiber.StatusBadRequest).JSON(models.APIError{
				Status:  fiber.StatusBadRequest,
				Message: "La cantidad de cada producto debe ser mayor que cero",
				Details: fiber.Map{"producto": item.ID},
			})
		}
	}

	// Guardamos el pedido: el repositorio abre la transacción y hace COMMIT o ROLLBACK.
	pedido, err := ctrl.repo.RegistrarPedido(req)
	if err != nil {
		// Si falló cualquier paso, la transacción ya se revirtió y no queda ningún pedido a medias.
		return errorInterno(c, err)
	}

	// La cuenta autenticada identifica el nodo del comprador en la red multinivel.
	// Si el frontend no la envía, se usa el correo del formulario.
	emailComprador := strings.TrimSpace(req.UsuarioEmail)
	if emailComprador == "" {
		emailComprador = strings.TrimSpace(req.Cliente.Email)
	}

	// Paso 4: el cálculo de comisiones no puede hacer esperar al cliente en la pantalla de pago.
	// Con "go" se ejecuta en una Goroutine aparte, después de que el COMMIT ya fue exitoso.
	go ctrl.calcularComisiones(emailComprador, pedido.Total)

	// Respondemos de inmediato con el pedido confirmado (201 Created).
	return c.Status(fiber.StatusCreated).JSON(pedido)
}

// calcularComisiones se ejecuta en segundo plano: acredita las comisiones de los niveles 1, 2 y 3
// y luego invalida la caché de la red de los usuarios afectados (Paso 5).
func (ctrl *CheckoutController) calcularComisiones(emailComprador string, total float64) {
	// La Goroutine no puede usar el contexto de Fiber: la petición HTTP ya terminó.
	ctx := context.Background()

	// El repositorio reparte las comisiones y devuelve los correos cuyo árbol cambió.
	afectados, err := ctrl.repo.AcreditarComisiones(emailComprador, total)
	if err != nil {
		log.Printf("no se pudieron acreditar las comisiones de %s: %v", emailComprador, err)
		return
	}

	// Invalidamos la clave red:tree:{email} de cada usuario afectado para que su próximo
	// Dashboard consulte la base de datos y muestre los KPIs actualizados.
	claves := make([]string, 0, len(afectados))
	for _, email := range afectados {
		claves = append(claves, "red:tree:"+email)
	}
	ctrl.cache.Eliminar(ctx, claves...)

	log.Printf("comisiones de %.2f acreditadas para %s; caché invalidada de %v", total, emailComprador, afectados)
}
