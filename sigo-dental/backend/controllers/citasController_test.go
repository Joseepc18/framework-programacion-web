// Pruebas de la agenda con un repositorio falso: el choque real de sillones lo garantiza la restricción
// de exclusión de PostgreSQL (migración 004) y se comprueba contra la base, no aquí.
package controllers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"sigo-dental-backend/models"
	"sigo-dental-backend/notificaciones"
)

var secretoDePrueba = []byte("secreto-de-prueba-con-32-bytes-o-mas")

type repoCitas struct {
	CitaRepository
	errCrear       error
	errCambio      error
	estadoActual   string
	inicioActual   time.Time
	simultaneas    int
	creada         models.DatosCita
	filtro         models.FiltroCitas
	estadoGuardado string
	motivoGuardado string
}

func (r *repoCitas) HorarioAgenda(context.Context) (models.HorarioAgenda, error) {
	return models.HorarioAgenda{Inicio: "07:00", Fin: "21:00", Zona: "America/Guayaquil"}, nil
}

func (r *repoCitas) CrearCita(_ context.Context, d models.DatosCita, _ int) (int, error) {
	r.creada = d
	return 3, r.errCrear
}

func (r *repoCitas) ReprogramarCita(_ context.Context, _ int, d models.DatosCita, _ int) error {
	r.creada = d
	return r.errCrear
}

func (r *repoCitas) ObtenerCita(_ context.Context, id int) (models.CitaDetalle, error) {
	estado := r.estadoActual
	if estado == "" {
		estado = models.EstadoProgramado
	}
	return models.CitaDetalle{Cita: models.Cita{ID: id, Estado: estado, Inicio: r.inicioActual}, Historial: []models.CambioEstado{}}, nil
}

func (r *repoCitas) CambiarEstado(_ context.Context, _ int, _, hasta, motivo string, _ int) error {
	r.estadoGuardado, r.motivoGuardado = hasta, motivo
	return r.errCambio
}

func (r *repoCitas) CitasSimultaneasDelProfesional(context.Context, int, time.Time, time.Time, int) (int, error) {
	return r.simultaneas, nil
}

func (r *repoCitas) ListarNotificaciones(context.Context, int) ([]models.Notificacion, error) {
	return []models.Notificacion{}, nil
}

func (r *repoCitas) ListarCitas(_ context.Context, f models.FiltroCitas) ([]models.Cita, error) {
	r.filtro = f
	return []models.Cita{}, nil
}

// ahoraDePrueba fija el reloj del controlador; las citas de los enlaces empiezan 3 días después.
var ahoraDePrueba = time.Date(2026, 11, 2, 12, 0, 0, 0, time.UTC)

func appCitas(repo *repoCitas, rol string) *fiber.App {
	ctrl := NuevoCitasController(repo, nil, secretoDePrueba)
	ctrl.ahora = func() time.Time { return ahoraDePrueba }
	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("rol", rol)
		c.Locals("usuario_id", 1)
		return c.Next()
	})
	app.Get("/api/citas", ctrl.ListarCitas)
	app.Post("/api/citas", ctrl.CrearCita)
	app.Put("/api/citas/:id", ctrl.ReprogramarCita)
	app.Post("/api/citas/:id/estado", ctrl.CambiarEstado)
	app.Post("/api/citas/confirmar", ctrl.ConfirmarPorEnlace)
	return app
}

// cuerpoDeCita arma una cita válida el 2 de noviembre de 2026 de 10:00 a 11:00 (hora de Guayaquil, UTC-5).
func cuerpoDeCita(reemplazos ...[2]string) string {
	cuerpo := `{"paciente_id":4,"profesional_id":1,"sillon_id":1,"procedimiento_id":null,"motivo":"Control",` +
		`"inicio":"2026-11-02T10:00:00-05:00","fin":"2026-11-02T11:00:00-05:00"}`
	for _, r := range reemplazos {
		cuerpo = strings.Replace(cuerpo, r[0], r[1], 1)
	}
	return cuerpo
}

func TestTransicionesDeEstado(t *testing.T) {
	permitidas := map[[2]string]bool{
		{models.EstadoProgramado, models.EstadoConfirmado}: true,
		{models.EstadoProgramado, models.EstadoEnSala}:     true,
		{models.EstadoProgramado, models.EstadoCancelado}:  true,
		{models.EstadoProgramado, models.EstadoNoAsistio}:  true,
		{models.EstadoConfirmado, models.EstadoEnSala}:     true,
		{models.EstadoConfirmado, models.EstadoCancelado}:  true,
		{models.EstadoConfirmado, models.EstadoNoAsistio}:  true,
		{models.EstadoEnSala, models.EstadoEnAtencion}:     true,
		{models.EstadoEnSala, models.EstadoCancelado}:      true,
		{models.EstadoEnAtencion, models.EstadoAtendido}:   true,
	}
	for _, desde := range estadosDeCita {
		for _, hasta := range estadosDeCita {
			if got := models.TransicionPermitida(desde, hasta); got != permitidas[[2]string{desde, hasta}] {
				t.Errorf("%s → %s: permitida = %v", desde, hasta, got)
			}
		}
	}
}

func TestCrearCita(t *testing.T) {
	casos := []struct {
		nombre, cuerpo string
		err            error
		estado         int
	}{
		{"cita válida", cuerpoDeCita(), nil, 201},
		{"sillón ocupado", cuerpoDeCita(), models.ErrConflicto, 409},
		{"referencia desactivada o inexistente", cuerpoDeCita(), models.ErrReferencia, 400},
		{"falla la base de datos", cuerpoDeCita(), errors.New("conexión rechazada"), 500},
		{"sin paciente", cuerpoDeCita([2]string{`"paciente_id":4`, `"paciente_id":0`}), nil, 400},
		{"sin sillón", cuerpoDeCita([2]string{`"sillon_id":1`, `"sillon_id":0`}), nil, 400},
		{"fin antes del inicio", cuerpoDeCita([2]string{`11:00:00-05:00`, `09:00:00-05:00`}), nil, 400},
		{"fin igual al inicio", cuerpoDeCita([2]string{`11:00:00-05:00`, `10:00:00-05:00`}), nil, 400},
		{"antes de abrir la agenda", cuerpoDeCita([2]string{`10:00:00-05:00`, `06:00:00-05:00`}, [2]string{`11:00:00-05:00`, `07:00:00-05:00`}), nil, 400},
		{"empieza justo al abrir", cuerpoDeCita([2]string{`10:00:00-05:00`, `07:00:00-05:00`}), nil, 201},
		{"termina justo al cerrar", cuerpoDeCita([2]string{`10:00:00-05:00`, `20:00:00-05:00`}, [2]string{`11:00:00-05:00`, `21:00:00-05:00`}), nil, 201},
		{"termina después de cerrar", cuerpoDeCita([2]string{`10:00:00-05:00`, `20:30:00-05:00`}, [2]string{`11:00:00-05:00`, `21:30:00-05:00`}), nil, 400},
		{"cruza la medianoche", cuerpoDeCita([2]string{`11:00:00-05:00`, `07:00:00-05:00`}, [2]string{`2026-11-02T07`, `2026-11-03T07`}), nil, 400},
		{"dura más de 12 horas", cuerpoDeCita([2]string{`10:00:00-05:00`, `07:00:00-05:00`}, [2]string{`11:00:00-05:00`, `20:00:00-05:00`}), nil, 400},
		{"motivo demasiado largo", cuerpoDeCita([2]string{`"Control"`, `"` + strings.Repeat("x", 201) + `"`}), nil, 400},
		{"fecha sin formato", cuerpoDeCita([2]string{`2026-11-02T10:00:00-05:00`, `mañana`}), nil, 400},
		{"cuerpo que no es JSON", `no-es-json`, nil, 400},
	}
	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			app := appCitas(&repoCitas{errCrear: caso.err}, models.RolRecepcionista)
			if estado := enviar(t, app, "POST", "/api/citas", caso.cuerpo); estado != caso.estado {
				t.Errorf("estado = %d, se esperaba %d", estado, caso.estado)
			}
		})
	}
}

func TestCrearCitaGuardaEnUTC(t *testing.T) {
	repo := &repoCitas{}
	app := appCitas(repo, models.RolRecepcionista)
	if estado := enviar(t, app, "POST", "/api/citas", cuerpoDeCita()); estado != 201 {
		t.Fatalf("estado = %d", estado)
	}
	quiere := time.Date(2026, 11, 2, 15, 0, 0, 0, time.UTC)
	if !repo.creada.Inicio.Equal(quiere) || repo.creada.Inicio.Location() != time.UTC {
		t.Errorf("inicio = %v, se esperaba %v en UTC", repo.creada.Inicio, quiere)
	}
}

func TestCambiarEstado(t *testing.T) {
	casos := []struct {
		nombre, rol, actual, nuevo, motivo string
		errRepo                            error
		estado                             int
	}{
		{"recepción confirma", models.RolRecepcionista, "programado", "confirmado", "", nil, 200},
		{"recepción cancela con motivo", models.RolRecepcionista, "confirmado", "cancelado", "El paciente avisó", nil, 200},
		{"cancelar sin motivo", models.RolRecepcionista, "programado", "cancelado", "  ", nil, 400},
		{"recepción marca no asistió", models.RolRecepcionista, "confirmado", "no_asistio", "", nil, 200},
		{"estado inexistente", models.RolRecepcionista, "programado", "volando", "", nil, 400},
		{"salto no permitido", models.RolRecepcionista, "programado", "atendido", "", nil, 409},
		{"atendido es final", models.RolRecepcionista, "atendido", "cancelado", "x", nil, 409},
		{"cancelado es final", models.RolRecepcionista, "cancelado", "programado", "", nil, 409},
		{"en atención no se cancela", models.RolRecepcionista, "en_atencion", "cancelado", "x", nil, 409},
		{"otra persona cambió el estado antes", models.RolRecepcionista, "programado", "confirmado", "", models.ErrEstadoCambiado, 409},
		{"odontólogo inicia la atención", models.RolOdontologo, "en_sala", "en_atencion", "", nil, 200},
		{"asistente termina la atención", models.RolAsistente, "en_atencion", "atendido", "", nil, 200},
		{"odontólogo no cancela", models.RolOdontologo, "programado", "cancelado", "x", nil, 403},
		{"asistente no confirma", models.RolAsistente, "programado", "confirmado", "", nil, 403},
	}
	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			repo := &repoCitas{estadoActual: caso.actual, errCambio: caso.errRepo}
			app := appCitas(repo, caso.rol)
			cuerpo := `{"estado":"` + caso.nuevo + `","motivo":"` + caso.motivo + `"}`
			if estado := enviar(t, app, "POST", "/api/citas/9/estado", cuerpo); estado != caso.estado {
				t.Errorf("estado = %d, se esperaba %d", estado, caso.estado)
			}
		})
	}
}

func TestCambiarEstadoGuardaElMotivoLimpio(t *testing.T) {
	repo := &repoCitas{estadoActual: "programado"}
	app := appCitas(repo, models.RolRecepcionista)
	enviar(t, app, "POST", "/api/citas/9/estado", `{"estado":"cancelado","motivo":"  Viaje  "}`)
	if repo.estadoGuardado != "cancelado" || repo.motivoGuardado != "Viaje" {
		t.Errorf("guardó %q con motivo %q", repo.estadoGuardado, repo.motivoGuardado)
	}
}

func TestListarCitasExigeUnRangoValido(t *testing.T) {
	casos := []struct {
		nombre, consulta string
		estado           int
	}{
		{"rango válido", "desde=2026-11-02T00:00:00Z&hasta=2026-11-09T00:00:00Z", 200},
		{"sin rango", "", 400},
		{"fecha inválida", "desde=ayer&hasta=hoy", 400},
		{"hasta antes de desde", "desde=2026-11-09T00:00:00Z&hasta=2026-11-02T00:00:00Z", 400},
		{"rango de más de 62 días", "desde=2026-01-01T00:00:00Z&hasta=2026-12-01T00:00:00Z", 400},
	}
	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			app := appCitas(&repoCitas{}, models.RolRecepcionista)
			if estado := enviar(t, app, "GET", "/api/citas?"+caso.consulta, ""); estado != caso.estado {
				t.Errorf("estado = %d, se esperaba %d", estado, caso.estado)
			}
		})
	}
}

func TestListarCitasPasaLosFiltros(t *testing.T) {
	repo := &repoCitas{}
	app := appCitas(repo, models.RolOdontologo)
	enviar(t, app, "GET", "/api/citas?desde=2026-11-02T00:00:00Z&hasta=2026-11-03T00:00:00Z&sillon_id=2&profesional_id=5&consultorio_id=1", "")
	if repo.filtro.SillonID != 2 || repo.filtro.ProfesionalID != 5 || repo.filtro.ConsultorioID != 1 || repo.filtro.PacienteID != 0 {
		t.Errorf("filtro = %+v", repo.filtro)
	}
}

func TestAvisoDeProfesionalSimultaneo(t *testing.T) {
	casos := []struct {
		nombre      string
		simultaneas int
		conAviso    bool
	}{
		{"sin otras citas no avisa", 0, false},
		{"otra cita a la misma hora avisa pero no bloquea", 1, true},
	}
	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			app := appCitas(&repoCitas{simultaneas: caso.simultaneas}, models.RolRecepcionista)
			req := httptest.NewRequest("POST", "/api/citas", strings.NewReader(cuerpoDeCita()))
			req.Header.Set("Content-Type", "application/json")
			resp, err := app.Test(req)
			if err != nil {
				t.Fatal(err)
			}
			if resp.StatusCode != 201 {
				t.Fatalf("estado = %d, se esperaba 201", resp.StatusCode)
			}
			var cuerpo map[string]any
			if err := json.NewDecoder(resp.Body).Decode(&cuerpo); err != nil {
				t.Fatal(err)
			}
			if _, hay := cuerpo["aviso"]; hay != caso.conAviso {
				t.Errorf("aviso presente = %v, se esperaba %v", hay, caso.conAviso)
			}
		})
	}
}

func TestReprogramarCita(t *testing.T) {
	casos := []struct {
		nombre string
		err    error
		estado int
	}{
		{"reprogramación válida", nil, 200},
		{"sillón ocupado", models.ErrConflicto, 409},
		{"la cita ya empezó o es final", models.ErrEstadoCambiado, 409},
		{"cita inexistente", models.ErrNoEncontrado, 404},
	}
	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			app := appCitas(&repoCitas{errCrear: caso.err}, models.RolRecepcionista)
			if estado := enviar(t, app, "PUT", "/api/citas/9", cuerpoDeCita()); estado != caso.estado {
				t.Errorf("estado = %d, se esperaba %d", estado, caso.estado)
			}
		})
	}
}

func TestConfirmarPorEnlace(t *testing.T) {
	inicio := ahoraDePrueba.Add(72 * time.Hour)
	token := func(clave []byte, id int, quien time.Time) string {
		return `{"token":"` + notificaciones.TokenDeConfirmacion(clave, id, quien) + `"}`
	}
	casos := []struct {
		nombre, cuerpo, estadoCita string
		inicioCita                 time.Time
		esperado                   int
		cambia                     bool
	}{
		{"enlace válido confirma", token(secretoDePrueba, 9, inicio), "programado", inicio, 200, true},
		{"confirmar dos veces es inofensivo", token(secretoDePrueba, 9, inicio), "confirmado", inicio, 200, false},
		{"firmado con otra clave", token([]byte("otra-clave-distinta-de-32-bytes-o-mas!!"), 9, inicio), "programado", inicio, 400, false},
		{"token basura", `{"token":"x.y"}`, "programado", inicio, 400, false},
		{"sin token", `{}`, "programado", inicio, 400, false},
		{"la cita se reprogramó", token(secretoDePrueba, 9, inicio), "programado", inicio.Add(24 * time.Hour), 410, false},
		{"la cita ya pasó", token(secretoDePrueba, 9, ahoraDePrueba.Add(-time.Hour)), "programado", ahoraDePrueba.Add(-time.Hour), 410, false},
		{"la cita está cancelada", token(secretoDePrueba, 9, inicio), "cancelado", inicio, 409, false},
		{"la cita ya se atendió", token(secretoDePrueba, 9, inicio), "atendido", inicio, 409, false},
	}
	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			repo := &repoCitas{estadoActual: caso.estadoCita, inicioActual: caso.inicioCita}
			// Es una ruta pública: se prueba sin rol ni usuario en la petición.
			ctrl := NuevoCitasController(repo, nil, secretoDePrueba)
			ctrl.ahora = func() time.Time { return ahoraDePrueba }
			app := fiber.New()
			app.Post("/api/citas/confirmar", ctrl.ConfirmarPorEnlace)

			if estado := enviar(t, app, "POST", "/api/citas/confirmar", caso.cuerpo); estado != caso.esperado {
				t.Errorf("estado = %d, se esperaba %d", estado, caso.esperado)
			}
			if cambio := repo.estadoGuardado == "confirmado"; cambio != caso.cambia {
				t.Errorf("cambió a confirmado = %v, se esperaba %v", cambio, caso.cambia)
			}
		})
	}
}

func TestLaRespuestaDelEnlaceNoFiltraDatosDelPaciente(t *testing.T) {
	inicio := ahoraDePrueba.Add(72 * time.Hour)
	ctrl := NuevoCitasController(&repoCitas{estadoActual: "programado", inicioActual: inicio}, nil, secretoDePrueba)
	ctrl.ahora = func() time.Time { return ahoraDePrueba }
	app := fiber.New()
	app.Post("/api/citas/confirmar", ctrl.ConfirmarPorEnlace)

	req := httptest.NewRequest("POST", "/api/citas/confirmar", strings.NewReader(`{"token":"`+notificaciones.TokenDeConfirmacion(secretoDePrueba, 9, inicio)+`"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	var cuerpo map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&cuerpo); err != nil {
		t.Fatal(err)
	}
	if len(cuerpo) != 2 || cuerpo["estado"] != "confirmado" || cuerpo["inicio"] == nil {
		t.Errorf("la respuesta pública debe traer solo estado e inicio: %v", cuerpo)
	}
}
