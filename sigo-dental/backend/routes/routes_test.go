// Pruebas de la matriz de permisos: cada rol solo entra a su propia ruta de prueba.
package routes

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"sigo-dental-backend/controllers"
	"sigo-dental-backend/middleware"
	"sigo-dental-backend/models"
)

var secretoPrueba = []byte("secreto-de-prueba-con-32-bytes-o-mas")

func appDePrueba(t *testing.T) *fiber.App {
	t.Helper()
	auth, err := controllers.NuevoAuthController(nil, secretoPrueba)
	if err != nil {
		t.Fatal(err)
	}
	app := fiber.New()
	SetupRoutes(app, Manejadores{
		Health:    controllers.NuevoHealthController(nil, nil),
		Auth:      auth,
		JWTSecret: secretoPrueba,
	})
	return app
}

func TestMatrizDePermisosPorRol(t *testing.T) {
	app := appDePrueba(t)
	roles := []string{
		models.RolAdminTI, models.RolRecepcionista, models.RolOdontologo, models.RolAsistente, models.RolAuditor,
	}
	if len(rutasPrueba) != len(roles) {
		t.Fatalf("hay %d rutas de prueba y %d roles", len(rutasPrueba), len(roles))
	}

	for _, rolDelToken := range roles {
		token, err := middleware.GenerarToken(secretoPrueba, 1, rolDelToken+"@sigo-dental.test", rolDelToken)
		if err != nil {
			t.Fatal(err)
		}
		for ruta, rolDeLaRuta := range rutasPrueba {
			quiere := fiber.StatusForbidden
			if rolDelToken == rolDeLaRuta {
				quiere = fiber.StatusOK
			}
			t.Run(rolDelToken+" en "+ruta, func(t *testing.T) {
				req := httptest.NewRequest("GET", ruta, nil)
				req.Header.Set("Authorization", "Bearer "+token)
				resp, err := app.Test(req)
				if err != nil {
					t.Fatal(err)
				}
				if resp.StatusCode != quiere {
					t.Errorf("estado = %d, se esperaba %d", resp.StatusCode, quiere)
				}
			})
		}
	}
}

func TestRutasDePruebaExigenToken(t *testing.T) {
	app := appDePrueba(t)
	for ruta := range rutasPrueba {
		resp, err := app.Test(httptest.NewRequest("GET", ruta, nil))
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != fiber.StatusUnauthorized {
			t.Errorf("%s sin token: estado = %d, se esperaba 401", ruta, resp.StatusCode)
		}
	}
}
