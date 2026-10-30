package controllers

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"sigo-dental-backend/models"
)

type repoOdontograma struct {
	OdontogramaRepository
	err      error
	al       *time.Time
	guardado models.NuevoHallazgo
}

func (r *repoOdontograma) ListarHallazgos(_ context.Context, _ int, al *time.Time) ([]models.Hallazgo, error) {
	r.al = al
	return []models.Hallazgo{}, r.err
}

func (r *repoOdontograma) RegistrarHallazgo(_ context.Context, n models.NuevoHallazgo, _ int) (models.Hallazgo, error) {
	r.guardado = n
	return models.Hallazgo{ID: 5, Pieza: n.Pieza}, r.err
}

func (r *repoOdontograma) CorregirHallazgo(_ context.Context, id int, n models.NuevoHallazgo, _ int) (models.Hallazgo, error) {
	r.guardado = n
	return models.Hallazgo{ID: id + 1, Pieza: n.Pieza}, r.err
}

func (r *repoOdontograma) RetirarHallazgo(context.Context, int, int, int) error { return r.err }

func appOdontograma(repo *repoOdontograma) *fiber.App {
	ctrl := NuevoOdontogramaController(repo)
	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("rol", models.RolOdontologo)
		c.Locals("usuario_id", 1)
		return c.Next()
	})
	app.Get("/api/pacientes/:id/odontograma", ctrl.ObtenerOdontograma)
	app.Post("/api/pacientes/:id/odontograma/hallazgos", ctrl.RegistrarHallazgo)
	app.Put("/api/pacientes/:id/odontograma/hallazgos/:hid", ctrl.CorregirHallazgo)
	app.Post("/api/pacientes/:id/odontograma/hallazgos/:hid/retirar", ctrl.RetirarHallazgo)
	return app
}

func TestRegistrarHallazgo(t *testing.T) {
	casos := []struct {
		nombre, cuerpo string
		err            error
		estado         int
	}{
		{"caries en una superficie", `{"pieza":16,"superficie":"oclusal","tipo":"caries","estado":"requerido"}`, nil, 201},
		{"endodoncia realizada", `{"pieza":36,"tipo":"endodoncia","estado":"realizado","observacion":"conducto único"}`, nil, 201},
		{"pieza temporal", `{"pieza":55,"tipo":"ausente","estado":"existente"}`, nil, 201},
		{"pieza inexistente", `{"pieza":19,"tipo":"ausente","estado":"existente"}`, nil, 400},
		{"caries sin superficie", `{"pieza":16,"tipo":"caries","estado":"requerido"}`, nil, 400},
		{"superficie en una endodoncia", `{"pieza":36,"superficie":"oclusal","tipo":"endodoncia","estado":"requerido"}`, nil, 400},
		{"tipo desconocido", `{"pieza":16,"tipo":"puente","estado":"requerido"}`, nil, 400},
		{"observación demasiado larga", `{"pieza":36,"tipo":"endodoncia","estado":"requerido","observacion":"` + strings.Repeat("a", 201) + `"}`, nil, 400},
		{"repetido", `{"pieza":16,"superficie":"oclusal","tipo":"caries","estado":"requerido"}`, models.ErrDuplicado, 409},
		{"paciente inexistente", `{"pieza":16,"superficie":"oclusal","tipo":"caries","estado":"requerido"}`, models.ErrNoEncontrado, 404},
		{"cuerpo que no es JSON", `no-es-json`, nil, 400},
	}
	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			app := appOdontograma(&repoOdontograma{err: caso.err})
			if estado := enviar(t, app, "POST", "/api/pacientes/4/odontograma/hallazgos", caso.cuerpo); estado != caso.estado {
				t.Errorf("estado = %d, se esperaba %d", estado, caso.estado)
			}
		})
	}
}

// RN-01: un hallazgo realizado no se corrige ni se retira, y la API lo dice con un 409 que remite a la adenda.
func TestUnHallazgoRealizadoNoSeCorrigeNiSeRetira(t *testing.T) {
	app := appOdontograma(&repoOdontograma{err: models.ErrInmutable})
	cuerpo := `{"pieza":36,"tipo":"endodoncia","estado":"requerido"}`
	if estado := enviar(t, app, "PUT", "/api/pacientes/4/odontograma/hallazgos/9", cuerpo); estado != 409 {
		t.Errorf("corregir: estado = %d, se esperaba 409", estado)
	}
	if estado := enviar(t, app, "POST", "/api/pacientes/4/odontograma/hallazgos/9/retirar", ""); estado != 409 {
		t.Errorf("retirar: estado = %d, se esperaba 409", estado)
	}
}

func TestCorregirYRetirarHallazgosVigentes(t *testing.T) {
	app := appOdontograma(&repoOdontograma{})
	if estado := enviar(t, app, "PUT", "/api/pacientes/4/odontograma/hallazgos/9", `{"pieza":16,"superficie":"mesial","tipo":"caries","estado":"requerido"}`); estado != 200 {
		t.Errorf("corregir: estado = %d", estado)
	}
	if estado := enviar(t, app, "POST", "/api/pacientes/4/odontograma/hallazgos/9/retirar", ""); estado != 204 {
		t.Errorf("retirar: estado = %d", estado)
	}
	for _, caso := range []struct {
		err    error
		estado int
	}{{models.ErrHallazgoRetirado, 409}, {models.ErrNoEncontrado, 404}, {errors.New("caída"), 500}} {
		app := appOdontograma(&repoOdontograma{err: caso.err})
		if estado := enviar(t, app, "POST", "/api/pacientes/4/odontograma/hallazgos/9/retirar", ""); estado != caso.estado {
			t.Errorf("retirar con %v: estado = %d, se esperaba %d", caso.err, estado, caso.estado)
		}
	}
}

func TestObtenerOdontogramaAceptaElInstanteHistorico(t *testing.T) {
	repo := &repoOdontograma{}
	app := appOdontograma(repo)
	if estado := enviar(t, app, "GET", "/api/pacientes/4/odontograma", ""); estado != 200 || repo.al != nil {
		t.Errorf("estado = %d, al = %v", estado, repo.al)
	}
	if estado := enviar(t, app, "GET", "/api/pacientes/4/odontograma?al=2026-10-01T10:00:00Z", ""); estado != 200 || repo.al == nil {
		t.Errorf("con fecha: estado = %d, al = %v", estado, repo.al)
	}
	if estado := enviar(t, app, "GET", "/api/pacientes/4/odontograma?al=ayer", ""); estado != 400 {
		t.Errorf("fecha inválida: estado = %d", estado)
	}
	if estado := enviar(t, app, "GET", "/api/pacientes/abc/odontograma", ""); estado != 400 {
		t.Errorf("id inválido: estado = %d", estado)
	}
}
