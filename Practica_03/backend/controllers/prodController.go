// Seguimos dentro del paquete controllers, ya que maneja la lógica de otra sección del negocio.
package controllers

import (
	// Importamos Fiber para manejar la respuesta HTTP.
	"github.com/gofiber/fiber/v2"
	// Importamos nuestro paquete de modelos para usar la estructura Producto.
	"multicatalogo-backend/models"
)

// GetProductos es la función controladora encargada de devolver el catálogo de artículos.
func GetProductos(c *fiber.Ctx) error {
	// Declaramos e inicializamos un 'slice' (arreglo dinámico) usando nuestro modelo models.Producto.
	productos := []models.Producto{
		// Agregamos el primer producto con sus respectivos valores para ID, Nombre, Precio e Img.
		{ID: 1, Nombre: "Serum Revitalizante", Precio: 45.00, Img: "https://picsum.photos/seed/serum/600/400"},
		// Agregamos el segundo producto a la lista.
		{ID: 2, Nombre: "Crema Hidratante Pro", Precio: 32.50, Img: "https://picsum.photos/seed/crema/600/400"},
		// Agregamos el tercer producto a la lista.
		{ID: 3, Nombre: "Tónico Purificante", Precio: 28.00, Img: "https://picsum.photos/seed/tonico/600/400"},
		// Agregamos el cuarto producto a la lista.
		{ID: 4, Nombre: "Mascarilla Nocturna", Precio: 50.00, Img: "https://picsum.photos/seed/mascarilla/600/400"},
	}
	
	// Fiber convierte automáticamente el slice de estructuras a formato JSON y lo envía como respuesta al cliente.
	return c.JSON(productos)
}