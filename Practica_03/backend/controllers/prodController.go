// Seguimos dentro del paquete controllers, ya que maneja la lógica de otra sección del negocio.
package controllers

import (
	// Importamos strconv para convertir el parámetro de la URL (texto) a número entero.
	"strconv"

	// Importamos Fiber para manejar la respuesta HTTP.
	"github.com/gofiber/fiber/v2"
	// Importamos nuestro paquete de modelos para usar la estructura APIError.
	"multicatalogo-backend/models"
	// Importamos el repositorio para consultar el catálogo.
	"multicatalogo-backend/repository"
)

// GetProductos es la función controladora encargada de devolver el catálogo de artículos.
func GetProductos(c *fiber.Ctx) error {
	// Consultamos el catálogo en la base de datos.
	productos, err := repository.ObtenerProductos()
	if err != nil {
		// Si la consulta falla, respondemos HTTP 500.
		return errorInterno(c, err)
	}
	// Fiber convierte automáticamente el slice de estructuras a formato JSON y lo envía como respuesta al cliente.
	return c.JSON(productos)
}

// GetProductoPorID devuelve un único producto según el ID recibido en la URL (/api/productos/:id).
func GetProductoPorID(c *fiber.Ctx) error {
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
	producto, ok, err := repository.ObtenerProductoPorID(id)
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
