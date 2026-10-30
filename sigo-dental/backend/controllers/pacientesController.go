// Este archivo atiende los pacientes y su anamnesis (F1-2: RF-01 y RF-02).
// La recepción da de alta y modifica la ficha; el odontólogo y el asistente editan la anamnesis;
// los tres consultan. El router fija esos permisos.
package controllers

import (
	"context"
	"errors"
	"net/mail"
	"slices"
	"time"

	"github.com/gofiber/fiber/v2"
	"sigo-dental-backend/models"
)

// PacienteRepository es el contrato de la capa de datos de pacientes y anamnesis.
type PacienteRepository interface {
	ListarPacientes(ctx context.Context, busqueda string, soloActivos bool) ([]models.PacienteResumen, error)
	ObtenerPaciente(ctx context.Context, id int) (models.PacienteDetalle, error)
	// CrearPaciente da de alta ficha y anamnesis juntas y devuelve el ID del paciente.
	CrearPaciente(ctx context.Context, p models.Paciente, a models.Anamnesis, usuarioID int) (int, error)
	ActualizarPaciente(ctx context.Context, id int, p models.Paciente) error
	ActualizarAnamnesis(ctx context.Context, id int, a models.Anamnesis, usuarioID int) error
}

// PacientesController atiende las rutas de pacientes con el repositorio inyectado.
type PacientesController struct {
	repo PacienteRepository
	// ahora permite fijar la fecha en las pruebas (la edad depende de ella).
	ahora func() time.Time
}

// NuevoPacientesController construye el controlador de pacientes.
func NuevoPacientesController(repo PacienteRepository) *PacientesController {
	return &PacientesController{repo: repo, ahora: time.Now}
}

const (
	mayoriaDeEdad  = 18
	maxDocumento   = 30
	maxTelefono    = 30
	maxEmail       = 150
	maxDireccion   = 200
	maxParentesco  = 50
	maxAnamnesis   = 2000
	maxCondiciones = 50
	formatoFecha   = "2006-01-02"
)

// edadEn calcula los años cumplidos a la fecha dada.
func edadEn(nacimiento, hoy time.Time) int {
	edad := hoy.Year() - nacimiento.Year()
	if hoy.Month() < nacimiento.Month() || (hoy.Month() == nacimiento.Month() && hoy.Day() < nacimiento.Day()) {
		edad--
	}
	return edad
}

type cuerpoPaciente struct {
	Documento            string `json:"documento"`
	Nombres              string `json:"nombres"`
	Apellidos            string `json:"apellidos"`
	FechaNacimiento      string `json:"fecha_nacimiento"`
	Sexo                 string `json:"sexo"`
	Telefono             string `json:"telefono"`
	Email                string `json:"email"`
	Direccion            string `json:"direccion"`
	EmergenciaNombre     string `json:"emergencia_nombre"`
	EmergenciaTelefono   string `json:"emergencia_telefono"`
	EmergenciaParentesco string `json:"emergencia_parentesco"`
	TutorNombre          string `json:"tutor_nombre"`
	TutorDocumento       string `json:"tutor_documento"`
	TutorTelefono        string `json:"tutor_telefono"`
	ConvenioID           *int   `json:"convenio_id"`
	CanalPreferido       string `json:"canal_preferido"`
	Activo               *bool  `json:"activo"`
	// Anamnesis solo se usa al dar de alta: obligatoria desde el primer registro.
	Anamnesis *models.Anamnesis `json:"anamnesis"`
}

// validar comprueba la ficha y la convierte al modelo; el segundo valor es el motivo si es inválida.
func (cuerpo cuerpoPaciente) validar(hoy time.Time) (models.Paciente, string) {
	var p models.Paciente
	var ok bool

	if p.Documento, ok = textoObligatorio(cuerpo.Documento, maxDocumento); !ok {
		return p, "El documento de identidad es obligatorio (hasta 30 caracteres)"
	}
	if p.Nombres, ok = textoObligatorio(cuerpo.Nombres, maxNombre); !ok {
		return p, "Los nombres son obligatorios (hasta 100 caracteres)"
	}
	if p.Apellidos, ok = textoObligatorio(cuerpo.Apellidos, maxNombre); !ok {
		return p, "Los apellidos son obligatorios (hasta 100 caracteres)"
	}

	nacimiento, err := time.Parse(formatoFecha, cuerpo.FechaNacimiento)
	if err != nil || nacimiento.After(hoy) || nacimiento.Year() < 1900 {
		return p, "La fecha de nacimiento debe tener el formato AAAA-MM-DD y no puede ser futura"
	}
	p.FechaNacimiento = cuerpo.FechaNacimiento

	if !slices.Contains(models.SexosValidos, cuerpo.Sexo) {
		return p, "El sexo debe ser femenino, masculino u otro"
	}
	p.Sexo = cuerpo.Sexo

	if p.Telefono, ok = textoObligatorio(cuerpo.Telefono, maxTelefono); !ok {
		return p, "El teléfono es obligatorio (hasta 30 caracteres)"
	}
	if p.Email, ok = textoOpcional(cuerpo.Email, maxEmail); !ok {
		return p, "El correo admite hasta 150 caracteres"
	}
	if p.Email != "" {
		if _, err := mail.ParseAddress(p.Email); err != nil {
			return p, "El correo electrónico no es válido"
		}
	}
	if p.Direccion, ok = textoOpcional(cuerpo.Direccion, maxDireccion); !ok {
		return p, "La dirección admite hasta 200 caracteres"
	}

	if p.EmergenciaNombre, ok = textoObligatorio(cuerpo.EmergenciaNombre, maxNombreLargo); !ok {
		return p, "El nombre del contacto de emergencia es obligatorio"
	}
	if p.EmergenciaTelefono, ok = textoObligatorio(cuerpo.EmergenciaTelefono, maxTelefono); !ok {
		return p, "El teléfono del contacto de emergencia es obligatorio"
	}
	if p.EmergenciaParentesco, ok = textoOpcional(cuerpo.EmergenciaParentesco, maxParentesco); !ok {
		return p, "El parentesco admite hasta 50 caracteres"
	}

	// El tutor legal es obligatorio para menores de edad; para adultos es opcional.
	tutorNombre, okNombre := textoOpcional(cuerpo.TutorNombre, maxNombreLargo)
	tutorDocumento, okDocumento := textoOpcional(cuerpo.TutorDocumento, maxDocumento)
	tutorTelefono, okTelefono := textoOpcional(cuerpo.TutorTelefono, maxTelefono)
	if !okNombre || !okDocumento || !okTelefono {
		return p, "Los datos del tutor legal exceden la longitud permitida"
	}
	if edadEn(nacimiento, hoy) < mayoriaDeEdad && (tutorNombre == "" || tutorDocumento == "" || tutorTelefono == "") {
		return p, "Un paciente menor de edad requiere nombre, documento y teléfono de su tutor legal"
	}
	p.TutorNombre, p.TutorDocumento, p.TutorTelefono = tutorNombre, tutorDocumento, tutorTelefono

	if cuerpo.ConvenioID != nil && *cuerpo.ConvenioID > 0 {
		p.ConvenioID = cuerpo.ConvenioID
	}
	p.CanalPreferido = cuerpo.CanalPreferido
	if p.CanalPreferido == "" {
		p.CanalPreferido = "whatsapp"
	}
	if !slices.Contains(models.CanalesValidos, p.CanalPreferido) {
		return p, "El canal preferido debe ser whatsapp, sms o email"
	}
	return p, ""
}

// validarAnamnesis exige los cuatro apartados con contenido (se escribe "Ninguna" si no hay nada que declarar)
// y deja la lista de condiciones críticas sin repetidos.
func validarAnamnesis(a models.Anamnesis) (models.Anamnesis, string) {
	var ok bool
	var limpia models.Anamnesis
	apartados := []struct {
		origen  string
		destino *string
		nombre  string
	}{
		{a.EnfermedadesSistemicas, &limpia.EnfermedadesSistemicas, "enfermedades sistémicas"},
		{a.AlergiasFarmacos, &limpia.AlergiasFarmacos, "alergias a fármacos"},
		{a.AntecedentesQuirurgicos, &limpia.AntecedentesQuirurgicos, "antecedentes quirúrgicos"},
		{a.Coagulopatias, &limpia.Coagulopatias, "coagulopatías"},
	}
	for _, apartado := range apartados {
		if *apartado.destino, ok = textoObligatorio(apartado.origen, maxAnamnesis); !ok {
			return limpia, "La anamnesis es obligatoria: completa " + apartado.nombre + " o escribe \"Ninguna\" (hasta 2000 caracteres)"
		}
	}

	if len(a.CondicionesCriticas) > maxCondiciones {
		return limpia, "Demasiadas condiciones críticas"
	}
	limpia.CondicionesCriticas = []int{}
	for _, id := range a.CondicionesCriticas {
		if id <= 0 {
			return limpia, "Una condición crítica no es válida"
		}
		if !slices.Contains(limpia.CondicionesCriticas, id) {
			limpia.CondicionesCriticas = append(limpia.CondicionesCriticas, id)
		}
	}
	return limpia, ""
}

// usuarioDeLaPeticion devuelve el ID de la cuenta que hizo la petición (lo fija RequireAuth).
func usuarioDeLaPeticion(c *fiber.Ctx) int {
	id, _ := c.Locals("usuario_id").(int)
	return id
}

// ListarPacientes responde GET /api/pacientes?q=texto. Solo la recepción puede pedir los desactivados (?todos=true).
func (ctrl *PacientesController) ListarPacientes(c *fiber.Ctx) error {
	rol, _ := c.Locals("rol").(string)
	soloActivos := !(c.QueryBool("todos") && rol == models.RolRecepcionista)
	lista, err := ctrl.repo.ListarPacientes(c.Context(), c.Query("q"), soloActivos)
	if err != nil {
		return errorDominio(c, err)
	}
	return c.JSON(lista)
}

// ObtenerPaciente responde GET /api/pacientes/:id con la ficha, la anamnesis y las alertas.
func (ctrl *PacientesController) ObtenerPaciente(c *fiber.Ctx) error {
	id, ok := idDeRuta(c)
	if !ok {
		return errorCliente(c, fiber.StatusBadRequest, "Identificador inválido")
	}
	detalle, err := ctrl.repo.ObtenerPaciente(c.Context(), id)
	if err != nil {
		return errorDominio(c, err)
	}
	return c.JSON(detalle)
}

// CrearPaciente responde POST /api/pacientes: ficha y anamnesis obligatoria en un solo alta.
func (ctrl *PacientesController) CrearPaciente(c *fiber.Ctx) error {
	var cuerpo cuerpoPaciente
	if err := c.BodyParser(&cuerpo); err != nil {
		return cuerpoInvalido(c)
	}
	paciente, motivo := cuerpo.validar(ctrl.ahora())
	if motivo != "" {
		return errorCliente(c, fiber.StatusBadRequest, motivo)
	}
	if cuerpo.Anamnesis == nil {
		return errorCliente(c, fiber.StatusBadRequest, "La anamnesis es obligatoria para registrar un paciente")
	}
	anamnesis, motivo := validarAnamnesis(*cuerpo.Anamnesis)
	if motivo != "" {
		return errorCliente(c, fiber.StatusBadRequest, motivo)
	}

	id, err := ctrl.repo.CrearPaciente(c.Context(), paciente, anamnesis, usuarioDeLaPeticion(c))
	if err != nil {
		return errorDocumentoRepetido(c, err)
	}
	detalle, err := ctrl.repo.ObtenerPaciente(c.Context(), id)
	if err != nil {
		return errorDominio(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(detalle)
}

// ActualizarPaciente responde PUT /api/pacientes/:id: reemplaza los datos de la ficha.
func (ctrl *PacientesController) ActualizarPaciente(c *fiber.Ctx) error {
	id, ok := idDeRuta(c)
	if !ok {
		return errorCliente(c, fiber.StatusBadRequest, "Identificador inválido")
	}
	var cuerpo cuerpoPaciente
	if err := c.BodyParser(&cuerpo); err != nil {
		return cuerpoInvalido(c)
	}
	paciente, motivo := cuerpo.validar(ctrl.ahora())
	if motivo != "" {
		return errorCliente(c, fiber.StatusBadRequest, motivo)
	}
	if cuerpo.Activo == nil {
		return errorCliente(c, fiber.StatusBadRequest, "El estado es obligatorio")
	}
	paciente.Activo = *cuerpo.Activo

	if err := ctrl.repo.ActualizarPaciente(c.Context(), id, paciente); err != nil {
		return errorDocumentoRepetido(c, err)
	}
	detalle, err := ctrl.repo.ObtenerPaciente(c.Context(), id)
	if err != nil {
		return errorDominio(c, err)
	}
	return c.JSON(detalle)
}

// ActualizarAnamnesis responde PUT /api/pacientes/:id/anamnesis.
func (ctrl *PacientesController) ActualizarAnamnesis(c *fiber.Ctx) error {
	id, ok := idDeRuta(c)
	if !ok {
		return errorCliente(c, fiber.StatusBadRequest, "Identificador inválido")
	}
	var cuerpo models.Anamnesis
	if err := c.BodyParser(&cuerpo); err != nil {
		return cuerpoInvalido(c)
	}
	anamnesis, motivo := validarAnamnesis(cuerpo)
	if motivo != "" {
		return errorCliente(c, fiber.StatusBadRequest, motivo)
	}
	if err := ctrl.repo.ActualizarAnamnesis(c.Context(), id, anamnesis, usuarioDeLaPeticion(c)); err != nil {
		return errorDominio(c, err)
	}
	detalle, err := ctrl.repo.ObtenerPaciente(c.Context(), id)
	if err != nil {
		return errorDominio(c, err)
	}
	return c.JSON(detalle)
}

// errorDocumentoRepetido da un mensaje claro cuando el documento ya pertenece a otro paciente.
func errorDocumentoRepetido(c *fiber.Ctx, err error) error {
	if errors.Is(err, models.ErrDuplicado) {
		return errorCliente(c, fiber.StatusConflict, "Ya existe un paciente con ese documento de identidad")
	}
	return errorDominio(c, err)
}
