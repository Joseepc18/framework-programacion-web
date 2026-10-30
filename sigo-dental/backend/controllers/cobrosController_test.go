// Pruebas de cobros con un repositorio falso. La numeración sin saltos ni repetidos y el límite del saldo bajo
// solicitudes simultáneas dependen de bloqueos de PostgreSQL y se comprueban contra la base real.
package controllers

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"sigo-dental-backend/models"
)

type repoCobros struct {
	CobroRepository
	err    error
	creado models.NuevoCobro
	pago   struct {
		metodo, referencia string
		monto              int
	}
	motivo string
}

func (r *repoCobros) CrearCobro(_ context.Context, n models.NuevoCobro, _ int) (int, error) {
	r.creado = n
	return 5, r.err
}

func (r *repoCobros) RegistrarPago(_ context.Context, _ int, metodo string, monto int, referencia string, _ int) error {
	r.pago.metodo, r.pago.monto, r.pago.referencia = metodo, monto, referencia
	return r.err
}

func (r *repoCobros) AnularCobro(_ context.Context, _ int, motivo string, _ int) error {
	r.motivo = motivo
	return r.err
}

func (r *repoCobros) ObtenerCobro(_ context.Context, id int) (models.Cobro, error) {
	return models.Cobro{CobroResumen: models.CobroResumen{ID: id}}, nil
}

func appCobros(repo *repoCobros) *fiber.App {
	ctrl := NuevoCobrosController(repo)
	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("rol", models.RolRecepcionista)
		c.Locals("usuario_id", 1)
		return c.Next()
	})
	app.Post("/api/cobros", ctrl.CrearCobro)
	app.Post("/api/cobros/:id/pagos", ctrl.RegistrarPago)
	app.Post("/api/cobros/:id/anular", ctrl.AnularCobro)
	return app
}

func TestCrearCobro(t *testing.T) {
	casos := []struct {
		nombre, cuerpo string
		err            error
		estado         int
	}{
		{"cobro a un paciente", `{"paciente_id":4,"lineas":[{"procedimiento_id":1,"cantidad":2}]}`, nil, 201},
		{"cobro desde una cita atendida", `{"cita_id":9,"lineas":[{"procedimiento_id":1,"cantidad":1}]}`, nil, 201},
		{"con descuento", `{"paciente_id":4,"descuento_bp":1500,"lineas":[{"procedimiento_id":1,"cantidad":1}]}`, nil, 201},
		{"sin paciente ni cita", `{"lineas":[{"procedimiento_id":1,"cantidad":1}]}`, nil, 400},
		{"cita inválida", `{"cita_id":0,"lineas":[{"procedimiento_id":1,"cantidad":1}]}`, nil, 400},
		{"sin líneas", `{"paciente_id":4,"lineas":[]}`, nil, 400},
		{"cantidad cero", `{"paciente_id":4,"lineas":[{"procedimiento_id":1,"cantidad":0}]}`, nil, 400},
		{"cantidad negativa", `{"paciente_id":4,"lineas":[{"procedimiento_id":1,"cantidad":-1}]}`, nil, 400},
		{"cantidad exagerada", `{"paciente_id":4,"lineas":[{"procedimiento_id":1,"cantidad":1000}]}`, nil, 400},
		{"procedimiento sin id", `{"paciente_id":4,"lineas":[{"cantidad":1}]}`, nil, 400},
		{"descuento mayor a 100 %", `{"paciente_id":4,"descuento_bp":10001,"lineas":[{"procedimiento_id":1,"cantidad":1}]}`, nil, 400},
		{"descuento negativo", `{"paciente_id":4,"descuento_bp":-1,"lineas":[{"procedimiento_id":1,"cantidad":1}]}`, nil, 400},
		{"procedimiento desactivado", `{"paciente_id":4,"lineas":[{"procedimiento_id":1,"cantidad":1}]}`, models.ErrReferencia, 400},
		{"cita sin atender", `{"cita_id":9,"lineas":[{"procedimiento_id":1,"cantidad":1}]}`, models.ErrCitaNoCobrable, 409},
		{"cita ya cobrada", `{"cita_id":9,"lineas":[{"procedimiento_id":1,"cantidad":1}]}`, models.ErrCitaYaCobrada, 409},
		{"cuerpo que no es JSON", `no-es-json`, nil, 400},
		{"falla la base de datos", `{"paciente_id":4,"lineas":[{"procedimiento_id":1,"cantidad":1}]}`, errors.New("conexión rechazada"), 500},
	}
	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			app := appCobros(&repoCobros{err: caso.err})
			if estado := enviar(t, app, "POST", "/api/cobros", caso.cuerpo); estado != caso.estado {
				t.Errorf("estado = %d, se esperaba %d", estado, caso.estado)
			}
		})
	}
}

// El cliente no puede fijar precios ni totales: aunque los envíe, el controlador solo pasa procedimiento y cantidad.
func TestCrearCobroIgnoraImportesDelCliente(t *testing.T) {
	repo := &repoCobros{}
	app := appCobros(repo)
	cuerpo := `{"paciente_id":4,"total_centavos":1,"subtotal_centavos":1,"descuento_centavos":99999,` +
		`"lineas":[{"procedimiento_id":3,"cantidad":2,"precio_unitario_centavos":1,"total_centavos":1}]}`
	if estado := enviar(t, app, "POST", "/api/cobros", cuerpo); estado != 201 {
		t.Fatalf("estado = %d", estado)
	}
	if len(repo.creado.Lineas) != 1 || repo.creado.Lineas[0] != (models.LineaPedida{ProcedimientoID: 3, Cantidad: 2}) {
		t.Errorf("líneas pasadas al repositorio = %+v", repo.creado.Lineas)
	}
}

func TestRegistrarPago(t *testing.T) {
	casos := []struct {
		nombre, cuerpo string
		err            error
		estado         int
	}{
		{"efectivo sin referencia", `{"metodo":"efectivo","monto_centavos":5000}`, nil, 201},
		{"tarjeta con referencia", `{"metodo":"tarjeta","monto_centavos":5000,"referencia":"AUT-123"}`, nil, 201},
		{"transferencia con referencia", `{"metodo":"transferencia","monto_centavos":5000,"referencia":"TRX-9"}`, nil, 201},
		{"tarjeta sin referencia", `{"metodo":"tarjeta","monto_centavos":5000}`, nil, 400},
		{"transferencia con referencia en blanco", `{"metodo":"transferencia","monto_centavos":5000,"referencia":"   "}`, nil, 400},
		{"método desconocido", `{"metodo":"cheque","monto_centavos":5000}`, nil, 400},
		{"monto cero", `{"metodo":"efectivo","monto_centavos":0}`, nil, 400},
		{"monto negativo", `{"metodo":"efectivo","monto_centavos":-5}`, nil, 400},
		{"monto absurdo", `{"metodo":"efectivo","monto_centavos":100000001}`, nil, 400},
		{"referencia demasiado larga", `{"metodo":"tarjeta","monto_centavos":5,"referencia":"` + strings.Repeat("a", 101) + `"}`, nil, 400},
		{"supera el saldo", `{"metodo":"efectivo","monto_centavos":5000}`, models.ErrExcedeSaldo, 409},
		{"cobro anulado", `{"metodo":"efectivo","monto_centavos":5000}`, models.ErrCobroAnulado, 409},
		{"cobro inexistente", `{"metodo":"efectivo","monto_centavos":5000}`, models.ErrNoEncontrado, 404},
	}
	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			app := appCobros(&repoCobros{err: caso.err})
			if estado := enviar(t, app, "POST", "/api/cobros/5/pagos", caso.cuerpo); estado != caso.estado {
				t.Errorf("estado = %d, se esperaba %d", estado, caso.estado)
			}
		})
	}
}

func TestAnularCobro(t *testing.T) {
	casos := []struct {
		nombre, cuerpo string
		err            error
		estado         int
	}{
		{"con motivo", `{"motivo":"Error en el precio"}`, nil, 200},
		{"sin motivo", `{"motivo":"  "}`, nil, 400},
		{"motivo demasiado largo", `{"motivo":"` + strings.Repeat("x", 301) + `"}`, nil, 400},
		{"ya estaba anulado", `{"motivo":"otra vez"}`, models.ErrCobroAnulado, 409},
		{"cobro inexistente", `{"motivo":"x"}`, models.ErrNoEncontrado, 404},
	}
	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			app := appCobros(&repoCobros{err: caso.err})
			if estado := enviar(t, app, "POST", "/api/cobros/5/anular", caso.cuerpo); estado != caso.estado {
				t.Errorf("estado = %d, se esperaba %d", estado, caso.estado)
			}
		})
	}
}

func TestElPagoGuardaLaReferenciaLimpia(t *testing.T) {
	repo := &repoCobros{}
	app := appCobros(repo)
	enviar(t, app, "POST", "/api/cobros/5/pagos", `{"metodo":"tarjeta","monto_centavos":2500,"referencia":"  AUT-1  "}`)
	if repo.pago.metodo != "tarjeta" || repo.pago.monto != 2500 || repo.pago.referencia != "AUT-1" {
		t.Errorf("pago guardado = %+v", repo.pago)
	}
}
