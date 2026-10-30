package notificaciones

import (
	"context"
	"fmt"
	"log"
	"time"

	"sigo-dental-backend/models"
)

// Repositorio es el contrato de la capa de datos que usa el servicio.
type Repositorio interface {
	DatosParaNotificar(ctx context.Context, citaID int) (models.DatosParaNotificar, error)
	// CitasParaRecordar devuelve las citas vigentes (programadas o confirmadas) que empiezan en el rango.
	CitasParaRecordar(ctx context.Context, desde, hasta time.Time) ([]models.DatosParaNotificar, error)
	// ReservarNotificacion inserta la notificación pendiente; creada es false si ya existía una vigente
	// de ese tipo para la cita (la clave única evita duplicados aunque haya ejecuciones simultáneas).
	ReservarNotificacion(ctx context.Context, n models.Notificacion) (id int, creada bool, err error)
	RegistrarResultado(ctx context.Context, id int, enviada bool, detalle string) error
	// InvalidarNotificaciones marca como no vigentes las de la cita, para regenerarlas tras reprogramarla.
	InvalidarNotificaciones(ctx context.Context, citaID int) error
}

// Servicio redacta, registra y entrega los avisos de las citas.
type Servicio struct {
	repo        Repositorio
	notificador Notificador
	secreto     []byte
	urlFrontend string
	ahora       func() time.Time
}

// NuevoServicio construye el servicio. urlFrontend es la dirección pública del sitio, para armar el enlace
// de confirmación.
func NuevoServicio(repo Repositorio, notificador Notificador, secreto []byte, urlFrontend string) *Servicio {
	return &Servicio{repo: repo, notificador: notificador, secreto: secreto, urlFrontend: urlFrontend, ahora: time.Now}
}

const (
	ventana48 = 48 * time.Hour
	ventana24 = 24 * time.Hour
)

// CitaAgendada envía la confirmación de una cita recién creada. Un fallo se registra pero no se propaga:
// no debe impedir que la cita exista.
func (s *Servicio) CitaAgendada(ctx context.Context, citaID int) {
	if err := s.enviar(ctx, citaID, models.NotifConfirmacion); err != nil {
		log.Printf("notificación de la cita %d: %v", citaID, err)
	}
}

// CitaReprogramada invalida los avisos anteriores (pertenecían al horario viejo) y envía una nueva
// confirmación; los recordatorios se regeneran solos con el horario nuevo.
func (s *Servicio) CitaReprogramada(ctx context.Context, citaID int) {
	if err := s.repo.InvalidarNotificaciones(ctx, citaID); err != nil {
		log.Printf("invalidar notificaciones de la cita %d: %v", citaID, err)
		return
	}
	s.CitaAgendada(ctx, citaID)
}

// Procesar revisa las citas próximas y envía los recordatorios que correspondan; devuelve cuántos envió.
// Es seguro ejecutarlo muchas veces: cada cita recibe a lo sumo un recordatorio de cada tipo. Una cita que
// entró al sistema (o fue reprogramada) ya dentro de una ventana no recibe ese recordatorio, porque su
// confirmación fue el aviso. Las citas canceladas, con inasistencia o ya iniciadas no se consideran.
func (s *Servicio) Procesar(ctx context.Context) (int, error) {
	ahora := s.ahora()
	citas, err := s.repo.CitasParaRecordar(ctx, ahora, ahora.Add(ventana48))
	if err != nil {
		return 0, err
	}
	enviados := 0
	for _, cita := range citas {
		tipo, ventana := models.NotifRecordatorio48, ventana48
		if cita.Inicio.Sub(ahora) <= ventana24 {
			tipo, ventana = models.NotifRecordatorio24, ventana24
		}
		if cita.Actualizada.After(cita.Inicio.Add(-ventana)) {
			continue
		}
		hecho, err := s.enviarSiFalta(ctx, cita, tipo)
		if err != nil {
			log.Printf("recordatorio de la cita %d: %v", cita.CitaID, err)
			continue
		}
		if hecho {
			enviados++
		}
	}
	return enviados, nil
}

// enviar carga la cita y entrega la notificación del tipo dado.
func (s *Servicio) enviar(ctx context.Context, citaID int, tipo string) error {
	cita, err := s.repo.DatosParaNotificar(ctx, citaID)
	if err != nil {
		return err
	}
	_, err = s.enviarSiFalta(ctx, cita, tipo)
	return err
}

// enviarSiFalta reserva la notificación (una sola por cita y tipo), la entrega y registra el resultado.
// El booleano indica si esta llamada fue la que la envió.
func (s *Servicio) enviarSiFalta(ctx context.Context, cita models.DatosParaNotificar, tipo string) (bool, error) {
	if cita.Estado != models.EstadoProgramado && cita.Estado != models.EstadoConfirmado {
		return false, nil
	}
	mensaje := s.redactar(cita, tipo)
	id, creada, err := s.repo.ReservarNotificacion(ctx, models.Notificacion{
		CitaID: cita.CitaID, Tipo: tipo, Canal: cita.Canal, Destino: mensaje.Destino, Mensaje: mensaje.Texto,
	})
	if err != nil || !creada {
		return false, err
	}
	if errEnvio := s.notificador.Enviar(ctx, mensaje); errEnvio != nil {
		return true, s.repo.RegistrarResultado(ctx, id, false, errEnvio.Error())
	}
	return true, s.repo.RegistrarResultado(ctx, id, true, "")
}

// destinoDe elige el teléfono o el correo según el canal preferido del paciente.
func destinoDe(cita models.DatosParaNotificar) string {
	if cita.Canal == "email" {
		return cita.Email
	}
	return cita.Telefono
}

func (s *Servicio) redactar(cita models.DatosParaNotificar, tipo string) Mensaje {
	zona, err := time.LoadLocation(cita.Zona)
	if err != nil {
		zona = time.UTC
	}
	local := cita.Inicio.In(zona)
	enlace := fmt.Sprintf("%s/confirmar?token=%s", s.urlFrontend, TokenDeConfirmacion(s.secreto, cita.CitaID, cita.Inicio))
	var introduccion string
	switch tipo {
	case models.NotifRecordatorio48:
		introduccion = "Recordatorio: en 2 días tienes una cita"
	case models.NotifRecordatorio24:
		introduccion = "Recordatorio: mañana tienes una cita"
	default:
		introduccion = "Tu cita fue agendada"
	}
	texto := fmt.Sprintf("Hola %s. %s en SIGO-Dental el %s a las %s con %s. Confirma tu asistencia aquí: %s",
		cita.PacienteNombre, introduccion, local.Format("02/01/2006"), local.Format("15:04"), cita.ProfesionalNombre, enlace)
	return Mensaje{Canal: cita.Canal, Destino: destinoDe(cita), Texto: texto}
}
