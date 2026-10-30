package models

import "testing"

// Casos del NIC calculados a mano con la convención del plan: NIC = profundidad + margen, margen positivo si hay recesión.
func TestNICConValoresCalculadosAMano(t *testing.T) {
	casos := []struct {
		nombre                        string
		profundidad, margen, esperado int
	}{
		{"encía sana al nivel de la unión cemento-esmalte", 3, 0, 3},
		{"recesión de 2 mm con surco de 3 mm", 3, 2, 5},
		{"agrandamiento: margen 1 mm coronal", 4, -1, 3},
		{"sin surco ni recesión", 0, 0, 0},
		{"bolsa profunda con recesión", 9, 4, 13},
		{"máximos", 15, 15, 30},
	}
	for _, c := range casos {
		if got := NIC(c.profundidad, c.margen); got != c.esperado {
			t.Errorf("%s: NIC(%d, %d) = %d, se esperaba %d", c.nombre, c.profundidad, c.margen, got, c.esperado)
		}
	}
}

func dienteValido(pieza int) DientePerio {
	d := DientePerio{Pieza: pieza}
	for _, s := range SitiosPeriodontales {
		d.Sitios = append(d.Sitios, Sitio{Sitio: s, Profundidad: 3, Margen: 0})
	}
	return d
}

func TestValidarDientePerio(t *testing.T) {
	con := func(cambio func(*DientePerio)) DientePerio { d := dienteValido(16); cambio(&d); return d }
	casos := []struct {
		nombre string
		diente DientePerio
		valido bool
	}{
		{"diente completo", dienteValido(16), true},
		{"molar con furca grado II y movilidad 1", con(func(d *DientePerio) { d.Furca, d.Movilidad = 2, 1 }), true},
		{"profundidad negativa", con(func(d *DientePerio) { d.Sitios[0].Profundidad = -1 }), false},
		{"profundidad mayor que 15 mm", con(func(d *DientePerio) { d.Sitios[2].Profundidad = 16 }), false},
		{"margen fuera de rango", con(func(d *DientePerio) { d.Sitios[1].Margen = 16 }), false},
		{"NIC negativo", con(func(d *DientePerio) { d.Sitios[1].Profundidad, d.Sitios[1].Margen = 2, -4 }), false},
		{"movilidad 4", con(func(d *DientePerio) { d.Movilidad = 4 }), false},
		{"furca grado IV", con(func(d *DientePerio) { d.Furca = 4 }), false},
		{"furca en un incisivo", func() DientePerio { d := dienteValido(11); d.Furca = 1; return d }(), false},
		{"faltan puntos", con(func(d *DientePerio) { d.Sitios = d.Sitios[:5] }), false},
		{"punto repetido", con(func(d *DientePerio) { d.Sitios[5].Sitio = "DV" }), false},
		{"punto inexistente", con(func(d *DientePerio) { d.Sitios[5].Sitio = "XX" }), false},
		{"diente temporal", dienteValido(55), false},
		{"pieza inexistente", dienteValido(19), false},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			if motivo := ValidarDientePerio(c.diente); (motivo == "") != c.valido {
				t.Errorf("motivo = %q, válido esperado = %v", motivo, c.valido)
			}
		})
	}
}
