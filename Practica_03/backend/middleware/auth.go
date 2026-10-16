// Declaramos el paquete middleware para funciones interceptoras de peticiones HTTP.
package middleware

import (
	"context"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"multicatalogo-backend/models"
)

// CuentaActiva indica si la cuenta con ese ID sigue existiendo y habilitada.
// Si no se puede comprobar (error de base de datos), debe devolver false: se falla cerrado.
type CuentaActiva func(ctx context.Context, usuarioID int) bool

// GenerarToken crea un token firmado con HS256 y duración de 2 horas.
func GenerarToken(secreto []byte, userID int, email, rol string) (string, error) {
	claims := jwt.MapClaims{
		"sub":   userID,
		"email": email,
		"rol":   rol,
		"exp":   time.Now().Add(2 * time.Hour).Unix(),
		"iat":   time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secreto)
}

// noAutorizado responde 401 con el contrato estándar de errores.
func noAutorizado(c *fiber.Ctx, mensaje string) error {
	return c.Status(fiber.StatusUnauthorized).JSON(models.APIError{
		Status:  fiber.StatusUnauthorized,
		Message: mensaje,
	})
}

// RequireAuth valida la firma y la vigencia del Bearer token JWT y, además, que la cuenta
// siga activa: desactivar a un usuario corta su sesión de inmediato, sin esperar a que el token expire.
// La identidad verificada queda en c.Locals para los controladores.
func RequireAuth(secreto []byte, activa CuentaActiva) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return noAutorizado(c, "Token de autorización no proporcionado")
		}

		partes := strings.SplitN(authHeader, " ", 2)
		if len(partes) != 2 || !strings.EqualFold(partes[0], "Bearer") {
			return noAutorizado(c, "Formato de token inválido. Use 'Bearer <token>'")
		}

		token, err := jwt.Parse(strings.TrimSpace(partes[1]), func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fiber.ErrUnauthorized
			}
			return secreto, nil
		})
		if err != nil || !token.Valid {
			return noAutorizado(c, "Token inválido o expirado")
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			return noAutorizado(c, "Claims de token inválidos")
		}

		sub, ok := claims["sub"].(float64)
		if !ok {
			return noAutorizado(c, "Claims de token inválidos")
		}
		usuarioID := int(sub)

		if activa != nil && !activa(c.Context(), usuarioID) {
			return noAutorizado(c, "Cuenta desactivada o inexistente")
		}

		// Guardamos la identidad autenticada en locals para uso en controladores
		c.Locals("email", claims["email"])
		c.Locals("rol", claims["rol"])
		c.Locals("usuario_id", usuarioID)

		return c.Next()
	}
}

// RequireAdmin verifica que el usuario autenticado posea el rol 'admin'.
func RequireAdmin() fiber.Handler {
	return func(c *fiber.Ctx) error {
		rol, _ := c.Locals("rol").(string)
		if rol != "admin" {
			return c.Status(fiber.StatusForbidden).JSON(models.APIError{
				Status:  fiber.StatusForbidden,
				Message: "Acceso denegado: se requieren permisos de administrador",
			})
		}
		return c.Next()
	}
}
