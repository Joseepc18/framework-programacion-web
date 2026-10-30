package models

import "time"

// Acciones que registra la auditoría de accesos (RNF-06).
const (
	AccionConsulta     = "consulta"
	AccionCreacion     = "creacion"
	AccionModificacion = "modificacion"
	AccionExportacion  = "exportacion"
	AccionFirma        = "firma"
	AccionAnulacion    = "anulacion"
)

// AccionesAuditoria es la lista cerrada de acciones válidas como filtro.
var AccionesAuditoria = []string{AccionConsulta, AccionCreacion, AccionModificacion, AccionExportacion, AccionFirma, AccionAnulacion}

// EventoAuditoria es un acceso a un registro clínico tal como lo capta el middleware.
type EventoAuditoria struct {
	UsuarioID  int
	Email      string
	Rol        string
	Accion     string
	Entidad    string
	EntidadID  *int
	PacienteID *int
	IP         string
	Metodo     string
	Ruta       string
}

// RegistroAuditoria es una fila de la bitácora de accesos lista para consultar.
type RegistroAuditoria struct {
	ID             int64     `json:"id"`
	OcurridoEn     time.Time `json:"ocurrido_en"`
	UsuarioEmail   string    `json:"usuario_email"`
	Rol            string    `json:"rol"`
	Accion         string    `json:"accion"`
	Entidad        string    `json:"entidad"`
	EntidadID      *int      `json:"entidad_id"`
	PacienteID     *int      `json:"paciente_id"`
	PacienteNombre string    `json:"paciente_nombre"`
	IP             string    `json:"ip"`
	Metodo         string    `json:"metodo"`
	Ruta           string    `json:"ruta"`
}

// FiltroAuditoria acota la consulta de la bitácora. Los campos vacíos no filtran.
type FiltroAuditoria struct {
	Usuario  string
	Paciente string
	Accion   string
	Desde    *time.Time
	Hasta    *time.Time
	Limite   int
	Pagina   int
}

// PaginaAuditoria es una página de la bitácora con el total que cumple el filtro.
type PaginaAuditoria struct {
	Total     int                 `json:"total"`
	Pagina    int                 `json:"pagina"`
	Limite    int                 `json:"limite"`
	Registros []RegistroAuditoria `json:"registros"`
}

// SignosVitales es una toma de signos vitales de una visita. Cada medición es opcional, pero hay al menos una.
type SignosVitales struct {
	ID                     int       `json:"id"`
	PacienteID             int       `json:"paciente_id"`
	CitaID                 *int      `json:"cita_id"`
	Sistolica              *int      `json:"sistolica"`
	Diastolica             *int      `json:"diastolica"`
	FrecuenciaCardiaca     *int      `json:"frecuencia_cardiaca"`
	FrecuenciaRespiratoria *int      `json:"frecuencia_respiratoria"`
	Temperatura            *float64  `json:"temperatura"`
	Saturacion             *int      `json:"saturacion"`
	PesoKg                 *float64  `json:"peso_kg"`
	TallaCm                *float64  `json:"talla_cm"`
	Observaciones          string    `json:"observaciones"`
	RegistradoPor          string    `json:"registrado_por"`
	RegistradoEn           time.Time `json:"registrado_en"`
}

// Categorías de archivo del paciente.
const (
	ArchivoRadiografia    = "radiografia"
	ArchivoFotografia     = "fotografia"
	ArchivoDocumento      = "documento"
	ArchivoConsentimiento = "consentimiento"
	ArchivoReceta         = "receta"
	ArchivoPresupuesto    = "presupuesto"
)

// CategoriasSubibles son las que el personal puede elegir al subir un archivo; las demás las genera el sistema.
var CategoriasSubibles = []string{ArchivoRadiografia, ArchivoFotografia, ArchivoDocumento}

// Archivo son los metadatos de un archivo del paciente; el contenido vive en el almacén.
type Archivo struct {
	ID             int       `json:"id"`
	PacienteID     int       `json:"paciente_id"`
	Categoria      string    `json:"categoria"`
	Descripcion    string    `json:"descripcion"`
	NombreOriginal string    `json:"nombre_original"`
	TipoMIME       string    `json:"tipo_mime"`
	TamanoBytes    int       `json:"tamano_bytes"`
	SHA256         string    `json:"sha256"`
	ClaveAlmacen   string    `json:"-"`
	SubidoPor      string    `json:"subido_por"`
	SubidoEn       time.Time `json:"subido_en"`
}

// NuevoArchivo es el registro de un archivo que ya se guardó en el almacén.
type NuevoArchivo struct {
	PacienteID     int
	Categoria      string
	Descripcion    string
	NombreOriginal string
	TipoMIME       string
	TamanoBytes    int
	SHA256         string
	ClaveAlmacen   string
	UsuarioID      int
}
