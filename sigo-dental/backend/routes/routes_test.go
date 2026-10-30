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
		Catalogo:  controllers.NuevoCatalogoController(nil),
		Pacientes: controllers.NuevoPacientesController(nil),
		Citas:     controllers.NuevoCitasController(nil, nil, secretoPrueba),
		Cobros:    controllers.NuevoCobrosController(nil),
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

// Cualquier cuenta con sesión consulta el catálogo, pero solo el administrador TI lo modifica:
// los demás roles reciben 403 antes de llegar al controlador, y sin token todo es 401.
func TestCatalogoSoloLoModificaElAdminTI(t *testing.T) {
	app := appDePrueba(t)
	escrituras := []struct{ metodo, ruta string }{
		{"POST", "/api/catalogo/consultorios"},
		{"PUT", "/api/catalogo/consultorios/1"},
		{"POST", "/api/sillones"},
		{"PUT", "/api/sillones/1"},
		{"POST", "/api/profesionales"},
		{"PUT", "/api/profesionales/1"},
		{"GET", "/api/profesionales/candidatos"},
		{"POST", "/api/procedimientos"},
		{"PUT", "/api/procedimientos/1"},
		{"PUT", "/api/parametros/agenda_hora_fin"},
	}
	lecturas := []string{
		"/api/catalogo/consultorios", "/api/sillones", "/api/profesionales", "/api/procedimientos", "/api/parametros",
	}

	for _, rol := range []string{models.RolRecepcionista, models.RolOdontologo, models.RolAsistente, models.RolAuditor} {
		token, err := middleware.GenerarToken(secretoPrueba, 1, rol+"@sigo-dental.test", rol)
		if err != nil {
			t.Fatal(err)
		}
		for _, e := range escrituras {
			t.Run(rol+" "+e.metodo+" "+e.ruta, func(t *testing.T) {
				req := httptest.NewRequest(e.metodo, e.ruta, nil)
				req.Header.Set("Authorization", "Bearer "+token)
				resp, err := app.Test(req)
				if err != nil {
					t.Fatal(err)
				}
				if resp.StatusCode != fiber.StatusForbidden {
					t.Errorf("estado = %d, se esperaba 403", resp.StatusCode)
				}
			})
		}
	}

	for _, e := range escrituras {
		resp, err := app.Test(httptest.NewRequest(e.metodo, e.ruta, nil))
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != fiber.StatusUnauthorized {
			t.Errorf("%s %s sin token: estado = %d, se esperaba 401", e.metodo, e.ruta, resp.StatusCode)
		}
	}
	for _, ruta := range lecturas {
		resp, err := app.Test(httptest.NewRequest("GET", ruta, nil))
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != fiber.StatusUnauthorized {
			t.Errorf("GET %s sin token: estado = %d, se esperaba 401", ruta, resp.StatusCode)
		}
	}
}

// Cada rol solo llega a lo que le corresponde en pacientes; el resto recibe 403 antes del controlador.
func TestPermisosDePacientes(t *testing.T) {
	app := appDePrueba(t)
	prohibidos := []struct{ rol, metodo, ruta string }{
		{models.RolAuditor, "GET", "/api/pacientes"},
		{models.RolAdminTI, "GET", "/api/pacientes"},
		{models.RolAuditor, "GET", "/api/pacientes/1"},
		{models.RolAdminTI, "GET", "/api/pacientes/1"},
		{models.RolOdontologo, "POST", "/api/pacientes"},
		{models.RolAsistente, "POST", "/api/pacientes"},
		{models.RolAuditor, "POST", "/api/pacientes"},
		{models.RolOdontologo, "PUT", "/api/pacientes/1"},
		{models.RolAsistente, "PUT", "/api/pacientes/1"},
		{models.RolRecepcionista, "PUT", "/api/pacientes/1/anamnesis"},
		{models.RolAuditor, "PUT", "/api/pacientes/1/anamnesis"},
		{models.RolAdminTI, "PUT", "/api/pacientes/1/anamnesis"},
	}
	for _, p := range prohibidos {
		t.Run(p.rol+" "+p.metodo+" "+p.ruta, func(t *testing.T) {
			token, err := middleware.GenerarToken(secretoPrueba, 1, p.rol+"@sigo-dental.test", p.rol)
			if err != nil {
				t.Fatal(err)
			}
			req := httptest.NewRequest(p.metodo, p.ruta, nil)
			req.Header.Set("Authorization", "Bearer "+token)
			resp, err := app.Test(req)
			if err != nil {
				t.Fatal(err)
			}
			if resp.StatusCode != fiber.StatusForbidden {
				t.Errorf("estado = %d, se esperaba 403", resp.StatusCode)
			}
		})
	}

	for _, ruta := range []string{"/api/pacientes", "/api/pacientes/1"} {
		resp, err := app.Test(httptest.NewRequest("GET", ruta, nil))
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != fiber.StatusUnauthorized {
			t.Errorf("GET %s sin token: estado = %d, se esperaba 401", ruta, resp.StatusCode)
		}
	}
}

// En la agenda solo la recepción agenda y reprograma; el auditor y el administrador TI no la ven.
func TestPermisosDeCitas(t *testing.T) {
	app := appDePrueba(t)
	prohibidos := []struct{ rol, metodo, ruta string }{
		{models.RolAuditor, "GET", "/api/citas"},
		{models.RolAdminTI, "GET", "/api/citas"},
		{models.RolAuditor, "GET", "/api/citas/1"},
		{models.RolAuditor, "GET", "/api/agenda/horario"},
		{models.RolOdontologo, "POST", "/api/citas"},
		{models.RolAsistente, "POST", "/api/citas"},
		{models.RolAuditor, "POST", "/api/citas"},
		{models.RolOdontologo, "PUT", "/api/citas/1"},
		{models.RolAsistente, "PUT", "/api/citas/1"},
		{models.RolAuditor, "POST", "/api/citas/1/estado"},
		{models.RolOdontologo, "GET", "/api/citas/1/notificaciones"},
		{models.RolAsistente, "GET", "/api/citas/1/notificaciones"},
		{models.RolAuditor, "GET", "/api/citas/1/notificaciones"},
		{models.RolAdminTI, "POST", "/api/citas/1/estado"},
	}
	for _, p := range prohibidos {
		t.Run(p.rol+" "+p.metodo+" "+p.ruta, func(t *testing.T) {
			token, err := middleware.GenerarToken(secretoPrueba, 1, p.rol+"@sigo-dental.test", p.rol)
			if err != nil {
				t.Fatal(err)
			}
			req := httptest.NewRequest(p.metodo, p.ruta, nil)
			req.Header.Set("Authorization", "Bearer "+token)
			resp, err := app.Test(req)
			if err != nil {
				t.Fatal(err)
			}
			if resp.StatusCode != fiber.StatusForbidden {
				t.Errorf("estado = %d, se esperaba 403", resp.StatusCode)
			}
		})
	}

	for _, ruta := range []string{"/api/citas", "/api/citas/1", "/api/agenda/horario"} {
		resp, err := app.Test(httptest.NewRequest("GET", ruta, nil))
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != fiber.StatusUnauthorized {
			t.Errorf("GET %s sin token: estado = %d, se esperaba 401", ruta, resp.StatusCode)
		}
	}
}

// Solo la recepción cobra; el auditor consulta; el resto de los roles no ve cobros.
func TestPermisosDeCobros(t *testing.T) {
	app := appDePrueba(t)
	prohibidos := []struct{ rol, metodo, ruta string }{
		{models.RolOdontologo, "GET", "/api/cobros"},
		{models.RolAsistente, "GET", "/api/cobros/1"},
		{models.RolAdminTI, "GET", "/api/cobros"},
		{models.RolAuditor, "POST", "/api/cobros"},
		{models.RolAuditor, "POST", "/api/cobros/1/pagos"},
		{models.RolAuditor, "POST", "/api/cobros/1/anular"},
		{models.RolOdontologo, "POST", "/api/cobros"},
		{models.RolAsistente, "POST", "/api/cobros/1/pagos"},
		{models.RolAdminTI, "POST", "/api/cobros/1/anular"},
	}
	for _, p := range prohibidos {
		t.Run(p.rol+" "+p.metodo+" "+p.ruta, func(t *testing.T) {
			token, err := middleware.GenerarToken(secretoPrueba, 1, p.rol+"@sigo-dental.test", p.rol)
			if err != nil {
				t.Fatal(err)
			}
			req := httptest.NewRequest(p.metodo, p.ruta, nil)
			req.Header.Set("Authorization", "Bearer "+token)
			resp, err := app.Test(req)
			if err != nil {
				t.Fatal(err)
			}
			if resp.StatusCode != fiber.StatusForbidden {
				t.Errorf("estado = %d, se esperaba 403", resp.StatusCode)
			}
		})
	}
	for _, ruta := range []string{"/api/cobros", "/api/cobros/1"} {
		resp, err := app.Test(httptest.NewRequest("GET", ruta, nil))
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != fiber.StatusUnauthorized {
			t.Errorf("GET %s sin token: estado = %d, se esperaba 401", ruta, resp.StatusCode)
		}
	}
}
