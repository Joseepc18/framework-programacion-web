package models

import (
	"errors"
	"strings"
	"time"
)

// Errores del consentimiento informado.
var (
	// ErrConsentimientoRequerido: el procedimiento exige un consentimiento firmado y vigente que el paciente no tiene.
	ErrConsentimientoRequerido = errors.New("el procedimiento requiere un consentimiento firmado y vigente")
	// ErrConsentimientoFirmado: el consentimiento ya está firmado y no admite otra firma ni cambios.
	ErrConsentimientoFirmado = errors.New("el consentimiento ya está firmado")
)

// Tipos y modalidades de firma del consentimiento.
const (
	ConsentTipoConsentimiento = "consentimiento"
	ConsentTipoRevocacion     = "revocacion"
	ConsentPendiente          = "pendiente"
	ConsentFirmado            = "firmado"
	FirmanteTutor             = "tutor"
	FirmantePaciente          = "paciente"
	ModalidadManuscrita       = "manuscrita"
	ModalidadOTP              = "otp"
)

// Plantilla es un formato legal de consentimiento para un procedimiento (o el general si no tiene procedimiento).
type Plantilla struct {
	ID                  int    `json:"id"`
	ProcedimientoID     *int   `json:"procedimiento_id"`
	ProcedimientoNombre string `json:"procedimiento_nombre"`
	Titulo              string `json:"titulo"`
	Cuerpo              string `json:"cuerpo"`
	Activo              bool   `json:"activo"`
}

// Punto de un trazo de la firma manuscrita: posición en el lienzo y milisegundos desde el inicio.
type Punto struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
	T int     `json:"t"`
}

// FirmaManuscrita es el trazo capturado en la pantalla táctil, con el tamaño del lienzo y el dispositivo.
type FirmaManuscrita struct {
	Ancho       float64   `json:"ancho"`
	Alto        float64   `json:"alto"`
	Trazos      [][]Punto `json:"trazos"`
	Dispositivo string    `json:"dispositivo"`
}

// Consentimiento es un consentimiento informado (o su revocación) de un paciente.
type Consentimiento struct {
	ID                  int        `json:"id"`
	PacienteID          int        `json:"paciente_id"`
	ProcedimientoID     int        `json:"procedimiento_id"`
	ProcedimientoNombre string     `json:"procedimiento_nombre"`
	ProfesionalID       int        `json:"profesional_id"`
	ProfesionalNombre   string     `json:"profesional_nombre"`
	PlantillaID         *int       `json:"plantilla_id"`
	Tipo                string     `json:"tipo"`
	RevocaA             *int       `json:"revoca_a"`
	Motivo              string     `json:"motivo"`
	Titulo              string     `json:"titulo"`
	Texto               string     `json:"texto"`
	HuellaTexto         string     `json:"huella_texto"`
	Estado              string     `json:"estado"`
	FirmanteTipo        string     `json:"firmante_tipo"`
	FirmanteNombre      string     `json:"firmante_nombre"`
	FirmanteDocumento   string     `json:"firmante_documento"`
	Modalidad           string     `json:"modalidad"`
	Trazo               []byte     `json:"-"`
	Dispositivo         string     `json:"dispositivo"`
	FirmadoEn           *time.Time `json:"firmado_en"`
	ArchivoID           *int       `json:"archivo_id"`
	CreadoPor           string     `json:"creado_por"`
	CreadoEn            time.Time  `json:"creado_en"`
	// Vigente: firmado, dentro de su vigencia y no revocado (solo aplica a los de tipo consentimiento).
	Vigente bool `json:"vigente"`
	// Revocado: existe una revocación firmada.
	Revocado bool `json:"revocado"`
}

// NuevoConsentimiento son los datos para preparar un consentimiento ya con el texto completado.
type NuevoConsentimiento struct {
	PacienteID        int
	ProcedimientoID   int
	ProfesionalID     int
	PlantillaID       *int
	Tipo              string
	RevocaA           *int
	Motivo            string
	Titulo            string
	Texto             string
	FirmanteTipo      string
	FirmanteNombre    string
	FirmanteDocumento string
}

// DatosFirmaConsentimiento son los datos con los que se firma un consentimiento pendiente.
type DatosFirmaConsentimiento struct {
	Modalidad   string
	Trazo       []byte
	Dispositivo string
	FirmadoEn   time.Time
	ArchivoID   int
}

// DatosPlantilla es lo que se reemplaza en una plantilla.
type DatosPlantilla struct {
	Paciente      string
	Documento     string
	Procedimiento string
	Pieza         string
	Profesional   string
	Fecha         string
}

// RellenarPlantilla completa los campos {{...}} de la plantilla con los datos del paciente, procedimiento y profesional.
func RellenarPlantilla(cuerpo string, d DatosPlantilla) string {
	return strings.NewReplacer(
		"{{paciente}}", d.Paciente, "{{documento}}", d.Documento, "{{procedimiento}}", d.Procedimiento,
		"{{pieza}}", d.Pieza, "{{profesional}}", d.Profesional, "{{fecha}}", d.Fecha,
	).Replace(cuerpo)
}

// ValidarFirmaManuscrita comprueba que el trazo sea una firma real: suficientes puntos, dentro del lienzo y con alguna extensión.
func ValidarFirmaManuscrita(f FirmaManuscrita) string {
	if f.Ancho <= 0 || f.Alto <= 0 || f.Ancho > 4000 || f.Alto > 4000 {
		return "El tamaño del lienzo de la firma no es válido"
	}
	puntos := 0
	minX, maxX, minY, maxY := f.Ancho, 0.0, f.Alto, 0.0
	if len(f.Trazos) == 0 || len(f.Trazos) > 100 {
		return "Dibuja la firma en la pantalla"
	}
	for _, trazo := range f.Trazos {
		for _, p := range trazo {
			if p.X < 0 || p.X > f.Ancho || p.Y < 0 || p.Y > f.Alto {
				return "El trazo de la firma sale del lienzo"
			}
			minX, maxX = min(minX, p.X), max(maxX, p.X)
			minY, maxY = min(minY, p.Y), max(maxY, p.Y)
			puntos++
		}
	}
	if puntos > 20000 {
		return "El trazo de la firma es demasiado largo"
	}
	if puntos < 8 || (maxX-minX) < f.Ancho*0.1 {
		return "La firma es demasiado corta: firma de nuevo"
	}
	return ""
}
