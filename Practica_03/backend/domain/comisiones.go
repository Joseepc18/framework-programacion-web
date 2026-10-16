// Este archivo concentra el reparto de comisiones multinivel (Fase 4).
// Es código puro sin I/O: ni base de datos, ni caché, ni peticiones HTTP,
// de modo que RepartirComisiones sea verificable con pruebas unitarias.
package domain

// Importamos math para redondear cada monto a dos decimales (plata).
import "math"

// TasasUpline define el porcentaje de cada nivel del patrocinador:
// nivel 1 (directo) 10 %, nivel 2 5 % y nivel 3 2 %.
var TasasUpline = []float64{0.10, 0.05, 0.02}

// Comision es el resultado del reparto para un nivel concreto del upline.
// El repositorio agrega después el referido_id real de cada nivel.
type Comision struct {
	// Nivel es la profundidad dentro de la red (1 = patrocinador directo).
	Nivel int `json:"nivel"`
	// Porcentaje es la tasa aplicada sobre el total del pedido.
	Porcentaje float64 `json:"porcentaje"`
	// Monto es la plata acreditada al referido de ese nivel.
	Monto float64 `json:"monto"`
}

// aPlata redondea un valor a dos decimales, el máximo que admite NUMERIC(10,2).
func aPlata(valor float64) float64 {
	return math.Round(valor*100) / 100
}

// RepartirComisiones devuelve una comisión por cada tasa recibida,
// calculada sobre el total del pedido. No tiene efectos secundarios.
//
// Ejemplo: RepartirComisiones(100, TasasUpline) → 10, 5 y 2.
// Devuelve una lista vacía si no hay nada que repartir.
func RepartirComisiones(total float64, tasas []float64) []Comision {
	if total <= 0 || len(tasas) == 0 {
		return []Comision{}
	}

	comisiones := make([]Comision, 0, len(tasas))
	for i, tasa := range tasas {
		if tasa <= 0 {
			continue
		}
		comisiones = append(comisiones, Comision{
			Nivel:      i + 1,
			Porcentaje: tasa,
			Monto:      aPlata(total * tasa),
		})
	}
	return comisiones
}
