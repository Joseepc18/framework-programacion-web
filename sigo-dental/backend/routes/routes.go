// Paquete routes: único responsable de registrar los endpoints de la API.
package routes

import (
	"github.com/gofiber/fiber/v2"
	// Controladores que se ejecutan en cada ruta.
	"sigo-dental-backend/controllers"
	"sigo-dental-backend/middleware"
	"sigo-dental-backend/models"
)

// Manejadores agrupa los controladores ya construidos con sus dependencias inyectadas.
type Manejadores struct {
	Health *controllers.HealthController
	Auth   *controllers.AuthController
	// JWTSecret es la clave con la que RequireAuth verifica los tokens.
	JWTSecret []byte
}

// rutasPrueba asocia cada ruta de demostración del RBAC con el único rol que puede usarla.
// Existen para comprobar los permisos; cuando haya módulos reales se reemplazan por sus rutas.
var rutasPrueba = map[string]string{
	"/api/prueba/admin-ti":   models.RolAdminTI,
	"/api/prueba/recepcion":  models.RolRecepcionista,
	"/api/prueba/odontologo": models.RolOdontologo,
	"/api/prueba/asistente":  models.RolAsistente,
	"/api/prueba/auditor":    models.RolAuditor,
}

// SetupRoutes registra los endpoints usando los manejadores recibidos.
func SetupRoutes(app *fiber.App, m Manejadores) {
	// Estado de PostgreSQL y Redis; lo usan Docker, Render y las pruebas de humo.
	app.Get("/api/health", m.Health.HealthCheck)

	// Inicio de sesión: devuelve un JWT con el rol de la cuenta.
	app.Post("/api/login", m.Auth.Login)

	// Rutas protegidas: RequireAuth responde 401 sin token válido y RequireRoles 403 si el rol no corresponde.
	autenticado := middleware.RequireAuth(m.JWTSecret)
	for ruta, rol := range rutasPrueba {
		app.Get(ruta, autenticado, middleware.RequireRoles(rol), func(c *fiber.Ctx) error {
			return c.JSON(fiber.Map{"acceso": "permitido", "rol": c.Locals("rol")})
		})
	}
}
