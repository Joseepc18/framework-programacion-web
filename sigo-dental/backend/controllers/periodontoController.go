// Este archivo atiende el periodontograma (F2-3: RF-04). El odontólogo y el asistente capturan y consultan los
// exámenes; el NIC lo calcula el servidor. Un examen no se modifica: una corrección es un examen nuevo.
package controllers

import (
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/gofiber/fiber/v2"
	"sigo-dental-backend/middleware"
	"sigo-dental-backend/models"
)

// PeriodontoRepository es el contrato de la capa de datos del periodontograma.
type PeriodontoRepository interface {
	ListarPeriodontogramas(ctx context.Context, pacienteID int) (models.ListaPerio, error)
	ObtenerPeriodontograma(ctx context.Context, pacienteID, id int) (models.ExamenPerio, error)
	CrearPeriodontograma(ctx context.Context, pacienteID int, fecha, notas string, dientes []models.DientePerio, usuarioID int) (int, error)
	PiezasAusentes(ctx context.Context, pacienteID int) ([]int, error)
}

// PeriodontoController atiende las rutas del periodontograma.
type PeriodontoController struct {
	repo  PeriodontoRepository
	ahora func() time.Time
}

// NuevoPeriodontoController construye el controlador del periodontograma.
func NuevoPeriodontoController(repo PeriodontoRepository) *PeriodontoController {
	return &PeriodontoController{repo: repo, ahora: time.Now}
}

type cuerpoExamenPerio struct {
	Fecha   string               `json:"fecha"`
	Notas   string               `json:"notas"`
	Dientes []models.DientePerio `json:"dientes"`
}

// validarExamen comprueba el examen contra las piezas ausentes del odontograma; devuelve el motivo si es inválido.
func (cuerpo cuerpoExamenPerio) validar(hoy time.Time, ausentes []int) (string, string) {
	fecha := cuerpo.Fecha
	if fecha == "" {
		fecha = hoy.Format(formatoFecha)
	}
	instante, err := time.Parse(formatoFecha, fecha)
	// Se tolera un día de diferencia por la zona horaria del cliente.
	if err != nil || instante.After(hoy.AddDate(0, 0, 1)) || instante.Year() < 2000 {
		return "", "La fecha del examen debe tener el formato AAAA-MM-DD y no puede ser futura"
	}
	if len(cuerpo.Dientes) == 0 {
		return "", "El examen necesita al menos un diente"
	}
	if len(cuerpo.Dientes) > 32 {
		return "", "Un examen tiene como máximo 32 dientes"
	}
	vistas := []int{}
	for _, d := range cuerpo.Dientes {
		if motivo := models.ValidarDientePerio(d); motivo != "" {
			return "", motivo
		}
		if slices.Contains(vistas, d.Pieza) {
			return "", fmt.Sprintf("La pieza %d está repetida", d.Pieza)
		}
		vistas = append(vistas, d.Pieza)
		if slices.Contains(ausentes, d.Pieza) {
			return "", fmt.Sprintf("La pieza %d figura como ausente en el odontograma", d.Pieza)
		}
	}
	return fecha, ""
}

// ListarPeriodontogramas responde GET /api/pacientes/:id/periodontogramas.
func (ctrl *PeriodontoController) ListarPeriodontogramas(c *fiber.Ctx) error {
	id, ok := idDeRuta(c)
	if !ok {
		return errorCliente(c, fiber.StatusBadRequest, "Identificador inválido")
	}
	lista, err := ctrl.repo.ListarPeriodontogramas(c.Context(), id)
	if err != nil {
		return errorDominio(c, err)
	}
	return c.JSON(lista)
}

// ObtenerPeriodontograma responde GET /api/pacientes/:id/periodontogramas/:eid con el examen y el NIC de cada punto.
func (ctrl *PeriodontoController) ObtenerPeriodontograma(c *fiber.Ctx) error {
	id, ok := idDeRuta(c)
	examenID, err := c.ParamsInt("eid")
	if !ok || err != nil || examenID <= 0 {
		return errorCliente(c, fiber.StatusBadRequest, "Identificador inválido")
	}
	examen, err := ctrl.repo.ObtenerPeriodontograma(c.Context(), id, examenID)
	if err != nil {
		return errorDominio(c, err)
	}
	return c.JSON(examen)
}

// CrearPeriodontograma responde POST /api/pacientes/:id/periodontogramas.
func (ctrl *PeriodontoController) CrearPeriodontograma(c *fiber.Ctx) error {
	id, ok := idDeRuta(c)
	if !ok {
		return errorCliente(c, fiber.StatusBadRequest, "Identificador inválido")
	}
	var cuerpo cuerpoExamenPerio
	if err := c.BodyParser(&cuerpo); err != nil {
		return cuerpoInvalido(c)
	}
	notas, okNotas := textoOpcional(cuerpo.Notas, 500)
	if !okNotas {
		return errorCliente(c, fiber.StatusBadRequest, "Las notas admiten hasta 500 caracteres")
	}
	ausentes, err := ctrl.repo.PiezasAusentes(c.Context(), id)
	if err != nil {
		return errorDominio(c, err)
	}
	fecha, motivo := cuerpo.validar(ctrl.ahora(), ausentes)
	if motivo != "" {
		return errorCliente(c, fiber.StatusBadRequest, motivo)
	}
	examenID, err := ctrl.repo.CrearPeriodontograma(c.Context(), id, fecha, notas, cuerpo.Dientes, usuarioDeLaPeticion(c))
	if err != nil {
		return errorDominio(c, err)
	}
	middleware.MarcarAuditoria(c, id, examenID)
	examen, err := ctrl.repo.ObtenerPeriodontograma(c.Context(), id, examenID)
	if err != nil {
		return errorDominio(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(examen)
}
