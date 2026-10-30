// Este archivo atiende las recetas (F2-6): el odontólogo las redacta, las firma con el servicio de firma de F2-4 y, si
// hay un error, las anula (con registro) y emite otra. Avisa si un medicamento coincide con una alergia de la anamnesis.
package controllers

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"sigo-dental-backend/documentos"
	"sigo-dental-backend/firma"
	"sigo-dental-backend/middleware"
	"sigo-dental-backend/models"
)

// RecetaRepository es el contrato de la capa de datos de las recetas.
type RecetaRepository interface {
	ObtenerPaciente(ctx context.Context, id int) (models.PacienteDetalle, error)
	ProfesionalDeUsuario(ctx context.Context, usuarioID int) (models.ProfesionalDeUsuario, bool, error)
	ListarRecetas(ctx context.Context, pacienteID int) ([]models.Receta, error)
	ObtenerReceta(ctx context.Context, pacienteID, id int) (models.Receta, error)
	CrearReceta(ctx context.Context, r models.Receta, usuarioID int) (int, error)
	ActualizarReceta(ctx context.Context, pacienteID, id int, indicaciones string, meds []models.Medicamento) error
	AnularReceta(ctx context.Context, pacienteID, id int, motivo string, usuarioID int, copiar bool) (int, error)
	GuardarFirma(ctx context.Context, f models.Firma) (models.Firma, error)
	ObtenerParametro(ctx context.Context, clave string) (string, error)
}

// RecetasController atiende las rutas de recetas.
type RecetasController struct {
	repo     RecetaRepository
	firmador Firmador
	ahora    func() time.Time
}

// NuevoRecetasController construye el controlador de recetas.
func NuevoRecetasController(repo RecetaRepository, firmador Firmador) *RecetasController {
	return &RecetasController{repo: repo, firmador: firmador, ahora: time.Now}
}

const maxMedicamentos = 20

type cuerpoReceta struct {
	CitaID       *int                 `json:"cita_id"`
	Indicaciones string               `json:"indicaciones"`
	Medicamentos []models.Medicamento `json:"medicamentos"`
}

// validar exige al menos un medicamento con nombre, dosis, frecuencia y duración.
func (cuerpo cuerpoReceta) validar() (models.Receta, string) {
	var r models.Receta
	if len(cuerpo.Medicamentos) == 0 || len(cuerpo.Medicamentos) > maxMedicamentos {
		return r, "La receta lleva entre 1 y 20 medicamentos"
	}
	for i, m := range cuerpo.Medicamentos {
		var ok bool
		var med models.Medicamento
		if med.Nombre, ok = textoObligatorio(m.Nombre, 150); !ok {
			return r, fmt.Sprintf("Medicamento %d: el nombre es obligatorio (hasta 150 caracteres)", i+1)
		}
		if med.Dosis, ok = textoObligatorio(m.Dosis, 100); !ok {
			return r, fmt.Sprintf("Medicamento %d: la dosis es obligatoria", i+1)
		}
		if med.Frecuencia, ok = textoObligatorio(m.Frecuencia, 100); !ok {
			return r, fmt.Sprintf("Medicamento %d: la frecuencia es obligatoria", i+1)
		}
		if med.Duracion, ok = textoObligatorio(m.Duracion, 100); !ok {
			return r, fmt.Sprintf("Medicamento %d: la duración es obligatoria", i+1)
		}
		if med.Indicaciones, ok = textoOpcional(m.Indicaciones, 300); !ok {
			return r, fmt.Sprintf("Medicamento %d: las indicaciones admiten hasta 300 caracteres", i+1)
		}
		r.Medicamentos = append(r.Medicamentos, med)
	}
	indicaciones, ok := textoOpcional(cuerpo.Indicaciones, 500)
	if !ok {
		return r, "Las indicaciones generales admiten hasta 500 caracteres"
	}
	if cuerpo.CitaID != nil && *cuerpo.CitaID <= 0 {
		return r, "La cita indicada no es válida"
	}
	r.Indicaciones, r.CitaID = indicaciones, cuerpo.CitaID
	return r, ""
}

// presentar completa los avisos de alergia y comprueba la huella de la firma.
func (ctrl *RecetasController) presentar(ctx context.Context, r models.Receta) models.Receta {
	if r.Firma != nil {
		r.Firma.Valida = firma.Verificar(*r.Firma, models.ContenidoCanonicoReceta(r))
	}
	r.Avisos = ctrl.avisos(ctx, r.PacienteID, nombresDe(r.Medicamentos))
	return r
}

func nombresDe(meds []models.Medicamento) []string {
	nombres := make([]string, len(meds))
	for i, m := range meds {
		nombres[i] = m.Nombre
	}
	return nombres
}

// avisos compara los medicamentos con las alergias registradas en la anamnesis del paciente.
func (ctrl *RecetasController) avisos(ctx context.Context, pacienteID int, nombres []string) []models.AvisoAlergia {
	paciente, err := ctrl.repo.ObtenerPaciente(ctx, pacienteID)
	if err != nil {
		return []models.AvisoAlergia{}
	}
	return models.AvisosDeAlergia(paciente.Anamnesis.AlergiasFarmacos, nombres)
}

func idsDeReceta(c *fiber.Ctx) (pacienteID, recetaID int, ok bool) {
	pacienteID, ok = idDeRuta(c)
	if !ok {
		return 0, 0, false
	}
	recetaID, err := c.ParamsInt("rid")
	return pacienteID, recetaID, err == nil && recetaID > 0
}

// ListarRecetas responde GET /api/pacientes/:id/recetas: el historial de recetas del paciente.
func (ctrl *RecetasController) ListarRecetas(c *fiber.Ctx) error {
	id, ok := idDeRuta(c)
	if !ok {
		return errorCliente(c, fiber.StatusBadRequest, "Identificador inválido")
	}
	lista, err := ctrl.repo.ListarRecetas(c.Context(), id)
	if err != nil {
		return errorDominio(c, err)
	}
	for i := range lista {
		lista[i] = ctrl.presentar(c.Context(), lista[i])
	}
	return c.JSON(lista)
}

// ObtenerReceta responde GET /api/pacientes/:id/recetas/:rid.
func (ctrl *RecetasController) ObtenerReceta(c *fiber.Ctx) error {
	pacienteID, recetaID, ok := idsDeReceta(c)
	if !ok {
		return errorCliente(c, fiber.StatusBadRequest, "Identificador inválido")
	}
	receta, err := ctrl.repo.ObtenerReceta(c.Context(), pacienteID, recetaID)
	if err != nil {
		return errorDominio(c, err)
	}
	return c.JSON(ctrl.presentar(c.Context(), receta))
}

type cuerpoVerificacion struct {
	Medicamentos []string `json:"medicamentos"`
}

// VerificarAlergias responde POST /api/pacientes/:id/recetas/verificar-alergias: el aviso en vivo mientras se escribe.
func (ctrl *RecetasController) VerificarAlergias(c *fiber.Ctx) error {
	id, ok := idDeRuta(c)
	if !ok {
		return errorCliente(c, fiber.StatusBadRequest, "Identificador inválido")
	}
	var cuerpo cuerpoVerificacion
	if err := c.BodyParser(&cuerpo); err != nil || len(cuerpo.Medicamentos) > maxMedicamentos {
		return cuerpoInvalido(c)
	}
	paciente, err := ctrl.repo.ObtenerPaciente(c.Context(), id)
	if err != nil {
		return errorDominio(c, err)
	}
	return c.JSON(fiber.Map{"avisos": models.AvisosDeAlergia(paciente.Anamnesis.AlergiasFarmacos, cuerpo.Medicamentos), "alergias": paciente.Anamnesis.AlergiasFarmacos})
}

// CrearReceta responde POST /api/pacientes/:id/recetas: crea un borrador numerado. Si un medicamento coincide con una
// alergia registrada, la respuesta trae el aviso (no impide recetar: la decisión es del profesional).
func (ctrl *RecetasController) CrearReceta(c *fiber.Ctx) error {
	pacienteID, ok := idDeRuta(c)
	if !ok {
		return errorCliente(c, fiber.StatusBadRequest, "Identificador inválido")
	}
	var cuerpo cuerpoReceta
	if err := c.BodyParser(&cuerpo); err != nil {
		return cuerpoInvalido(c)
	}
	receta, motivo := cuerpo.validar()
	if motivo != "" {
		return errorCliente(c, fiber.StatusBadRequest, motivo)
	}
	usuarioID := usuarioDeLaPeticion(c)
	profesional, ok, err := ctrl.repo.ProfesionalDeUsuario(c.Context(), usuarioID)
	if err != nil {
		return errorDominio(c, err)
	}
	if !ok {
		return errorEvolucion(c, models.ErrNoEsProfesional)
	}
	receta.PacienteID, receta.ProfesionalID = pacienteID, profesional.ID
	id, err := ctrl.repo.CrearReceta(c.Context(), receta, usuarioID)
	if err != nil {
		return errorEvolucion(c, err)
	}
	middleware.MarcarAuditoria(c, pacienteID, id)
	creada, err := ctrl.repo.ObtenerReceta(c.Context(), pacienteID, id)
	if err != nil {
		return errorDominio(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(ctrl.presentar(c.Context(), creada))
}

// ActualizarReceta responde PUT /api/pacientes/:id/recetas/:rid. Solo un borrador del propio profesional.
func (ctrl *RecetasController) ActualizarReceta(c *fiber.Ctx) error {
	pacienteID, recetaID, ok := idsDeReceta(c)
	if !ok {
		return errorCliente(c, fiber.StatusBadRequest, "Identificador inválido")
	}
	var cuerpo cuerpoReceta
	if err := c.BodyParser(&cuerpo); err != nil {
		return cuerpoInvalido(c)
	}
	nueva, motivo := cuerpo.validar()
	if motivo != "" {
		return errorCliente(c, fiber.StatusBadRequest, motivo)
	}
	actual, err := ctrl.repo.ObtenerReceta(c.Context(), pacienteID, recetaID)
	if err != nil {
		return errorDominio(c, err)
	}
	if actual.Estado != models.RecetaBorrador {
		return errorCliente(c, fiber.StatusConflict, "Una receta firmada o anulada no se edita: anúlala y emite otra")
	}
	if err := ctrl.repo.ActualizarReceta(c.Context(), pacienteID, recetaID, nueva.Indicaciones, nueva.Medicamentos); err != nil {
		return errorEvolucion(c, err)
	}
	middleware.MarcarAuditoria(c, pacienteID, recetaID)
	editada, err := ctrl.repo.ObtenerReceta(c.Context(), pacienteID, recetaID)
	if err != nil {
		return errorDominio(c, err)
	}
	return c.JSON(ctrl.presentar(c.Context(), editada))
}

// FirmarReceta responde POST /api/pacientes/:id/recetas/:rid/firmar con la reautenticación del odontólogo.
func (ctrl *RecetasController) FirmarReceta(c *fiber.Ctx) error {
	pacienteID, recetaID, ok := idsDeReceta(c)
	if !ok {
		return errorCliente(c, fiber.StatusBadRequest, "Identificador inválido")
	}
	var cuerpo cuerpoFirma
	if err := c.BodyParser(&cuerpo); err != nil || cuerpo.Password == "" {
		return errorCliente(c, fiber.StatusBadRequest, "Escribe tu contraseña para firmar")
	}
	receta, err := ctrl.repo.ObtenerReceta(c.Context(), pacienteID, recetaID)
	if err != nil {
		return errorDominio(c, err)
	}
	if receta.Estado == models.RecetaAnulada {
		return errorCliente(c, fiber.StatusConflict, "La receta está anulada")
	}
	if receta.Firma != nil {
		return errorEvolucion(c, models.ErrYaFirmado)
	}
	firmante, err := ctrl.firmador.Autenticar(c.Context(), usuarioDeLaPeticion(c), cuerpo.Password)
	if err != nil {
		return errorEvolucion(c, err)
	}
	if firmante.ProfesionalID != receta.ProfesionalID {
		return errorEvolucion(c, models.ErrFirmanteNoCorresponde)
	}
	if _, err := ctrl.repo.GuardarFirma(c.Context(), firma.Construir(firmante, models.FirmaReceta, recetaID, models.ContenidoCanonicoReceta(receta))); err != nil {
		if errors.Is(err, models.ErrDuplicado) {
			err = models.ErrYaFirmado
		}
		return errorEvolucion(c, err)
	}
	middleware.MarcarAuditoria(c, pacienteID, recetaID)
	firmada, err := ctrl.repo.ObtenerReceta(c.Context(), pacienteID, recetaID)
	if err != nil {
		return errorDominio(c, err)
	}
	return c.JSON(ctrl.presentar(c.Context(), firmada))
}

type cuerpoAnulacion struct {
	Motivo string `json:"motivo"`
	Copiar bool   `json:"copiar"`
}

// AnularReceta responde POST /api/pacientes/:id/recetas/:rid/anular. La receta no se borra ni se modifica: queda el
// registro de quién la anuló y por qué. Con "copiar" se crea un borrador nuevo con los mismos medicamentos.
func (ctrl *RecetasController) AnularReceta(c *fiber.Ctx) error {
	pacienteID, recetaID, ok := idsDeReceta(c)
	if !ok {
		return errorCliente(c, fiber.StatusBadRequest, "Identificador inválido")
	}
	var cuerpo cuerpoAnulacion
	if err := c.BodyParser(&cuerpo); err != nil {
		return cuerpoInvalido(c)
	}
	motivo, okMotivo := textoObligatorio(cuerpo.Motivo, 300)
	if !okMotivo {
		return errorCliente(c, fiber.StatusBadRequest, "Indica el motivo de la anulación (hasta 300 caracteres)")
	}
	nuevoID, err := ctrl.repo.AnularReceta(c.Context(), pacienteID, recetaID, motivo, usuarioDeLaPeticion(c), cuerpo.Copiar)
	if err != nil {
		return errorEvolucion(c, err)
	}
	middleware.MarcarAuditoria(c, pacienteID, recetaID)
	anulada, err := ctrl.repo.ObtenerReceta(c.Context(), pacienteID, recetaID)
	if err != nil {
		return errorDominio(c, err)
	}
	respuesta := fiber.Map{"receta": ctrl.presentar(c.Context(), anulada)}
	if nuevoID > 0 {
		respuesta["reemplazo_id"] = nuevoID
	}
	return c.JSON(respuesta)
}

// GenerarPDF arma el PDF de la receta del paciente; lo usan la vista del personal y el portal del paciente.
func (ctrl *RecetasController) GenerarPDF(ctx context.Context, pacienteID, recetaID int) ([]byte, models.Receta, error) {
	receta, err := ctrl.repo.ObtenerReceta(ctx, pacienteID, recetaID)
	if err != nil {
		return nil, receta, err
	}
	paciente, err := ctrl.repo.ObtenerPaciente(ctx, pacienteID)
	if err != nil {
		return nil, receta, err
	}
	zona, errZona := time.LoadLocation(parametroTexto(ctx, ctrl.repo, "zona_horaria", "America/Guayaquil"))
	if errZona != nil {
		zona = time.UTC
	}
	datos := documentos.DatosRecetaPDF{
		Clinica: clinicaDe(ctx, ctrl.repo), Receta: ctrl.presentar(ctx, receta),
		Paciente: nombreCompleto(paciente.Nombres, paciente.Apellidos), Documento: paciente.Documento,
		Fecha: receta.CreadaEn.In(zona).Format("02/01/2006 15:04"),
	}
	if receta.Firma != nil {
		datos.Firmada = receta.Firma.FirmadoEn.In(zona).Format("02/01/2006 15:04:05")
	}
	contenido, err := documentos.RecetaPDF(datos)
	return contenido, receta, err
}

// enviarPDF responde con el PDF sin guardarlo en cachés compartidas.
func enviarPDF(c *fiber.Ctx, nombre string, contenido []byte) error {
	c.Set(fiber.HeaderContentType, "application/pdf")
	c.Set(fiber.HeaderContentDisposition, fmt.Sprintf(`inline; filename="%s"`, nombre))
	c.Set(fiber.HeaderCacheControl, "private, no-store")
	return c.Send(contenido)
}

// PDFReceta responde GET /api/pacientes/:id/recetas/:rid/pdf: la receta con el encabezado de la clínica. Un borrador o
// una receta anulada se identifican como tales en el documento.
func (ctrl *RecetasController) PDFReceta(c *fiber.Ctx) error {
	pacienteID, recetaID, ok := idsDeReceta(c)
	if !ok {
		return errorCliente(c, fiber.StatusBadRequest, "Identificador inválido")
	}
	contenido, receta, err := ctrl.GenerarPDF(c.Context(), pacienteID, recetaID)
	if errors.Is(err, models.ErrNoEncontrado) {
		return errorDominio(c, err)
	}
	if err != nil {
		return errorInterno(c, err)
	}
	middleware.MarcarAuditoria(c, pacienteID, recetaID)
	return enviarPDF(c, "receta-"+receta.Numero+".pdf", contenido)
}
