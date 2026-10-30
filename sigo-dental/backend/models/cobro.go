package models

import (
	"errors"
	"time"
)

// Métodos de pago aceptados.
const (
	PagoEfectivo      = "efectivo"
	PagoTarjeta       = "tarjeta"
	PagoTransferencia = "transferencia"
)

// Estados de pago de un cobro, calculados a partir de los pagos registrados.
const (
	CobroPendiente = "pendiente"
	CobroParcial   = "parcial"
	CobroPagado    = "pagado"
)

// MetodosDePago es la lista cerrada de métodos válidos.
var MetodosDePago = []string{PagoEfectivo, PagoTarjeta, PagoTransferencia}

// LineaPedida es una línea de cobro tal como la envía el cliente: el precio nunca viaja, lo pone el servidor.
type LineaPedida struct {
	ProcedimientoID int `json:"procedimiento_id"`
	Cantidad        int `json:"cantidad"`
}

// NuevoCobro es la solicitud de cobro ya validada. Los importes los calcula el servidor.
type NuevoCobro struct {
	PacienteID  int
	CitaID      *int
	DescuentoBP int
	Lineas      []LineaPedida
}

// CobroLinea es una línea del cobro con el precio del catálogo vigente al cobrar.
type CobroLinea struct {
	ID                     int    `json:"id"`
	ProcedimientoID        int    `json:"procedimiento_id"`
	Descripcion            string `json:"descripcion"`
	Cantidad               int    `json:"cantidad"`
	PrecioUnitarioCentavos int    `json:"precio_unitario_centavos"`
	TotalCentavos          int    `json:"total_centavos"`
}

// Pago es un abono registrado contra un cobro.
type Pago struct {
	ID            int       `json:"id"`
	Metodo        string    `json:"metodo"`
	MontoCentavos int       `json:"monto_centavos"`
	Referencia    string    `json:"referencia"`
	UsuarioNombre string    `json:"usuario_nombre"`
	CreadoEn      time.Time `json:"creado_en"`
}

// Anulacion es el registro que deja sin efecto un cobro.
type Anulacion struct {
	Motivo        string    `json:"motivo"`
	UsuarioNombre string    `json:"usuario_nombre"`
	CreadoEn      time.Time `json:"creado_en"`
}

// CobroResumen es una fila del listado de cobros.
type CobroResumen struct {
	ID             int       `json:"id"`
	Numero         string    `json:"numero"`
	PacienteID     int       `json:"paciente_id"`
	PacienteNombre string    `json:"paciente_nombre"`
	TotalCentavos  int       `json:"total_centavos"`
	PagadoCentavos int       `json:"pagado_centavos"`
	SaldoCentavos  int       `json:"saldo_centavos"`
	Estado         string    `json:"estado"`
	Anulado        bool      `json:"anulado"`
	CreadoEn       time.Time `json:"creado_en"`
}

// Cobro es el cobro completo con sus líneas, pagos y, si se anuló, el registro de anulación.
type Cobro struct {
	CobroResumen
	CitaID            *int         `json:"cita_id"`
	PacienteDocumento string       `json:"paciente_documento"`
	SubtotalCentavos  int          `json:"subtotal_centavos"`
	DescuentoBP       int          `json:"descuento_bp"`
	DescuentoCentavos int          `json:"descuento_centavos"`
	Moneda            string       `json:"moneda"`
	CreadoPor         string       `json:"creado_por"`
	Lineas            []CobroLinea `json:"lineas"`
	Pagos             []Pago       `json:"pagos"`
	Anulacion         *Anulacion   `json:"anulacion"`
}

// Errores propios de la facturación que la capa de datos devuelve y el controlador traduce a HTTP.
var (
	// ErrCobroAnulado: el cobro ya fue anulado y no admite más cambios.
	ErrCobroAnulado = errors.New("el cobro está anulado")
	// ErrExcedeSaldo: el pago supera lo que falta por cobrar.
	ErrExcedeSaldo = errors.New("el pago excede el saldo del cobro")
	// ErrCitaNoCobrable: la cita no está Atendida.
	ErrCitaNoCobrable = errors.New("la cita no está atendida")
	// ErrCitaYaCobrada: la cita ya tiene un cobro vigente.
	ErrCitaYaCobrada = errors.New("la cita ya tiene un cobro")
)

// CalcularTotales devuelve el descuento y el total en centavos. El descuento va en puntos básicos
// (10000 = 100 %) y se redondea a la mitad hacia arriba al centavo: todo en enteros, sin decimales flotantes.
func CalcularTotales(subtotal, descuentoBP int) (descuento, total int) {
	descuento = int((int64(subtotal)*int64(descuentoBP) + 5000) / 10000)
	return descuento, subtotal - descuento
}

// EstadoDePago deriva el estado de un cobro de lo pagado: con total cero queda pagado desde el inicio.
func EstadoDePago(total, pagado int) string {
	switch {
	case pagado >= total:
		return CobroPagado
	case pagado == 0:
		return CobroPendiente
	}
	return CobroParcial
}
