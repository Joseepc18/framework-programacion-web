package models

// Paciente es la ficha única del paciente (RF-01). Las fechas viajan como texto AAAA-MM-DD.
type Paciente struct {
	ID                   int    `json:"id"`
	NumeroHistoria       string `json:"numero_historia"`
	Documento            string `json:"documento"`
	Nombres              string `json:"nombres"`
	Apellidos            string `json:"apellidos"`
	FechaNacimiento      string `json:"fecha_nacimiento"`
	Edad                 int    `json:"edad"`
	Sexo                 string `json:"sexo"`
	Telefono             string `json:"telefono"`
	Email                string `json:"email"`
	Direccion            string `json:"direccion"`
	EmergenciaNombre     string `json:"emergencia_nombre"`
	EmergenciaTelefono   string `json:"emergencia_telefono"`
	EmergenciaParentesco string `json:"emergencia_parentesco"`
	TutorNombre          string `json:"tutor_nombre"`
	TutorDocumento       string `json:"tutor_documento"`
	TutorTelefono        string `json:"tutor_telefono"`
	// ConvenioID es nil cuando el paciente no tiene seguro ni convenio.
	ConvenioID     *int   `json:"convenio_id"`
	ConvenioNombre string `json:"convenio_nombre"`
	CanalPreferido string `json:"canal_preferido"`
	Activo         bool   `json:"activo"`
}

// Anamnesis es el historial médico general (RF-02). Los cuatro apartados son obligatorios.
type Anamnesis struct {
	EnfermedadesSistemicas  string `json:"enfermedades_sistemicas"`
	AlergiasFarmacos        string `json:"alergias_farmacos"`
	AntecedentesQuirurgicos string `json:"antecedentes_quirurgicos"`
	Coagulopatias           string `json:"coagulopatias"`
	// CondicionesCriticas son los IDs de las condiciones críticas que tiene el paciente.
	CondicionesCriticas []int `json:"condiciones_criticas"`
}

// PacienteDetalle es la ficha completa: datos, anamnesis y las alertas médicas con su nombre.
type PacienteDetalle struct {
	Paciente
	Anamnesis Anamnesis `json:"anamnesis"`
	// Alertas son los nombres de las condiciones críticas, para el badge de la cabecera.
	Alertas []string `json:"alertas"`
}

// PacienteResumen es una fila del listado de pacientes.
type PacienteResumen struct {
	ID             int      `json:"id"`
	NumeroHistoria string   `json:"numero_historia"`
	Documento      string   `json:"documento"`
	Nombres        string   `json:"nombres"`
	Apellidos      string   `json:"apellidos"`
	Edad           int      `json:"edad"`
	Telefono       string   `json:"telefono"`
	ConvenioNombre string   `json:"convenio_nombre"`
	Activo         bool     `json:"activo"`
	Alertas        []string `json:"alertas"`
}

// Canales de notificación y valores de sexo admitidos.
var (
	SexosValidos   = []string{"femenino", "masculino", "otro"}
	CanalesValidos = []string{"whatsapp", "sms", "email"}
)
