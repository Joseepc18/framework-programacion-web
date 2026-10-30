package models

import "testing"

func TestCalcularTotalesRedondeaAlCentavo(t *testing.T) {
	casos := []struct {
		nombre                         string
		subtotal, bp, descuento, total int
	}{
		{"sin descuento", 12345, 0, 0, 12345},
		{"10 % exacto", 10000, 1000, 1000, 9000},
		{"100 % deja el total en cero", 5000, 10000, 5000, 0},
		{"cuarto de centavo se queda abajo", 1005, 500, 50, 955},
		{"mitad exacta redondea hacia arriba", 1010, 500, 51, 959},
		{"tercio se redondea al centavo", 1000, 3333, 333, 667},              // 333,3 -> 333
		{"7,5 % de 1,99", 199, 750, 15, 184},                                 // 14,925 -> 15
		{"0,01 % de un importe pequeño no genera descuento", 100, 1, 0, 100}, // 0,01 -> 0
		{"importe grande sin desbordar", 100_000_000, 2550, 25_500_000, 74_500_000},
	}
	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			descuento, total := CalcularTotales(caso.subtotal, caso.bp)
			if descuento != caso.descuento || total != caso.total {
				t.Errorf("CalcularTotales(%d, %d) = %d, %d; se esperaba %d, %d",
					caso.subtotal, caso.bp, descuento, total, caso.descuento, caso.total)
			}
			if descuento+total != caso.subtotal {
				t.Errorf("descuento + total = %d, debe sumar el subtotal %d", descuento+total, caso.subtotal)
			}
		})
	}
}

func TestEstadoDePago(t *testing.T) {
	casos := []struct {
		total, pagado int
		estado        string
	}{
		{10000, 0, CobroPendiente},
		{10000, 1, CobroParcial},
		{10000, 9999, CobroParcial},
		{10000, 10000, CobroPagado},
		{0, 0, CobroPagado},
	}
	for _, caso := range casos {
		if got := EstadoDePago(caso.total, caso.pagado); got != caso.estado {
			t.Errorf("EstadoDePago(%d, %d) = %s, se esperaba %s", caso.total, caso.pagado, got, caso.estado)
		}
	}
}
