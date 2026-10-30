// Este archivo atiende el catálogo médico (F1-1): consultorios, sillones, profesionales,
// procedimientos, convenios, condiciones críticas y los parámetros del sistema.
// Todos los roles con sesión consultan; solo el administrador TI crea y modifica (lo exige el router).
package controllers

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"sigo-dental-backend/models"
)

// CatalogoRepository es el contrato de la capa de datos del catálogo médico.
type CatalogoRepository interface {
	ListarSimples(ctx context.Context, recurso string, soloActivos bool) ([]models.ItemCatalogo, error)
	CrearSimple(ctx context.Context, recurso, nombre string) (models.ItemCatalogo, error)
	ActualizarSimple(ctx context.Context, recurso string, id int, nombre string, activo bool) (models.ItemCatalogo, error)

	ListarSillones(ctx context.Context, consultorioID int, soloActivos bool) ([]models.Sillon, error)
	CrearSillon(ctx context.Context, consultorioID int, nombre string) (models.Sillon, error)
	ActualizarSillon(ctx context.Context, id int, nombre string, activo bool) (models.Sillon, error)

	ListarProfesionales(ctx context.Context, soloActivos bool) ([]models.Profesional, error)
	CrearProfesional(ctx context.Context, usuarioID int, especialidad, registro string) (models.Profesional, error)
	ActualizarProfesional(ctx context.Context, id int, especialidad, registro string, activo bool) (models.Profesional, error)
	ListarUsuariosOdontologos(ctx context.Context) ([]models.UsuarioResumen, error)

	ListarProcedimientos(ctx context.Context, soloActivos bool) ([]models.Procedimiento, error)
	CrearProcedimiento(ctx context.Context, p models.Procedimiento) (models.Procedimiento, error)
	ActualizarProcedimiento(ctx context.Context, id int, p models.Procedimiento) (models.Procedimiento, error)

	ListarParametros(ctx context.Context) ([]models.Parametro, error)
	ObtenerParametro(ctx context.Context, clave string) (string, error)
	ActualizarParametro(ctx context.Context, clave, valor string) (models.Parametro, error)
}

// CatalogoController atiende las rutas del catálogo con el repositorio inyectado.
type CatalogoController struct {
	repo CatalogoRepository
}

// NuevoCatalogoController construye el controlador del catálogo.
func NuevoCatalogoController(repo CatalogoRepository) *CatalogoController {
	return &CatalogoController{repo: repo}
}

// soloActivos decide si el listado oculta los elementos desactivados: solo el administrador TI
// puede pedirlos todos (?todos=true); para el resto siempre se ofrecen los activos.
func soloActivos(c *fiber.Ctx) bool {
	return !(c.QueryBool("todos") && esAdminTI(c))
}

const (
	maxNombre       = 100
	maxNombreLargo  = 150
	maxCodigo       = 20
	maxEspecialidad = 100
	maxRegistro     = 50
)

// --- Catálogos simples: consultorios, convenios y condiciones críticas ---

type cuerpoSimple struct {
	Nombre string `json:"nombre"`
	Activo *bool  `json:"activo"`
}

// ListarSimples responde GET /api/catalogo/:recurso.
func (ctrl *CatalogoController) ListarSimples(c *fiber.Ctx) error {
	items, err := ctrl.repo.ListarSimples(c.Context(), c.Params("recurso"), soloActivos(c))
	if err != nil {
		return errorDominio(c, err)
	}
	return c.JSON(items)
}

// CrearSimple responde POST /api/catalogo/:recurso.
func (ctrl *CatalogoController) CrearSimple(c *fiber.Ctx) error {
	var cuerpo cuerpoSimple
	if err := c.BodyParser(&cuerpo); err != nil {
		return cuerpoInvalido(c)
	}
	nombre, ok := textoObligatorio(cuerpo.Nombre, maxNombre)
	if !ok {
		return errorCliente(c, fiber.StatusBadRequest, "El nombre es obligatorio y admite hasta 100 caracteres")
	}
	item, err := ctrl.repo.CrearSimple(c.Context(), c.Params("recurso"), nombre)
	if err != nil {
		return errorDominio(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(item)
}

// ActualizarSimple responde PUT /api/catalogo/:recurso/:id.
func (ctrl *CatalogoController) ActualizarSimple(c *fiber.Ctx) error {
	id, ok := idDeRuta(c)
	if !ok {
		return errorCliente(c, fiber.StatusBadRequest, "Identificador inválido")
	}
	var cuerpo cuerpoSimple
	if err := c.BodyParser(&cuerpo); err != nil {
		return cuerpoInvalido(c)
	}
	nombre, valido := textoObligatorio(cuerpo.Nombre, maxNombre)
	if !valido || cuerpo.Activo == nil {
		return errorCliente(c, fiber.StatusBadRequest, "El nombre (hasta 100 caracteres) y el estado son obligatorios")
	}
	item, err := ctrl.repo.ActualizarSimple(c.Context(), c.Params("recurso"), id, nombre, *cuerpo.Activo)
	if err != nil {
		return errorDominio(c, err)
	}
	return c.JSON(item)
}

// --- Sillones ---

type cuerpoSillon struct {
	ConsultorioID int    `json:"consultorio_id"`
	Nombre        string `json:"nombre"`
	Activo        *bool  `json:"activo"`
}

// ListarSillones responde GET /api/sillones (?consultorio_id=N filtra por consultorio).
func (ctrl *CatalogoController) ListarSillones(c *fiber.Ctx) error {
	sillones, err := ctrl.repo.ListarSillones(c.Context(), c.QueryInt("consultorio_id", 0), soloActivos(c))
	if err != nil {
		return errorDominio(c, err)
	}
	return c.JSON(sillones)
}

// CrearSillon responde POST /api/sillones.
func (ctrl *CatalogoController) CrearSillon(c *fiber.Ctx) error {
	var cuerpo cuerpoSillon
	if err := c.BodyParser(&cuerpo); err != nil {
		return cuerpoInvalido(c)
	}
	nombre, ok := textoObligatorio(cuerpo.Nombre, maxNombre)
	if !ok || cuerpo.ConsultorioID <= 0 {
		return errorCliente(c, fiber.StatusBadRequest, "El consultorio y el nombre (hasta 100 caracteres) son obligatorios")
	}
	sillon, err := ctrl.repo.CrearSillon(c.Context(), cuerpo.ConsultorioID, nombre)
	if err != nil {
		return errorDominio(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(sillon)
}

// ActualizarSillon responde PUT /api/sillones/:id.
func (ctrl *CatalogoController) ActualizarSillon(c *fiber.Ctx) error {
	id, ok := idDeRuta(c)
	if !ok {
		return errorCliente(c, fiber.StatusBadRequest, "Identificador inválido")
	}
	var cuerpo cuerpoSillon
	if err := c.BodyParser(&cuerpo); err != nil {
		return cuerpoInvalido(c)
	}
	nombre, valido := textoObligatorio(cuerpo.Nombre, maxNombre)
	if !valido || cuerpo.Activo == nil {
		return errorCliente(c, fiber.StatusBadRequest, "El nombre (hasta 100 caracteres) y el estado son obligatorios")
	}
	sillon, err := ctrl.repo.ActualizarSillon(c.Context(), id, nombre, *cuerpo.Activo)
	if err != nil {
		return errorDominio(c, err)
	}
	return c.JSON(sillon)
}

// --- Profesionales ---

type cuerpoProfesional struct {
	UsuarioID           int    `json:"usuario_id"`
	Especialidad        string `json:"especialidad"`
	RegistroProfesional string `json:"registro_profesional"`
	Activo              *bool  `json:"activo"`
}

// ListarProfesionales responde GET /api/profesionales.
func (ctrl *CatalogoController) ListarProfesionales(c *fiber.Ctx) error {
	lista, err := ctrl.repo.ListarProfesionales(c.Context(), soloActivos(c))
	if err != nil {
		return errorDominio(c, err)
	}
	return c.JSON(lista)
}

// ListarUsuariosOdontologos responde GET /api/profesionales/candidatos: las cuentas de odontólogo
// que todavía no son profesionales (alimenta el formulario de alta).
func (ctrl *CatalogoController) ListarUsuariosOdontologos(c *fiber.Ctx) error {
	lista, err := ctrl.repo.ListarUsuariosOdontologos(c.Context())
	if err != nil {
		return errorDominio(c, err)
	}
	return c.JSON(lista)
}

// CrearProfesional responde POST /api/profesionales.
func (ctrl *CatalogoController) CrearProfesional(c *fiber.Ctx) error {
	var cuerpo cuerpoProfesional
	if err := c.BodyParser(&cuerpo); err != nil {
		return cuerpoInvalido(c)
	}
	especialidad, okEsp := textoObligatorio(cuerpo.Especialidad, maxEspecialidad)
	registro, okReg := textoObligatorio(cuerpo.RegistroProfesional, maxRegistro)
	if !okEsp || !okReg || cuerpo.UsuarioID <= 0 {
		return errorCliente(c, fiber.StatusBadRequest, "La cuenta, la especialidad y el registro profesional son obligatorios")
	}
	profesional, err := ctrl.repo.CrearProfesional(c.Context(), cuerpo.UsuarioID, especialidad, registro)
	if err != nil {
		return errorDominio(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(profesional)
}

// ActualizarProfesional responde PUT /api/profesionales/:id.
func (ctrl *CatalogoController) ActualizarProfesional(c *fiber.Ctx) error {
	id, ok := idDeRuta(c)
	if !ok {
		return errorCliente(c, fiber.StatusBadRequest, "Identificador inválido")
	}
	var cuerpo cuerpoProfesional
	if err := c.BodyParser(&cuerpo); err != nil {
		return cuerpoInvalido(c)
	}
	especialidad, okEsp := textoObligatorio(cuerpo.Especialidad, maxEspecialidad)
	registro, okReg := textoObligatorio(cuerpo.RegistroProfesional, maxRegistro)
	if !okEsp || !okReg || cuerpo.Activo == nil {
		return errorCliente(c, fiber.StatusBadRequest, "La especialidad, el registro profesional y el estado son obligatorios")
	}
	profesional, err := ctrl.repo.ActualizarProfesional(c.Context(), id, especialidad, registro, *cuerpo.Activo)
	if err != nil {
		return errorDominio(c, err)
	}
	return c.JSON(profesional)
}

// --- Procedimientos ---

type cuerpoProcedimiento struct {
	Codigo                 string `json:"codigo"`
	Nombre                 string `json:"nombre"`
	Especialidad           string `json:"especialidad"`
	PrecioCentavos         int    `json:"precio_centavos"`
	DuracionMin            int    `json:"duracion_min"`
	RequiereConsentimiento bool   `json:"requiere_consentimiento"`
	Activo                 *bool  `json:"activo"`
}

// validar comprueba el cuerpo y lo convierte al modelo; el segundo valor es el motivo si es inválido.
func (cuerpo cuerpoProcedimiento) validar() (models.Procedimiento, string) {
	codigo, okCodigo := textoObligatorio(cuerpo.Codigo, maxCodigo)
	nombre, okNombre := textoObligatorio(cuerpo.Nombre, maxNombreLargo)
	especialidad, okEsp := textoObligatorio(cuerpo.Especialidad, maxEspecialidad)
	switch {
	case !okCodigo:
		return models.Procedimiento{}, "El código es obligatorio y admite hasta 20 caracteres"
	case !okNombre:
		return models.Procedimiento{}, "El nombre es obligatorio y admite hasta 150 caracteres"
	case !okEsp:
		return models.Procedimiento{}, "La especialidad es obligatoria y admite hasta 100 caracteres"
	case cuerpo.PrecioCentavos < 0:
		return models.Procedimiento{}, "El precio no puede ser negativo"
	case cuerpo.DuracionMin <= 0:
		return models.Procedimiento{}, "La duración debe ser de al menos un minuto"
	}
	return models.Procedimiento{
		Codigo: codigo, Nombre: nombre, Especialidad: especialidad,
		PrecioCentavos: cuerpo.PrecioCentavos, DuracionMin: cuerpo.DuracionMin,
		RequiereConsentimiento: cuerpo.RequiereConsentimiento,
	}, ""
}

// ListarProcedimientos responde GET /api/procedimientos.
func (ctrl *CatalogoController) ListarProcedimientos(c *fiber.Ctx) error {
	lista, err := ctrl.repo.ListarProcedimientos(c.Context(), soloActivos(c))
	if err != nil {
		return errorDominio(c, err)
	}
	return c.JSON(lista)
}

// CrearProcedimiento responde POST /api/procedimientos.
func (ctrl *CatalogoController) CrearProcedimiento(c *fiber.Ctx) error {
	var cuerpo cuerpoProcedimiento
	if err := c.BodyParser(&cuerpo); err != nil {
		return cuerpoInvalido(c)
	}
	procedimiento, motivo := cuerpo.validar()
	if motivo != "" {
		return errorCliente(c, fiber.StatusBadRequest, motivo)
	}
	creado, err := ctrl.repo.CrearProcedimiento(c.Context(), procedimiento)
	if err != nil {
		return errorDominio(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(creado)
}

// ActualizarProcedimiento responde PUT /api/procedimientos/:id.
func (ctrl *CatalogoController) ActualizarProcedimiento(c *fiber.Ctx) error {
	id, ok := idDeRuta(c)
	if !ok {
		return errorCliente(c, fiber.StatusBadRequest, "Identificador inválido")
	}
	var cuerpo cuerpoProcedimiento
	if err := c.BodyParser(&cuerpo); err != nil {
		return cuerpoInvalido(c)
	}
	procedimiento, motivo := cuerpo.validar()
	if motivo != "" {
		return errorCliente(c, fiber.StatusBadRequest, motivo)
	}
	if cuerpo.Activo == nil {
		return errorCliente(c, fiber.StatusBadRequest, "El estado es obligatorio")
	}
	procedimiento.Activo = *cuerpo.Activo
	actualizado, err := ctrl.repo.ActualizarProcedimiento(c.Context(), id, procedimiento)
	if err != nil {
		return errorDominio(c, err)
	}
	return c.JSON(actualizado)
}

// --- Parámetros del sistema ---

// horaValida comprueba el formato HH:MM de 24 horas con dos dígitos en cada parte
// (así los textos se ordenan igual que las horas).
func horaValida(valor string) bool {
	if len(valor) != 5 {
		return false
	}
	_, err := time.Parse("15:04", valor)
	return err == nil
}

// ListarParametros responde GET /api/parametros.
func (ctrl *CatalogoController) ListarParametros(c *fiber.Ctx) error {
	lista, err := ctrl.repo.ListarParametros(c.Context())
	if err != nil {
		return errorDominio(c, err)
	}
	return c.JSON(lista)
}

// ActualizarParametro responde PUT /api/parametros/:clave. Las horas de la agenda deben ser HH:MM
// y la apertura tiene que ser anterior al cierre.
func (ctrl *CatalogoController) ActualizarParametro(c *fiber.Ctx) error {
	clave := c.Params("clave")
	var cuerpo struct {
		Valor string `json:"valor"`
	}
	if err := c.BodyParser(&cuerpo); err != nil {
		return cuerpoInvalido(c)
	}
	valor, ok := textoObligatorio(cuerpo.Valor, 200)
	if !ok {
		return errorCliente(c, fiber.StatusBadRequest, "El valor es obligatorio")
	}

	if clave == ClaveAgendaHoraInicio || clave == ClaveAgendaHoraFin {
		if !horaValida(valor) {
			return errorCliente(c, fiber.StatusBadRequest, "La hora debe tener el formato HH:MM")
		}
		otra, orden := ClaveAgendaHoraFin, -1
		if clave == ClaveAgendaHoraFin {
			otra, orden = ClaveAgendaHoraInicio, 1
		}
		valorOtra, err := ctrl.repo.ObtenerParametro(c.Context(), otra)
		if err != nil {
			return errorDominio(c, err)
		}
		// Comparar los textos HH:MM equivale a compararlos como horas.
		if (orden < 0 && valor >= valorOtra) || (orden > 0 && valor <= valorOtra) {
			return errorCliente(c, fiber.StatusBadRequest, "La hora de apertura debe ser anterior a la de cierre")
		}
	}

	parametro, err := ctrl.repo.ActualizarParametro(c.Context(), clave, valor)
	if err != nil {
		return errorDominio(c, err)
	}
	return c.JSON(parametro)
}

// Claves de los parámetros de la agenda (sembradas en la migración 002).
const (
	ClaveAgendaHoraInicio = "agenda_hora_inicio"
	ClaveAgendaHoraFin    = "agenda_hora_fin"
)
