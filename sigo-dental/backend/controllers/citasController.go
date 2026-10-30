// Este archivo atiende la agenda de citas (F1-3: RF-09, RF-10 y RN-03).
// La recepción agenda, reprograma y cambia cualquier estado; el odontólogo y el asistente solo pasan
// la cita a En Atención y Atendido; los tres consultan. El router fija quién entra a cada ruta.
package controllers

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"sigo-dental-backend/models"
	"sigo-dental-backend/notificaciones"
)

// CitaRepository es el contrato de la capa de datos de la agenda.
type CitaRepository interface {
	ListarCitas(ctx context.Context, filtro models.FiltroCitas) ([]models.Cita, error)
	ObtenerCita(ctx context.Context, id int) (models.CitaDetalle, error)
	// CrearCita agenda la cita; devuelve models.ErrConflicto si el sillón está ocupado (RN-03).
	CrearCita(ctx context.Context, datos models.DatosCita, usuarioID int) (int, error)
	ReprogramarCita(ctx context.Context, id int, datos models.DatosCita, usuarioID int) error
	CambiarEstado(ctx context.Context, id int, desde, hasta, motivo string, usuarioID int) error
	// CitasSimultaneasDelProfesional cuenta las citas vigentes del profesional que se cruzan con el rango.
	CitasSimultaneasDelProfesional(ctx context.Context, profesionalID int, inicio, fin time.Time, excluirID int) (int, error)
	HorarioAgenda(ctx context.Context) (models.HorarioAgenda, error)
	// ListarNotificaciones devuelve la bitácora de mensajes de la cita.
	ListarNotificaciones(ctx context.Context, citaID int) ([]models.Notificacion, error)
}

// AvisosDeCita es lo que el controlador necesita del servicio de notificaciones (F1-4). Es opcional: sin él
// la agenda funciona igual, solo que no se envían avisos.
type AvisosDeCita interface {
	CitaAgendada(ctx context.Context, citaID int)
	CitaReprogramada(ctx context.Context, citaID int)
}

// CitasController atiende las rutas de la agenda con el repositorio inyectado.
type CitasController struct {
	repo   CitaRepository
	avisos AvisosDeCita
	// secreto firma y verifica los enlaces de confirmación.
	secreto []byte
	ahora   func() time.Time
}

// NuevoCitasController construye el controlador de la agenda. avisos puede ser nil.
func NuevoCitasController(repo CitaRepository, avisos AvisosDeCita, secreto []byte) *CitasController {
	return &CitasController{repo: repo, avisos: avisos, secreto: secreto, ahora: time.Now}
}

const (
	maxMotivoCita   = 200
	maxNotasCita    = 1000
	maxMotivoCambio = 300
	// duracionMaximaCita evita citas de varios días por un error de captura.
	duracionMaximaCita = 12 * time.Hour
	// rangoMaximoAgenda acota cuánto calendario se puede pedir de una vez.
	rangoMaximoAgenda = 62 * 24 * time.Hour
)

var estadosDeCita = []string{
	models.EstadoProgramado, models.EstadoConfirmado, models.EstadoEnSala, models.EstadoEnAtencion,
	models.EstadoAtendido, models.EstadoCancelado, models.EstadoNoAsistio,
}

var nombreDeEstado = map[string]string{
	models.EstadoProgramado: "Programado", models.EstadoConfirmado: "Confirmado", models.EstadoEnSala: "En sala de espera",
	models.EstadoEnAtencion: "En atención", models.EstadoAtendido: "Atendido", models.EstadoCancelado: "Cancelado",
	models.EstadoNoAsistio: "No asistió",
}

type cuerpoCita struct {
	PacienteID      int       `json:"paciente_id"`
	ProfesionalID   int       `json:"profesional_id"`
	SillonID        int       `json:"sillon_id"`
	ProcedimientoID *int      `json:"procedimiento_id"`
	Motivo          string    `json:"motivo"`
	Notas           string    `json:"notas"`
	Inicio          time.Time `json:"inicio"`
	Fin             time.Time `json:"fin"`
	// PresupuestoItemID: tarea aprobada de un presupuesto que esta cita atiende (RF-08).
	PresupuestoItemID *int `json:"presupuesto_item_id"`
}

// minutosDelDia convierte "HH:MM" en minutos desde la medianoche.
func minutosDelDia(hhmm string) (int, bool) {
	hora, minuto, ok := strings.Cut(hhmm, ":")
	if !ok {
		return 0, false
	}
	h, errH := strconv.Atoi(hora)
	m, errM := strconv.Atoi(minuto)
	if errH != nil || errM != nil || h < 0 || h > 24 || m < 0 || m > 59 {
		return 0, false
	}
	return h*60 + m, true
}

// dentroDelHorario comprueba que la cita empiece y termine el mismo día local y dentro del horario de la agenda.
func dentroDelHorario(inicio, fin time.Time, horario models.HorarioAgenda) (bool, error) {
	zona, err := time.LoadLocation(horario.Zona)
	if err != nil {
		return false, fmt.Errorf("zona horaria %q inválida: %w", horario.Zona, err)
	}
	abre, okAbre := minutosDelDia(horario.Inicio)
	cierra, okCierra := minutosDelDia(horario.Fin)
	if !okAbre || !okCierra {
		return false, fmt.Errorf("horario de agenda inválido: %q a %q", horario.Inicio, horario.Fin)
	}
	i, f := inicio.In(zona), fin.In(zona)
	if i.YearDay() != f.YearDay() || i.Year() != f.Year() {
		return false, nil
	}
	desde, hasta := i.Hour()*60+i.Minute(), f.Hour()*60+f.Minute()
	return desde >= abre && hasta <= cierra, nil
}

// validar comprueba los datos de una cita y los convierte al modelo; el segundo valor es el motivo si son inválidos.
func (cuerpo cuerpoCita) validar(horario models.HorarioAgenda) (models.DatosCita, string, error) {
	var d models.DatosCita
	var ok bool
	if cuerpo.ProfesionalID <= 0 || cuerpo.SillonID <= 0 {
		return d, "El profesional y el sillón son obligatorios", nil
	}
	if cuerpo.ProcedimientoID != nil && *cuerpo.ProcedimientoID <= 0 {
		return d, "El procedimiento no es válido", nil
	}
	if d.Motivo, ok = textoOpcional(cuerpo.Motivo, maxMotivoCita); !ok {
		return d, "El motivo admite hasta 200 caracteres", nil
	}
	if d.Notas, ok = textoOpcional(cuerpo.Notas, maxNotasCita); !ok {
		return d, "Las notas admiten hasta 1000 caracteres", nil
	}
	if cuerpo.Inicio.IsZero() || cuerpo.Fin.IsZero() || !cuerpo.Fin.After(cuerpo.Inicio) {
		return d, "El inicio y el fin son obligatorios y el fin debe ser posterior al inicio", nil
	}
	if cuerpo.Fin.Sub(cuerpo.Inicio) > duracionMaximaCita {
		return d, "Una cita no puede durar más de 12 horas", nil
	}
	dentro, err := dentroDelHorario(cuerpo.Inicio, cuerpo.Fin, horario)
	if err != nil {
		return d, "", err
	}
	if !dentro {
		return d, fmt.Sprintf("La cita debe quedar dentro del horario de la agenda (%s a %s) en un mismo día", horario.Inicio, horario.Fin), nil
	}
	d.PacienteID, d.ProfesionalID, d.SillonID, d.ProcedimientoID = cuerpo.PacienteID, cuerpo.ProfesionalID, cuerpo.SillonID, cuerpo.ProcedimientoID
	d.Inicio, d.Fin = cuerpo.Inicio.UTC(), cuerpo.Fin.UTC()
	if cuerpo.PresupuestoItemID != nil && *cuerpo.PresupuestoItemID > 0 {
		d.PresupuestoItemID = cuerpo.PresupuestoItemID
	}
	return d, "", nil
}

// respuestaCita añade un aviso (no bloqueante) al detalle de la cita.
type respuestaCita struct {
	models.CitaDetalle
	Aviso string `json:"aviso,omitempty"`
}

// errorAgenda traduce los errores de la agenda: un choque de sillón o de estado tiene su propio mensaje.
func errorAgenda(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, models.ErrConflicto):
		return errorCliente(c, fiber.StatusConflict, "El sillón ya tiene una cita en ese horario")
	case errors.Is(err, models.ErrEstadoCambiado):
		return errorCliente(c, fiber.StatusConflict, "La cita cambió de estado o ya no se puede modificar; recarga la agenda")
	case errors.Is(err, models.ErrReferencia):
		return errorCliente(c, fiber.StatusBadRequest, "El paciente, el profesional, el sillón o el procedimiento no existe o está desactivado")
	}
	return errorDominio(c, err)
}

// avisoDeSimultaneidad avisa si el profesional ya tiene otra cita a la misma hora en otro sillón.
func (ctrl *CitasController) avisoDeSimultaneidad(ctx context.Context, d models.DatosCita, excluirID int) string {
	n, err := ctrl.repo.CitasSimultaneasDelProfesional(ctx, d.ProfesionalID, d.Inicio, d.Fin, excluirID)
	if err != nil || n == 0 {
		return ""
	}
	return "El profesional ya tiene otra cita a esa misma hora en otro sillón"
}

// ListarCitas responde GET /api/citas?desde=&hasta=&consultorio_id=&sillon_id=&profesional_id=&paciente_id=.
// El rango (RFC 3339) es obligatorio y no puede superar 62 días.
func (ctrl *CitasController) ListarCitas(c *fiber.Ctx) error {
	desde, errDesde := time.Parse(time.RFC3339, c.Query("desde"))
	hasta, errHasta := time.Parse(time.RFC3339, c.Query("hasta"))
	if errDesde != nil || errHasta != nil || !hasta.After(desde) {
		return errorCliente(c, fiber.StatusBadRequest, "Indica un rango válido con desde y hasta (RFC 3339)")
	}
	if hasta.Sub(desde) > rangoMaximoAgenda {
		return errorCliente(c, fiber.StatusBadRequest, "El rango de la agenda no puede superar 62 días")
	}
	filtro := models.FiltroCitas{
		Desde: desde.UTC(), Hasta: hasta.UTC(),
		ConsultorioID: c.QueryInt("consultorio_id"), SillonID: c.QueryInt("sillon_id"),
		ProfesionalID: c.QueryInt("profesional_id"), PacienteID: c.QueryInt("paciente_id"),
	}
	lista, err := ctrl.repo.ListarCitas(c.Context(), filtro)
	if err != nil {
		return errorDominio(c, err)
	}
	return c.JSON(lista)
}

// ObtenerCita responde GET /api/citas/:id con el historial de estados.
func (ctrl *CitasController) ObtenerCita(c *fiber.Ctx) error {
	id, ok := idDeRuta(c)
	if !ok {
		return errorCliente(c, fiber.StatusBadRequest, "Identificador inválido")
	}
	cita, err := ctrl.repo.ObtenerCita(c.Context(), id)
	if err != nil {
		return errorDominio(c, err)
	}
	return c.JSON(cita)
}

// HorarioAgenda responde GET /api/agenda/horario con el horario operativo y la zona horaria.
func (ctrl *CitasController) HorarioAgenda(c *fiber.Ctx) error {
	horario, err := ctrl.repo.HorarioAgenda(c.Context())
	if err != nil {
		return errorInterno(c, err)
	}
	return c.JSON(horario)
}

// CrearCita responde POST /api/citas.
func (ctrl *CitasController) CrearCita(c *fiber.Ctx) error {
	var cuerpo cuerpoCita
	if err := c.BodyParser(&cuerpo); err != nil {
		return cuerpoInvalido(c)
	}
	if cuerpo.PacienteID <= 0 {
		return errorCliente(c, fiber.StatusBadRequest, "El paciente es obligatorio")
	}
	horario, err := ctrl.repo.HorarioAgenda(c.Context())
	if err != nil {
		return errorInterno(c, err)
	}
	datos, motivo, err := cuerpo.validar(horario)
	if err != nil {
		return errorInterno(c, err)
	}
	if motivo != "" {
		return errorCliente(c, fiber.StatusBadRequest, motivo)
	}
	datos.PacienteID = cuerpo.PacienteID

	id, err := ctrl.repo.CrearCita(c.Context(), datos, usuarioDeLaPeticion(c))
	if err != nil {
		return errorAgenda(c, err)
	}
	if ctrl.avisos != nil {
		ctrl.avisos.CitaAgendada(c.Context(), id)
	}
	detalle, err := ctrl.repo.ObtenerCita(c.Context(), id)
	if err != nil {
		return errorDominio(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(respuestaCita{detalle, ctrl.avisoDeSimultaneidad(c.Context(), datos, id)})
}

// ReprogramarCita responde PUT /api/citas/:id: cambia horario, sillón, profesional o detalle.
// El paciente de la cita no se puede cambiar.
func (ctrl *CitasController) ReprogramarCita(c *fiber.Ctx) error {
	id, ok := idDeRuta(c)
	if !ok {
		return errorCliente(c, fiber.StatusBadRequest, "Identificador inválido")
	}
	var cuerpo cuerpoCita
	if err := c.BodyParser(&cuerpo); err != nil {
		return cuerpoInvalido(c)
	}
	horario, err := ctrl.repo.HorarioAgenda(c.Context())
	if err != nil {
		return errorInterno(c, err)
	}
	datos, motivo, err := cuerpo.validar(horario)
	if err != nil {
		return errorInterno(c, err)
	}
	if motivo != "" {
		return errorCliente(c, fiber.StatusBadRequest, motivo)
	}

	if err := ctrl.repo.ReprogramarCita(c.Context(), id, datos, usuarioDeLaPeticion(c)); err != nil {
		return errorAgenda(c, err)
	}
	if ctrl.avisos != nil {
		ctrl.avisos.CitaReprogramada(c.Context(), id)
	}
	detalle, err := ctrl.repo.ObtenerCita(c.Context(), id)
	if err != nil {
		return errorDominio(c, err)
	}
	return c.JSON(respuestaCita{detalle, ctrl.avisoDeSimultaneidad(c.Context(), datos, id)})
}

// CambiarEstado responde POST /api/citas/:id/estado con {"estado": "...", "motivo": "..."}.
// Cancelar exige motivo. La recepción puede pasar a cualquier estado permitido por las transiciones;
// el odontólogo y el asistente solo a En Atención y Atendido.
func (ctrl *CitasController) CambiarEstado(c *fiber.Ctx) error {
	id, ok := idDeRuta(c)
	if !ok {
		return errorCliente(c, fiber.StatusBadRequest, "Identificador inválido")
	}
	var cuerpo struct {
		Estado string `json:"estado"`
		Motivo string `json:"motivo"`
	}
	if err := c.BodyParser(&cuerpo); err != nil {
		return cuerpoInvalido(c)
	}
	if !slices.Contains(estadosDeCita, cuerpo.Estado) {
		return errorCliente(c, fiber.StatusBadRequest, "El estado no es válido")
	}
	motivo, ok := textoOpcional(cuerpo.Motivo, maxMotivoCambio)
	if !ok {
		return errorCliente(c, fiber.StatusBadRequest, "El motivo admite hasta 300 caracteres")
	}
	if cuerpo.Estado == models.EstadoCancelado && motivo == "" {
		return errorCliente(c, fiber.StatusBadRequest, "Cancelar una cita exige indicar el motivo")
	}
	rol, _ := c.Locals("rol").(string)
	if rol != models.RolRecepcionista && cuerpo.Estado != models.EstadoEnAtencion && cuerpo.Estado != models.EstadoAtendido {
		return errorCliente(c, fiber.StatusForbidden, "Tu rol solo puede pasar la cita a En atención o Atendido")
	}

	cita, err := ctrl.repo.ObtenerCita(c.Context(), id)
	if err != nil {
		return errorDominio(c, err)
	}
	if !models.TransicionPermitida(cita.Estado, cuerpo.Estado) {
		return errorCliente(c, fiber.StatusConflict,
			fmt.Sprintf("No se puede pasar una cita de %s a %s", nombreDeEstado[cita.Estado], nombreDeEstado[cuerpo.Estado]))
	}
	if err := ctrl.repo.CambiarEstado(c.Context(), id, cita.Estado, cuerpo.Estado, motivo, usuarioDeLaPeticion(c)); err != nil {
		return errorAgenda(c, err)
	}
	actualizada, err := ctrl.repo.ObtenerCita(c.Context(), id)
	if err != nil {
		return errorDominio(c, err)
	}
	return c.JSON(actualizada)
}

// ListarNotificaciones responde GET /api/citas/:id/notificaciones con la bitácora de mensajes de la cita.
func (ctrl *CitasController) ListarNotificaciones(c *fiber.Ctx) error {
	id, ok := idDeRuta(c)
	if !ok {
		return errorCliente(c, fiber.StatusBadRequest, "Identificador inválido")
	}
	if _, err := ctrl.repo.ObtenerCita(c.Context(), id); err != nil {
		return errorDominio(c, err)
	}
	lista, err := ctrl.repo.ListarNotificaciones(c.Context(), id)
	if err != nil {
		return errorInterno(c, err)
	}
	return c.JSON(lista)
}

// ConfirmarPorEnlace responde POST /api/citas/confirmar con {"token": "..."}: es público porque el paciente no
// tiene cuenta, y por eso solo devuelve el estado y la hora, nunca datos clínicos. El enlace debe estar firmado,
// corresponder al horario actual de la cita y no haber vencido (la cita no ha empezado). Repetirlo es inofensivo.
func (ctrl *CitasController) ConfirmarPorEnlace(c *fiber.Ctx) error {
	var cuerpo struct {
		Token string `json:"token"`
	}
	if err := c.BodyParser(&cuerpo); err != nil {
		return cuerpoInvalido(c)
	}
	id, inicio, err := notificaciones.LeerToken(ctrl.secreto, cuerpo.Token)
	if err != nil {
		return errorCliente(c, fiber.StatusBadRequest, "El enlace de confirmación no es válido")
	}
	cita, err := ctrl.repo.ObtenerCita(c.Context(), id)
	if err != nil {
		return errorCliente(c, fiber.StatusBadRequest, "El enlace de confirmación no es válido")
	}
	if !cita.Inicio.Equal(inicio) || !ctrl.ahora().Before(cita.Inicio) {
		return errorCliente(c, fiber.StatusGone, "Este enlace ya no está vigente: la cita cambió de horario o ya pasó. Comunícate con la clínica")
	}

	respuesta := fiber.Map{"estado": cita.Estado, "inicio": cita.Inicio}
	switch cita.Estado {
	case models.EstadoConfirmado:
		return c.JSON(respuesta)
	case models.EstadoProgramado:
		motivo := "Confirmada por el paciente desde el enlace"
		if err := ctrl.repo.CambiarEstado(c.Context(), id, models.EstadoProgramado, models.EstadoConfirmado, motivo, 0); err != nil {
			return errorAgenda(c, err)
		}
		respuesta["estado"] = models.EstadoConfirmado
		return c.JSON(respuesta)
	}
	return errorCliente(c, fiber.StatusConflict, "La cita ya no se puede confirmar. Comunícate con la clínica")
}
