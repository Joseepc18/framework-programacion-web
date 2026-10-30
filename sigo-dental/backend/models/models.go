// Paquete models: tipos compartidos entre capas (peticiones, entidades y errores de la API).
package models

// Códigos de rol: coinciden con roles.codigo en la base de datos y viajan en el claim "rol" del JWT.
const (
	RolAdminTI       = "admin_ti"
	RolRecepcionista = "recepcionista"
	RolOdontologo    = "odontologo"
	RolAsistente     = "asistente"
	RolAuditor       = "auditor"
	// RolPaciente es la cuenta del portal del paciente (F2-8); no figura entre los cinco roles del documento.
	RolPaciente = "paciente"
)

// APIError estandariza el formato de todas las respuestas de error HTTP de la API.
type APIError struct {
	// Status es el código de estado HTTP del error (400, 401, 403, ...).
	Status int `json:"status"`
	// Message es la descripción legible del error.
	Message string `json:"message"`
	// Details contiene información adicional opcional; omitempty lo oculta cuando está vacío.
	Details interface{} `json:"details,omitempty"`
}

// LoginRequest es el cuerpo JSON esperado en POST /api/login.
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Usuario es la cuenta con los datos que necesita el login. No se serializa a JSON.
type Usuario struct {
	ID           int
	Email        string
	PasswordHash string
	Rol          string
	Activo       bool
}
