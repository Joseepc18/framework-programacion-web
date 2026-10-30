// Este archivo atiende el consentimiento informado digital (F2-5: RF-06): plantillas por procedimiento con campos que se
// completan solos, firma manuscrita en pantalla táctil o código OTP enviado al paciente (al tutor si es menor), PDF
// firmado e inmutable, y revocación como documento nuevo. El router fija los permisos.
package controllers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"sigo-dental-backend/almacenamiento"
	"sigo-dental-backend/documentos"
	"sigo-dental-backend/middleware"
	"sigo-dental-backend/models"
	"sigo-dental-backend/notificaciones"
	"sigo-dental-backend/otp"
)

// ConsentimientoRepository es el contrato de la capa de datos de los consentimientos.
type ConsentimientoRepository interface {
	ObtenerPaciente(ctx context.Context, id int) (models.PacienteDetalle, error)
	ProfesionalDeUsuario(ctx context.Context, usuarioID int) (models.ProfesionalDeUsuario, bool, error)
	ProcedimientoDeCatalogo(ctx context.Context, id int) (nombre string, requiere bool, err error)
	PlantillaDe(ctx context.Context, procedimientoID int) (models.Plantilla, error)
	ListarPlantillas(ctx context.Context) ([]models.Plantilla, error)
	GuardarPlantilla(ctx context.Context, id int, p models.Plantilla) (int, error)
	CrearConsentimiento(ctx context.Context, n models.NuevoConsentimiento, usuarioID int) (int, error)
	ListarConsentimientos(ctx context.Context, pacienteID int) ([]models.Consentimiento, error)
	ObtenerConsentimiento(ctx context.Context, pacienteID, id int) (models.Consentimiento, error)
	FirmarConsentimiento(ctx context.Context, pacienteID, id int, f models.DatosFirmaConsentimiento) error
	ListarNotificacionesPaciente(ctx context.Context, pacienteID int) ([]models.Notificacion, error)
	ObtenerParametro(ctx context.Context, clave string) (string, error)
}

// CodigosOTP emite y valida los códigos de un solo uso.
type CodigosOTP interface {
	Emitir(ctx context.Context, proposito string, referenciaID int) (string, time.Time, error)
	Validar(ctx context.Context, proposito string, referenciaID int, codigo string) error
}

// EnvioDeCodigos entrega un código al paciente (o a su tutor) y lo deja en la bitácora de notificaciones.
type EnvioDeCodigos interface {
	EnviarCodigo(ctx context.Context, d notificaciones.DestinoCodigo, codigo string, vigencia time.Duration) (string, error)
}

// GuardadorDeArchivos guarda el PDF firmado en el almacén y lo registra como archivo del paciente.
type GuardadorDeArchivos interface {
	Guardar(ctx context.Context, n models.NuevoArchivo, contenido []byte) (int, error)
}

var _ GuardadorDeArchivos = (*almacenamiento.Servicio)(nil)

// ConsentimientosController atiende plantillas, consentimientos y revocaciones.
type ConsentimientosController struct {
	repo     ConsentimientoRepository
	codigos  CodigosOTP
	envio    EnvioDeCodigos
	archivos GuardadorDeArchivos
	ahora    func() time.Time
}

// NuevoConsentimientosController construye el controlador de consentimientos.
func NuevoConsentimientosController(repo ConsentimientoRepository, codigos CodigosOTP, envio EnvioDeCodigos, archivos GuardadorDeArchivos) *ConsentimientosController {
	return &ConsentimientosController{repo: repo, codigos: codigos, envio: envio, archivos: archivos, ahora: time.Now}
}

// --- Plantillas ---

// ListarPlantillas responde GET /api/consentimiento-plantillas.
func (ctrl *ConsentimientosController) ListarPlantillas(c *fiber.Ctx) error {
	lista, err := ctrl.repo.ListarPlantillas(c.Context())
	if err != nil {
		return errorDominio(c, err)
	}
	return c.JSON(lista)
}

type cuerpoPlantilla struct {
	ProcedimientoID *int   `json:"procedimiento_id"`
	Titulo          string `json:"titulo"`
	Cuerpo          string `json:"cuerpo"`
	Activo          *bool  `json:"activo"`
}

func (ctrl *ConsentimientosController) guardarPlantilla(c *fiber.Ctx, id int) error {
	var cuerpo cuerpoPlantilla
	if err := c.BodyParser(&cuerpo); err != nil {
		return cuerpoInvalido(c)
	}
	titulo, okTitulo := textoObligatorio(cuerpo.Titulo, 150)
	texto, okCuerpo := textoObligatorio(cuerpo.Cuerpo, 8000)
	if !okTitulo || !okCuerpo {
		return errorCliente(c, fiber.StatusBadRequest, "El título (hasta 150 caracteres) y el texto (hasta 8000) son obligatorios")
	}
	if cuerpo.ProcedimientoID != nil && *cuerpo.ProcedimientoID <= 0 {
		cuerpo.ProcedimientoID = nil
	}
	activo := true
	if cuerpo.Activo != nil {
		activo = *cuerpo.Activo
	}
	guardada, err := ctrl.repo.GuardarPlantilla(c.Context(), id, models.Plantilla{ProcedimientoID: cuerpo.ProcedimientoID, Titulo: titulo, Cuerpo: texto, Activo: activo})
	if errors.Is(err, models.ErrDuplicado) {
		return errorCliente(c, fiber.StatusConflict, "Ya hay una plantilla activa para ese procedimiento: desactívala primero")
	}
	if err != nil {
		return errorDominio(c, err)
	}
	status := fiber.StatusOK
	if id == 0 {
		status = fiber.StatusCreated
	}
	lista, err := ctrl.repo.ListarPlantillas(c.Context())
	if err != nil {
		return errorDominio(c, err)
	}
	for _, p := range lista {
		if p.ID == guardada {
			return c.Status(status).JSON(p)
		}
	}
	return c.SendStatus(status)
}

// CrearPlantilla responde POST /api/consentimiento-plantillas.
func (ctrl *ConsentimientosController) CrearPlantilla(c *fiber.Ctx) error {
	return ctrl.guardarPlantilla(c, 0)
}

// ActualizarPlantilla responde PUT /api/consentimiento-plantillas/:id. Los consentimientos ya preparados conservan el
// texto con el que se mostraron: cambiar la plantilla no altera lo firmado.
func (ctrl *ConsentimientosController) ActualizarPlantilla(c *fiber.Ctx) error {
	id, ok := idDeRuta(c)
	if !ok {
		return errorCliente(c, fiber.StatusBadRequest, "Identificador inválido")
	}
	return ctrl.guardarPlantilla(c, id)
}

// --- Consentimientos ---

func (ctrl *ConsentimientosController) zona(ctx context.Context) *time.Location {
	zona, err := time.LoadLocation(parametroTexto(ctx, ctrl.repo, "zona_horaria", "America/Guayaquil"))
	if err != nil {
		return time.UTC
	}
	return zona
}

// ListarConsentimientos responde GET /api/pacientes/:id/consentimientos.
func (ctrl *ConsentimientosController) ListarConsentimientos(c *fiber.Ctx) error {
	id, ok := idDeRuta(c)
	if !ok {
		return errorCliente(c, fiber.StatusBadRequest, "Identificador inválido")
	}
	lista, err := ctrl.repo.ListarConsentimientos(c.Context(), id)
	if err != nil {
		return errorDominio(c, err)
	}
	return c.JSON(lista)
}

// ListarNotificaciones responde GET /api/pacientes/:id/notificaciones: la bitácora de códigos enviados. Con el
// adaptador simulado el código es visible para poder probar; con uno real se guarda enmascarado.
func (ctrl *ConsentimientosController) ListarNotificaciones(c *fiber.Ctx) error {
	id, ok := idDeRuta(c)
	if !ok {
		return errorCliente(c, fiber.StatusBadRequest, "Identificador inválido")
	}
	lista, err := ctrl.repo.ListarNotificacionesPaciente(c.Context(), id)
	if err != nil {
		return errorDominio(c, err)
	}
	return c.JSON(lista)
}

type cuerpoConsentimiento struct {
	ProcedimientoID int  `json:"procedimiento_id"`
	Pieza           *int `json:"pieza"`
}

// firmanteDe decide quién firma: el paciente, o su tutor legal si es menor de edad.
func firmanteDe(p models.PacienteDetalle) (tipo, nombre, documento string, motivo string) {
	if p.Edad < mayoriaDeEdad {
		if p.TutorNombre == "" || p.TutorDocumento == "" {
			return "", "", "", "El paciente es menor de edad y no tiene tutor legal registrado: complétalo en su ficha"
		}
		return models.FirmanteTutor, p.TutorNombre, p.TutorDocumento, ""
	}
	return models.FirmantePaciente, nombreCompleto(p.Nombres, p.Apellidos), p.Documento, ""
}

func nombreCompleto(nombres, apellidos string) string {
	return strings.TrimSpace(nombres + " " + apellidos)
}

// CrearConsentimiento responde POST /api/pacientes/:id/consentimientos: prepara el documento con la plantilla del
// procedimiento y los campos completados. Queda pendiente hasta que se firma.
func (ctrl *ConsentimientosController) CrearConsentimiento(c *fiber.Ctx) error {
	pacienteID, ok := idDeRuta(c)
	if !ok {
		return errorCliente(c, fiber.StatusBadRequest, "Identificador inválido")
	}
	var cuerpo cuerpoConsentimiento
	if err := c.BodyParser(&cuerpo); err != nil || cuerpo.ProcedimientoID <= 0 {
		return errorCliente(c, fiber.StatusBadRequest, "Elige el procedimiento del consentimiento")
	}
	if cuerpo.Pieza != nil && !models.PiezaFDIValida(*cuerpo.Pieza) {
		return errorCliente(c, fiber.StatusBadRequest, "La pieza dental no es válida (numeración FDI)")
	}
	usuarioID := usuarioDeLaPeticion(c)
	profesional, ok, err := ctrl.repo.ProfesionalDeUsuario(c.Context(), usuarioID)
	if err != nil {
		return errorDominio(c, err)
	}
	if !ok {
		return errorEvolucion(c, models.ErrNoEsProfesional)
	}
	paciente, err := ctrl.repo.ObtenerPaciente(c.Context(), pacienteID)
	if err != nil {
		return errorDominio(c, err)
	}
	nombreProc, _, err := ctrl.repo.ProcedimientoDeCatalogo(c.Context(), cuerpo.ProcedimientoID)
	if err != nil {
		return errorCliente(c, fiber.StatusBadRequest, "El procedimiento no existe o está desactivado")
	}
	plantilla, err := ctrl.repo.PlantillaDe(c.Context(), cuerpo.ProcedimientoID)
	if err != nil {
		return errorCliente(c, fiber.StatusConflict, "No hay una plantilla de consentimiento activa para este procedimiento")
	}
	tipoFirmante, nombreFirmante, documentoFirmante, motivo := firmanteDe(paciente)
	if motivo != "" {
		return errorCliente(c, fiber.StatusConflict, motivo)
	}

	pieza := ""
	if cuerpo.Pieza != nil {
		pieza = fmt.Sprintf("en la pieza %d", *cuerpo.Pieza)
	}
	texto := models.RellenarPlantilla(plantilla.Cuerpo, models.DatosPlantilla{
		Paciente: nombreFirmante, Documento: documentoFirmante, Procedimiento: nombreProc, Pieza: pieza,
		Profesional: profesional.Nombre, Fecha: ctrl.ahora().In(ctrl.zona(c.Context())).Format("02/01/2006"),
	})
	if tipoFirmante == models.FirmanteTutor {
		texto = fmt.Sprintf("Consentimiento otorgado por el tutor legal %s (documento %s) en representación del paciente menor de edad %s.\n\n%s",
			nombreFirmante, documentoFirmante, nombreCompleto(paciente.Nombres, paciente.Apellidos), texto)
	}
	plantillaID := plantilla.ID
	id, err := ctrl.repo.CrearConsentimiento(c.Context(), models.NuevoConsentimiento{
		PacienteID: pacienteID, ProcedimientoID: cuerpo.ProcedimientoID, ProfesionalID: profesional.ID, PlantillaID: &plantillaID,
		Tipo: models.ConsentTipoConsentimiento, Titulo: plantilla.Titulo, Texto: texto,
		FirmanteTipo: tipoFirmante, FirmanteNombre: nombreFirmante, FirmanteDocumento: documentoFirmante,
	}, usuarioID)
	if err != nil {
		return errorDominio(c, err)
	}
	middleware.MarcarAuditoria(c, pacienteID, id)
	creado, err := ctrl.repo.ObtenerConsentimiento(c.Context(), pacienteID, id)
	if err != nil {
		return errorDominio(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(creado)
}

type cuerpoRevocacion struct {
	Motivo string `json:"motivo"`
}

// RevocarConsentimiento responde POST /api/pacientes/:id/consentimientos/:cid/revocar. Una revocación es un documento
// nuevo, enlazado al consentimiento, que también se firma; el consentimiento original no se modifica.
func (ctrl *ConsentimientosController) RevocarConsentimiento(c *fiber.Ctx) error {
	pacienteID, consentimientoID, ok := idsDeConsentimiento(c)
	if !ok {
		return errorCliente(c, fiber.StatusBadRequest, "Identificador inválido")
	}
	var cuerpo cuerpoRevocacion
	if err := c.BodyParser(&cuerpo); err != nil {
		return cuerpoInvalido(c)
	}
	motivo, okMotivo := textoObligatorio(cuerpo.Motivo, 300)
	if !okMotivo {
		return errorCliente(c, fiber.StatusBadRequest, "Indica el motivo de la revocación (hasta 300 caracteres)")
	}
	original, err := ctrl.repo.ObtenerConsentimiento(c.Context(), pacienteID, consentimientoID)
	if err != nil {
		return errorDominio(c, err)
	}
	if original.Tipo != models.ConsentTipoConsentimiento || original.Estado != models.ConsentFirmado || original.Revocado {
		return errorCliente(c, fiber.StatusConflict, "Solo se revoca un consentimiento firmado que no esté revocado")
	}
	usuarioID := usuarioDeLaPeticion(c)
	profesional, ok, err := ctrl.repo.ProfesionalDeUsuario(c.Context(), usuarioID)
	if err != nil {
		return errorDominio(c, err)
	}
	if !ok {
		return errorEvolucion(c, models.ErrNoEsProfesional)
	}
	fecha := ctrl.ahora().In(ctrl.zona(c.Context())).Format("02/01/2006")
	firmadoEl := fecha
	if original.FirmadoEn != nil {
		firmadoEl = original.FirmadoEn.In(ctrl.zona(c.Context())).Format("02/01/2006")
	}
	texto := fmt.Sprintf("Yo, %s, con documento de identidad %s, revoco el consentimiento informado que otorgué el %s para el procedimiento «%s». Motivo: %s.\n\nEntiendo que el procedimiento no se realizará mientras no otorgue un nuevo consentimiento.\n\nFecha: %s.",
		original.FirmanteNombre, original.FirmanteDocumento, firmadoEl, original.ProcedimientoNombre, motivo, fecha)
	revoca := original.ID
	id, err := ctrl.repo.CrearConsentimiento(c.Context(), models.NuevoConsentimiento{
		PacienteID: pacienteID, ProcedimientoID: original.ProcedimientoID, ProfesionalID: profesional.ID,
		Tipo: models.ConsentTipoRevocacion, RevocaA: &revoca, Motivo: motivo, Titulo: "Revocación del consentimiento informado",
		Texto: texto, FirmanteTipo: original.FirmanteTipo, FirmanteNombre: original.FirmanteNombre, FirmanteDocumento: original.FirmanteDocumento,
	}, usuarioID)
	if err != nil {
		return errorDominio(c, err)
	}
	middleware.MarcarAuditoria(c, pacienteID, id)
	creada, err := ctrl.repo.ObtenerConsentimiento(c.Context(), pacienteID, id)
	if err != nil {
		return errorDominio(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(creada)
}

func idsDeConsentimiento(c *fiber.Ctx) (pacienteID, consentimientoID int, ok bool) {
	pacienteID, ok = idDeRuta(c)
	if !ok {
		return 0, 0, false
	}
	consentimientoID, err := c.ParamsInt("cid")
	return pacienteID, consentimientoID, err == nil && consentimientoID > 0
}

// destinoDelCodigo decide a dónde va el código: al teléfono del tutor (por SMS) si firma el tutor; si no, al canal
// preferido del paciente.
func destinoDelCodigo(p models.PacienteDetalle, firmanteTipo string) (canal, destino string) {
	if firmanteTipo == models.FirmanteTutor {
		return "sms", p.TutorTelefono
	}
	if p.CanalPreferido == "email" {
		return "email", p.Email
	}
	return p.CanalPreferido, p.Telefono
}

// EnviarCodigo responde POST /api/pacientes/:id/consentimientos/:cid/otp: envía al paciente (o su tutor) un código de
// un solo uso para firmar. Un código nuevo inutiliza el anterior.
func (ctrl *ConsentimientosController) EnviarCodigo(c *fiber.Ctx) error {
	pacienteID, consentimientoID, ok := idsDeConsentimiento(c)
	if !ok {
		return errorCliente(c, fiber.StatusBadRequest, "Identificador inválido")
	}
	consentimiento, err := ctrl.repo.ObtenerConsentimiento(c.Context(), pacienteID, consentimientoID)
	if err != nil {
		return errorDominio(c, err)
	}
	if consentimiento.Estado != models.ConsentPendiente {
		return errorEvolucion(c, models.ErrConsentimientoFirmado)
	}
	paciente, err := ctrl.repo.ObtenerPaciente(c.Context(), pacienteID)
	if err != nil {
		return errorDominio(c, err)
	}
	canal, destino := destinoDelCodigo(paciente, consentimiento.FirmanteTipo)
	if strings.TrimSpace(destino) == "" {
		return errorCliente(c, fiber.StatusConflict, "No hay un teléfono o correo registrado para enviar el código: firma en pantalla")
	}
	codigo, expira, err := ctrl.codigos.Emitir(c.Context(), otp.Consentimiento, consentimientoID)
	if err != nil {
		return errorInterno(c, err)
	}
	enmascarado, err := ctrl.envio.EnviarCodigo(c.Context(), notificaciones.DestinoCodigo{
		PacienteID: pacienteID, Canal: canal, Destino: destino, Tipo: models.NotifOTPConsentimiento,
		Referencia: "firmar el consentimiento de " + consentimiento.ProcedimientoNombre,
	}, codigo, otp.Vigencia)
	if err != nil {
		return errorCliente(c, fiber.StatusBadGateway, "No se pudo enviar el código: "+err.Error())
	}
	middleware.MarcarAuditoria(c, pacienteID, consentimientoID)
	return c.JSON(fiber.Map{"enviado_a": enmascarado, "canal": canal, "expira_en": expira})
}

type cuerpoFirmaConsentimiento struct {
	Modalidad   string                  `json:"modalidad"`
	Codigo      string                  `json:"codigo"`
	Firma       *models.FirmaManuscrita `json:"firma"`
	Dispositivo string                  `json:"dispositivo"`
}

// errorCodigo traduce los errores de validación del OTP.
func errorCodigo(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, otp.ErrInvalido):
		return errorCliente(c, fiber.StatusBadRequest, "El código no es correcto")
	case errors.Is(err, otp.ErrAgotado):
		return errorCliente(c, fiber.StatusTooManyRequests, "Se agotaron los 3 intentos del código: solicita uno nuevo")
	case errors.Is(err, otp.ErrExpirado):
		return errorCliente(c, fiber.StatusGone, "El código venció (dura 10 minutos): solicita uno nuevo")
	case errors.Is(err, otp.ErrUsado):
		return errorCliente(c, fiber.StatusGone, "El código ya se usó o fue reemplazado: solicita uno nuevo")
	case errors.Is(err, otp.ErrNoExiste):
		return errorCliente(c, fiber.StatusConflict, "Primero solicita el código")
	}
	return errorInterno(c, err)
}

// FirmarConsentimiento responde POST /api/pacientes/:id/consentimientos/:cid/firmar. Con modalidad "manuscrita" guarda
// el trazo, la marca de tiempo y el dispositivo; con "otp" valida el código enviado. Genera el PDF, lo guarda en el
// almacén (inmutable, queda en la auditoría) y cierra el consentimiento.
func (ctrl *ConsentimientosController) FirmarConsentimiento(c *fiber.Ctx) error {
	pacienteID, consentimientoID, ok := idsDeConsentimiento(c)
	if !ok {
		return errorCliente(c, fiber.StatusBadRequest, "Identificador inválido")
	}
	var cuerpo cuerpoFirmaConsentimiento
	if err := c.BodyParser(&cuerpo); err != nil {
		return cuerpoInvalido(c)
	}
	consentimiento, err := ctrl.repo.ObtenerConsentimiento(c.Context(), pacienteID, consentimientoID)
	if err != nil {
		return errorDominio(c, err)
	}
	if consentimiento.Estado != models.ConsentPendiente {
		return errorEvolucion(c, models.ErrConsentimientoFirmado)
	}

	firmaDatos := models.DatosFirmaConsentimiento{Modalidad: cuerpo.Modalidad, FirmadoEn: ctrl.ahora()}
	dispositivo, _ := textoOpcional(cuerpo.Dispositivo, 300)
	var trazo *models.FirmaManuscrita
	switch cuerpo.Modalidad {
	case models.ModalidadManuscrita:
		if cuerpo.Firma == nil {
			return errorCliente(c, fiber.StatusBadRequest, "Dibuja la firma en la pantalla")
		}
		if motivo := models.ValidarFirmaManuscrita(*cuerpo.Firma); motivo != "" {
			return errorCliente(c, fiber.StatusBadRequest, motivo)
		}
		trazo = cuerpo.Firma
		trazo.Dispositivo = dispositivo
		datos, err := json.Marshal(trazo)
		if err != nil {
			return errorInterno(c, err)
		}
		firmaDatos.Trazo = datos
	case models.ModalidadOTP:
		if err := ctrl.codigos.Validar(c.Context(), otp.Consentimiento, consentimientoID, strings.TrimSpace(cuerpo.Codigo)); err != nil {
			return errorCodigo(c, err)
		}
	default:
		return errorCliente(c, fiber.StatusBadRequest, "La modalidad debe ser manuscrita u otp")
	}
	firmaDatos.Dispositivo = strings.TrimSpace(dispositivo + " | IP " + middleware.IPDelCliente(c))

	// El PDF refleja lo firmado: el texto tal cual se mostró, quién firmó, cómo y cuándo.
	paciente, err := ctrl.repo.ObtenerPaciente(c.Context(), pacienteID)
	if err != nil {
		return errorDominio(c, err)
	}
	firmado := consentimiento
	firmado.Modalidad, firmado.Dispositivo = cuerpo.Modalidad, firmaDatos.Dispositivo
	destinoOTP := ""
	if cuerpo.Modalidad == models.ModalidadOTP {
		_, destino := destinoDelCodigo(paciente, consentimiento.FirmanteTipo)
		destinoOTP = notificaciones.EnmascararDestino(destino)
	}
	pdf, err := documentos.ConsentimientoPDF(documentos.DatosConsentimientoPDF{
		Clinica:        ctrl.clinica(c.Context()),
		Consentimiento: firmado,
		Paciente:       nombreCompleto(paciente.Nombres, paciente.Apellidos) + " (doc. " + paciente.Documento + ")",
		FirmadoEn:      firmaDatos.FirmadoEn.In(ctrl.zona(c.Context())).Format("02/01/2006 15:04:05 -07:00"),
		Trazo:          trazo, CodigoOTPEnviadoA: destinoOTP,
	})
	if err != nil {
		return errorInterno(c, err)
	}
	descripcion := "Consentimiento informado firmado: " + consentimiento.ProcedimientoNombre
	if consentimiento.Tipo == models.ConsentTipoRevocacion {
		descripcion = "Revocación de consentimiento: " + consentimiento.ProcedimientoNombre
	}
	archivoID, err := ctrl.archivos.Guardar(c.Context(), models.NuevoArchivo{
		PacienteID: pacienteID, Categoria: models.ArchivoConsentimiento, Descripcion: descripcion,
		NombreOriginal: fmt.Sprintf("consentimiento-%d.pdf", consentimientoID), TipoMIME: "application/pdf", UsuarioID: usuarioDeLaPeticion(c),
	}, pdf)
	if err != nil {
		return errorDominio(c, err)
	}
	firmaDatos.ArchivoID = archivoID
	if err := ctrl.repo.FirmarConsentimiento(c.Context(), pacienteID, consentimientoID, firmaDatos); err != nil {
		return errorEvolucion(c, err)
	}
	middleware.MarcarAuditoria(c, pacienteID, consentimientoID)
	cerrado, err := ctrl.repo.ObtenerConsentimiento(c.Context(), pacienteID, consentimientoID)
	if err != nil {
		return errorDominio(c, err)
	}
	return c.JSON(cerrado)
}

// clinica lee los parámetros del encabezado de los documentos.
func (ctrl *ConsentimientosController) clinica(ctx context.Context) documentos.Clinica {
	return clinicaDe(ctx, ctrl.repo)
}

func clinicaDe(ctx context.Context, lector LectorParametros) documentos.Clinica {
	return documentos.Clinica{
		Nombre:    parametroTexto(ctx, lector, "clinica_nombre", "Clínica Odontológica"),
		Direccion: parametroTexto(ctx, lector, "clinica_direccion", ""),
		Telefono:  parametroTexto(ctx, lector, "clinica_telefono", ""),
	}
}
