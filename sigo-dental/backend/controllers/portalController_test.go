package controllers

import (
	"context"
	"io"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"sigo-dental-backend/models"
	"sigo-dental-backend/notificaciones"
	"sigo-dental-backend/otp"
)

type repoPortal struct {
	PortalRepository
	citas       map[int]models.CitaDetalle
	estadoNuevo string
	historial   string
	usuarioEn   int
}

func (r *repoPortal) BuscarPacientePorDocumento(_ context.Context, documento string) (int, bool, error) {
	if documento == "0911111111" {
		return 5, true, nil
	}
	return 0, false, nil
}
func (r *repoPortal) ObtenerPaciente(context.Context, int) (models.PacienteDetalle, error) {
	return models.PacienteDetalle{Paciente: models.Paciente{ID: 5, Edad: 30, Telefono: "0990000000", CanalPreferido: "sms"}}, nil
}
func (r *repoPortal) PacienteDePortal(_ context.Context, usuarioID int) (int, bool, error) {
	if usuarioID == 50 {
		return 5, true, nil
	}
	return 0, false, nil
}
func (r *repoPortal) UsuarioDePaciente(context.Context, int, string) (int, string, error) {
	return 50, "paciente.HC-000005@portal.invalid", nil
}
func (r *repoPortal) ObtenerCita(_ context.Context, id int) (models.CitaDetalle, error) {
	c, ok := r.citas[id]
	if !ok {
		return c, models.ErrNoEncontrado
	}
	return c, nil
}
func (r *repoPortal) CambiarEstado(_ context.Context, _ int, desde, hasta, motivo string, usuarioID int) error {
	r.estadoNuevo, r.historial, r.usuarioEn = hasta, motivo, usuarioID
	return nil
}

type otpFalso struct {
	valido string
	err    error
}

func (o otpFalso) Emitir(context.Context, string, int) (string, time.Time, error) {
	return "123456", time.Now().Add(otp.Vigencia), nil
}
func (o otpFalso) Validar(_ context.Context, _ string, _ int, codigo string) error {
	if o.err != nil {
		return o.err
	}
	if codigo != o.valido {
		return otp.ErrInvalido
	}
	return nil
}

type envioFalso struct{ enviados int }

func (e *envioFalso) EnviarCodigo(context.Context, notificaciones.DestinoCodigo, string, time.Duration) (string, error) {
	e.enviados++
	return "•••••••000", nil
}

func appPortal(repo *repoPortal, envio *envioFalso, ahora time.Time) *fiber.App {
	ctrl := NuevoPortalController(repo, nil, otpFalso{valido: "123456"}, envio, nil, []byte("secreto-de-prueba-con-32-bytes-o-mas"))
	ctrl.ahora = func() time.Time { return ahora }
	app := fiber.New()
	app.Post("/api/portal/codigo", ctrl.SolicitarCodigo)
	app.Post("/api/portal/ingresar", ctrl.Ingresar)
	// La sesión del paciente (usuario 50 es el paciente 5; el 77 no es un paciente).
	con := func(usuario int) fiber.Handler {
		return func(c *fiber.Ctx) error {
			c.Locals("rol", models.RolPaciente)
			c.Locals("usuario_id", usuario)
			return c.Next()
		}
	}
	app.Post("/api/portal/citas/:id/confirmar", con(50), ctrl.ConfirmarCita)
	app.Post("/api/portal/citas/:id/cancelar", con(50), ctrl.CancelarCita)
	app.Get("/api/portal/citas", con(77), ctrl.ListarCitas)
	return app
}

func citaDePrueba(id, pacienteID int, estado string, inicio time.Time) models.CitaDetalle {
	return models.CitaDetalle{Cita: models.Cita{ID: id, PacienteID: pacienteID, Estado: estado, Inicio: inicio, Fin: inicio.Add(time.Hour)}}
}

// Un paciente solo ve y modifica sus propios datos; confirmar pasa la cita a Confirmado y lo deja en el historial.
func TestElPacienteSoloTocaSusCitas(t *testing.T) {
	ahora := time.Date(2026, 10, 30, 10, 0, 0, 0, time.UTC)
	futuro := ahora.Add(48 * time.Hour)
	repo := &repoPortal{citas: map[int]models.CitaDetalle{
		1: citaDePrueba(1, 5, models.EstadoProgramado, futuro),
		2: citaDePrueba(2, 9, models.EstadoProgramado, futuro),
		3: citaDePrueba(3, 5, models.EstadoProgramado, ahora.Add(-time.Hour)),
		4: citaDePrueba(4, 5, models.EstadoConfirmado, futuro),
		5: citaDePrueba(5, 5, models.EstadoAtendido, futuro),
	}}
	app := appPortal(repo, &envioFalso{}, ahora)

	if estado := enviar(t, app, "POST", "/api/portal/citas/1/confirmar", ""); estado != 200 {
		t.Fatalf("confirmar la propia: estado = %d", estado)
	}
	if repo.estadoNuevo != models.EstadoConfirmado || repo.usuarioEn != 50 || !strings.Contains(repo.historial, "portal") {
		t.Errorf("no quedó en el historial: estado = %q, usuario = %d, motivo = %q", repo.estadoNuevo, repo.usuarioEn, repo.historial)
	}
	repo.estadoNuevo = ""
	casos := []struct {
		nombre, ruta string
		estado       int
	}{
		{"confirmar la cita de otro paciente", "/api/portal/citas/2/confirmar", 404},
		{"cancelar la cita de otro paciente", "/api/portal/citas/2/cancelar", 404},
		{"confirmar una cita que ya pasó", "/api/portal/citas/3/confirmar", 409},
		{"confirmar una ya confirmada", "/api/portal/citas/4/confirmar", 409},
		{"cancelar una ya atendida", "/api/portal/citas/5/cancelar", 409},
		{"cancelar la propia futura", "/api/portal/citas/4/cancelar", 200},
		{"cita inexistente", "/api/portal/citas/99/confirmar", 404},
	}
	for _, c := range casos {
		repo.estadoNuevo = ""
		if estado := enviar(t, app, "POST", c.ruta, `{"motivo":"viajo"}`); estado != c.estado {
			t.Errorf("%s: estado = %d, se esperaba %d", c.nombre, estado, c.estado)
		}
		if c.estado != 200 && repo.estadoNuevo != "" {
			t.Errorf("%s: no debía cambiar el estado", c.nombre)
		}
	}
	// Una cuenta que no está ligada a ningún paciente no ve nada.
	if estado := enviar(t, app, "GET", "/api/portal/citas", ""); estado != 403 {
		t.Errorf("cuenta sin paciente: estado = %d, se esperaba 403", estado)
	}
}

func TestIngresoAlPortalConDocumentoYCodigo(t *testing.T) {
	ahora := time.Date(2026, 10, 30, 10, 0, 0, 0, time.UTC)
	envio := &envioFalso{}
	app := appPortal(&repoPortal{}, envio, ahora)

	// La respuesta es la misma exista o no el documento, y solo se envía código a quien existe.
	for _, documento := range []string{"0911111111", "0900000000"} {
		if estado := enviar(t, app, "POST", "/api/portal/codigo", `{"documento":"`+documento+`"}`); estado != 200 {
			t.Errorf("documento %s: estado = %d", documento, estado)
		}
	}
	if envio.enviados != 1 {
		t.Errorf("códigos enviados = %d, se esperaba 1 (solo al documento registrado)", envio.enviados)
	}
	// Pedirlo otra vez enseguida no envía otro mensaje (espera de 60 s).
	enviar(t, app, "POST", "/api/portal/codigo", `{"documento":"0911111111"}`)
	if envio.enviados != 1 {
		t.Errorf("segundo pedido inmediato: enviados = %d, se esperaba 1", envio.enviados)
	}
	if enviar(t, app, "POST", "/api/portal/codigo", `{"documento":""}`) != 400 {
		t.Error("un documento vacío debe ser 400")
	}

	req := httptest.NewRequest("POST", "/api/portal/ingresar", strings.NewReader(`{"documento":"0911111111","codigo":"123456"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil || resp.StatusCode != 200 {
		t.Fatalf("ingreso correcto: %v %v", resp, err)
	}
	cuerpo, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(cuerpo), `"rol":"paciente"`) || !strings.Contains(string(cuerpo), `"token":"`) {
		t.Errorf("respuesta = %s", cuerpo)
	}
	for nombre, cuerpo := range map[string]string{
		"código incorrecto":     `{"documento":"0911111111","codigo":"000000"}`,
		"documento inexistente": `{"documento":"0900000000","codigo":"123456"}`,
	} {
		if estado := enviar(t, app, "POST", "/api/portal/ingresar", cuerpo); estado != 400 {
			t.Errorf("%s: estado = %d, se esperaba 400 (y nunca 401, que cerraría una sesión)", nombre, estado)
		}
	}
}
