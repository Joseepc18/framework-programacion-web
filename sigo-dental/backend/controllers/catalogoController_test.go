// Pruebas del catálogo médico con un repositorio falso: no hace falta PostgreSQL.
package controllers

import (
	"context"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"sigo-dental-backend/models"
)

// repoCatalogo implementa solo lo que cada prueba necesita; el resto de CatalogoRepository
// queda en nil y fallaría si una prueba lo llamara sin querer.
type repoCatalogo struct {
	CatalogoRepository
	err           error
	soloActivos   *bool
	parametros    map[string]string
	procedimiento models.Procedimiento
}

func (r *repoCatalogo) ListarSimples(_ context.Context, _ string, soloActivos bool) ([]models.ItemCatalogo, error) {
	r.soloActivos = &soloActivos
	return []models.ItemCatalogo{}, r.err
}

func (r *repoCatalogo) CrearSimple(_ context.Context, _, nombre string) (models.ItemCatalogo, error) {
	return models.ItemCatalogo{ID: 1, Nombre: nombre, Activo: true}, r.err
}

func (r *repoCatalogo) ActualizarSimple(_ context.Context, _ string, id int, nombre string, activo bool) (models.ItemCatalogo, error) {
	return models.ItemCatalogo{ID: id, Nombre: nombre, Activo: activo}, r.err
}

func (r *repoCatalogo) CrearProcedimiento(_ context.Context, p models.Procedimiento) (models.Procedimiento, error) {
	r.procedimiento = p
	return p, r.err
}

func (r *repoCatalogo) ObtenerParametro(_ context.Context, clave string) (string, error) {
	valor, ok := r.parametros[clave]
	if !ok {
		return "", models.ErrNoEncontrado
	}
	return valor, nil
}

func (r *repoCatalogo) ActualizarParametro(_ context.Context, clave, valor string) (models.Parametro, error) {
	return models.Parametro{Clave: clave, Valor: valor}, r.err
}

// appCatalogo monta las rutas del catálogo que se prueban, con el rol de la cuenta ya fijado.
func appCatalogo(repo *repoCatalogo, rol string) *fiber.App {
	ctrl := NuevoCatalogoController(repo)
	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("rol", rol)
		return c.Next()
	})
	app.Get("/api/catalogo/:recurso", ctrl.ListarSimples)
	app.Post("/api/catalogo/:recurso", ctrl.CrearSimple)
	app.Put("/api/catalogo/:recurso/:id", ctrl.ActualizarSimple)
	app.Post("/api/procedimientos", ctrl.CrearProcedimiento)
	app.Put("/api/parametros/:clave", ctrl.ActualizarParametro)
	return app
}

func enviar(t *testing.T, app *fiber.App, metodo, ruta, cuerpo string) int {
	t.Helper()
	req := httptest.NewRequest(metodo, ruta, strings.NewReader(cuerpo))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	return resp.StatusCode
}

func TestCatalogoSimple(t *testing.T) {
	casos := []struct {
		nombre, metodo, ruta, cuerpo string
		err                          error
		estado                       int
	}{
		{"crear válido", "POST", "/api/catalogo/consultorios", `{"nombre":"Consultorio 3"}`, nil, 201},
		{"crear con espacios alrededor", "POST", "/api/catalogo/consultorios", `{"nombre":"  Consultorio 3  "}`, nil, 201},
		{"crear sin nombre", "POST", "/api/catalogo/consultorios", `{"nombre":"   "}`, nil, 400},
		{"crear con nombre demasiado largo", "POST", "/api/catalogo/convenios", `{"nombre":"` + strings.Repeat("a", 101) + `"}`, nil, 400},
		{"crear duplicado", "POST", "/api/catalogo/consultorios", `{"nombre":"Consultorio 1"}`, models.ErrDuplicado, 409},
		{"crear en un recurso que no existe", "POST", "/api/catalogo/otro", `{"nombre":"x"}`, models.ErrNoEncontrado, 404},
		{"cuerpo que no es JSON", "POST", "/api/catalogo/consultorios", `no-es-json`, nil, 400},
		{"actualizar válido", "PUT", "/api/catalogo/convenios/2", `{"nombre":"Seguro A","activo":false}`, nil, 200},
		{"actualizar sin estado", "PUT", "/api/catalogo/convenios/2", `{"nombre":"Seguro A"}`, nil, 400},
		{"actualizar con id inválido", "PUT", "/api/catalogo/convenios/abc", `{"nombre":"x","activo":true}`, nil, 400},
		{"actualizar inexistente", "PUT", "/api/catalogo/convenios/99", `{"nombre":"x","activo":true}`, models.ErrNoEncontrado, 404},
		{"falla la base de datos", "POST", "/api/catalogo/consultorios", `{"nombre":"x"}`, context.DeadlineExceeded, 500},
	}
	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			app := appCatalogo(&repoCatalogo{err: caso.err}, models.RolAdminTI)
			if estado := enviar(t, app, caso.metodo, caso.ruta, caso.cuerpo); estado != caso.estado {
				t.Errorf("estado = %d, se esperaba %d", estado, caso.estado)
			}
		})
	}
}

// Solo el administrador TI puede pedir los elementos desactivados; el resto siempre recibe solo los activos.
func TestListadoOcultaInactivosSalvoAdmin(t *testing.T) {
	casos := []struct {
		rol, ruta   string
		soloActivos bool
	}{
		{models.RolAdminTI, "/api/catalogo/convenios?todos=true", false},
		{models.RolAdminTI, "/api/catalogo/convenios", true},
		{models.RolRecepcionista, "/api/catalogo/convenios?todos=true", true},
		{models.RolOdontologo, "/api/catalogo/convenios?todos=true", true},
	}
	for _, caso := range casos {
		t.Run(caso.rol+" "+caso.ruta, func(t *testing.T) {
			repo := &repoCatalogo{}
			enviar(t, appCatalogo(repo, caso.rol), "GET", caso.ruta, "")
			if repo.soloActivos == nil || *repo.soloActivos != caso.soloActivos {
				t.Errorf("soloActivos = %v, se esperaba %v", repo.soloActivos, caso.soloActivos)
			}
		})
	}
}

func TestCrearProcedimiento(t *testing.T) {
	base := func(cambio string) string {
		cuerpo := `{"codigo":"LIMP-02","nombre":"Limpieza","especialidad":"General","precio_centavos":4000,"duracion_min":45,"requiere_consentimiento":false}`
		if cambio == "" {
			return cuerpo
		}
		return strings.Replace(cuerpo, cambio[:strings.Index(cambio, "=>")], cambio[strings.Index(cambio, "=>")+2:], 1)
	}
	casos := []struct {
		nombre, cuerpo string
		err            error
		estado         int
	}{
		{"válido", base(""), nil, 201},
		{"sin código", base(`"codigo":"LIMP-02"=>"codigo":""`), nil, 400},
		{"precio negativo", base(`"precio_centavos":4000=>"precio_centavos":-1`), nil, 400},
		{"precio cero es válido", base(`"precio_centavos":4000=>"precio_centavos":0`), nil, 201},
		{"duración cero", base(`"duracion_min":45=>"duracion_min":0`), nil, 400},
		{"código repetido", base(""), models.ErrDuplicado, 409},
	}
	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			app := appCatalogo(&repoCatalogo{err: caso.err}, models.RolAdminTI)
			if estado := enviar(t, app, "POST", "/api/procedimientos", caso.cuerpo); estado != caso.estado {
				t.Errorf("estado = %d, se esperaba %d", estado, caso.estado)
			}
		})
	}
}

func TestActualizarParametroHorasDeAgenda(t *testing.T) {
	casos := []struct {
		nombre, clave, cuerpo string
		estado                int
	}{
		{"cierre válido", ClaveAgendaHoraFin, `{"valor":"20:00"}`, 200},
		{"apertura válida", ClaveAgendaHoraInicio, `{"valor":"08:30"}`, 200},
		{"hora con un solo dígito", ClaveAgendaHoraFin, `{"valor":"9:00"}`, 400},
		{"hora imposible", ClaveAgendaHoraFin, `{"valor":"25:00"}`, 400},
		{"texto que no es hora", ClaveAgendaHoraInicio, `{"valor":"temprano"}`, 400},
		{"apertura igual al cierre", ClaveAgendaHoraInicio, `{"valor":"21:00"}`, 400},
		{"apertura después del cierre", ClaveAgendaHoraInicio, `{"valor":"22:00"}`, 400},
		{"cierre antes de la apertura", ClaveAgendaHoraFin, `{"valor":"06:00"}`, 400},
		{"valor vacío", ClaveAgendaHoraFin, `{"valor":""}`, 400},
		{"clave sin reglas de formato", "otra_clave", `{"valor":"x"}`, 200},
	}
	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			repo := &repoCatalogo{parametros: map[string]string{ClaveAgendaHoraInicio: "07:00", ClaveAgendaHoraFin: "21:00"}}
			app := appCatalogo(repo, models.RolAdminTI)
			if estado := enviar(t, app, "PUT", "/api/parametros/"+caso.clave, caso.cuerpo); estado != caso.estado {
				t.Errorf("estado = %d, se esperaba %d", estado, caso.estado)
			}
		})
	}
}
