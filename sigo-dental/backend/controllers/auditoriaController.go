// Este archivo atiende la consulta de la bitácora de accesos (F2-1: RNF-06). La bitácora solo se lee: no existe
// ninguna ruta que la modifique, y la base de datos lo impide además con un disparador.
package controllers

import (
	"context"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"sigo-dental-backend/models"
)

// AuditoriaRepository es el contrato de la capa de datos de la bitácora.
type AuditoriaRepository interface {
	ListarAuditoria(ctx context.Context, f models.FiltroAuditoria) (models.PaginaAuditoria, error)
}

// AuditoriaController atiende la consulta de la auditoría.
type AuditoriaController struct {
	repo AuditoriaRepository
}

// NuevoAuditoriaController construye el controlador de la auditoría.
func NuevoAuditoriaController(repo AuditoriaRepository) *AuditoriaController {
	return &AuditoriaController{repo: repo}
}

const (
	limiteAuditoriaPorDefecto = 50
	limiteAuditoriaMaximo     = 200
	maxFiltroTexto            = 100
)

// ListarAuditoria responde GET /api/auditoria con filtros por usuario (correo), paciente (nombre, documento o
// historia), acción y rango de fechas (RFC 3339; "hasta" es exclusivo), paginado.
func (ctrl *AuditoriaController) ListarAuditoria(c *fiber.Ctx) error {
	filtro := models.FiltroAuditoria{
		Usuario:  strings.TrimSpace(c.Query("usuario")),
		Paciente: strings.TrimSpace(c.Query("paciente")),
		Accion:   strings.TrimSpace(c.Query("accion")),
		Limite:   limiteAuditoriaPorDefecto,
		Pagina:   1,
	}
	if len(filtro.Usuario) > maxFiltroTexto || len(filtro.Paciente) > maxFiltroTexto {
		return errorCliente(c, fiber.StatusBadRequest, "Los filtros de texto admiten hasta 100 caracteres")
	}
	if filtro.Accion != "" && !slices.Contains(models.AccionesAuditoria, filtro.Accion) {
		return errorCliente(c, fiber.StatusBadRequest, "La acción del filtro no es válida")
	}
	for _, fecha := range []struct {
		nombre  string
		destino **time.Time
	}{{"desde", &filtro.Desde}, {"hasta", &filtro.Hasta}} {
		texto := c.Query(fecha.nombre)
		if texto == "" {
			continue
		}
		instante, err := time.Parse(time.RFC3339, texto)
		if err != nil {
			return errorCliente(c, fiber.StatusBadRequest, "La fecha \""+fecha.nombre+"\" debe tener el formato ISO 8601 (RFC 3339)")
		}
		*fecha.destino = &instante
	}
	if filtro.Desde != nil && filtro.Hasta != nil && !filtro.Hasta.After(*filtro.Desde) {
		return errorCliente(c, fiber.StatusBadRequest, "La fecha final debe ser posterior a la inicial")
	}
	if texto := c.Query("limite"); texto != "" {
		limite, err := strconv.Atoi(texto)
		if err != nil || limite < 1 || limite > limiteAuditoriaMaximo {
			return errorCliente(c, fiber.StatusBadRequest, "El límite debe estar entre 1 y 200")
		}
		filtro.Limite = limite
	}
	if texto := c.Query("pagina"); texto != "" {
		pagina, err := strconv.Atoi(texto)
		if err != nil || pagina < 1 {
			return errorCliente(c, fiber.StatusBadRequest, "La página debe ser un entero positivo")
		}
		filtro.Pagina = pagina
	}

	pagina, err := ctrl.repo.ListarAuditoria(c.Context(), filtro)
	if err != nil {
		return errorDominio(c, err)
	}
	return c.JSON(pagina)
}
