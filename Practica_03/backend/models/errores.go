// Este archivo agrupa los errores de dominio compartidos entre repositorio y controladores.
// Viven en models para que los controladores los reconozcan sin importar la capa de datos.
package models

import (
	"errors"
	"time"
)

// Errores de dominio del alta de cuentas.
var (
	ErrEmailYaExiste     = errors.New("el correo electrónico ya está registrado")
	ErrCodigoInvalido    = errors.New("el código de referido no existe")
	ErrRegistroInvalido  = errors.New("nombre, correo y contraseña son obligatorios")
	ErrCuentaDesactivada = errors.New("cuenta desactivada")
)

// Errores de dominio del checkout.
var (
	ErrStockInsuficiente = errors.New("stock insuficiente")
	ErrCantidadInvalida  = errors.New("cantidad debe ser mayor que cero")
	ErrCarritoVacio      = errors.New("el carrito no contiene artículos")
)

// Errores de dominio de pedidos y liquidaciones, traducibles a HTTP.
var (
	// ErrPedidoNoEncontrado indica que el pedido no existe o no pertenece al token.
	ErrPedidoNoEncontrado = errors.New("pedido no encontrado")
	// ErrTransicionInvalida indica un salto de estado fuera del ciclo de vida.
	ErrTransicionInvalida = errors.New("transición de estado inválida")
	// ErrSinComisionesPendientes indica que el distribuidor no debe nada.
	ErrSinComisionesPendientes = errors.New("el distribuidor no tiene comisiones pendientes")
)

// RangoFechas delimita el periodo [Desde, Hasta) de un reporte; un límite nil significa sin límite.
type RangoFechas struct {
	Desde *time.Time
	Hasta *time.Time
}
