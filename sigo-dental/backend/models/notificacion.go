package models

import "time"

// Tipos de notificación de una cita (RF-11); coinciden con notificaciones.tipo en la base de datos.
const (
	NotifConfirmacion      = "confirmacion"
	NotifRecordatorio48    = "recordatorio_48h"
	NotifRecordatorio24    = "recordatorio_24h"
	NotifOTPConsentimiento = "otp_consentimiento"
	NotifOTPPortal         = "otp_portal"
)

// Estados de una notificación.
const (
	NotifPendiente = "pendiente"
	NotifEnviada   = "enviada"
	NotifFallida   = "fallida"
)

// Notificacion es una línea de la bitácora de mensajes enviados al paciente por una cita.
type Notificacion struct {
	ID        int        `json:"id"`
	CitaID    int        `json:"cita_id"`
	Tipo      string     `json:"tipo"`
	Canal     string     `json:"canal"`
	Destino   string     `json:"destino"`
	Mensaje   string     `json:"mensaje"`
	Estado    string     `json:"estado"`
	Detalle   string     `json:"detalle"`
	Vigente   bool       `json:"vigente"`
	CreadoEn  time.Time  `json:"creado_en"`
	EnviadoEn *time.Time `json:"enviado_en"`
}

// DatosParaNotificar es lo que hace falta de una cita para redactar y enviar sus mensajes.
type DatosParaNotificar struct {
	CitaID            int
	Estado            string
	Inicio            time.Time
	Actualizada       time.Time
	PacienteNombre    string
	Telefono          string
	Email             string
	Canal             string
	ProfesionalNombre string
	Zona              string
}

// NotificacionCodigo es un código de un solo uso enviado al paciente, que se registra en la bitácora sin pertenecer a una cita.
type NotificacionCodigo struct {
	PacienteID int
	Tipo       string
	Canal      string
	Destino    string
	Mensaje    string
}
