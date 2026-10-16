// Este archivo carga la clave de firma de los tokens JWT desde el entorno.
package config

import (
	// Importamos errors para avisar claramente cuando falta la clave.
	"errors"
	// Importamos os para leer las variables de entorno.
	"os"
)

// secretoDesarrollo solo se usa con APP_ENV=dev, para no obligar a configurar una clave en local.
const secretoDesarrollo = "multicatalogo-solo-desarrollo"

// CargarJWTSecret devuelve la clave de firma (JWT_SECRET). Sin ella la API no arranca,
// salvo en desarrollo (APP_ENV=dev), donde se usa una clave de desarrollo fija.
// Así un despliegue olvidado nunca queda firmando con una clave conocida.
func CargarJWTSecret() ([]byte, error) {
	if secreto := os.Getenv("JWT_SECRET"); secreto != "" {
		return []byte(secreto), nil
	}
	if os.Getenv("APP_ENV") == "dev" {
		return []byte(secretoDesarrollo), nil
	}
	return nil, errors.New("falta la variable JWT_SECRET (defínala en .env, o use APP_ENV=dev solo en desarrollo)")
}
