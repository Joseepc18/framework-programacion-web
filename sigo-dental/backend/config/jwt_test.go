package config

import (
	"strings"
	"testing"
)

func TestCargarJWTSecret(t *testing.T) {
	casos := []struct {
		nombre string
		valor  string
		falla  bool
	}{
		{"vacía", "", true},
		{"muy corta", strings.Repeat("a", LongitudMinimaJWT-1), true},
		{"longitud mínima", strings.Repeat("a", LongitudMinimaJWT), false},
		{"larga", strings.Repeat("a", 64), false},
	}
	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Setenv("JWT_SECRET", caso.valor)
			secreto, err := CargarJWTSecret()
			if (err != nil) != caso.falla {
				t.Fatalf("error = %v, se esperaba falla=%v", err, caso.falla)
			}
			if !caso.falla && string(secreto) != caso.valor {
				t.Fatal("el secreto devuelto no coincide con JWT_SECRET")
			}
		})
	}
}
