// Este archivo reúne utilidades compartidas para leer parámetros del sistema desde los controladores.
package controllers

import (
	"context"
	"strconv"
	"strings"
)

// LectorParametros es lo mínimo que un controlador necesita para leer un parámetro de la tabla parametros.
type LectorParametros interface {
	ObtenerParametro(ctx context.Context, clave string) (string, error)
}

// parametroEntero lee un parámetro entero positivo; si falta o no es válido devuelve el valor por defecto.
func parametroEntero(ctx context.Context, lector LectorParametros, clave string, defecto int) int {
	valor, err := lector.ObtenerParametro(ctx, clave)
	if err != nil {
		return defecto
	}
	n, err := strconv.Atoi(strings.TrimSpace(valor))
	if err != nil || n <= 0 {
		return defecto
	}
	return n
}

// parametroTexto lee un parámetro de texto; si falta devuelve el valor por defecto.
func parametroTexto(ctx context.Context, lector LectorParametros, clave, defecto string) string {
	valor, err := lector.ObtenerParametro(ctx, clave)
	if err != nil {
		return defecto
	}
	return strings.TrimSpace(valor)
}
