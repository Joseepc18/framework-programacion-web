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
	Health    *controllers.HealthController
	Auth      *controllers.AuthController
	Catalogo  *controllers.CatalogoController
	Pacientes *controllers.PacientesController
	Citas     *controllers.CitasController
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
	registrarPacientes(app, m, autenticado)
	registrarCitas(app, m, autenticado)
}

// registrarCitas declara las rutas de la agenda: la recepción agenda y reprograma; la recepción, el
// odontólogo y el asistente consultan y cambian el estado (el controlador limita a los dos últimos a
// En Atención y Atendido). El auditor no accede a la agenda clínica.
func registrarCitas(app *fiber.App, m Manejadores, autenticado fiber.Handler) {
	consulta := middleware.RequireRoles(models.RolRecepcionista, models.RolOdontologo, models.RolAsistente)
	recepcion := middleware.RequireRoles(models.RolRecepcionista)
	cit := m.Citas

	app.Get("/api/agenda/horario", autenticado, consulta, cit.HorarioAgenda)
	app.Get("/api/citas", autenticado, consulta, cit.ListarCitas)
	app.Get("/api/citas/:id", autenticado, consulta, cit.ObtenerCita)
	app.Post("/api/citas", autenticado, recepcion, cit.CrearCita)
	app.Put("/api/citas/:id", autenticado, recepcion, cit.ReprogramarCita)
	app.Post("/api/citas/:id/estado", autenticado, consulta, cit.CambiarEstado)
}

// registrarPacientes declara las rutas de pacientes: la recepción registra y modifica la ficha,
// el odontólogo y el asistente editan la anamnesis, y los tres consultan. El auditor no accede
// a datos clínicos (menor privilegio).
func registrarPacientes(app *fiber.App, m Manejadores, autenticado fiber.Handler) {
	consulta := middleware.RequireRoles(models.RolRecepcionista, models.RolOdontologo, models.RolAsistente)
	recepcion := middleware.RequireRoles(models.RolRecepcionista)
	clinico := middleware.RequireRoles(models.RolOdontologo, models.RolAsistente)
	pac := m.Pacientes

	app.Get("/api/pacientes", autenticado, consulta, pac.ListarPacientes)
	app.Get("/api/pacientes/:id", autenticado, consulta, pac.ObtenerPaciente)
	app.Post("/api/pacientes", autenticado, recepcion, pac.CrearPaciente)
	app.Put("/api/pacientes/:id", autenticado, recepcion, pac.ActualizarPaciente)
	app.Put("/api/pacientes/:id/anamnesis", autenticado, clinico, pac.ActualizarAnamnesis)
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
