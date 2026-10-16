// Pruebas de los controladores con un repositorio falso: no hace falta base de datos
// porque dependen de interfaces (inyección de dependencias).
package controllers

import (
	"context"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
	"multicatalogo-backend/models"
)

// repoFalso embebe la interfaz (nil) y solo implementa lo que cada prueba necesita;
// llamar a un método no implementado provoca pánico y delata una dependencia inesperada.
type repoFalso struct {
	DBRepository
	usuario      models.Usuario
	existe       bool
	errCrear     error
	errLiquidar  error
	errEstado    error
	rangoVisto   bool
	actualizados int
}

func (r *repoFalso) BuscarUsuarioPorEmail(string) (models.Usuario, bool, error) {
	return r.usuario, r.existe, nil
}

func (r *repoFalso) CrearUsuario(_ context.Context, req models.RegisterRequest, _ string) (models.Usuario, error) {
	if r.errCrear != nil {
		return models.Usuario{}, r.errCrear
	}
	return models.Usuario{ID: 9, Email: req.Email, Nombre: req.Nombre, Rol: "cliente"}, nil
}

func (r *repoFalso) CrearLiquidacion(context.Context, int) (models.Liquidacion, error) {
	return models.Liquidacion{}, r.errLiquidar
}

func (r *repoFalso) CambiarEstadoPedido(context.Context, int, string) (models.Pedido, error) {
	return models.Pedido{}, r.errEstado
}

func (r *repoFalso) ActualizarUsuario(context.Context, int, *string, *bool) (models.UsuarioAdmin, bool, error) {
	r.actualizados++
	return models.UsuarioAdmin{}, true, nil
}

func (r *repoFalso) ObtenerResumenVentas(context.Context, models.RangoFechas) (models.ResumenVentas, error) {
	r.rangoVisto = true
	return models.ResumenVentas{}, nil
}

// cacheFalsa descarta todo: las pruebas no dependen de Redis.
type cacheFalsa struct{ CacheRepository }

func (cacheFalsa) Eliminar(context.Context, ...string) {}

// correoFalso cuenta los envíos para comprobar que el controlador avisa al mailer.
type correoFalso struct{ bienvenidas int }

func (m *correoFalso) EnviarBienvenida(context.Context, string, string) { m.bienvenidas++ }
func (m *correoFalso) EnviarConfirmacionPedido(context.Context, string, string, float64) {
}

// peticion monta una app mínima, inyecta la identidad (como haría RequireAuth) y devuelve el código HTTP.
func peticion(t *testing.T, metodo, ruta, patron, cuerpo string, usuarioID int, manejador fiber.Handler) int {
	t.Helper()
	app := fiber.New()
	app.Add(metodo, patron, func(c *fiber.Ctx) error {
		c.Locals("usuario_id", usuarioID)
		c.Locals("email", "x@upse.edu.ec")
		return c.Next()
	}, manejador)
	req := httptest.NewRequest(metodo, ruta, strings.NewReader(cuerpo))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	return resp.StatusCode
}

func hash(t *testing.T, clave string) string {
	t.Helper()
	h, err := bcrypt.GenerateFromPassword([]byte(clave), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}
	return string(h)
}

func TestLogin(t *testing.T) {
	casos := []struct {
		nombre string
		repo   *repoFalso
		cuerpo string
		quiere int
	}{
		{"correcto", &repoFalso{existe: true, usuario: models.Usuario{ID: 1, Email: "a@b.ec", Rol: "cliente", Activo: true, Password: hash(t, "123456")}}, `{"email":"a@b.ec","password":"123456"}`, 200},
		{"clave incorrecta", &repoFalso{existe: true, usuario: models.Usuario{Activo: true, Password: hash(t, "123456")}}, `{"email":"a@b.ec","password":"mala"}`, 401},
		{"usuario inexistente", &repoFalso{}, `{"email":"a@b.ec","password":"123456"}`, 401},
		{"cuenta inactiva", &repoFalso{existe: true, usuario: models.Usuario{Activo: false, Password: hash(t, "123456")}}, `{"email":"a@b.ec","password":"123456"}`, 403},
		{"campos vacios", &repoFalso{}, `{"email":"","password":""}`, 400},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			ctrl := NuevoAuthController(c.repo, []byte("secreto-de-prueba"), &correoFalso{})
			if got := peticion(t, "POST", "/login", "/login", c.cuerpo, 0, ctrl.Login); got != c.quiere {
				t.Errorf("esperaba %d, llegó %d", c.quiere, got)
			}
		})
	}
}

func TestRegister(t *testing.T) {
	casos := []struct {
		nombre string
		err    error
		cuerpo string
		quiere int
	}{
		{"creado", nil, `{"nombre":"Ana","email":"ana@upse.edu.ec","password":"123456"}`, 201},
		{"email duplicado", models.ErrEmailYaExiste, `{"nombre":"Ana","email":"ana@upse.edu.ec","password":"123456"}`, 409},
		{"ref inexistente", models.ErrCodigoInvalido, `{"nombre":"Ana","email":"ana@upse.edu.ec","password":"123456","ref":"NOEXISTE"}`, 400},
		{"sin nombre", nil, `{"nombre":"","email":"ana@upse.edu.ec","password":"123456"}`, 400},
		{"email mal formado", nil, `{"nombre":"Ana","email":"ana","password":"123456"}`, 400},
		{"clave corta", nil, `{"nombre":"Ana","email":"ana@upse.edu.ec","password":"123"}`, 400},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			correo := &correoFalso{}
			ctrl := NuevoAuthController(&repoFalso{errCrear: c.err}, nil, correo)
			got := peticion(t, "POST", "/register", "/register", c.cuerpo, 0, ctrl.Register)
			if got != c.quiere {
				t.Errorf("esperaba %d, llegó %d", c.quiere, got)
			}
			if (got == 201) != (correo.bienvenidas == 1) {
				t.Errorf("el correo de bienvenida solo debe enviarse al crear la cuenta (envíos: %d)", correo.bienvenidas)
			}
		})
	}
}

func TestActualizarUsuarioNoPermiteAutoDegradarse(t *testing.T) {
	repo := &repoFalso{}
	ctrl := NuevoUsuarioController(repo)
	for _, cuerpo := range []string{`{"activo":false}`, `{"rol":"cliente"}`} {
		if got := peticion(t, "PATCH", "/usuarios/1", "/usuarios/:id", cuerpo, 1, ctrl.ActualizarUsuario); got != 409 {
			t.Errorf("%s: esperaba 409, llegó %d", cuerpo, got)
		}
	}
	if repo.actualizados != 0 {
		t.Error("no debía llegar al repositorio")
	}
	// Sobre otra cuenta sí está permitido.
	if got := peticion(t, "PATCH", "/usuarios/2", "/usuarios/:id", `{"activo":false}`, 1, ctrl.ActualizarUsuario); got != 200 {
		t.Errorf("esperaba 200, llegó %d", got)
	}
}

func TestLiquidacionSinSaldoDevuelve409(t *testing.T) {
	ctrl := NuevoLiquidacionController(&repoFalso{errLiquidar: models.ErrSinComisionesPendientes})
	if got := peticion(t, "POST", "/liquidaciones", "/liquidaciones", `{"referido_id":2}`, 1, ctrl.CrearLiquidacion); got != 409 {
		t.Errorf("esperaba 409, llegó %d", got)
	}
}

func TestReporteRechazaFechaInvalida(t *testing.T) {
	repo := &repoFalso{}
	ctrl := NuevoReporteController(repo)
	if got := peticion(t, "GET", "/r?desde=ayer", "/r", "", 1, ctrl.GetReporteResumen); got != 400 {
		t.Errorf("esperaba 400, llegó %d", got)
	}
	if repo.rangoVisto {
		t.Error("con fecha inválida no debe consultarse el repositorio")
	}
	if got := peticion(t, "GET", "/r", "/r", "", 1, ctrl.GetReporteResumen); got != 200 {
		t.Errorf("sin filtros esperaba 200, llegó %d", got)
	}
}

func TestPedidoTransicionInvalidaDevuelve400(t *testing.T) {
	ctrl := NuevoPedidoController(&repoFalso{errEstado: models.ErrTransicionInvalida}, cacheFalsa{})
	got := peticion(t, "PATCH", "/pedidos/1/estado", "/pedidos/:id/estado", `{"estado":"pendiente"}`, 1, ctrl.CambiarEstadoPedido)
	if got != 400 {
		t.Errorf("esperaba 400, llegó %d", got)
	}
}
