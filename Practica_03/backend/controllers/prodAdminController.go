// Este archivo agrupa la administración del catálogo (Fase 2 del back-office).
// Todas sus rutas exigen sesión y rol admin (RequireAuth + RequireAdmin).
package controllers

import (
	// Importamos context para el contexto de la invalidación de caché.
	"context"
	// Importamos strconv para convertir el ID de la URL a número.
	"strconv"
	// Importamos strings para validar y limpiar los campos recibidos.
	"strings"

	// Importamos el framework Fiber para manejar la petición y la respuesta.
	"github.com/gofiber/fiber/v2"
	// Importamos config para invalidar la caché de Redis tras cada escritura.
	"multicatalogo-backend/config"
	// Importamos nuestros modelos para Producto y APIError.
	"multicatalogo-backend/models"
	// Importamos el repositorio para leer y escribir el catálogo.
	"multicatalogo-backend/repository"
)

// invalidarCacheCatalogo elimina la clave del catálogo público para que la
// siguiente lectura (cache-aside) vuelva a consultar PostgreSQL.
func invalidarCacheCatalogo() {
	if config.Cache != nil {
		_ = config.Cache.Del(context.Background(), "productos:all").Err()
	}
}

// GetProductosAdmin devuelve una página de productos (activos e inactivos) con
// búsqueda server-side mediante ?q= (ILIKE sobre nombre, categoría y descripción).
func GetProductosAdmin(c *fiber.Ctx) error {
	pagina, limite := paginaDe(c)
	resultado, err := repository.ObtenerProductosPagina(c.Context(), c.Query("q"), false, pagina, limite)
	if err != nil {
		return errorInterno(c, err)
	}
	return c.JSON(resultado)
}

// CrearProducto valida y da de alta un producto del catálogo (POST /api/productos).
func CrearProducto(c *fiber.Ctx) error {
	var p models.Producto
	if err := c.BodyParser(&p); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.APIError{
			Status:  fiber.StatusBadRequest,
			Message: "Cuerpo de petición inválido",
		})
	}

	if msg := validarProducto(p); msg != "" {
		return c.Status(fiber.StatusBadRequest).JSON(models.APIError{
			Status:  fiber.StatusBadRequest,
			Message: msg,
		})
	}

	// Normalizamos: la galería vacía queda como lista vacía, no como null.
	if p.Galeria == nil {
		p.Galeria = []string{}
	}
	// El stock no puede nacer negativo.
	if p.Stock < 0 {
		p.Stock = 0
	}

	creado, err := repository.CrearProducto(c.Context(), p)
	if err != nil {
		return errorInterno(c, err)
	}

	invalidarCacheCatalogo()
	return c.Status(fiber.StatusCreated).JSON(creado)
}

// ActualizarProducto edita un producto existente (PUT /api/productos/:id).
func ActualizarProducto(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.APIError{
			Status:  fiber.StatusBadRequest,
			Message: "El ID del producto debe ser un número entero",
			Details: fiber.Map{"id": c.Params("id")},
		})
	}

	var p models.Producto
	if err := c.BodyParser(&p); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.APIError{
			Status:  fiber.StatusBadRequest,
			Message: "Cuerpo de petición inválido",
		})
	}

	if msg := validarProducto(p); msg != "" {
		return c.Status(fiber.StatusBadRequest).JSON(models.APIError{
			Status:  fiber.StatusBadRequest,
			Message: msg,
		})
	}
	if p.Galeria == nil {
		p.Galeria = []string{}
	}
	if p.Stock < 0 {
		p.Stock = 0
	}

	existe, err := repository.ActualizarProducto(c.Context(), id, p)
	if err != nil {
		return errorInterno(c, err)
	}
	if !existe {
		return c.Status(fiber.StatusNotFound).JSON(models.APIError{
			Status:  fiber.StatusNotFound,
			Message: "Producto no encontrado",
			Details: fiber.Map{"id": id},
		})
	}

	invalidarCacheCatalogo()

	actualizado, _, err := repository.ObtenerProductoPorID(id)
	if err != nil {
		// La edición ya se guardó; devolvemos solo el ID si la relectura falla.
		return c.JSON(fiber.Map{"id": id, "message": "Producto actualizado"})
	}
	return c.JSON(actualizado)
}

// DesactivarProducto aplica el borrado lógico (DELETE /api/productos/:id).
// No borra la fila: los pedido_items históricos siguen siendo válidos.
func DesactivarProducto(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.APIError{
			Status:  fiber.StatusBadRequest,
			Message: "El ID del producto debe ser un número entero",
			Details: fiber.Map{"id": c.Params("id")},
		})
	}

	// ?activar=true reactiva el producto; por defecto lo desactiva.
	activo := c.Query("activar") == "true"

	existe, err := repository.CambiarEstadoProducto(c.Context(), id, activo)
	if err != nil {
		return errorInterno(c, err)
	}
	if !existe {
		return c.Status(fiber.StatusNotFound).JSON(models.APIError{
			Status:  fiber.StatusNotFound,
			Message: "Producto no encontrado",
			Details: fiber.Map{"id": id},
		})
	}

	invalidarCacheCatalogo()

	mensaje := "Producto desactivado del catálogo"
	if activo {
		mensaje = "Producto activado en el catálogo"
	}
	return c.JSON(fiber.Map{"id": id, "activo": activo, "message": mensaje})
}

// validarProducto aplica las mismas reglas en alta y edición.
func validarProducto(p models.Producto) string {
	camposVacios := []string{}
	if strings.TrimSpace(p.Nombre) == "" {
		camposVacios = append(camposVacios, "nombre")
	}
	if strings.TrimSpace(p.Categoria) == "" {
		camposVacios = append(camposVacios, "categoria")
	}
	if strings.TrimSpace(p.Img) == "" {
		camposVacios = append(camposVacios, "img")
	}
	if len(camposVacios) > 0 {
		return "Nombre, categoría e imagen son obligatorios"
	}
	if p.Precio <= 0 {
		return "El precio debe ser mayor que cero"
	}
	return ""
}
