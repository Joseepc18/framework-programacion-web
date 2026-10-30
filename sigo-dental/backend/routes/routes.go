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
	Cobros    *controllers.CobrosController
	// Expediente (F2-1): signos vitales, archivos y consulta de la auditoría.
	Expediente *controllers.ExpedienteController
	Auditoria  *controllers.AuditoriaController
	// Módulos clínicos de la Fase 2.
	Odontograma *controllers.OdontogramaController
	// Auditor guarda cada acceso a datos clínicos; si es nil las rutas funcionan sin registrarlos (solo pruebas).
	Auditor middleware.Auditor
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
	registrarCobros(app, m, autenticado)
	registrarExpediente(app, m, autenticado)
	registrarOdontograma(app, m, autenticado)
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
	// La bitácora de mensajes la ve solo la recepción.
	app.Get("/api/citas/:id/notificaciones", autenticado, recepcion, cit.ListarNotificaciones)
	// Público a propósito: el paciente confirma desde el enlace que recibe, sin cuenta. Lo protege la firma del token.
	app.Post("/api/citas/confirmar", cit.ConfirmarPorEnlace)
}

// registrarPacientes declara las rutas de pacientes: la recepción registra y modifica la ficha,
// el odontólogo y el asistente editan la anamnesis, y los tres consultan. El auditor no accede
// a datos clínicos (menor privilegio).
func registrarPacientes(app *fiber.App, m Manejadores, autenticado fiber.Handler) {
	consulta := middleware.RequireRoles(models.RolRecepcionista, models.RolOdontologo, models.RolAsistente)
	recepcion := middleware.RequireRoles(models.RolRecepcionista)
	clinico := middleware.RequireRoles(models.RolOdontologo, models.RolAsistente)
	pac := m.Pacientes

	paciente, entidadPaciente := middleware.PacienteDeRuta("id"), middleware.EntidadDeRuta("id")
	app.Get("/api/pacientes", autenticado, consulta, auditar(m, models.AccionConsulta, "pacientes"), pac.ListarPacientes)
	app.Get("/api/pacientes/:id", autenticado, consulta, auditar(m, models.AccionConsulta, "paciente", paciente, entidadPaciente), pac.ObtenerPaciente)
	app.Post("/api/pacientes", autenticado, recepcion, auditar(m, models.AccionCreacion, "paciente"), pac.CrearPaciente)
	app.Put("/api/pacientes/:id", autenticado, recepcion, auditar(m, models.AccionModificacion, "paciente", paciente, entidadPaciente), pac.ActualizarPaciente)
	app.Put("/api/pacientes/:id/anamnesis", autenticado, clinico, auditar(m, models.AccionModificacion, "anamnesis", paciente), pac.ActualizarAnamnesis)
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

// registrarCobros declara las rutas de facturación: la recepción cobra, registra pagos y anula; la recepción y el
// auditor consultan. Los demás roles no ven cobros.
func registrarCobros(app *fiber.App, m Manejadores, autenticado fiber.Handler) {
	consulta := middleware.RequireRoles(models.RolRecepcionista, models.RolAuditor)
	recepcion := middleware.RequireRoles(models.RolRecepcionista)
	cob := m.Cobros

	app.Get("/api/cobros", autenticado, consulta, cob.ListarCobros)
	app.Get("/api/cobros/:id", autenticado, consulta, cob.ObtenerCobro)
	app.Post("/api/cobros", autenticado, recepcion, cob.CrearCobro)
	app.Post("/api/cobros/:id/pagos", autenticado, recepcion, cob.RegistrarPago)
	app.Post("/api/cobros/:id/anular", autenticado, recepcion, cob.AnularCobro)
}

// auditar devuelve el middleware que registra el acceso en la bitácora (RNF-06), o uno que no hace nada si no hay auditor.
func auditar(m Manejadores, accion, entidad string, opciones ...middleware.OpcionAuditoria) fiber.Handler {
	if m.Auditor == nil {
		return func(c *fiber.Ctx) error { return c.Next() }
	}
	return middleware.Auditar(m.Auditor, accion, entidad, opciones...)
}

// registrarExpediente declara las rutas de signos vitales, archivos y auditoría. El personal clínico (odontólogo y
// asistente) registra y consulta; el administrador TI y el auditor leen la bitácora, que nadie puede modificar.
func registrarExpediente(app *fiber.App, m Manejadores, autenticado fiber.Handler) {
	clinico := middleware.RequireRoles(models.RolOdontologo, models.RolAsistente)
	lectorAuditoria := middleware.RequireRoles(models.RolAdminTI, models.RolAuditor)
	exp, aud := m.Expediente, m.Auditoria
	paciente := middleware.PacienteDeRuta("id")

	app.Get("/api/pacientes/:id/signos-vitales", autenticado, clinico, auditar(m, models.AccionConsulta, "signos_vitales", paciente), exp.ListarSignosVitales)
	app.Post("/api/pacientes/:id/signos-vitales", autenticado, clinico, auditar(m, models.AccionCreacion, "signos_vitales", paciente), exp.RegistrarSignosVitales)
	app.Get("/api/pacientes/:id/archivos", autenticado, clinico, auditar(m, models.AccionConsulta, "archivos", paciente), exp.ListarArchivos)
	app.Post("/api/pacientes/:id/archivos", autenticado, clinico, auditar(m, models.AccionCreacion, "archivo", paciente), exp.SubirArchivo)
	// El paciente de una descarga se conoce al leer los metadatos: el controlador lo anota para la auditoría.
	app.Get("/api/archivos/:id/descargar", autenticado, clinico, auditar(m, models.AccionExportacion, "archivo", middleware.EntidadDeRuta("id")), exp.DescargarArchivo)

	app.Get("/api/auditoria", autenticado, lectorAuditoria, auditar(m, models.AccionConsulta, "auditoria"), aud.ListarAuditoria)
}

// registrarOdontograma declara las rutas del odontograma (F2-2): el odontólogo registra, corrige y retira hallazgos;
// el asistente solo consulta.
func registrarOdontograma(app *fiber.App, m Manejadores, autenticado fiber.Handler) {
	lectura := middleware.RequireRoles(models.RolOdontologo, models.RolAsistente)
	escritura := middleware.RequireRoles(models.RolOdontologo)
	odo := m.Odontograma
	paciente := middleware.PacienteDeRuta("id")
	base := "/api/pacientes/:id/odontograma"

	app.Get(base, autenticado, lectura, auditar(m, models.AccionConsulta, "odontograma", paciente), odo.ObtenerOdontograma)
	app.Get(base+"/historial", autenticado, lectura, auditar(m, models.AccionConsulta, "odontograma_historial", paciente), odo.HistorialOdontograma)
	app.Post(base+"/hallazgos", autenticado, escritura, auditar(m, models.AccionCreacion, "hallazgo", paciente), odo.RegistrarHallazgo)
	app.Put(base+"/hallazgos/:hid", autenticado, escritura, auditar(m, models.AccionModificacion, "hallazgo", paciente), odo.CorregirHallazgo)
	app.Post(base+"/hallazgos/:hid/retirar", autenticado, escritura, auditar(m, models.AccionModificacion, "hallazgo", paciente, middleware.EntidadDeRuta("hid")), odo.RetirarHallazgo)
}
