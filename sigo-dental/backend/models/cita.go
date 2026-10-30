package models

import "time"

// Estados de una cita (RF-10); coinciden con citas.estado en la base de datos.
const (
	EstadoProgramado = "programado"
	EstadoConfirmado = "confirmado"
	EstadoEnSala     = "en_sala"
	EstadoEnAtencion = "en_atencion"
	EstadoAtendido   = "atendido"
	EstadoCancelado  = "cancelado"
	EstadoNoAsistio  = "no_asistio"
)

// transiciones define a qué estados puede pasar una cita desde cada uno. Los estados finales
// (atendido, cancelado y no_asistio) no aparecen como origen.
var transiciones = map[string][]string{
	EstadoProgramado: {EstadoConfirmado, EstadoEnSala, EstadoCancelado, EstadoNoAsistio},
	EstadoConfirmado: {EstadoEnSala, EstadoCancelado, EstadoNoAsistio},
	EstadoEnSala:     {EstadoEnAtencion, EstadoCancelado},
	EstadoEnAtencion: {EstadoAtendido},
}

// TransicionPermitida indica si una cita puede pasar del estado desde al estado hasta.
func TransicionPermitida(desde, hasta string) bool {
	for _, destino := range transiciones[desde] {
		if destino == hasta {
			return true
		}
	}
	return false
}

// EstadoLiberaSillon indica si una cita en ese estado deja libre su sillón (RN-03).
func EstadoLiberaSillon(estado string) bool {
	return estado == EstadoCancelado || estado == EstadoNoAsistio
}

// DatosCita son los campos que la recepción define al agendar o reprogramar.
type DatosCita struct {
	PacienteID      int       `json:"paciente_id"`
	ProfesionalID   int       `json:"profesional_id"`
	SillonID        int       `json:"sillon_id"`
	ProcedimientoID *int      `json:"procedimiento_id"`
	Motivo          string    `json:"motivo"`
	Notas           string    `json:"notas"`
	Inicio          time.Time `json:"inicio"`
	Fin             time.Time `json:"fin"`
}

// Cita es una cita con los nombres de lo que referencia, lista para pintar en el calendario.
type Cita struct {
	ID                  int       `json:"id"`
	Estado              string    `json:"estado"`
	PacienteID          int       `json:"paciente_id"`
	PacienteNombre      string    `json:"paciente_nombre"`
	NumeroHistoria      string    `json:"numero_historia"`
	Alertas             []string  `json:"alertas"`
	ProfesionalID       int       `json:"profesional_id"`
	ProfesionalNombre   string    `json:"profesional_nombre"`
	SillonID            int       `json:"sillon_id"`
	SillonNombre        string    `json:"sillon_nombre"`
	ConsultorioID       int       `json:"consultorio_id"`
	ConsultorioNombre   string    `json:"consultorio_nombre"`
	ProcedimientoID     *int      `json:"procedimiento_id"`
	ProcedimientoNombre string    `json:"procedimiento_nombre"`
	Motivo              string    `json:"motivo"`
	Notas               string    `json:"notas"`
	Inicio              time.Time `json:"inicio"`
	Fin                 time.Time `json:"fin"`
}

// CambioEstado es una línea del historial de una cita.
type CambioEstado struct {
	EstadoAnterior string    `json:"estado_anterior"`
	EstadoNuevo    string    `json:"estado_nuevo"`
	Motivo         string    `json:"motivo"`
	UsuarioNombre  string    `json:"usuario_nombre"`
	Fecha          time.Time `json:"fecha"`
}

// CitaDetalle es la cita con su historial de cambios de estado.
type CitaDetalle struct {
	Cita
	Historial []CambioEstado `json:"historial"`
}

// FiltroCitas acota el calendario: el rango es obligatorio y el resto de filtros es opcional (0 = todos).
type FiltroCitas struct {
	Desde         time.Time
	Hasta         time.Time
	ConsultorioID int
	SillonID      int
	ProfesionalID int
	PacienteID    int
}

// HorarioAgenda es el horario operativo diario de la clínica en su zona horaria.
type HorarioAgenda struct {
	Inicio string `json:"inicio"`
	Fin    string `json:"fin"`
	Zona   string `json:"zona"`
}
