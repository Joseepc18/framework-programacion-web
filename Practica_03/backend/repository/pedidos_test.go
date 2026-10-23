package repository

import (
	"testing"
	"time"
)

func TestNumeroDePedido(t *testing.T) {
	fecha := time.Date(2026, time.October, 23, 16, 24, 32, 0, time.UTC)

	if got, want := numeroDePedido(fecha, 7), "MC-261023-7"; got != want {
		t.Fatalf("numeroDePedido() = %q, se esperaba %q", got, want)
	}
}

func TestNumeroDePedidoNoSeRepiteEnElMismoInstante(t *testing.T) {
	fecha := time.Date(2026, time.October, 23, 16, 24, 32, 0, time.UTC)

	if numeroDePedido(fecha, 7) == numeroDePedido(fecha, 8) {
		t.Fatal("dos pedidos creados en el mismo instante no pueden compartir número")
	}
}

func TestNumeroDePedidoCabeEnLaColumna(t *testing.T) {
	// pedidos.numero es VARCHAR(20): debe alcanzar para millones de pedidos.
	fecha := time.Date(2026, time.October, 23, 0, 0, 0, 0, time.UTC)

	if largo := len(numeroDePedido(fecha, 9999999)); largo > 20 {
		t.Fatalf("el número mide %d caracteres y la columna admite 20", largo)
	}
}
