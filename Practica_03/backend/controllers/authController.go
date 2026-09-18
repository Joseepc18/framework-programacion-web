// Declaramos el paquete controllers para agrupar las funciones que manejan la lógica de negocio de las rutas.
package controllers

import (
	// Importamos el framework Fiber para tener acceso al contexto (c *fiber.Ctx) de la petición HTTP.
	"github.com/gofiber/fiber/v2"
	// Importamos nuestro paquete de modelos para poder usar la estructura LoginRequest.
	"multicatalogo-backend/models"
)

// Login es la función controladora que se ejecutará cuando el cliente envíe sus credenciales.
func Login(c *fiber.Ctx) error {
	// Creamos una variable 'req' del tipo LoginRequest (ubicada en nuestro paquete models) para almacenar los datos.
	var req models.LoginRequest
	
	// Intentamos parsear (transformar) el cuerpo JSON entrante y guardarlo en la variable 'req'.
	if err := c.BodyParser(&req); err != nil {
		// Si ocurre un error al parsear (ej. JSON mal formado), retornamos un estado HTTP 400 (Bad Request).
		return c.Status(400).JSON(fiber.Map{"error": "Cuerpo de petición inválido"})
	}

	// Evaluamos si el email y la contraseña coinciden con las credenciales predefinidas.
	if req.Email == "admin@upse.edu.ec" && req.Password == "123456" {
		// Si coinciden, retornamos un estado HTTP 200 (por defecto) enviando un JSON con un token ficticio y el correo.
		return c.JSON(fiber.Map{"token": "fake-jwt-token-123", "email": req.Email})
	}
	
	// Si las credenciales son incorrectas, retornamos un estado HTTP 401 (No autorizado) con un mensaje de error.
	return c.Status(401).JSON(fiber.Map{"error": "Credenciales incorrectas"})
}