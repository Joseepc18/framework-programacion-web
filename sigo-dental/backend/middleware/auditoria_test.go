// Pruebas del middleware de auditoría: registra lo que debe, con el usuario y la IP correctos, y falla cerrado.
package middleware

import (
	"context"
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"sigo-dental-backend/models"
)

type auditorFalso struct {
	eventos []models.EventoAuditoria
	err     error
}

func (a *auditorFalso) RegistrarAcceso(_ context.Context, e models.EventoAuditoria) error {
	if a.err != nil {
		return a.err
	}
	a.eventos = append(a.eventos, e)
	return nil
}

func appAuditada(a *auditorFalso, accion string, estado int) *fiber.App {
	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("usuario_id", 7)
		c.Locals("email", "odontologo@sigo-dental.test")
		c.Locals("rol", models.RolOdontologo)
		return c.Next()
	})
	app.Get("/api/pacientes/:id", Auditar(a, accion, "paciente", PacienteDeRuta("id"), EntidadDeRuta("id")), func(c *fiber.Ctx) error {
		return c.Status(estado).SendString("datos clínicos")
	})
	app.Get("/api/archivos/:id", Auditar(a, models.AccionExportacion, "archivo", EntidadDeRuta("id")), func(c *fiber.Ctx) error {
		MarcarAuditoria(c, 42, 0)
		return c.SendString("contenido")
	})
	return app
}

func TestAuditarRegistraQuienQueYDesdeDonde(t *testing.T) {
	a := &auditorFalso{}
	app := appAuditada(a, models.AccionConsulta, 200)
	resp, err := app.Test(httptest.NewRequest("GET", "/api/pacientes/15", nil))
	if err != nil || resp.StatusCode != 200 {
		t.Fatalf("respuesta = %v, %v", resp, err)
	}
	if len(a.eventos) != 1 {
		t.Fatalf("eventos = %d, se esperaba 1", len(a.eventos))
	}
	e := a.eventos[0]
	if e.UsuarioID != 7 || e.Email != "odontologo@sigo-dental.test" || e.Rol != models.RolOdontologo {
		t.Errorf("usuario mal registrado: %+v", e)
	}
	if e.Accion != models.AccionConsulta || e.Entidad != "paciente" || e.Metodo != "GET" || e.Ruta != "/api/pacientes/15" {
		t.Errorf("acción mal registrada: %+v", e)
	}
	if e.PacienteID == nil || *e.PacienteID != 15 || e.EntidadID == nil || *e.EntidadID != 15 {
		t.Errorf("paciente o entidad mal registrados: %+v", e)
	}
	if e.IP == "" {
		t.Error("la IP de origen no se registró")
	}
}

func TestAuditarNoRegistraLasPeticionesFallidas(t *testing.T) {
	a := &auditorFalso{}
	app := appAuditada(a, models.AccionConsulta, 404)
	if _, err := app.Test(httptest.NewRequest("GET", "/api/pacientes/15", nil)); err != nil {
		t.Fatal(err)
	}
	if len(a.eventos) != 0 {
		t.Errorf("se registró una petición fallida: %+v", a.eventos)
	}
}

func TestAuditarTomaElPacienteQueAnotaElControlador(t *testing.T) {
	a := &auditorFalso{}
	app := appAuditada(a, models.AccionConsulta, 200)
	if _, err := app.Test(httptest.NewRequest("GET", "/api/archivos/9", nil)); err != nil {
		t.Fatal(err)
	}
	e := a.eventos[0]
	if e.PacienteID == nil || *e.PacienteID != 42 || e.EntidadID == nil || *e.EntidadID != 9 {
		t.Errorf("paciente o entidad mal tomados: %+v", e)
	}
}

func TestAuditarFallaCerradoEnLasLecturas(t *testing.T) {
	a := &auditorFalso{err: errors.New("base caída")}
	app := appAuditada(a, models.AccionConsulta, 200)
	resp, err := app.Test(httptest.NewRequest("GET", "/api/pacientes/15", nil))
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 500 {
		t.Errorf("estado = %d: sin auditoría una consulta no debe entregar datos", resp.StatusCode)
	}
	cuerpo := make([]byte, 64)
	n, _ := resp.Body.Read(cuerpo)
	if string(cuerpo[:n]) == "datos clínicos" {
		t.Error("la respuesta entregó los datos clínicos sin dejar rastro")
	}
}

func TestAuditarNoAnulaUnaEscrituraYaAplicada(t *testing.T) {
	a := &auditorFalso{err: errors.New("base caída")}
	app := appAuditada(a, models.AccionCreacion, 201)
	resp, err := app.Test(httptest.NewRequest("GET", "/api/pacientes/15", nil))
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 201 {
		t.Errorf("estado = %d, la escritura ya estaba aplicada", resp.StatusCode)
	}
}

func TestIPDeProxyTomaLaEntradaDeConfianzaYNoLaFalseable(t *testing.T) {
	t.Cleanup(func() { ConfigurarIPDeProxy("", 1) })
	ConfigurarIPDeProxy("X-Forwarded-For", 1)
	a := &auditorFalso{}
	app := appAuditada(a, models.AccionConsulta, 200)
	req := httptest.NewRequest("GET", "/api/pacientes/1", nil)
	// El cliente falsea la primera entrada; el proxy de confianza añadió la última.
	req.Header.Set("X-Forwarded-For", "6.6.6.6, 203.0.113.9")
	if _, err := app.Test(req); err != nil {
		t.Fatal(err)
	}
	if ip := a.eventos[0].IP; ip != "203.0.113.9" {
		t.Errorf("IP = %q, se esperaba la entrada que añadió el proxy", ip)
	}
}
