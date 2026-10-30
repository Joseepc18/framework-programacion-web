// Este archivo atiende el portal del paciente (F2-8). El paciente ingresa con su documento de identidad y un código OTP,
// y solo ve y modifica lo suyo: el paciente se deduce siempre de la cuenta del token, nunca de la petición.
package controllers

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
	"sigo-dental-backend/almacenamiento"
	"sigo-dental-backend/middleware"
	"sigo-dental-backend/models"
	"sigo-dental-backend/notificaciones"
	"sigo-dental-backend/otp"
)

// PortalRepository es el contrato de la capa de datos del portal.
type PortalRepository interface {
	BuscarPacientePorDocumento(ctx context.Context, documento string) (int, bool, error)
	ObtenerPaciente(ctx context.Context, id int) (models.PacienteDetalle, error)
	UsuarioDePaciente(ctx context.Context, pacienteID int, hashInicial string) (int, string, error)
	PacienteDePortal(ctx context.Context, usuarioID int) (int, bool, error)
	CitasDePaciente(ctx context.Context, pacienteID int) ([]models.CitaPortal, error)
	ObtenerCita(ctx context.Context, id int) (models.CitaDetalle, error)
	CambiarEstado(ctx context.Context, id int, desde, hasta, motivo string, usuarioID int) error
	ListarRecetas(ctx context.Context, pacienteID int) ([]models.Receta, error)
	ListarConsentimientos(ctx context.Context, pacienteID int) ([]models.Consentimiento, error)
	ObtenerConsentimiento(ctx context.Context, pacienteID, id int) (models.Consentimiento, error)
	ObtenerArchivo(ctx context.Context, id int) (models.Archivo, error)
}

// PortalController atiende las rutas del portal.
type PortalController struct {
	repo    PortalRepository
	recetas *RecetasController
	codigos CodigosOTP
	envio   EnvioDeCodigos
	almacen almacenamiento.AlmacenArchivos
	secreto []byte
	ahora   func() time.Time

	mu     sync.Mutex
	envios map[int]time.Time
}

// NuevoPortalController construye el controlador del portal.
func NuevoPortalController(repo PortalRepository, recetas *RecetasController, codigos CodigosOTP, envio EnvioDeCodigos, almacen almacenamiento.AlmacenArchivos, secreto []byte) *PortalController {
	return &PortalController{repo: repo, recetas: recetas, codigos: codigos, envio: envio, almacen: almacen, secreto: secreto, ahora: time.Now, envios: map[int]time.Time{}}
}

// esperaEntreCodigos evita que alguien use el portal para inundar de mensajes al teléfono de un paciente.
const esperaEntreCodigos = 60 * time.Second

type cuerpoDocumento struct {
	Documento string `json:"documento"`
	Codigo    string `json:"codigo"`
}

// SolicitarCodigo responde POST /api/portal/codigo (público). La respuesta es siempre la misma, exista o no el
// documento, para no revelar quién es paciente de la clínica.
func (ctrl *PortalController) SolicitarCodigo(c *fiber.Ctx) error {
	var cuerpo cuerpoDocumento
	if err := c.BodyParser(&cuerpo); err != nil {
		return cuerpoInvalido(c)
	}
	documento, ok := textoObligatorio(cuerpo.Documento, maxDocumento)
	if !ok {
		return errorCliente(c, fiber.StatusBadRequest, "Escribe tu documento de identidad")
	}
	respuesta := fiber.Map{"mensaje": "Si el documento está registrado, enviamos un código al teléfono o correo de la ficha. Vence en 10 minutos."}

	pacienteID, existe, err := ctrl.repo.BuscarPacientePorDocumento(c.Context(), documento)
	if err != nil {
		return errorInterno(c, err)
	}
	if !existe || !ctrl.puedeEnviar(pacienteID) {
		return c.JSON(respuesta)
	}
	paciente, err := ctrl.repo.ObtenerPaciente(c.Context(), pacienteID)
	if err != nil {
		return errorInterno(c, err)
	}
	// Un menor recibe el código su tutor legal.
	firmante := models.FirmantePaciente
	if paciente.Edad < mayoriaDeEdad {
		firmante = models.FirmanteTutor
	}
	canal, destino := destinoDelCodigo(paciente, firmante)
	if strings.TrimSpace(destino) == "" {
		return c.JSON(respuesta)
	}
	codigo, _, err := ctrl.codigos.Emitir(c.Context(), otp.Portal, pacienteID)
	if err != nil {
		return errorInterno(c, err)
	}
	if _, err := ctrl.envio.EnviarCodigo(c.Context(), notificaciones.DestinoCodigo{
		PacienteID: pacienteID, Canal: canal, Destino: destino, Tipo: models.NotifOTPPortal, Referencia: "ingresar a tu portal de SIGO-Dental",
	}, codigo, otp.Vigencia); err != nil {
		return errorInterno(c, err)
	}
	return c.JSON(respuesta)
}

// puedeEnviar aplica la espera mínima entre dos códigos del mismo paciente.
func (ctrl *PortalController) puedeEnviar(pacienteID int) bool {
	ctrl.mu.Lock()
	defer ctrl.mu.Unlock()
	if ultimo, ok := ctrl.envios[pacienteID]; ok && ctrl.ahora().Sub(ultimo) < esperaEntreCodigos {
		return false
	}
	ctrl.envios[pacienteID] = ctrl.ahora()
	return true
}

// Ingresar responde POST /api/portal/ingresar (público): valida el código y entrega un token con el rol paciente.
func (ctrl *PortalController) Ingresar(c *fiber.Ctx) error {
	var cuerpo cuerpoDocumento
	if err := c.BodyParser(&cuerpo); err != nil {
		return cuerpoInvalido(c)
	}
	documento, ok := textoObligatorio(cuerpo.Documento, maxDocumento)
	codigo := strings.TrimSpace(cuerpo.Codigo)
	if !ok || codigo == "" {
		return errorCliente(c, fiber.StatusBadRequest, "Escribe tu documento y el código que recibiste")
	}
	pacienteID, existe, err := ctrl.repo.BuscarPacientePorDocumento(c.Context(), documento)
	if err != nil {
		return errorInterno(c, err)
	}
	if !existe {
		// Misma respuesta que un código incorrecto.
		return errorCliente(c, fiber.StatusBadRequest, "El código no es correcto")
	}
	if err := ctrl.codigos.Validar(c.Context(), otp.Portal, pacienteID, codigo); err != nil {
		return errorCodigo(c, err)
	}
	// La cuenta del portal nace con una contraseña aleatoria que nadie conoce: solo se entra con código.
	aleatorio := make([]byte, 24)
	if _, err := rand.Read(aleatorio); err != nil {
		return errorInterno(c, err)
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(hex.EncodeToString(aleatorio)), bcrypt.DefaultCost)
	if err != nil {
		return errorInterno(c, err)
	}
	usuarioID, email, err := ctrl.repo.UsuarioDePaciente(c.Context(), pacienteID, string(hash))
	if err != nil {
		return errorInterno(c, err)
	}
	token, err := middleware.GenerarToken(ctrl.secreto, usuarioID, email, models.RolPaciente)
	if err != nil {
		return errorInterno(c, err)
	}
	return c.JSON(fiber.Map{"token": token, "email": email, "rol": models.RolPaciente})
}

// pacienteDelToken devuelve el paciente de la cuenta con sesión; la identidad sale del token, nunca de la petición.
func (ctrl *PortalController) pacienteDelToken(c *fiber.Ctx) (int, bool) {
	id, ok, err := ctrl.repo.PacienteDePortal(c.Context(), usuarioDeLaPeticion(c))
	if err != nil {
		_ = errorInterno(c, err)
		return 0, false
	}
	if !ok {
		_ = errorCliente(c, fiber.StatusForbidden, "Esta cuenta no tiene un paciente asociado")
		return 0, false
	}
	return id, true
}

// Perfil responde GET /api/portal/yo.
func (ctrl *PortalController) Perfil(c *fiber.Ctx) error {
	pacienteID, ok := ctrl.pacienteDelToken(c)
	if !ok {
		return nil
	}
	p, err := ctrl.repo.ObtenerPaciente(c.Context(), pacienteID)
	if err != nil {
		return errorDominio(c, err)
	}
	middleware.MarcarAuditoria(c, pacienteID, pacienteID)
	return c.JSON(fiber.Map{"nombres": p.Nombres, "apellidos": p.Apellidos, "numero_historia": p.NumeroHistoria})
}

// ListarCitas responde GET /api/portal/citas.
func (ctrl *PortalController) ListarCitas(c *fiber.Ctx) error {
	pacienteID, ok := ctrl.pacienteDelToken(c)
	if !ok {
		return nil
	}
	lista, err := ctrl.repo.CitasDePaciente(c.Context(), pacienteID)
	if err != nil {
		return errorDominio(c, err)
	}
	middleware.MarcarAuditoria(c, pacienteID, 0)
	return c.JSON(lista)
}

// citaPropia carga la cita y comprueba que sea del paciente de la sesión. Si es de otro responde 404, igual que si no
// existiera: no se revela que existe.
func (ctrl *PortalController) citaPropia(c *fiber.Ctx) (models.CitaDetalle, int, bool) {
	pacienteID, ok := ctrl.pacienteDelToken(c)
	if !ok {
		return models.CitaDetalle{}, 0, false
	}
	id, okID := idDeRuta(c)
	if !okID {
		_ = errorCliente(c, fiber.StatusBadRequest, "Identificador inválido")
		return models.CitaDetalle{}, 0, false
	}
	cita, err := ctrl.repo.ObtenerCita(c.Context(), id)
	if errors.Is(err, models.ErrNoEncontrado) || (err == nil && cita.PacienteID != pacienteID) {
		_ = errorDominio(c, models.ErrNoEncontrado)
		return models.CitaDetalle{}, 0, false
	}
	if err != nil {
		_ = errorDominio(c, err)
		return models.CitaDetalle{}, 0, false
	}
	return cita, pacienteID, true
}

// ConfirmarCita responde POST /api/portal/citas/:id/confirmar: pasa a Confirmado y queda en el historial de la cita.
func (ctrl *PortalController) ConfirmarCita(c *fiber.Ctx) error {
	cita, pacienteID, ok := ctrl.citaPropia(c)
	if !ok {
		return nil
	}
	if cita.Estado != models.EstadoProgramado || !cita.Inicio.After(ctrl.ahora()) {
		return errorCliente(c, fiber.StatusConflict, "Esta cita ya no se puede confirmar")
	}
	if err := ctrl.repo.CambiarEstado(c.Context(), cita.ID, models.EstadoProgramado, models.EstadoConfirmado, "Confirmada por el paciente desde el portal", usuarioDeLaPeticion(c)); err != nil {
		return errorEstadoPortal(c, err)
	}
	middleware.MarcarAuditoria(c, pacienteID, cita.ID)
	return ctrl.citaActualizada(c, cita.ID)
}

type cuerpoCancelacion struct {
	Motivo string `json:"motivo"`
}

// CancelarCita responde POST /api/portal/citas/:id/cancelar: solo una cita futura sin atender.
func (ctrl *PortalController) CancelarCita(c *fiber.Ctx) error {
	cita, pacienteID, ok := ctrl.citaPropia(c)
	if !ok {
		return nil
	}
	var cuerpo cuerpoCancelacion
	_ = c.BodyParser(&cuerpo)
	motivo, okMotivo := textoOpcional(cuerpo.Motivo, 200)
	if !okMotivo {
		return errorCliente(c, fiber.StatusBadRequest, "El motivo admite hasta 200 caracteres")
	}
	if !slices.Contains([]string{models.EstadoProgramado, models.EstadoConfirmado}, cita.Estado) || !cita.Inicio.After(ctrl.ahora()) {
		return errorCliente(c, fiber.StatusConflict, "Esta cita ya no se puede cancelar desde el portal: llama a la clínica")
	}
	texto := "Cancelada por el paciente desde el portal"
	if motivo != "" {
		texto += ": " + motivo
	}
	if err := ctrl.repo.CambiarEstado(c.Context(), cita.ID, cita.Estado, models.EstadoCancelado, texto, usuarioDeLaPeticion(c)); err != nil {
		return errorEstadoPortal(c, err)
	}
	middleware.MarcarAuditoria(c, pacienteID, cita.ID)
	return ctrl.citaActualizada(c, cita.ID)
}

func errorEstadoPortal(c *fiber.Ctx, err error) error {
	if errors.Is(err, models.ErrEstadoCambiado) {
		return errorCliente(c, fiber.StatusConflict, "La cita cambió de estado: actualiza la pantalla")
	}
	return errorDominio(c, err)
}

func (ctrl *PortalController) citaActualizada(c *fiber.Ctx, id int) error {
	actualizada, err := ctrl.repo.ObtenerCita(c.Context(), id)
	if err != nil {
		return errorDominio(c, err)
	}
	return c.JSON(models.CitaPortal{
		ID: actualizada.ID, Estado: actualizada.Estado, Inicio: actualizada.Inicio, Fin: actualizada.Fin,
		ProfesionalNombre: actualizada.ProfesionalNombre, ConsultorioNombre: actualizada.ConsultorioNombre,
		ProcedimientoNombre: actualizada.ProcedimientoNombre, Motivo: actualizada.Motivo,
	})
}

// ListarRecetas responde GET /api/portal/recetas: solo las firmadas o anuladas (un borrador no es del paciente).
func (ctrl *PortalController) ListarRecetas(c *fiber.Ctx) error {
	pacienteID, ok := ctrl.pacienteDelToken(c)
	if !ok {
		return nil
	}
	todas, err := ctrl.repo.ListarRecetas(c.Context(), pacienteID)
	if err != nil {
		return errorDominio(c, err)
	}
	visibles := []models.Receta{}
	for _, r := range todas {
		if r.Estado != models.RecetaBorrador {
			r.Avisos = []models.AvisoAlergia{}
			visibles = append(visibles, r)
		}
	}
	middleware.MarcarAuditoria(c, pacienteID, 0)
	return c.JSON(visibles)
}

// PDFReceta responde GET /api/portal/recetas/:id/pdf: solo de recetas propias y firmadas.
func (ctrl *PortalController) PDFReceta(c *fiber.Ctx) error {
	pacienteID, ok := ctrl.pacienteDelToken(c)
	if !ok {
		return nil
	}
	id, okID := idDeRuta(c)
	if !okID {
		return errorCliente(c, fiber.StatusBadRequest, "Identificador inválido")
	}
	contenido, receta, err := ctrl.recetas.GenerarPDF(c.Context(), pacienteID, id)
	if errors.Is(err, models.ErrNoEncontrado) || (err == nil && receta.Estado == models.RecetaBorrador) {
		return errorDominio(c, models.ErrNoEncontrado)
	}
	if err != nil {
		return errorInterno(c, err)
	}
	middleware.MarcarAuditoria(c, pacienteID, id)
	return enviarPDF(c, "receta-"+receta.Numero+".pdf", contenido)
}

// ListarConsentimientos responde GET /api/portal/consentimientos: solo los firmados.
func (ctrl *PortalController) ListarConsentimientos(c *fiber.Ctx) error {
	pacienteID, ok := ctrl.pacienteDelToken(c)
	if !ok {
		return nil
	}
	todos, err := ctrl.repo.ListarConsentimientos(c.Context(), pacienteID)
	if err != nil {
		return errorDominio(c, err)
	}
	firmados := []models.Consentimiento{}
	for _, k := range todos {
		if k.Estado == models.ConsentFirmado {
			firmados = append(firmados, k)
		}
	}
	middleware.MarcarAuditoria(c, pacienteID, 0)
	return c.JSON(firmados)
}

// PDFConsentimiento responde GET /api/portal/consentimientos/:id/pdf: el PDF firmado de un consentimiento propio.
func (ctrl *PortalController) PDFConsentimiento(c *fiber.Ctx) error {
	pacienteID, ok := ctrl.pacienteDelToken(c)
	if !ok {
		return nil
	}
	id, okID := idDeRuta(c)
	if !okID {
		return errorCliente(c, fiber.StatusBadRequest, "Identificador inválido")
	}
	consentimiento, err := ctrl.repo.ObtenerConsentimiento(c.Context(), pacienteID, id)
	if err != nil || consentimiento.Estado != models.ConsentFirmado || consentimiento.ArchivoID == nil {
		return errorDominio(c, models.ErrNoEncontrado)
	}
	archivo, err := ctrl.repo.ObtenerArchivo(c.Context(), *consentimiento.ArchivoID)
	if err != nil || archivo.PacienteID != pacienteID {
		return errorDominio(c, models.ErrNoEncontrado)
	}
	return enviarArchivo(c, ctrl.almacen, archivo)
}
