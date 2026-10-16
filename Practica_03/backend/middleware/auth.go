// Declaramos el paquete middleware para funciones interceptoras de peticiones HTTP.
package middleware

import (
	"os"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"multicatalogo-backend/models"
)

// ObtenerSecretoJWT devuelve la clave para firmar y validar tokens JWT.
func ObtenerSecretoJWT() []byte {
	secreto := os.Getenv("JWT_SECRET")
	if secreto == "" {
		return []byte("multicatalogo-jwt-super-secret-key-2026")
	}
	return []byte(secreto)
}

// GenerarToken crea un token firmado con HS256 y duración de 2 horas.
func GenerarToken(userID int, email, rol string) (string, error) {
	claims := jwt.MapClaims{
		"sub":   userID,
		"email": email,
		"rol":   rol,
		"exp":   time.Now().Add(2 * time.Hour).Unix(),
		"iat":   time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(ObtenerSecretoJWT())
}

// RequireAuth intercepta la petición y valida la existencia y firma del Bearer token JWT.
func RequireAuth() fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(models.APIError{
				Status:  fiber.StatusUnauthorized,
				Message: "Token de autorización no proporcionado",
			})
		}

		partes := strings.SplitN(authHeader, " ", 2)
		if len(partes) != 2 || !strings.EqualFold(partes[0], "Bearer") {
			return c.Status(fiber.StatusUnauthorized).JSON(models.APIError{
				Status:  fiber.StatusUnauthorized,
				Message: "Formato de token inválido. Use 'Bearer <token>'",
			})
		}

		tokenStr := strings.TrimSpace(partes[1])
		token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fiber.ErrUnauthorized
			}
			return ObtenerSecretoJWT(), nil
		})

		if err != nil || !token.Valid {
			return c.Status(fiber.StatusUnauthorized).JSON(models.APIError{
				Status:  fiber.StatusUnauthorized,
				Message: "Token inválido o expirado",
			})
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(models.APIError{
				Status:  fiber.StatusUnauthorized,
				Message: "Claims de token inválidos",
			})
		}

		// Guardamos la identidad autenticada en locals para uso en controladores
		c.Locals("email", claims["email"])
		c.Locals("rol", claims["rol"])
		if sub, ok := claims["sub"].(float64); ok {
			c.Locals("usuario_id", int(sub))
		}

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
