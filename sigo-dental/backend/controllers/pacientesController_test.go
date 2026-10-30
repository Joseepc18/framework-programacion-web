// Pruebas de pacientes y anamnesis con un repositorio falso: no hace falta PostgreSQL.
package controllers

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"sigo-dental-backend/models"
)

// repoPacientes implementa solo lo que las pruebas usan; el resto de PacienteRepository queda en nil.
type repoPacientes struct {
	PacienteRepository
	errCrear    error
	errBuscar   error
	soloActivos *bool
	creado      models.Paciente
	anamnesis   models.Anamnesis
}

func (r *repoPacientes) ListarPacientes(_ context.Context, _ string, soloActivos bool) ([]models.PacienteResumen, error) {
	r.soloActivos = &soloActivos
	return []models.PacienteResumen{}, nil
}

func (r *repoPacientes) CrearPaciente(_ context.Context, p models.Paciente, a models.Anamnesis, _ int) (int, error) {
	r.creado, r.anamnesis = p, a
	return 7, r.errCrear
}

func (r *repoPacientes) ActualizarPaciente(_ context.Context, _ int, p models.Paciente) error {
	r.creado = p
	return r.errCrear
}

func (r *repoPacientes) ActualizarAnamnesis(_ context.Context, _ int, a models.Anamnesis, _ int) error {
	r.anamnesis = a
	return r.errCrear
}

func (r *repoPacientes) ObtenerPaciente(_ context.Context, id int) (models.PacienteDetalle, error) {
	if r.errBuscar != nil {
		return models.PacienteDetalle{}, r.errBuscar
	}
	return models.PacienteDetalle{Paciente: models.Paciente{ID: id}, Alertas: []string{}}, nil
}

// El día de las pruebas queda fijo para que la edad no dependa del reloj.
var hoyDePrueba = time.Date(2026, 10, 29, 12, 0, 0, 0, time.UTC)

func appPacientes(repo *repoPacientes, rol string) *fiber.App {
	ctrl := NuevoPacientesController(repo)
	ctrl.ahora = func() time.Time { return hoyDePrueba }
	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("rol", rol)
		c.Locals("usuario_id", 1)
		return c.Next()
	})
	app.Get("/api/pacientes", ctrl.ListarPacientes)
	app.Post("/api/pacientes", ctrl.CrearPaciente)
	app.Put("/api/pacientes/:id", ctrl.ActualizarPaciente)
	app.Put("/api/pacientes/:id/anamnesis", ctrl.ActualizarAnamnesis)
	return app
}

const anamnesisCompleta = `"anamnesis":{"enfermedades_sistemicas":"Hipertensión","alergias_farmacos":"Penicilina","antecedentes_quirurgicos":"Ninguna","coagulopatias":"Ninguna","condiciones_criticas":[1,2]}`

// cuerpoPaciente arma un alta válida de un adulto y aplica los reemplazos [texto, nuevo] para variarla.
func cuerpoDePaciente(reemplazos ...[2]string) string {
	cuerpo := `{"documento":"0912345678","nombres":"Ana","apellidos":"Pérez","fecha_nacimiento":"1990-05-20","sexo":"femenino",` +
		`"telefono":"0991112233","email":"ana@correo.test","direccion":"Calle 1","emergencia_nombre":"Luis Pérez",` +
		`"emergencia_telefono":"0994445566","emergencia_parentesco":"Hermano","convenio_id":null,"canal_preferido":"whatsapp",` +
		anamnesisCompleta + `}`
	for _, r := range reemplazos {
		cuerpo = strings.Replace(cuerpo, r[0], r[1], 1)
	}
	return cuerpo
}

func TestCrearPaciente(t *testing.T) {
	menor := [2]string{`"fecha_nacimiento":"1990-05-20"`, `"fecha_nacimiento":"2015-01-10"`}
	conTutor := [2]string{`"emergencia_parentesco":"Hermano"`, `"emergencia_parentesco":"Hermano","tutor_nombre":"Luis Pérez","tutor_documento":"0911111111","tutor_telefono":"0990000000"`}
	casos := []struct {
		nombre, cuerpo string
		err            error
		estado         int
	}{
		{"adulto válido", cuerpoDePaciente(), nil, 201},
		{"menor sin tutor", cuerpoDePaciente(menor), nil, 400},
		{"menor con tutor", cuerpoDePaciente(menor, conTutor), nil, 201},
		{"cumple 18 hoy no necesita tutor", cuerpoDePaciente([2]string{`"1990-05-20"`, `"2008-10-29"`}), nil, 201},
		{"cumple 18 mañana necesita tutor", cuerpoDePaciente([2]string{`"1990-05-20"`, `"2008-10-30"`}), nil, 400},
		{"sin anamnesis", strings.Replace(cuerpoDePaciente(), `,`+anamnesisCompleta, ``, 1), nil, 400},
		{"anamnesis con un apartado vacío", cuerpoDePaciente([2]string{`"coagulopatias":"Ninguna"`, `"coagulopatias":"   "`}), nil, 400},
		{"fecha futura", cuerpoDePaciente([2]string{`1990-05-20`, `2030-01-01`}), nil, 400},
		{"fecha con otro formato", cuerpoDePaciente([2]string{`1990-05-20`, `20/05/1990`}), nil, 400},
		{"sin documento", cuerpoDePaciente([2]string{`"documento":"0912345678"`, `"documento":""`}), nil, 400},
		{"correo inválido", cuerpoDePaciente([2]string{`ana@correo.test`, `no-es-correo`}), nil, 400},
		{"correo vacío es válido", cuerpoDePaciente([2]string{`ana@correo.test`, ``}), nil, 201},
		{"sexo inválido", cuerpoDePaciente([2]string{`"femenino"`, `"x"`}), nil, 400},
		{"canal inválido", cuerpoDePaciente([2]string{`"whatsapp"`, `"paloma"`}), nil, 400},
		{"condición crítica inválida", cuerpoDePaciente([2]string{`[1,2]`, `[0]`}), nil, 400},
		{"documento repetido", cuerpoDePaciente(), models.ErrDuplicado, 409},
		{"convenio inexistente", cuerpoDePaciente(), models.ErrReferencia, 400},
		{"cuerpo que no es JSON", `no-es-json`, nil, 400},
		{"falla la base de datos", cuerpoDePaciente(), errors.New("conexión rechazada"), 500},
	}
	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			app := appPacientes(&repoPacientes{errCrear: caso.err}, models.RolRecepcionista)
			if estado := enviar(t, app, "POST", "/api/pacientes", caso.cuerpo); estado != caso.estado {
				t.Errorf("estado = %d, se esperaba %d", estado, caso.estado)
			}
		})
	}
}

func TestCrearPacienteGuardaLosDatosLimpios(t *testing.T) {
	repo := &repoPacientes{}
	app := appPacientes(repo, models.RolRecepcionista)
	cuerpo := cuerpoDePaciente(
		[2]string{`"nombres":"Ana"`, `"nombres":"  Ana  "`},
		[2]string{`"canal_preferido":"whatsapp"`, `"canal_preferido":""`},
		[2]string{`[1,2]`, `[2,2,3]`},
	)
	if estado := enviar(t, app, "POST", "/api/pacientes", cuerpo); estado != 201 {
		t.Fatalf("estado = %d, se esperaba 201", estado)
	}
	if repo.creado.Nombres != "Ana" {
		t.Errorf("nombres = %q, se esperaba sin espacios sobrantes", repo.creado.Nombres)
	}
	if repo.creado.CanalPreferido != "whatsapp" {
		t.Errorf("canal = %q, se esperaba el valor por defecto whatsapp", repo.creado.CanalPreferido)
	}
	if want := []int{2, 3}; !reflect.DeepEqual(repo.anamnesis.CondicionesCriticas, want) {
		t.Errorf("condiciones = %v, se esperaban %v sin repetidos", repo.anamnesis.CondicionesCriticas, want)
	}
}

func TestActualizarPaciente(t *testing.T) {
	casos := []struct {
		nombre, cuerpo string
		estado         int
	}{
		{"válido", strings.Replace(cuerpoDePaciente(), `,`+anamnesisCompleta, `,"activo":true`, 1), 200},
		{"sin estado", strings.Replace(cuerpoDePaciente(), `,`+anamnesisCompleta, ``, 1), 400},
	}
	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			app := appPacientes(&repoPacientes{}, models.RolRecepcionista)
			if estado := enviar(t, app, "PUT", "/api/pacientes/7", caso.cuerpo); estado != caso.estado {
				t.Errorf("estado = %d, se esperaba %d", estado, caso.estado)
			}
		})
	}
}

func TestActualizarAnamnesis(t *testing.T) {
	completa := strings.TrimSuffix(strings.TrimPrefix(anamnesisCompleta, `"anamnesis":`), ``)
	casos := []struct {
		nombre, cuerpo string
		err            error
		estado         int
	}{
		{"válida", completa, nil, 200},
		{"apartado vacío", strings.Replace(completa, `"Penicilina"`, `""`, 1), nil, 400},
		{"paciente inexistente", completa, models.ErrNoEncontrado, 404},
		{"condición inexistente", completa, models.ErrReferencia, 400},
	}
	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			repo := &repoPacientes{errCrear: caso.err}
			app := appPacientes(repo, models.RolOdontologo)
			if estado := enviar(t, app, "PUT", "/api/pacientes/7/anamnesis", caso.cuerpo); estado != caso.estado {
				t.Errorf("estado = %d, se esperaba %d", estado, caso.estado)
			}
		})
	}
}

func TestEdadEn(t *testing.T) {
	casos := []struct {
		nacimiento string
		edad       int
	}{
		{"2026-10-29", 0},
		{"2008-10-29", 18},
		{"2008-10-30", 17},
		{"2008-11-01", 17},
		{"1990-02-28", 36},
		{"2000-02-29", 26},
	}
	for _, caso := range casos {
		nacimiento, err := time.Parse(formatoFecha, caso.nacimiento)
		if err != nil {
			t.Fatal(err)
		}
		if edad := edadEn(nacimiento, hoyDePrueba); edad != caso.edad {
			t.Errorf("edadEn(%s) = %d, se esperaba %d", caso.nacimiento, edad, caso.edad)
		}
	}
}

// Solo la recepción puede pedir los pacientes desactivados.
func TestListadoDePacientesOcultaInactivosSalvoRecepcion(t *testing.T) {
	casos := []struct {
		rol         string
		soloActivos bool
	}{
		{models.RolRecepcionista, false},
		{models.RolOdontologo, true},
		{models.RolAsistente, true},
	}
	for _, caso := range casos {
		t.Run(caso.rol, func(t *testing.T) {
			repo := &repoPacientes{}
			enviar(t, appPacientes(repo, caso.rol), "GET", "/api/pacientes?todos=true", "")
			if repo.soloActivos == nil || *repo.soloActivos != caso.soloActivos {
				t.Errorf("soloActivos = %v, se esperaba %v", repo.soloActivos, caso.soloActivos)
			}
		})
	}
}
