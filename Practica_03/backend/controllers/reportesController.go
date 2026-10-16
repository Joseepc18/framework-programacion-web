// Este archivo expone los reportes de negocio del admin (Fase 6).
package controllers

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"multicatalogo-backend/models"
	"multicatalogo-backend/repository"
)

// rangoDeQuery lee ?desde y ?hasta (YYYY-MM-DD). El rango es inclusivo en ambos extremos.
// Devuelve false si alguna fecha viene mal formada.
func rangoDeQuery(c *fiber.Ctx) (repository.RangoFechas, bool) {
	var r repository.RangoFechas
	if v := c.Query("desde"); v != "" {
		t, err := time.ParseInLocation("2006-01-02", v, time.Local)
		if err != nil {
			return r, false
		}
		r.Desde = &t
	}
	if v := c.Query("hasta"); v != "" {
		t, err := time.ParseInLocation("2006-01-02", v, time.Local)
		if err != nil {
			return r, false
		}
		sig := t.AddDate(0, 0, 1)
		r.Hasta = &sig
	}
	return r, true
}

// fechaInvalida responde 400 cuando el rango de fechas no se puede interpretar.
func fechaInvalida(c *fiber.Ctx) error {
	return c.Status(fiber.StatusBadRequest).JSON(models.APIError{
		Status:  fiber.StatusBadRequest,
		Message: "Las fechas deben tener el formato YYYY-MM-DD",
	})
}

// limiteTop acota el tamaño de los rankings (por defecto 5, máximo 50).
func limiteTop(c *fiber.Ctx) int {
	return min(max(atoiSeguro(c.Query("limit"), 5), 1), 50)
}

// GetReporteResumen devuelve los KPIs de ventas del periodo.
func GetReporteResumen(c *fiber.Ctx) error {
	rango, ok := rangoDeQuery(c)
	if !ok {
		return fechaInvalida(c)
	}
	res, err := repository.ObtenerResumenVentas(c.Context(), rango)
	if err != nil {
		return errorInterno(c, err)
	}
	return c.JSON(res)
}

// GetReporteTopProductos devuelve los productos más vendidos por ingresos.
func GetReporteTopProductos(c *fiber.Ctx) error {
	rango, ok := rangoDeQuery(c)
	if !ok {
		return fechaInvalida(c)
	}
	top, err := repository.ObtenerTopProductos(c.Context(), rango, limiteTop(c))
	if err != nil {
		return errorInterno(c, err)
	}
	return c.JSON(top)
}

// GetReporteTopReferidos devuelve los distribuidores con más ventas en su red.
func GetReporteTopReferidos(c *fiber.Ctx) error {
	rango, ok := rangoDeQuery(c)
	if !ok {
		return fechaInvalida(c)
	}
	top, err := repository.ObtenerTopReferidos(c.Context(), rango, limiteTop(c))
	if err != nil {
		return errorInterno(c, err)
	}
	return c.JSON(top)
}

// GetReporteComisionesPendientes devuelve la deuda por distribuidor (misma cola de liquidaciones).
func GetReporteComisionesPendientes(c *fiber.Ctx) error {
	deudas, err := repository.ObtenerDeudas(c.Context())
	if err != nil {
		return errorInterno(c, err)
	}
	return c.JSON(deudas)
}
