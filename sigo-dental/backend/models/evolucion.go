package models

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"time"
)

// Errores de la evolución clínica y de la firma digital.
var (
	// ErrNotaFirmada: la nota ya está firmada y no se edita (RN-01).
	ErrNotaFirmada = errors.New("la nota está firmada")
	// ErrNotaBloqueada: venció el plazo de edición de la nota (RF-05).
	ErrNotaBloqueada = errors.New("venció el plazo de edición de la nota")
	// ErrYaFirmado: el documento ya tiene una firma.
	ErrYaFirmado = errors.New("el documento ya está firmado")
	// ErrPasswordIncorrecta: la reautenticación de la firma falló.
	ErrPasswordIncorrecta = errors.New("contraseña incorrecta")
	// ErrNoEsProfesional: la cuenta no es un profesional activo y no puede firmar.
	ErrNoEsProfesional = errors.New("la cuenta no es un profesional que pueda firmar")
	// ErrFirmanteNoCorresponde: otro profesional es el responsable del documento.
	ErrFirmanteNoCorresponde = errors.New("el documento corresponde a otro profesional")
	// ErrDemasiadosIntentos: demasiadas contraseñas erróneas seguidas al firmar.
	ErrDemasiadosIntentos = errors.New("demasiados intentos de firma")
	// ErrProcedimientoNoEnCurso: el procedimiento no está en curso.
	ErrProcedimientoNoEnCurso = errors.New("el procedimiento no está en curso")
)

// Entidades que se pueden firmar.
const (
	FirmaNota           = "nota"
	FirmaAdenda         = "adenda"
	FirmaReceta         = "receta"
	FirmaConsentimiento = "consentimiento"
)

// Firma es la firma digital de un documento: quién, cuándo y la huella del contenido que firmó.
type Firma struct {
	ID                  int       `json:"id"`
	Entidad             string    `json:"entidad"`
	EntidadID           int       `json:"entidad_id"`
	UsuarioID           int       `json:"-"`
	ProfesionalID       int       `json:"profesional_id"`
	NombreFirmante      string    `json:"nombre_firmante"`
	RegistroProfesional string    `json:"registro_profesional"`
	HuellaSHA256        string    `json:"huella_sha256"`
	ContenidoCanonico   string    `json:"-"`
	FirmadoEn           time.Time `json:"firmado_en"`
	// Valida indica si la huella guardada coincide con el contenido actual del documento; si alguien lo alterara
	// por fuera de la API, dejaría de coincidir.
	Valida bool `json:"valida"`
}

// Huella calcula el SHA-256 en hexadecimal del contenido canónico.
func Huella(canonico string) string {
	suma := sha256.Sum256([]byte(canonico))
	return hex.EncodeToString(suma[:])
}

// canonico serializa en JSON, con los campos en un orden fijo, lo que se firma. Cualquier cambio en el contenido
// cambia la huella.
func canonico(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		return ""
	}
	return string(b)
}

// Nota es una nota de evolución SOEP (RF-05).
type Nota struct {
	ID                int       `json:"id"`
	PacienteID        int       `json:"paciente_id"`
	CitaID            *int      `json:"cita_id"`
	ProfesionalID     int       `json:"profesional_id"`
	ProfesionalNombre string    `json:"profesional_nombre"`
	AutorID           int       `json:"autor_id"`
	AutorNombre       string    `json:"autor_nombre"`
	Subjetivo         string    `json:"subjetivo"`
	Objetivo          string    `json:"objetivo"`
	Evaluacion        string    `json:"evaluacion"`
	Plan              string    `json:"plan"`
	CreadaEn          time.Time `json:"creada_en"`
	ActualizadaEn     time.Time `json:"actualizada_en"`
}

// ContenidoCanonicoNota es lo que firma el profesional al firmar una nota.
func ContenidoCanonicoNota(n Nota) string {
	return canonico(struct {
		Version       string `json:"version"`
		NotaID        int    `json:"nota_id"`
		PacienteID    int    `json:"paciente_id"`
		CitaID        *int   `json:"cita_id"`
		ProfesionalID int    `json:"profesional_id"`
		AutorID       int    `json:"autor_id"`
		CreadaEn      string `json:"creada_en"`
		Subjetivo     string `json:"subjetivo"`
		Objetivo      string `json:"objetivo"`
		Evaluacion    string `json:"evaluacion"`
		Plan          string `json:"plan"`
	}{"soep-v1", n.ID, n.PacienteID, n.CitaID, n.ProfesionalID, n.AutorID, n.CreadaEn.UTC().Format(time.RFC3339Nano),
		n.Subjetivo, n.Objetivo, n.Evaluacion, n.Plan})
}

// Estados derivados de una nota.
const (
	NotaEditable       = "editable"
	NotaFirmada        = "firmada"
	NotaPendienteFirma = "pendiente_firma"
)

// EstadoDeNota indica si la nota se puede editar, está firmada o venció el plazo sin firma ("pendiente de firma").
func EstadoDeNota(creada time.Time, firmada bool, ahora time.Time, plazo time.Duration) string {
	switch {
	case firmada:
		return NotaFirmada
	case !ahora.Before(creada.Add(plazo)):
		return NotaPendienteFirma
	}
	return NotaEditable
}

// Adenda corrige una nota firmada o un procedimiento realizado sin modificarlos (RN-01).
type Adenda struct {
	ID                       int       `json:"id"`
	PacienteID               int       `json:"paciente_id"`
	NotaID                   *int      `json:"nota_id"`
	ProcedimientoRealizadoID *int      `json:"procedimiento_realizado_id"`
	ProfesionalID            int       `json:"profesional_id"`
	ProfesionalNombre        string    `json:"profesional_nombre"`
	AutorID                  int       `json:"autor_id"`
	Texto                    string    `json:"texto"`
	CreadaEn                 time.Time `json:"creada_en"`
	Firma                    *Firma    `json:"firma"`
}

// ContenidoCanonicoAdenda es lo que firma el profesional al crear una adenda. No incluye ID ni fecha porque se
// calcula antes de guardarla; la fecha de la firma queda en la propia firma.
func ContenidoCanonicoAdenda(a Adenda) string {
	return canonico(struct {
		Version         string `json:"version"`
		PacienteID      int    `json:"paciente_id"`
		NotaID          *int   `json:"nota_id"`
		ProcedimientoID *int   `json:"procedimiento_realizado_id"`
		ProfesionalID   int    `json:"profesional_id"`
		AutorID         int    `json:"autor_id"`
		Texto           string `json:"texto"`
	}{"adenda-v1", a.PacienteID, a.NotaID, a.ProcedimientoRealizadoID, a.ProfesionalID, a.AutorID, a.Texto})
}

// NotaCompleta es una nota con su estado, su firma y sus adendas, lista para mostrar en la bitácora.
type NotaCompleta struct {
	Nota
	Estado        string    `json:"estado"`
	EditableHasta time.Time `json:"editable_hasta"`
	Firma         *Firma    `json:"firma"`
	Adendas       []Adenda  `json:"adendas"`
	// PuedeEditar y PuedeFirmar dicen qué puede hacer la cuenta que consulta; el servidor lo vuelve a comprobar al actuar.
	PuedeEditar bool `json:"puede_editar"`
	PuedeFirmar bool `json:"puede_firmar"`
}

// Estados de un procedimiento realizado.
const (
	ProcEnCurso   = "en_curso"
	ProcRealizado = "realizado"
	ProcCancelado = "cancelado"
)

// ProcedimientoRealizado es un procedimiento hecho (o en curso) al paciente. Una vez realizado no se borra ni se cambia.
type ProcedimientoRealizado struct {
	ID                  int        `json:"id"`
	PacienteID          int        `json:"paciente_id"`
	ProcedimientoID     int        `json:"procedimiento_id"`
	ProcedimientoNombre string     `json:"procedimiento_nombre"`
	Pieza               *int       `json:"pieza"`
	Superficie          string     `json:"superficie"`
	ProfesionalID       int        `json:"profesional_id"`
	ProfesionalNombre   string     `json:"profesional_nombre"`
	CitaID              *int       `json:"cita_id"`
	PresupuestoItemID   *int       `json:"presupuesto_item_id"`
	Estado              string     `json:"estado"`
	Notas               string     `json:"notas"`
	IniciadoEn          time.Time  `json:"iniciado_en"`
	RealizadoEn         *time.Time `json:"realizado_en"`
	Adendas             []Adenda   `json:"adendas"`
}

// NuevoProcedimientoRealizado son los datos validados para iniciar un procedimiento.
type NuevoProcedimientoRealizado struct {
	PacienteID        int
	ProcedimientoID   int
	Pieza             *int
	Superficie        string
	CitaID            *int
	PresupuestoItemID *int
	Notas             string
}

// Profesional es el profesional de la clínica asociado a una cuenta.
type ProfesionalDeUsuario struct {
	ID                  int
	Nombre              string
	RegistroProfesional string
}
