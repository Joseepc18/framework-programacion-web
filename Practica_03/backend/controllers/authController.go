// Declaramos el paquete controllers para agrupar las funciones que manejan la lógica de negocio de las rutas.
package controllers

import (
	// Importamos strings para limpiar espacios en blanco de los campos recibidos.
	"strings"

	// Importamos bcrypt para la verificación criptográfica de contraseñas.
	"golang.org/x/crypto/bcrypt"

	// Importamos el framework Fiber para tener acceso al contexto (c *fiber.Ctx) de la petición HTTP.
	"github.com/gofiber/fiber/v2"
	// Importamos el middleware para generar el token JWT.
	"multicatalogo-backend/middleware"
	// Importamos nuestro paquete de modelos para usar LoginRequest y APIError.
	"multicatalogo-backend/models"
	// Importamos el repositorio para consultar las cuentas registradas.
	"multicatalogo-backend/repository"
)

// Login es la función controladora que se ejecutará cuando el cliente envíe sus credenciales.
func Login(c *fiber.Ctx) error {
	// Creamos una variable 'req' del tipo LoginRequest para almacenar los datos.
	var req models.LoginRequest

	// Intentamos parsear el cuerpo JSON entrante.
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.APIError{
			Status:  fiber.StatusBadRequest,
			Message: "Cuerpo de petición inválido",
		})
	}

	// Acumulamos los campos obligatorios que llegaron vacíos.
	camposVacios := []string{}
	if strings.TrimSpace(req.Email) == "" {
		camposVacios = append(camposVacios, "email")
	}
	if strings.TrimSpace(req.Password) == "" {
		camposVacios = append(camposVacios, "password")
	}

	if len(camposVacios) > 0 {
		return c.Status(fiber.StatusBadRequest).JSON(models.APIError{
			Status:  fiber.StatusBadRequest,
			Message: "El email y la contraseña son obligatorios",
			Details: fiber.Map{"camposVacios": camposVacios},
		})
	}

	// Buscamos el usuario por correo electrónico en la base de datos.
	usuario, ok, err := repository.BuscarUsuarioPorEmail(req.Email)
	if err != nil {
		return errorInterno(c, err)
	}
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(models.APIError{
			Status:  fiber.StatusUnauthorized,
			Message: "Credenciales incorrectas",
		})
	}

	// Una cuenta desactivada por un administrador no puede iniciar sesión.
	if !usuario.Activo {
		return c.Status(fiber.StatusForbidden).JSON(models.APIError{
			Status:  fiber.StatusForbidden,
			Message: "Cuenta desactivada: contacte al administrador",
		})
	}

	// Comparamos el hash de la contraseña usando bcrypt.
	if err := bcrypt.CompareHashAndPassword([]byte(usuario.Password), []byte(req.Password)); err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(models.APIError{
			Status:  fiber.StatusUnauthorized,
			Message: "Credenciales incorrectas",
		})
	}

	// Generamos el token JWT real firmado por 2 horas.
	token, err := middleware.GenerarToken(usuario.ID, usuario.Email, usuario.Rol)
	if err != nil {
		return errorInterno(c, err)
	}

	// Retornamos HTTP 200 con el token JWT real, email y rol del usuario.
	return c.JSON(fiber.Map{"token": token, "email": usuario.Email, "rol": usuario.Rol})
}
