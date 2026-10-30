// Matriz de permisos de los módulos clínicos de la Fase 2: cada ruta declara los roles que la usan y todos los
// demás reciben 403 antes de llegar al controlador; sin token, 401. (El camino permitido se prueba en los controladores.)
package routes

import (
	"net/http/httptest"
	"slices"
	"testing"

	"github.com/gofiber/fiber/v2"
	"sigo-dental-backend/middleware"
	"sigo-dental-backend/models"
)

type permisoRuta struct {
	metodo, ruta string
	roles        []string
}

var (
	rolesClinicos  = []string{models.RolOdontologo, models.RolAsistente}
	soloOdontologo = []string{models.RolOdontologo}
)

var permisosClinicos = []permisoRuta{
	{"GET", "/api/pacientes/1/signos-vitales", rolesClinicos},
	{"POST", "/api/pacientes/1/signos-vitales", rolesClinicos},
	{"GET", "/api/pacientes/1/archivos", rolesClinicos},
	{"POST", "/api/pacientes/1/archivos", rolesClinicos},
	{"GET", "/api/archivos/1/descargar", rolesClinicos},
	{"GET", "/api/auditoria", []string{models.RolAdminTI, models.RolAuditor}},
	{"GET", "/api/pacientes/1/periodontogramas", rolesClinicos},
	{"GET", "/api/pacientes/1/periodontogramas/1", rolesClinicos},
	{"POST", "/api/pacientes/1/periodontogramas", rolesClinicos},
	// F2-2: el odontólogo escribe el odontograma; el asistente solo lo consulta.
	{"GET", "/api/pacientes/1/odontograma", rolesClinicos},
	{"GET", "/api/pacientes/1/odontograma/historial", rolesClinicos},
	{"POST", "/api/pacientes/1/odontograma/hallazgos", soloOdontologo},
	{"PUT", "/api/pacientes/1/odontograma/hallazgos/1", soloOdontologo},
	{"POST", "/api/pacientes/1/odontograma/hallazgos/1/retirar", soloOdontologo},
}

func TestPermisosDeLosModulosClinicos(t *testing.T) {
	app := appDePrueba(t)
	roles := []string{models.RolAdminTI, models.RolRecepcionista, models.RolOdontologo, models.RolAsistente, models.RolAuditor}
	for _, p := range permisosClinicos {
		for _, rol := range roles {
			if slices.Contains(p.roles, rol) {
				continue
			}
			token, err := middleware.GenerarToken(secretoPrueba, 1, rol+"@sigo-dental.test", rol)
			if err != nil {
				t.Fatal(err)
			}
			t.Run(rol+" "+p.metodo+" "+p.ruta, func(t *testing.T) {
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
		resp, err := app.Test(httptest.NewRequest(p.metodo, p.ruta, nil))
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != fiber.StatusUnauthorized {
			t.Errorf("%s %s sin token: estado = %d, se esperaba 401", p.metodo, p.ruta, resp.StatusCode)
		}
	}
}
