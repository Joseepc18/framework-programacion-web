// Este archivo atiende los cobros y comprobantes internos (F1-5: parte básica de RF-12).
// La recepción cobra y registra pagos; la recepción y el auditor consultan. El router fija esos permisos.
// Los importes los calcula siempre el servidor: el cliente solo dice qué procedimientos, cuántos y el descuento.
package controllers

import (
	"context"
	"errors"
	"slices"

	"github.com/gofiber/fiber/v2"
	"sigo-dental-backend/models"
)

// CobroRepository es el contrato de la capa de datos de la facturación.
type CobroRepository interface {
	ListarCobros(ctx context.Context, busqueda string, pacienteID int) ([]models.CobroResumen, error)
	ObtenerCobro(ctx context.Context, id int) (models.Cobro, error)
	// CrearCobro calcula importes y numeración y devuelve el ID del cobro.
	CrearCobro(ctx context.Context, cobro models.NuevoCobro, usuarioID int) (int, error)
	RegistrarPago(ctx context.Context, cobroID int, metodo string, monto int, referencia string, usuarioID int) error
	AnularCobro(ctx context.Context, cobroID int, motivo string, usuarioID int) error
}

// CobrosController atiende las rutas de cobros con el repositorio inyectado.
type CobrosController struct {
	repo CobroRepository
}

// NuevoCobrosController construye el controlador de cobros.
func NuevoCobrosController(repo CobroRepository) *CobrosController {
	return &CobrosController{repo: repo}
}

const (
	maxLineasCobro   = 30
	maxCantidadLinea = 999
	// maxPagoCentavos evita importes absurdos por un error de captura (1.000.000,00).
	maxPagoCentavos  = 100_000_000
	maxReferencia    = 100
	maxMotivoAnulado = 300
)

// errorCobro traduce los errores propios de la facturación; el resto sigue el camino común.
func errorCobro(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, models.ErrCobroAnulado):
		return errorCliente(c, fiber.StatusConflict, "El cobro está anulado y no admite cambios")
	case errors.Is(err, models.ErrExcedeSaldo):
		return errorCliente(c, fiber.StatusConflict, "El pago supera el saldo pendiente del cobro")
	case errors.Is(err, models.ErrCitaNoCobrable):
		return errorCliente(c, fiber.StatusConflict, "Solo se puede cobrar una cita en estado Atendido")
	case errors.Is(err, models.ErrCitaYaCobrada):
		return errorCliente(c, fiber.StatusConflict, "La cita ya tiene un cobro; anúlalo para generar otro")
	}
	return errorDominio(c, err)
}

// ListarCobros responde GET /api/cobros?q=texto&paciente_id=N.
func (ctrl *CobrosController) ListarCobros(c *fiber.Ctx) error {
	lista, err := ctrl.repo.ListarCobros(c.Context(), c.Query("q"), c.QueryInt("paciente_id"))
	if err != nil {
		return errorDominio(c, err)
	}
	return c.JSON(lista)
}

// ObtenerCobro responde GET /api/cobros/:id con líneas, pagos y anulación.
func (ctrl *CobrosController) ObtenerCobro(c *fiber.Ctx) error {
	id, ok := idDeRuta(c)
	if !ok {
		return errorCliente(c, fiber.StatusBadRequest, "Identificador inválido")
	}
	cobro, err := ctrl.repo.ObtenerCobro(c.Context(), id)
	if err != nil {
		return errorDominio(c, err)
	}
	return c.JSON(cobro)
}

// CrearCobro responde POST /api/cobros. Se cobra a un paciente o, mejor, desde una cita Atendida (cita_id), en
// cuyo caso el paciente sale de la cita. Un cobro no se modifica después: para corregirlo se anula y se crea otro.
func (ctrl *CobrosController) CrearCobro(c *fiber.Ctx) error {
	var cuerpo struct {
		PacienteID  int                  `json:"paciente_id"`
		CitaID      *int                 `json:"cita_id"`
		DescuentoBP int                  `json:"descuento_bp"`
		Lineas      []models.LineaPedida `json:"lineas"`
	}
	if err := c.BodyParser(&cuerpo); err != nil {
		return cuerpoInvalido(c)
	}
	if cuerpo.CitaID != nil && *cuerpo.CitaID <= 0 {
		return errorCliente(c, fiber.StatusBadRequest, "La cita no es válida")
	}
	if cuerpo.CitaID == nil && cuerpo.PacienteID <= 0 {
		return errorCliente(c, fiber.StatusBadRequest, "Indica el paciente o la cita que se cobra")
	}
	if cuerpo.DescuentoBP < 0 || cuerpo.DescuentoBP > 10000 {
		return errorCliente(c, fiber.StatusBadRequest, "El descuento debe estar entre 0 % y 100 %")
	}
	if len(cuerpo.Lineas) == 0 || len(cuerpo.Lineas) > maxLineasCobro {
		return errorCliente(c, fiber.StatusBadRequest, "El cobro necesita entre 1 y 30 líneas")
	}
	for _, linea := range cuerpo.Lineas {
		if linea.ProcedimientoID <= 0 || linea.Cantidad <= 0 || linea.Cantidad > maxCantidadLinea {
			return errorCliente(c, fiber.StatusBadRequest, "Cada línea necesita un procedimiento y una cantidad entre 1 y 999")
		}
	}

	id, err := ctrl.repo.CrearCobro(c.Context(), models.NuevoCobro{
		PacienteID: cuerpo.PacienteID, CitaID: cuerpo.CitaID, DescuentoBP: cuerpo.DescuentoBP, Lineas: cuerpo.Lineas,
	}, usuarioDeLaPeticion(c))
	if err != nil {
		return errorCobro(c, err)
	}
	cobro, err := ctrl.repo.ObtenerCobro(c.Context(), id)
	if err != nil {
		return errorDominio(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(cobro)
}

// RegistrarPago responde POST /api/cobros/:id/pagos. Admite pagos parciales, pero nunca más que el saldo.
func (ctrl *CobrosController) RegistrarPago(c *fiber.Ctx) error {
	id, ok := idDeRuta(c)
	if !ok {
		return errorCliente(c, fiber.StatusBadRequest, "Identificador inválido")
	}
	var cuerpo struct {
		Metodo        string `json:"metodo"`
		MontoCentavos int    `json:"monto_centavos"`
		Referencia    string `json:"referencia"`
	}
	if err := c.BodyParser(&cuerpo); err != nil {
		return cuerpoInvalido(c)
	}
	if !slices.Contains(models.MetodosDePago, cuerpo.Metodo) {
		return errorCliente(c, fiber.StatusBadRequest, "El método de pago debe ser efectivo, tarjeta o transferencia")
	}
	if cuerpo.MontoCentavos <= 0 || cuerpo.MontoCentavos > maxPagoCentavos {
		return errorCliente(c, fiber.StatusBadRequest, "El monto debe ser mayor que cero")
	}
	referencia, ok := textoOpcional(cuerpo.Referencia, maxReferencia)
	if !ok {
		return errorCliente(c, fiber.StatusBadRequest, "La referencia admite hasta 100 caracteres")
	}
	if cuerpo.Metodo != models.PagoEfectivo && referencia == "" {
		return errorCliente(c, fiber.StatusBadRequest, "Los pagos con tarjeta o transferencia exigen su número de referencia")
	}

	if err := ctrl.repo.RegistrarPago(c.Context(), id, cuerpo.Metodo, cuerpo.MontoCentavos, referencia, usuarioDeLaPeticion(c)); err != nil {
		return errorCobro(c, err)
	}
	cobro, err := ctrl.repo.ObtenerCobro(c.Context(), id)
	if err != nil {
		return errorDominio(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(cobro)
}

// AnularCobro responde POST /api/cobros/:id/anular con el motivo. Es la única forma de corregir un cobro.
func (ctrl *CobrosController) AnularCobro(c *fiber.Ctx) error {
	id, ok := idDeRuta(c)
	if !ok {
		return errorCliente(c, fiber.StatusBadRequest, "Identificador inválido")
	}
	var cuerpo struct {
		Motivo string `json:"motivo"`
	}
	if err := c.BodyParser(&cuerpo); err != nil {
		return cuerpoInvalido(c)
	}
	motivo, ok := textoObligatorio(cuerpo.Motivo, maxMotivoAnulado)
	if !ok {
		return errorCliente(c, fiber.StatusBadRequest, "El motivo de la anulación es obligatorio (hasta 300 caracteres)")
	}
	if err := ctrl.repo.AnularCobro(c.Context(), id, motivo, usuarioDeLaPeticion(c)); err != nil {
		return errorCobro(c, err)
	}
	cobro, err := ctrl.repo.ObtenerCobro(c.Context(), id)
	if err != nil {
		return errorDominio(c, err)
	}
	return c.JSON(cobro)
}
