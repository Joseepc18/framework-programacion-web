// Declaramos el paquete controllers para agrupar las funciones que manejan la lógica de negocio de las rutas.
package controllers

import (
	// Importamos bcrypt para la verificación criptográfica de contraseñas.
	"golang.org/x/crypto/bcrypt"

	// Importamos el framework Fiber para tener acceso al contexto (c *fiber.Ctx) de la petición HTTP.
	"github.com/gofiber/fiber/v2"
	// Importamos el middleware para generar el token JWT.
	"multicatalogo-backend/middleware"
	// Importamos nuestro paquete de modelos para usar LoginRequest y APIError.
	"multicatalogo-backend/models"
)

// Login es la función controladora que se ejecutará cuando el cliente envíe sus credenciales.
func (ctrl *AuthController) Login(c *fiber.Ctx) error {
	// Creamos una variable 'req' del tipo LoginRequest para almacenar los datos.
	var req models.LoginRequest

	// Intentamos parsear el cuerpo JSON entrante.
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.APIError{
			Status:  fiber.StatusBadRequest,
			Message: "Cuerpo de petición inválido",
		})
	}

	// Validación central de campos obligatorios.
	if vacios := camposVacios(campo{"email", req.Email}, campo{"password", req.Password}); len(vacios) > 0 {
		return respuestaCamposVacios(c, "El email y la contraseña son obligatorios", vacios)
	}

	// Buscamos el usuario por correo electrónico en la base de datos.
	usuario, ok, err := ctrl.repo.BuscarUsuarioPorEmail(req.Email)
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
	token, err := middleware.GenerarToken(ctrl.secreto, usuario.ID, usuario.Email, usuario.Rol)
	if err != nil {
		return errorInterno(c, err)
	}

	// Retornamos HTTP 200 con el token JWT real, email y rol del usuario.
	return c.JSON(fiber.Map{"token": token, "email": usuario.Email, "rol": usuario.Rol})
}

// AuthController agrupa login y registro con sus dependencias inyectadas al arrancar.
type AuthController struct {
	repo    DBRepository
	secreto []byte
	mailer  Mailer
}

// NuevoAuthController construye el controlador con el repositorio, la clave JWT y el mailer.
func NuevoAuthController(repo DBRepository, secreto []byte, mailer Mailer) *AuthController {
	return &AuthController{repo: repo, secreto: secreto, mailer: mailer}
}
