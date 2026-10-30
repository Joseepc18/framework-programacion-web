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
	// F2-4: el asistente redacta notas pero no las firma ni completa procedimientos.
	{"GET", "/api/pacientes/1/notas", rolesClinicos},
	{"POST", "/api/pacientes/1/notas", rolesClinicos},
	{"PUT", "/api/pacientes/1/notas/1", rolesClinicos},
	{"POST", "/api/pacientes/1/notas/1/firmar", soloOdontologo},
	{"POST", "/api/pacientes/1/adendas", soloOdontologo},
	{"GET", "/api/pacientes/1/procedimientos-realizados", rolesClinicos},
	{"POST", "/api/pacientes/1/procedimientos-realizados", soloOdontologo},
	{"POST", "/api/pacientes/1/procedimientos-realizados/1/completar", soloOdontologo},
	{"POST", "/api/pacientes/1/procedimientos-realizados/1/cancelar", soloOdontologo},
	// F2-5: consentimientos; las plantillas las mantiene el administrador TI.
	{"GET", "/api/consentimiento-plantillas", []string{models.RolOdontologo, models.RolAdminTI}},
	{"POST", "/api/consentimiento-plantillas", []string{models.RolAdminTI}},
	{"PUT", "/api/consentimiento-plantillas/1", []string{models.RolAdminTI}},
	{"GET", "/api/pacientes/1/consentimientos", rolesClinicos},
	{"POST", "/api/pacientes/1/consentimientos", soloOdontologo},
	{"POST", "/api/pacientes/1/consentimientos/1/otp", rolesClinicos},
	{"POST", "/api/pacientes/1/consentimientos/1/firmar", rolesClinicos},
	{"POST", "/api/pacientes/1/consentimientos/1/revocar", soloOdontologo},
	{"GET", "/api/pacientes/1/notificaciones", []string{models.RolOdontologo, models.RolAsistente, models.RolRecepcionista}},
	// F2-6: solo el odontólogo emite, firma y anula recetas; el asistente las consulta.
	{"GET", "/api/pacientes/1/recetas", rolesClinicos},
	{"POST", "/api/pacientes/1/recetas", soloOdontologo},
	{"POST", "/api/pacientes/1/recetas/verificar-alergias", soloOdontologo},
	{"GET", "/api/pacientes/1/recetas/1", rolesClinicos},
	{"PUT", "/api/pacientes/1/recetas/1", soloOdontologo},
	{"POST", "/api/pacientes/1/recetas/1/firmar", soloOdontologo},
	{"POST", "/api/pacientes/1/recetas/1/anular", soloOdontologo},
	{"GET", "/api/pacientes/1/recetas/1/pdf", rolesClinicos},
	// F2-7: el odontólogo arma el plan; la recepción y el odontólogo consultan, presentan y registran la aprobación.
	{"GET", "/api/pacientes/1/presupuestos", []string{models.RolOdontologo, models.RolRecepcionista}},
	{"POST", "/api/pacientes/1/presupuestos", soloOdontologo},
	{"PUT", "/api/pacientes/1/presupuestos/1", soloOdontologo},
	{"POST", "/api/pacientes/1/presupuestos/1/presentar", []string{models.RolOdontologo, models.RolRecepcionista}},
	{"POST", "/api/pacientes/1/presupuestos/1/resolver", []string{models.RolOdontologo, models.RolRecepcionista}},
	{"POST", "/api/pacientes/1/presupuestos/1/nueva-version", soloOdontologo},
	{"GET", "/api/pacientes/1/presupuestos/1/pdf", []string{models.RolOdontologo, models.RolRecepcionista}},
	{"GET", "/api/pacientes/1/tareas", []string{models.RolOdontologo, models.RolRecepcionista}},
	// F2-8: el portal es solo del paciente; ningún rol del personal entra a sus rutas.
	{"GET", "/api/portal/yo", []string{models.RolPaciente}},
	{"GET", "/api/portal/citas", []string{models.RolPaciente}},
	{"POST", "/api/portal/citas/1/confirmar", []string{models.RolPaciente}},
	{"POST", "/api/portal/citas/1/cancelar", []string{models.RolPaciente}},
	{"GET", "/api/portal/recetas", []string{models.RolPaciente}},
	{"GET", "/api/portal/recetas/1/pdf", []string{models.RolPaciente}},
	{"GET", "/api/portal/consentimientos", []string{models.RolPaciente}},
	{"GET", "/api/portal/consentimientos/1/pdf", []string{models.RolPaciente}},
	// El paciente no entra a las listas del catálogo: son para el personal.
	{"GET", "/api/catalogo/consultorios", []string{models.RolAdminTI, models.RolRecepcionista, models.RolOdontologo, models.RolAsistente, models.RolAuditor}},
	{"GET", "/api/profesionales", []string{models.RolAdminTI, models.RolRecepcionista, models.RolOdontologo, models.RolAsistente, models.RolAuditor}},
	{"GET", "/api/parametros", []string{models.RolAdminTI, models.RolRecepcionista, models.RolOdontologo, models.RolAsistente, models.RolAuditor}},
	// F2-2: el odontólogo escribe el odontograma; el asistente solo lo consulta.
	{"GET", "/api/pacientes/1/odontograma", rolesClinicos},
	{"GET", "/api/pacientes/1/odontograma/historial", rolesClinicos},
	{"POST", "/api/pacientes/1/odontograma/hallazgos", soloOdontologo},
	{"PUT", "/api/pacientes/1/odontograma/hallazgos/1", soloOdontologo},
	{"POST", "/api/pacientes/1/odontograma/hallazgos/1/retirar", soloOdontologo},
}

func TestPermisosDeLosModulosClinicos(t *testing.T) {
	app := appDePrueba(t)
	roles := []string{models.RolAdminTI, models.RolRecepcionista, models.RolOdontologo, models.RolAsistente, models.RolAuditor, models.RolPaciente}
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
