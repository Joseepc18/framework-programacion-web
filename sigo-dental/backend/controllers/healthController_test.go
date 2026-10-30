// Pruebas del health check con dobles: no hace falta PostgreSQL ni Redis
// porque el controlador depende de interfaces.
package controllers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
)

// pingFalso simula una dependencia que responde (err nil) o falla.
type pingFalso struct{ err error }

func (p pingFalso) Ping(context.Context) error { return p.err }

func TestHealthCheck(t *testing.T) {
	caida := errors.New("caída")
	casos := []struct {
		nombre   string
		db       error
		redis    error
		codigo   int
		estado   string
		database string
		cache    string
	}{
		{"todo operativo", nil, nil, fiber.StatusOK, "ok", "up", "up"},
		{"redis caído", nil, caida, fiber.StatusOK, "degraded", "up", "down"},
		{"postgres caído", caida, nil, fiber.StatusServiceUnavailable, "error", "down", ""},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			ctrl := NuevoHealthController(pingFalso{caso.db}, pingFalso{caso.redis})
			app := fiber.New()
			app.Get("/api/health", ctrl.HealthCheck)

			resp, err := app.Test(httptest.NewRequest("GET", "/api/health", nil))
			if err != nil {
				t.Fatal(err)
			}
			if resp.StatusCode != caso.codigo {
				t.Fatalf("código = %d, se esperaba %d", resp.StatusCode, caso.codigo)
			}

			var cuerpo map[string]string
			if err := json.NewDecoder(resp.Body).Decode(&cuerpo); err != nil {
				t.Fatal(err)
			}
			if cuerpo["status"] != caso.estado || cuerpo["database"] != caso.database || cuerpo["redis"] != caso.cache {
				t.Fatalf("cuerpo inesperado: %v", cuerpo)
			}
		})
	}
}
