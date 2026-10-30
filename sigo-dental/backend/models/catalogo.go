package models

import "errors"

// Errores de dominio que la capa de datos devuelve y que los controladores traducen a HTTP.
var (
	// ErrNoEncontrado: el registro pedido no existe.
	ErrNoEncontrado = errors.New("registro no encontrado")
	// ErrDuplicado: viola una restricción de unicidad (código, documento, nombre repetido).
	ErrDuplicado = errors.New("registro duplicado")
	// ErrReferencia: apunta a un registro que no existe o no es válido para esa relación.
	ErrReferencia = errors.New("referencia inválida")
	// ErrConflicto: choca con el estado actual (por ejemplo un sillón ya ocupado a esa hora).
	ErrConflicto = errors.New("conflicto con el estado actual")
)

// ItemCatalogo es un elemento de un catálogo simple (consultorios, convenios y condiciones críticas).
type ItemCatalogo struct {
	ID     int    `json:"id"`
	Nombre string `json:"nombre"`
	Activo bool   `json:"activo"`
}

// Sillon es un sillón dental físico, siempre dentro de un consultorio.
type Sillon struct {
	ID                int    `json:"id"`
	ConsultorioID     int    `json:"consultorio_id"`
	ConsultorioNombre string `json:"consultorio_nombre"`
	Nombre            string `json:"nombre"`
	Activo            bool   `json:"activo"`
}

// Profesional es un odontólogo de la clínica: su cuenta de usuario más sus datos de ejercicio.
type Profesional struct {
	ID                  int    `json:"id"`
	UsuarioID           int    `json:"usuario_id"`
	Nombre              string `json:"nombre"`
	Email               string `json:"email"`
	Especialidad        string `json:"especialidad"`
	RegistroProfesional string `json:"registro_profesional"`
	Activo              bool   `json:"activo"`
}

// Procedimiento es un tratamiento del catálogo con su precio en centavos y duración estimada.
type Procedimiento struct {
	ID                     int    `json:"id"`
	Codigo                 string `json:"codigo"`
	Nombre                 string `json:"nombre"`
	Especialidad           string `json:"especialidad"`
	PrecioCentavos         int    `json:"precio_centavos"`
	DuracionMin            int    `json:"duracion_min"`
	RequiereConsentimiento bool   `json:"requiere_consentimiento"`
	Activo                 bool   `json:"activo"`
}

// UsuarioResumen es la cuenta que se ofrece al dar de alta un profesional.
type UsuarioResumen struct {
	ID     int    `json:"id"`
	Nombre string `json:"nombre"`
	Email  string `json:"email"`
}

// Parametro es un valor de configuración del sistema.
type Parametro struct {
	Clave       string `json:"clave"`
	Valor       string `json:"valor"`
	Descripcion string `json:"descripcion"`
}
