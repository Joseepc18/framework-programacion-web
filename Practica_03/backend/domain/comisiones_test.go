// Pruebas unitarias del reparto de comisiones: son puras, sin base de datos.
package domain

import "testing"

// TestRepartirComisionesTasasPorDefecto verifica los tres niveles sobre $100.
func TestRepartirComisionesTasasPorDefecto(t *testing.T) {
	comisiones := RepartirComisiones(100, TasasUpline)

	if len(comisiones) != 3 {
		t.Fatalf("se esperaban 3 comisiones y llegaron %d", len(comisiones))
	}

	esperado := []struct {
		nivel      int
		porcentaje float64
		monto      float64
	}{
		{1, 0.10, 10},
		{2, 0.05, 5},
		{3, 0.02, 2},
	}

	for i, e := range esperado {
		if comisiones[i].Nivel != e.nivel {
			t.Errorf("comisión %d: nivel %d, se esperaba %d", i, comisiones[i].Nivel, e.nivel)
		}
		if comisiones[i].Porcentaje != e.porcentaje {
			t.Errorf("comisión %d: porcentaje %v, se esperaba %v", i, comisiones[i].Porcentaje, e.porcentaje)
		}
		if comisiones[i].Monto != e.monto {
			t.Errorf("comisión %d: monto %v, se esperaba %v", i, comisiones[i].Monto, e.monto)
		}
	}
}

// TestRepartirComisionesRedondeoAPlata asegura dos decimales con totales nones.
func TestRepartirComisionesRedondeoAPlata(t *testing.T) {
	comisiones := RepartirComisiones(90, TasasUpline)

	esperado := []float64{9, 4.5, 1.8}
	for i, monto := range esperado {
		if comisiones[i].Monto != monto {
			t.Errorf("nivel %d: monto %v, se esperaba %v", i+1, comisiones[i].Monto, monto)
		}
	}
}

// TestRepartirComisionesSinNadaQueRepartir cubre los casos borde del contrato.
func TestRepartirComisionesSinNadaQueRepartir(t *testing.T) {
	casos := []struct {
		nombre string
		total  float64
		tasas  []float64
	}{
		{"total cero", 0, TasasUpline},
		{"total negativo", -10, TasasUpline},
		{"sin tasas", 100, nil},
	}

	for _, c := range casos {
		if got := RepartirComisiones(c.total, c.tasas); len(got) != 0 {
			t.Errorf("%s: se esperaba una lista vacía y llegaron %d filas", c.nombre, len(got))
		}
	}
}
