// Pruebas de las reglas de la evolución clínica con el reloj controlado: plazo de edición, firma, adendas y RN-01.
package controllers

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"sigo-dental-backend/firma"
	"sigo-dental-backend/models"
)

var creadaPrueba = time.Date(2026, 10, 30, 9, 0, 0, 0, time.UTC)

type repoEvolucion struct {
	EvolucionRepository
	nota         models.NotaCompleta
	proc         models.ProcedimientoRealizado
	plazoHoras   string
	firmada      bool
	actualizada  bool
	errFirma     error
	errProc      error
	adendaCreada bool
}

func (r *repoEvolucion) ObtenerParametro(context.Context, string) (string, error) {
	if r.plazoHoras == "" {
		return "", models.ErrNoEncontrado
	}
	return r.plazoHoras, nil
}
func (r *repoEvolucion) ObtenerNota(context.Context, int, int) (models.NotaCompleta, error) {
	return r.nota, nil
}
func (r *repoEvolucion) ActualizarNota(context.Context, int, int, models.Nota) error {
	r.actualizada = true
	return nil
}
func (r *repoEvolucion) GuardarFirma(_ context.Context, f models.Firma) (models.Firma, error) {
	r.firmada = r.errFirma == nil
	return f, r.errFirma
}
func (r *repoEvolucion) ObtenerProcedimientoRealizado(context.Context, int, int) (models.ProcedimientoRealizado, error) {
	return r.proc, nil
}
func (r *repoEvolucion) CambiarEstadoProcedimiento(context.Context, int, int, string) error {
	return r.errProc
}
func (r *repoEvolucion) CrearAdendaFirmada(_ context.Context, a models.Adenda, _ func(models.Adenda) models.Firma) (models.Adenda, error) {
	r.adendaCreada = true
	a.ID = 1
	return a, nil
}

type firmadorFalso struct {
	err      error
	firmante firma.Firmante
}

func (f firmadorFalso) Autenticar(context.Context, int, string) (firma.Firmante, error) {
	return f.firmante, f.err
}

func appEvolucion(repo *repoEvolucion, firmador firmadorFalso, ahora time.Time) *fiber.App {
	ctrl := NuevoEvolucionController(repo, firmador)
	ctrl.ahora = func() time.Time { return ahora }
	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("rol", models.RolOdontologo)
		c.Locals("usuario_id", 1)
		return c.Next()
	})
	app.Put("/api/pacientes/:id/notas/:nid", ctrl.ActualizarNota)
	app.Post("/api/pacientes/:id/notas/:nid/firmar", ctrl.FirmarNota)
	app.Post("/api/pacientes/:id/adendas", ctrl.CrearAdenda)
	app.Post("/api/pacientes/:id/procedimientos-realizados/:pid/completar", ctrl.CompletarProcedimiento)
	return app
}

const notaValida = `{"subjetivo":"Dolor al masticar","objetivo":"Caries profunda en 16","evaluacion":"Pulpitis","plan":"Endodoncia"}`

func notaPrueba(firmada bool) models.NotaCompleta {
	n := models.NotaCompleta{Nota: models.Nota{ID: 3, PacienteID: 2, ProfesionalID: 7, AutorID: 1, CreadaEn: creadaPrueba,
		Subjetivo: "a", Objetivo: "b", Evaluacion: "c", Plan: "d"}}
	if firmada {
		f := firma.Construir(firma.Firmante{UsuarioID: 1, ProfesionalID: 7}, models.FirmaNota, 3, models.ContenidoCanonicoNota(n.Nota))
		n.Firma = &f
	}
	return n
}

// RF-05: pasado el plazo (24 h por defecto) el servidor rechaza cualquier edición, y se prueba con el reloj controlado.
func TestEditarNotaRespetaElPlazoYLaFirma(t *testing.T) {
	casos := []struct {
		nombre  string
		ahora   time.Time
		plazo   string
		firmada bool
		estado  int
	}{
		{"dentro del plazo", creadaPrueba.Add(23 * time.Hour), "", false, 200},
		{"un instante antes de vencer", creadaPrueba.Add(24*time.Hour - time.Second), "", false, 200},
		{"vencido el plazo de 24 h", creadaPrueba.Add(24 * time.Hour), "", false, 409},
		{"mucho después", creadaPrueba.Add(10 * 24 * time.Hour), "", false, 409},
		{"plazo configurado en 48 h", creadaPrueba.Add(30 * time.Hour), "48", false, 200},
		{"plazo configurado, ya vencido", creadaPrueba.Add(49 * time.Hour), "48", false, 409},
		{"firmada dentro del plazo", creadaPrueba.Add(time.Hour), "", true, 409},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			repo := &repoEvolucion{nota: notaPrueba(c.firmada), plazoHoras: c.plazo}
			app := appEvolucion(repo, firmadorFalso{}, c.ahora)
			if estado := enviar(t, app, "PUT", "/api/pacientes/2/notas/3", notaValida); estado != c.estado {
				t.Errorf("estado = %d, se esperaba %d", estado, c.estado)
			}
			if repo.actualizada != (c.estado == 200) {
				t.Errorf("actualizada = %v: una nota rechazada no debe tocar la base", repo.actualizada)
			}
		})
	}
}

func TestSoloElAutorEditaSuNotaYLosCuatroApartadosSonObligatorios(t *testing.T) {
	repo := &repoEvolucion{nota: notaPrueba(false)}
	repo.nota.AutorID = 99
	app := appEvolucion(repo, firmadorFalso{}, creadaPrueba.Add(time.Hour))
	if estado := enviar(t, app, "PUT", "/api/pacientes/2/notas/3", notaValida); estado != 403 {
		t.Errorf("otro autor: estado = %d, se esperaba 403", estado)
	}
	repo.nota.AutorID = 1
	for _, cuerpo := range []string{
		`{"subjetivo":"","objetivo":"b","evaluacion":"c","plan":"d"}`,
		`{"subjetivo":"a","objetivo":"b","evaluacion":"c","plan":"   "}`,
		`{"subjetivo":"a"}`,
	} {
		if estado := enviar(t, app, "PUT", "/api/pacientes/2/notas/3", cuerpo); estado != 400 {
			t.Errorf("%s: estado = %d, se esperaba 400", cuerpo, estado)
		}
	}
}

func TestFirmarNota(t *testing.T) {
	profesional := firma.Firmante{UsuarioID: 1, ProfesionalID: 7, Nombre: "Dra. Prueba", RegistroProfesional: "R-1"}
	casos := []struct {
		nombre   string
		ahora    time.Time
		firmador firmadorFalso
		firmada  bool
		cuerpo   string
		estado   int
		guarda   bool
	}{
		{"firma el responsable", creadaPrueba.Add(time.Hour), firmadorFalso{firmante: profesional}, false, `{"password":"x"}`, 200, true},
		// La nota vencida sin firmar queda "pendiente de firma" y su autor todavía puede firmarla.
		{"firma una nota con el plazo vencido", creadaPrueba.Add(72 * time.Hour), firmadorFalso{firmante: profesional}, false, `{"password":"x"}`, 200, true},
		{"contraseña incorrecta es 403 y no 401", creadaPrueba, firmadorFalso{err: models.ErrPasswordIncorrecta}, false, `{"password":"mal"}`, 403, false},
		{"demasiados intentos", creadaPrueba, firmadorFalso{err: models.ErrDemasiadosIntentos}, false, `{"password":"mal"}`, 429, false},
		{"cuenta sin ficha de profesional", creadaPrueba, firmadorFalso{err: models.ErrNoEsProfesional}, false, `{"password":"x"}`, 403, false},
		{"otro profesional no firma la nota ajena", creadaPrueba, firmadorFalso{firmante: firma.Firmante{ProfesionalID: 8}}, false, `{"password":"x"}`, 403, false},
		{"ya firmada", creadaPrueba, firmadorFalso{firmante: profesional}, true, `{"password":"x"}`, 409, false},
		{"sin contraseña", creadaPrueba, firmadorFalso{firmante: profesional}, false, `{}`, 400, false},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			repo := &repoEvolucion{nota: notaPrueba(c.firmada)}
			app := appEvolucion(repo, c.firmador, c.ahora)
			if estado := enviar(t, app, "POST", "/api/pacientes/2/notas/3/firmar", c.cuerpo); estado != c.estado {
				t.Errorf("estado = %d, se esperaba %d", estado, c.estado)
			}
			if repo.firmada != c.guarda {
				t.Errorf("firma guardada = %v, se esperaba %v", repo.firmada, c.guarda)
			}
		})
	}
}

func TestUnaNotaAlteradaDejaDeCoincidirConSuHuella(t *testing.T) {
	repo := &repoEvolucion{nota: notaPrueba(true)}
	ctrl := NuevoEvolucionController(repo, firmadorFalso{})
	ctrl.ahora = func() time.Time { return creadaPrueba.Add(time.Hour) }
	if !ctrl.presentarNota(repo.nota, 24*time.Hour).Firma.Valida {
		t.Fatal("la nota sin cambios debe tener firma válida")
	}
	alterada := notaPrueba(true)
	f := *alterada.Firma
	alterada.Firma = &f
	alterada.Plan = "Exodoncia de todas las piezas" // alguien la modificó por fuera de la API
	if ctrl.presentarNota(alterada, 24*time.Hour).Firma.Valida {
		t.Error("la firma de una nota alterada debe marcarse como no válida")
	}
}

// RN-01: la adenda es la única corrección; no procede mientras la nota sigue editable ni la firma otro profesional.
func TestAdendaSoloCorrigeLoYaCerradoYLaFirmaElResponsable(t *testing.T) {
	profesional := firma.Firmante{UsuarioID: 1, ProfesionalID: 7}
	cuerpo := `{"nota_id":3,"texto":"Corrección: era la pieza 26","password":"x"}`
	casos := []struct {
		nombre   string
		ahora    time.Time
		firmada  bool
		firmador firmadorFalso
		cuerpo   string
		estado   int
	}{
		{"nota firmada", creadaPrueba.Add(time.Hour), true, firmadorFalso{firmante: profesional}, cuerpo, 201},
		{"nota vencida sin firmar", creadaPrueba.Add(48 * time.Hour), false, firmadorFalso{firmante: profesional}, cuerpo, 201},
		{"nota todavía editable", creadaPrueba.Add(time.Hour), false, firmadorFalso{firmante: profesional}, cuerpo, 409},
		{"firma otro profesional", creadaPrueba.Add(time.Hour), true, firmadorFalso{firmante: firma.Firmante{ProfesionalID: 9}}, cuerpo, 403},
		{"contraseña incorrecta", creadaPrueba.Add(time.Hour), true, firmadorFalso{err: models.ErrPasswordIncorrecta}, cuerpo, 403},
		{"sin destino", creadaPrueba, true, firmadorFalso{firmante: profesional}, `{"texto":"x","password":"x"}`, 400},
		{"con dos destinos", creadaPrueba, true, firmadorFalso{firmante: profesional}, `{"nota_id":3,"procedimiento_realizado_id":4,"texto":"x","password":"x"}`, 400},
		{"sin texto", creadaPrueba, true, firmadorFalso{firmante: profesional}, `{"nota_id":3,"texto":" ","password":"x"}`, 400},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			repo := &repoEvolucion{nota: notaPrueba(c.firmada)}
			app := appEvolucion(repo, c.firmador, c.ahora)
			if estado := enviar(t, app, "POST", "/api/pacientes/2/adendas", c.cuerpo); estado != c.estado {
				t.Errorf("estado = %d, se esperaba %d", estado, c.estado)
			}
			if repo.adendaCreada != (c.estado == 201) {
				t.Errorf("adenda creada = %v", repo.adendaCreada)
			}
		})
	}
}

func TestAdendaDeProcedimientoSoloSiEstaRealizado(t *testing.T) {
	profesional := firma.Firmante{UsuarioID: 1, ProfesionalID: 7}
	cuerpo := `{"procedimiento_realizado_id":4,"texto":"Se omitió la anestesia","password":"x"}`
	for estado, esperado := range map[string]int{models.ProcRealizado: 201, models.ProcEnCurso: 409, models.ProcCancelado: 409} {
		repo := &repoEvolucion{proc: models.ProcedimientoRealizado{ID: 4, ProfesionalID: 7, Estado: estado}}
		app := appEvolucion(repo, firmadorFalso{firmante: profesional}, creadaPrueba)
		if got := enviar(t, app, "POST", "/api/pacientes/2/adendas", cuerpo); got != esperado {
			t.Errorf("procedimiento %s: estado = %d, se esperaba %d", estado, got, esperado)
		}
	}
}

func TestUnProcedimientoRealizadoNoVuelveACambiar(t *testing.T) {
	for err, esperado := range map[error]int{nil: 200, models.ErrProcedimientoNoEnCurso: 409, models.ErrInmutable: 409, models.ErrNoEncontrado: 404, errors.New("caída"): 500} {
		app := appEvolucion(&repoEvolucion{errProc: err, proc: models.ProcedimientoRealizado{ID: 4}}, firmadorFalso{}, creadaPrueba)
		if got := enviar(t, app, "POST", "/api/pacientes/2/procedimientos-realizados/4/completar", ""); got != esperado {
			t.Errorf("error %v: estado = %d, se esperaba %d", err, got, esperado)
		}
	}
}
