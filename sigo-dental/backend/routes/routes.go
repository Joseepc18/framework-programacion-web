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
	Health   *controllers.HealthController
	Auth     *controllers.AuthController
	Catalogo *controllers.CatalogoController
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

	registrarCatalogo(app, m, autenticado)
}

// registrarCatalogo declara las rutas del catálogo médico: cualquier cuenta con sesión consulta
// (las listas alimentan los selectores de los demás módulos) y solo el administrador TI modifica.
func registrarCatalogo(app *fiber.App, m Manejadores, autenticado fiber.Handler) {
	admin := middleware.RequireRoles(models.RolAdminTI)
	cat := m.Catalogo

	app.Get("/api/catalogo/:recurso", autenticado, cat.ListarSimples)
	app.Post("/api/catalogo/:recurso", autenticado, admin, cat.CrearSimple)
	app.Put("/api/catalogo/:recurso/:id", autenticado, admin, cat.ActualizarSimple)

	app.Get("/api/sillones", autenticado, cat.ListarSillones)
	app.Post("/api/sillones", autenticado, admin, cat.CrearSillon)
	app.Put("/api/sillones/:id", autenticado, admin, cat.ActualizarSillon)

	app.Get("/api/profesionales", autenticado, cat.ListarProfesionales)
	app.Get("/api/profesionales/candidatos", autenticado, admin, cat.ListarUsuariosOdontologos)
	app.Post("/api/profesionales", autenticado, admin, cat.CrearProfesional)
	app.Put("/api/profesionales/:id", autenticado, admin, cat.ActualizarProfesional)

	app.Get("/api/procedimientos", autenticado, cat.ListarProcedimientos)
	app.Post("/api/procedimientos", autenticado, admin, cat.CrearProcedimiento)
	app.Put("/api/procedimientos/:id", autenticado, admin, cat.ActualizarProcedimiento)

	app.Get("/api/parametros", autenticado, cat.ListarParametros)
	app.Put("/api/parametros/:clave", autenticado, admin, cat.ActualizarParametro)
}
