// Este archivo carga la clave de firma de los tokens JWT desde el entorno.
package config

import (
	"errors"
	"fmt"
	"os"
)

// LongitudMinimaJWT son los bytes mínimos de JWT_SECRET: HS256 firma con 256 bits.
const LongitudMinimaJWT = 32

// CargarJWTSecret devuelve la clave de firma (JWT_SECRET). Sin ella, o si es demasiado corta,
// la API no arranca: así un despliegue olvidado nunca firma con una clave vacía o débil.
func CargarJWTSecret() ([]byte, error) {
	secreto := os.Getenv("JWT_SECRET")
	if secreto == "" {
		return nil, errors.New("falta la variable JWT_SECRET (defínala en .env)")
	}
	if len(secreto) < LongitudMinimaJWT {
		return nil, fmt.Errorf("JWT_SECRET es muy corta: requiere al menos %d caracteres", LongitudMinimaJWT)
	}
	return []byte(secreto), nil
}
