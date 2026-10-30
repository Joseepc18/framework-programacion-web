package models

import (
	"slices"
	"testing"
)

func TestHayExactamente32PermanentesY20Temporales(t *testing.T) {
	permanentes, temporales := 0, 0
	for pieza := 0; pieza <= 99; pieza++ {
		if !PiezaFDIValida(pieza) {
			continue
		}
		if PiezaPermanente(pieza) {
			permanentes++
		} else {
			temporales++
		}
	}
	if permanentes != 32 || temporales != 20 {
		t.Errorf("permanentes = %d, temporales = %d; se esperaban 32 y 20", permanentes, temporales)
	}
	for _, invalida := range []int{0, 10, 19, 29, 49, 56, 66, 76, 86, 90, 100, -11} {
		if PiezaFDIValida(invalida) {
			t.Errorf("la pieza %d no debería ser válida", invalida)
		}
	}
}

func TestNumeracionUniversalCoincideConLaTablaEstandar(t *testing.T) {
	casos := map[int]string{
		18: "1", 11: "8", 21: "9", 28: "16", 38: "17", 31: "24", 41: "25", 48: "32",
		55: "A", 51: "E", 61: "F", 65: "J", 75: "K", 71: "O", 81: "P", 85: "T",
	}
	for fdi, universal := range casos {
		if got := NumeroUniversal(fdi); got != universal {
			t.Errorf("NumeroUniversal(%d) = %q, se esperaba %q", fdi, got, universal)
		}
	}
	// La conversión es biyectiva: ningún número universal se repite.
	vistos := map[string]int{}
	for pieza := 11; pieza <= 85; pieza++ {
		if !PiezaFDIValida(pieza) {
			continue
		}
		u := NumeroUniversal(pieza)
		if previa, repetido := vistos[u]; repetido {
			t.Errorf("las piezas %d y %d comparten el número universal %s", previa, pieza, u)
		}
		vistos[u] = pieza
	}
	if len(vistos) != 52 {
		t.Errorf("hay %d números universales, se esperaban 52", len(vistos))
	}
	if NumeroUniversal(99) != "" {
		t.Error("una pieza inválida debe dar cadena vacía")
	}
}

func TestValidarHallazgo(t *testing.T) {
	casos := []struct {
		nombre                   string
		pieza                    int
		superficie, tipo, estado string
		valido                   bool
	}{
		{"caries en oclusal", 16, "oclusal", "caries", "requerido", true},
		{"restauración existente en mesial", 21, "mesial", "restauracion", "existente", true},
		{"endodoncia realizada en la pieza completa", 36, "", "endodoncia", "realizado", true},
		{"implante requerido", 46, "", "implante", "requerido", true},
		{"exodoncia realizada", 48, "", "exodoncia", "realizado", true},
		{"ausente en un temporal", 55, "", "ausente", "existente", true},
		{"sellante en oclusal", 26, "oclusal", "sellante", "realizado", true},
		{"prótesis removible existente", 14, "", "protesis_removible", "existente", true},
		{"caries sin superficie", 16, "", "caries", "requerido", false},
		{"caries ya realizada", 16, "oclusal", "caries", "realizado", false},
		{"endodoncia con superficie", 36, "oclusal", "endodoncia", "requerido", false},
		{"superficie inexistente", 16, "palatino", "caries", "requerido", false},
		{"pieza inexistente", 19, "", "ausente", "existente", false},
		{"tipo inexistente", 16, "", "puente", "requerido", false},
		{"estado inexistente", 16, "oclusal", "caries", "curado", false},
		{"ausente requerido", 16, "", "ausente", "requerido", false},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			motivo := ValidarHallazgo(c.pieza, c.superficie, c.tipo, c.estado)
			if (motivo == "") != c.valido {
				t.Errorf("motivo = %q, válido esperado = %v", motivo, c.valido)
			}
		})
	}
}

func TestTiposDeHallazgoSonLosNueveDelCatalogo(t *testing.T) {
	tipos := TiposDeHallazgo()
	for _, esperado := range []string{"caries", "restauracion", "endodoncia", "implante", "protesis_fija", "protesis_removible", "exodoncia", "sellante", "ausente"} {
		if !slices.Contains(tipos, esperado) {
			t.Errorf("falta el tipo %s", esperado)
		}
	}
	if len(tipos) != 9 {
		t.Errorf("hay %d tipos, se esperaban 9", len(tipos))
	}
}
