// Seguimos dentro del paquete controllers, ya que maneja la lógica de otra sección del negocio.
package controllers

import (
	// Importamos context para propagar el contexto de las operaciones de caché.
	"context"
	// Importamos encoding/json para serializar y deserializar el catálogo cacheado.
	"encoding/json"
	// Importamos strconv para convertir el parámetro de la URL (texto) a número entero.
	"strconv"
	// Importamos time para definir el tiempo de vida (TTL) de la caché.
	"time"

	// Importamos Fiber para manejar la respuesta HTTP.
	"github.com/gofiber/fiber/v2"
	// Importamos el paquete config para acceder al cliente de Redis.
	// Importamos nuestro paquete de modelos para usar la estructura APIError.
	"multicatalogo-backend/models"
)

// GetProductos es la función controladora encargada de devolver el catálogo de artículos.
// Aplica el patrón Cache-Aside: primero busca en Redis y solo consulta PostgreSQL si no hay caché.
func (ctrl *ProductoController) GetProductos(c *fiber.Ctx) error {
	// Con ?q=, ?page= o ?limit= respondemos paginado desde PostgreSQL ({data,page,limit,total}).
	// Sin parámetros se mantiene el catálogo completo cacheado que consume la tienda.
	if c.Query("q") != "" || c.Query("page") != "" || c.Query("limit") != "" {
		pagina, limite := paginaDe(c)
		resultado, err := ctrl.repo.ObtenerProductosPagina(c.Context(), c.Query("q"), true, pagina, limite)
		if err != nil {
			return errorInterno(c, err)
		}
		return c.JSON(resultado)
	}

	// Clave y tiempo de vida con los que guardamos el catálogo en Redis.
	const clave = "productos:all"
	const ttl = 5 * time.Minute

	ctx := context.Background()

	// 1) Cache Hit: si el catálogo está en Redis, lo devolvemos sin tocar PostgreSQL.
	if datos, ok := ctrl.cache.Obtener(ctx, clave); ok {
		c.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
		return c.Send(datos)
	}

	// 2) Cache Miss: consultamos el catálogo en la base de datos.
	productos, err := ctrl.repo.ObtenerProductos()
	if err != nil {
		// Si la consulta falla, respondemos HTTP 500.
		return errorInterno(c, err)
	}

	// 3) Guardamos el resultado serializado en Redis durante 5 minutos para las siguientes peticiones.
	if datos, err := json.Marshal(productos); err == nil {
		ctrl.cache.Guardar(ctx, clave, datos, ttl)
	}

	// Fiber convierte automáticamente el slice de estructuras a formato JSON y lo envía como respuesta al cliente.
	return c.JSON(productos)
}

// GetProductoPorID devuelve un único producto según el ID recibido en la URL (/api/productos/:id).
func (ctrl *ProductoController) GetProductoPorID(c *fiber.Ctx) error {
	// Convertimos el parámetro "id" a entero; si el usuario envía texto, Atoi devuelve un error.
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		// Si el ID no es numérico, retornamos HTTP 400 indicando el valor recibido.
		return c.Status(fiber.StatusBadRequest).JSON(models.APIError{
			Status:  fiber.StatusBadRequest,
			Message: "El ID del producto debe ser un número entero",
			Details: fiber.Map{"id": c.Params("id")},
		})
	}

	// Buscamos el producto en el repositorio.
	producto, ok, err := ctrl.repo.ObtenerProductoPorID(id)
	if err != nil {
		// Si la consulta falla, respondemos HTTP 500.
		return errorInterno(c, err)
	}
	if !ok {
		// Si no existe un producto con ese ID, retornamos HTTP 404 (No encontrado).
		return c.Status(fiber.StatusNotFound).JSON(models.APIError{
			Status:  fiber.StatusNotFound,
			Message: "Producto no encontrado",
			Details: fiber.Map{"id": id},
		})
	}

	// Si existe, lo devolvemos con HTTP 200.
	return c.JSON(producto)
}

// ProductoController agrupa los manejadores de su área con el repositorio y la caché inyectados al arrancar.
type ProductoController struct {
	repo  DBRepository
	cache CacheRepository
}

// NuevoProductoController construye el controlador con las dependencias que recibirá.
func NuevoProductoController(repo DBRepository, cache CacheRepository) *ProductoController {
	return &ProductoController{repo: repo, cache: cache}
}
