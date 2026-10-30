package controllers

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
	"sigo-dental-backend/middleware"
	"sigo-dental-backend/models"
)

// AuthController atiende el inicio de sesión con sus dependencias inyectadas al arrancar.
type AuthController struct {
	repo    UsuarioRepository
	secreto []byte
	// hashFalso se compara cuando el correo no existe, para que responder tarde lo mismo
	// y no revelar por el tiempo de respuesta qué correos están registrados.
	hashFalso []byte
}

// NuevoAuthController construye el controlador con el repositorio y la clave de firma del JWT.
func NuevoAuthController(repo UsuarioRepository, secreto []byte) (*AuthController, error) {
	hashFalso, err := bcrypt.GenerateFromPassword([]byte("contraseña-que-no-corresponde-a-nadie"), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	return &AuthController{repo: repo, secreto: secreto, hashFalso: hashFalso}, nil
}

// credencialesIncorrectas responde 401 con el mismo mensaje sea cual sea el motivo
// (correo inexistente, contraseña errónea o cuenta desactivada).
func credencialesIncorrectas(c *fiber.Ctx) error {
	return c.Status(fiber.StatusUnauthorized).JSON(models.APIError{
		Status:  fiber.StatusUnauthorized,
		Message: "Credenciales incorrectas",
	})
}

// Login valida el correo y la contraseña contra la tabla usuarios y devuelve un JWT con el rol.
func (ctrl *AuthController) Login(c *fiber.Ctx) error {
	var req models.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.APIError{
			Status:  fiber.StatusBadRequest,
			Message: "Cuerpo de petición inválido",
		})
	}

	email := strings.TrimSpace(req.Email)
	if email == "" || req.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(models.APIError{
			Status:  fiber.StatusBadRequest,
			Message: "El email y la contraseña son obligatorios",
		})
	}

	usuario, existe, err := ctrl.repo.BuscarUsuarioPorEmail(c.Context(), email)
	if err != nil {
		return errorInterno(c, err)
	}

	hash := ctrl.hashFalso
	if existe {
		hash = []byte(usuario.PasswordHash)
	}
	coincide := bcrypt.CompareHashAndPassword(hash, []byte(req.Password)) == nil
	if !existe || !coincide || !usuario.Activo {
		return credencialesIncorrectas(c)
	}

	token, err := middleware.GenerarToken(ctrl.secreto, usuario.ID, usuario.Email, usuario.Rol)
	if err != nil {
		return errorInterno(c, err)
	}
	return c.JSON(fiber.Map{"token": token, "email": usuario.Email, "rol": usuario.Rol})
}
