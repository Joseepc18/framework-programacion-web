package controllers

import (
	"context"
	"testing"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
	"sigo-dental-backend/models"
)

type repoPresupuestos struct {
	PresupuestoRepository
	usuarios  map[string]models.Usuario
	previo    models.Presupuesto
	guardados []models.ItemParaGuardar
}

func (r *repoPresupuestos) ProfesionalDeUsuario(context.Context, int) (models.ProfesionalDeUsuario, bool, error) {
	return models.ProfesionalDeUsuario{ID: 1, Nombre: "Dra. Prueba"}, true, nil
}

func (r *repoPresupuestos) BuscarUsuarioPorEmail(_ context.Context, email string) (models.Usuario, bool, error) {
	u, ok := r.usuarios[email]
	return u, ok, nil
}

func (r *repoPresupuestos) ObtenerParametro(context.Context, string) (string, error) {
	return "odontologo,admin_ti", nil
}

func (r *repoPresupuestos) ObtenerPresupuesto(_ context.Context, _, id int) (models.Presupuesto, error) {
	if r.previo.ID == id {
		return r.previo, nil
	}
	return models.Presupuesto{ID: id, Estado: models.PresupBorrador}, nil
}

func (r *repoPresupuestos) GuardarPresupuesto(_ context.Context, _, id, _ int, _ string, items []models.ItemParaGuardar, _ int) (int, error) {
	r.guardados = items
	if id == 0 {
		id = 9
	}
	return id, nil
}

func appPresupuestos(t *testing.T, repo *repoPresupuestos) *fiber.App {
	t.Helper()
	hash, err := bcrypt.GenerateFromPassword([]byte("clave"), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}
	repo.usuarios = map[string]models.Usuario{
		"doctor@x.test": {ID: 10, Email: "doctor@x.test", PasswordHash: string(hash), Rol: models.RolOdontologo, Activo: true},
		"recep@x.test":  {ID: 11, Email: "recep@x.test", PasswordHash: string(hash), Rol: models.RolRecepcionista, Activo: true},
		"baja@x.test":   {ID: 12, Email: "baja@x.test", PasswordHash: string(hash), Rol: models.RolOdontologo, Activo: false},
	}
	ctrl := NuevoPresupuestosController(repo)
	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("rol", models.RolOdontologo)
		c.Locals("usuario_id", 1)
		return c.Next()
	})
	app.Post("/api/pacientes/:id/presupuestos", ctrl.CrearPresupuesto)
	app.Put("/api/pacientes/:id/presupuestos/:pid", ctrl.ActualizarPresupuesto)
	return app
}

const itemSinDescuento = `{"fase":1,"sesion":1,"procedimiento_id":3,"descuento_centavos":0}`

// Un descuento sin autorización válida se rechaza; con la de un rol habilitado se acepta y queda quién autorizó.
func TestDescuentosRequierenAutorizacionDeUnRolHabilitado(t *testing.T) {
	descuento := `{"fase":1,"sesion":1,"procedimiento_id":3,"honorario_centavos":10000,"descuento_centavos":1500}`
	casos := []struct {
		nombre, cuerpo string
		estado         int
		autoriza       int
	}{
		{"sin descuentos no hace falta autorizar", `{"items":[` + itemSinDescuento + `]}`, 201, 0},
		{"descuento sin autorización", `{"items":[` + descuento + `]}`, 422, 0},
		{"autoriza un odontólogo", `{"items":[` + descuento + `],"autorizacion":{"email":"doctor@x.test","password":"clave"}}`, 201, 10},
		{"contraseña de quien autoriza incorrecta", `{"items":[` + descuento + `],"autorizacion":{"email":"doctor@x.test","password":"mala"}}`, 403, 0},
		{"rol no habilitado (recepción)", `{"items":[` + descuento + `],"autorizacion":{"email":"recep@x.test","password":"clave"}}`, 403, 0},
		{"cuenta desactivada", `{"items":[` + descuento + `],"autorizacion":{"email":"baja@x.test","password":"clave"}}`, 403, 0},
		{"correo inexistente", `{"items":[` + descuento + `],"autorizacion":{"email":"nadie@x.test","password":"clave"}}`, 403, 0},
		{"descuento mayor que el ítem", `{"items":[{"fase":1,"sesion":1,"procedimiento_id":3,"honorario_centavos":1000,"descuento_centavos":1001}],"autorizacion":{"email":"doctor@x.test","password":"clave"}}`, 400, 0},
		{"importe negativo", `{"items":[{"fase":1,"sesion":1,"procedimiento_id":3,"laboratorio_centavos":-5}]}`, 400, 0},
		{"fase fuera de rango", `{"items":[{"fase":0,"sesion":1,"procedimiento_id":3}]}`, 400, 0},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			repo := &repoPresupuestos{}
			app := appPresupuestos(t, repo)
			if estado := enviar(t, app, "POST", "/api/pacientes/4/presupuestos", c.cuerpo); estado != c.estado {
				t.Fatalf("estado = %d, se esperaba %d", estado, c.estado)
			}
			if c.autoriza != 0 && (len(repo.guardados) != 1 || repo.guardados[0].DescuentoAutorizadoPorID == nil || *repo.guardados[0].DescuentoAutorizadoPorID != c.autoriza) {
				t.Errorf("no quedó registrado quién autorizó: %+v", repo.guardados)
			}
			if c.estado != 201 && len(repo.guardados) != 0 {
				t.Error("se guardó un presupuesto rechazado")
			}
		})
	}
}

// Al editar un borrador, un descuento que no cambió conserva su autorización sin pedirla otra vez; uno modificado la pide.
func TestUnDescuentoSinCambiosConservaSuAutorizacion(t *testing.T) {
	autorizador := 10
	previo := models.Presupuesto{ID: 5, Estado: models.PresupBorrador, Items: []models.ItemPresupuesto{
		{ID: 21, DescuentoCentavos: 1500, DescuentoAutorizadoPorID: &autorizador},
	}}
	repo := &repoPresupuestos{previo: previo}
	app := appPresupuestos(t, repo)
	igual := `{"items":[{"id":21,"fase":1,"sesion":1,"procedimiento_id":3,"honorario_centavos":10000,"descuento_centavos":1500}]}`
	if estado := enviar(t, app, "PUT", "/api/pacientes/4/presupuestos/5", igual); estado != 200 {
		t.Fatalf("sin cambios: estado = %d, se esperaba 200", estado)
	}
	if len(repo.guardados) != 1 || repo.guardados[0].DescuentoAutorizadoPorID == nil || *repo.guardados[0].DescuentoAutorizadoPorID != 10 {
		t.Errorf("no se conservó la autorización: %+v", repo.guardados)
	}
	mayor := `{"items":[{"id":21,"fase":1,"sesion":1,"procedimiento_id":3,"honorario_centavos":10000,"descuento_centavos":3000}]}`
	if estado := enviar(t, app, "PUT", "/api/pacientes/4/presupuestos/5", mayor); estado != 422 {
		t.Errorf("descuento modificado sin autorizar: estado = %d, se esperaba 422", estado)
	}
}

func TestSoloUnBorradorSeEdita(t *testing.T) {
	for _, estado := range []string{models.PresupPresentado, models.PresupAprobadoTotal, models.PresupAprobadoParcial, models.PresupRechazado} {
		repo := &repoPresupuestos{previo: models.Presupuesto{ID: 5, Estado: estado}}
		app := appPresupuestos(t, repo)
		if got := enviar(t, app, "PUT", "/api/pacientes/4/presupuestos/5", `{"items":[`+itemSinDescuento+`]}`); got != 409 {
			t.Errorf("presupuesto %s: estado = %d, se esperaba 409", estado, got)
		}
	}
}
