package middleware

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
)

var secretoPrueba = []byte("secreto-de-prueba")

// codigo ejecuta una petición contra una ruta protegida por RequireAuth y devuelve el estado HTTP.
func codigo(t *testing.T, cabecera string, activa CuentaActiva, admin bool) int {
	t.Helper()
	app := fiber.New()
	cadena := []fiber.Handler{RequireAuth(secretoPrueba, activa)}
	if admin {
		cadena = append(cadena, RequireAdmin())
	}
	cadena = append(cadena, func(c *fiber.Ctx) error { return c.SendStatus(200) })
	app.Get("/p", cadena...)
	req := httptest.NewRequest("GET", "/p", nil)
	if cabecera != "" {
		req.Header.Set("Authorization", cabecera)
	}
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	return resp.StatusCode
}

func siempre(bool) CuentaActiva {
	return func(context.Context, int) bool { return true }
}

func TestRequireAuth(t *testing.T) {
	activa := siempre(true)
	cliente, _ := GenerarToken(secretoPrueba, 2, "c@upse.edu.ec", "cliente")
	admin, _ := GenerarToken(secretoPrueba, 1, "a@upse.edu.ec", "admin")
	ajeno, _ := GenerarToken([]byte("otra-clave"), 1, "a@upse.edu.ec", "admin")
	inactiva := func(context.Context, int) bool { return false }

	casos := []struct {
		nombre, cabecera string
		cuenta           CuentaActiva
		soloAdmin        bool
		quiere           int
	}{
		{"sin token", "", activa, false, 401},
		{"token mal firmado", "Bearer " + ajeno, activa, false, 401},
		{"basura", "Bearer abc.def", activa, false, 401},
		{"cliente valido", "Bearer " + cliente, activa, false, 200},
		{"cuenta desactivada", "Bearer " + cliente, inactiva, false, 401},
		{"cliente en ruta admin", "Bearer " + cliente, activa, true, 403},
		{"admin en ruta admin", "Bearer " + admin, activa, true, 200},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			if got := codigo(t, c.cabecera, c.cuenta, c.soloAdmin); got != c.quiere {
				t.Errorf("esperaba %d, llegó %d", c.quiere, got)
			}
		})
	}
}
