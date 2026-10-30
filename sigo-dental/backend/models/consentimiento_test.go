package models

import (
	"strings"
	"testing"
)

func TestRellenarPlantillaCompletaLosCamposSolos(t *testing.T) {
	texto := RellenarPlantilla("Yo, {{paciente}} ({{documento}}), autorizo «{{procedimiento}}» {{pieza}} con {{profesional}} el {{fecha}}.",
		DatosPlantilla{Paciente: "Ana Pérez", Documento: "0912345678", Procedimiento: "Exodoncia simple", Pieza: "en la pieza 38", Profesional: "Dr. Soto", Fecha: "30/10/2026"})
	esperado := "Yo, Ana Pérez (0912345678), autorizo «Exodoncia simple» en la pieza 38 con Dr. Soto el 30/10/2026."
	if texto != esperado {
		t.Errorf("texto = %q", texto)
	}
	if strings.Contains(texto, "{{") {
		t.Error("quedaron campos sin completar")
	}
}

func firmaDePrueba(puntos int, ancho float64) FirmaManuscrita {
	trazo := []Punto{}
	for i := 0; i < puntos; i++ {
		trazo = append(trazo, Punto{X: float64(i) * ancho / float64(max(puntos, 1)), Y: 20, T: i * 10})
	}
	return FirmaManuscrita{Ancho: 300, Alto: 120, Trazos: [][]Punto{trazo}}
}

func TestValidarFirmaManuscrita(t *testing.T) {
	casos := []struct {
		nombre string
		firma  FirmaManuscrita
		valida bool
	}{
		{"firma con trazo real", firmaDePrueba(40, 250), true},
		{"sin trazos", FirmaManuscrita{Ancho: 300, Alto: 120}, false},
		{"un solo toque", firmaDePrueba(1, 0), false},
		{"muy pocos puntos", firmaDePrueba(5, 250), false},
		{"trazo diminuto", firmaDePrueba(40, 10), false},
		{"lienzo inválido", FirmaManuscrita{Ancho: 0, Alto: 0, Trazos: firmaDePrueba(40, 250).Trazos}, false},
		{"sale del lienzo", func() FirmaManuscrita { f := firmaDePrueba(40, 250); f.Trazos[0][3].X = 999; return f }(), false},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			if motivo := ValidarFirmaManuscrita(c.firma); (motivo == "") != c.valida {
				t.Errorf("motivo = %q, válida esperada = %v", motivo, c.valida)
			}
		})
	}
}
