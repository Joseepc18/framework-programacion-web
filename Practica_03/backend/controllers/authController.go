// Declaramos el paquete controllers para agrupar las funciones que manejan la lógica de negocio de las rutas.
package controllers

import (
	// Importamos strconv para armar el token ficticio a partir del ID del usuario.
	"strconv"
	// Importamos strings para limpiar espacios en blanco de los campos recibidos.
	"strings"

	// Importamos el framework Fiber para tener acceso al contexto (c *fiber.Ctx) de la petición HTTP.
	"github.com/gofiber/fiber/v2"
	// Importamos nuestro paquete de modelos para usar LoginRequest y APIError.
	"multicatalogo-backend/models"
	// Importamos el repositorio para consultar las cuentas registradas.
	"multicatalogo-backend/repository"
)

// Login es la función controladora que se ejecutará cuando el cliente envíe sus credenciales.
func Login(c *fiber.Ctx) error {
	// Creamos una variable 'req' del tipo LoginRequest (ubicada en nuestro paquete models) para almacenar los datos.
	var req models.LoginRequest

	// Intentamos parsear (transformar) el cuerpo JSON entrante y guardarlo en la variable 'req'.
	if err := c.BodyParser(&req); err != nil {
		// Si ocurre un error al parsear (ej. JSON mal formado), retornamos un estado HTTP 400 (Bad Request).
		return c.Status(fiber.StatusBadRequest).JSON(models.APIError{
			Status:  fiber.StatusBadRequest,
			Message: "Cuerpo de petición inválido",
		})
	}

	// Acumulamos los campos obligatorios que llegaron vacíos (o solo con espacios).
	camposVacios := []string{}
	if strings.TrimSpace(req.Email) == "" {
		camposVacios = append(camposVacios, "email")
	}
	if strings.TrimSpace(req.Password) == "" {
		camposVacios = append(camposVacios, "password")
	}

	// Si falta algún campo, retornamos HTTP 400 indicando en Details cuáles faltan, sin consultar el repositorio.
	if len(camposVacios) > 0 {
		return c.Status(fiber.StatusBadRequest).JSON(models.APIError{
			Status:  fiber.StatusBadRequest,
			Message: "El email y la contraseña son obligatorios",
			Details: fiber.Map{"camposVacios": camposVacios},
		})
	}

	// Consultamos al repositorio si existe una cuenta con esas credenciales.
	usuario, ok, err := repository.BuscarUsuarioPorCredenciales(req.Email, req.Password)
	if err != nil {
		// Si la consulta falla, respondemos HTTP 500.
		return errorInterno(c, err)
	}
	if !ok {
		// Si las credenciales son incorrectas, retornamos un estado HTTP 401 (No autorizado).
		return c.Status(fiber.StatusUnauthorized).JSON(models.APIError{
			Status:  fiber.StatusUnauthorized,
			Message: "Credenciales incorrectas",
		})
	}

	// Generamos un token ficticio a partir del ID (la autenticación real llegará en una fase posterior).
	token := "fake-jwt-token-" + strconv.Itoa(usuario.ID)

	// Si las credenciales son válidas, retornamos HTTP 200 con el mismo JSON que ya espera el frontend.
	return c.JSON(fiber.Map{"token": token, "email": usuario.Email, "rol": usuario.Rol})
}
