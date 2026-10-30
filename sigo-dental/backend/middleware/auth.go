// Paquete middleware: funciones que interceptan las peticiones HTTP antes del controlador.
package middleware

import (
	"slices"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"sigo-dental-backend/models"
)

// DuracionToken es la vigencia de un token desde que se emite.
const DuracionToken = 2 * time.Hour

// GenerarToken crea un token firmado con HS256 que incluye el rol de la cuenta.
func GenerarToken(secreto []byte, usuarioID int, email, rol string) (string, error) {
	ahora := time.Now()
	claims := jwt.MapClaims{
		"sub":   usuarioID,
		"email": email,
		"rol":   rol,
		"iat":   ahora.Unix(),
		"exp":   ahora.Add(DuracionToken).Unix(),
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(secreto)
}

// errorAPI responde con el contrato estándar de errores.
func errorAPI(c *fiber.Ctx, estado int, mensaje string) error {
	return c.Status(estado).JSON(models.APIError{Status: estado, Message: mensaje})
}

// RequireAuth valida la firma y la vigencia del Bearer token y deja la identidad verificada
// en c.Locals ("usuario_id", "email" y "rol"). Responde 401 si falta o es inválido.
func RequireAuth(secreto []byte) fiber.Handler {
	return func(c *fiber.Ctx) error {
		cabecera := c.Get("Authorization")
		if cabecera == "" {
			return errorAPI(c, fiber.StatusUnauthorized, "Token de autorización no proporcionado")
		}

		partes := strings.SplitN(cabecera, " ", 2)
		if len(partes) != 2 || !strings.EqualFold(partes[0], "Bearer") {
			return errorAPI(c, fiber.StatusUnauthorized, "Formato de token inválido. Use 'Bearer <token>'")
		}

		// Solo se acepta HS256: rechaza "none" y cualquier otro algoritmo que no use nuestra clave.
		token, err := jwt.Parse(strings.TrimSpace(partes[1]), func(*jwt.Token) (interface{}, error) {
			return secreto, nil
		}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}), jwt.WithExpirationRequired())
		if err != nil || !token.Valid {
			return errorAPI(c, fiber.StatusUnauthorized, "Token inválido o expirado")
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			return errorAPI(c, fiber.StatusUnauthorized, "Claims de token inválidos")
		}
		sub, okSub := claims["sub"].(float64)
		rol, okRol := claims["rol"].(string)
		if !okSub || !okRol || rol == "" {
			return errorAPI(c, fiber.StatusUnauthorized, "Claims de token inválidos")
		}
		email, _ := claims["email"].(string)

		c.Locals("usuario_id", int(sub))
		c.Locals("email", email)
		c.Locals("rol", rol)
		return c.Next()
	}
}

// RequireRoles deja pasar solo a las cuentas cuyo rol está en la lista; el resto recibe 403.
// Debe ir después de RequireAuth. Falla cerrado: sin rol en el contexto o con la lista vacía, 403.
func RequireRoles(permitidos ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		rol, _ := c.Locals("rol").(string)
		if rol == "" || !slices.Contains(permitidos, rol) {
			return errorAPI(c, fiber.StatusForbidden, "Acceso denegado: su rol no tiene permiso para este recurso")
		}
		return c.Next()
	}
}
