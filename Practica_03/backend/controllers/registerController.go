// Este archivo agrupa la lógica del alta de cuentas públicas (registro de clientes).
package controllers

import (
	// Importamos strings para validar y limpiar los campos recibidos.
	"strings"

	// Importamos bcrypt para hashear la contraseña antes de guardarla.
	"golang.org/x/crypto/bcrypt"

	// Importamos el framework Fiber para manejar la petición y la respuesta.
	"github.com/gofiber/fiber/v2"
	// Importamos nuestros modelos para RegisterRequest y APIError.
	"multicatalogo-backend/models"
)

// Register crea una cuenta de cliente y, si llega código de referido,
// su nodo queda colgado del patrocinador dentro de la misma transacción.
// Respuestas: 201 creado, 400 datos faltantes o ref inválido, 409 email duplicado.
func (ctrl *AuthController) Register(c *fiber.Ctx) error {
	var req models.RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.APIError{
			Status:  fiber.StatusBadRequest,
			Message: "Cuerpo de petición inválido",
		})
	}

	// El código de invitación también puede llegar como query (?ref=CODIGO), igual que el link del frontend.
	if strings.TrimSpace(req.Ref) == "" {
		req.Ref = c.Query("ref")
	}

	// Validamos los campos obligatorios antes de tocar la base de datos.
	camposVacios := []string{}
	if strings.TrimSpace(req.Nombre) == "" {
		camposVacios = append(camposVacios, "nombre")
	}
	if strings.TrimSpace(req.Email) == "" {
		camposVacios = append(camposVacios, "email")
	}
	if strings.TrimSpace(req.Password) == "" {
		camposVacios = append(camposVacios, "password")
	}
	if len(camposVacios) > 0 {
		return c.Status(fiber.StatusBadRequest).JSON(models.APIError{
			Status:  fiber.StatusBadRequest,
			Message: "Nombre, correo y contraseña son obligatorios",
			Details: fiber.Map{"camposVacios": camposVacios},
		})
	}

	// Hasheamos la contraseña con bcrypt (el costo por defecto, 10 rondas).
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return errorInterno(c, err)
	}

	usuario, err := ctrl.repo.CrearUsuario(c.Context(), req, string(hash))
	if err != nil {
		// Mapeamos los errores de dominio a sus códigos HTTP.
		switch err {
		case models.ErrEmailYaExiste:
			return c.Status(fiber.StatusConflict).JSON(models.APIError{
				Status:  fiber.StatusConflict,
				Message: "El correo electrónico ya está registrado",
			})
		case models.ErrCodigoInvalido:
			return c.Status(fiber.StatusBadRequest).JSON(models.APIError{
				Status:  fiber.StatusBadRequest,
				Message: "El código de referido no existe",
				Details: fiber.Map{"ref": strings.ToUpper(strings.TrimSpace(req.Ref))},
			})
		case models.ErrRegistroInvalido:
			return c.Status(fiber.StatusBadRequest).JSON(models.APIError{
				Status:  fiber.StatusBadRequest,
				Message: "Nombre, correo y contraseña son obligatorios",
			})
		default:
			return errorInterno(c, err)
		}
	}

	// 201 Created: devolvemos la cuenta creada (sin la contraseña).
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"id":         usuario.ID,
		"email":      usuario.Email,
		"nombre":     usuario.Nombre,
		"rol":        usuario.Rol,
		"referidoId": usuario.ReferidoID,
		"message":    "Cuenta creada con éxito. Ya puede iniciar sesión.",
	})
}
