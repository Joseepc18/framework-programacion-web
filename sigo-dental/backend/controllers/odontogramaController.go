// Este archivo atiende el odontograma (F2-2: RF-03). El odontólogo registra, corrige y retira hallazgos; el asistente
// solo consulta. Un hallazgo realizado es inmutable (RN-01). El router fija los permisos.
package controllers

import (
	"context"
	"errors"
	"time"

	"github.com/gofiber/fiber/v2"
	"sigo-dental-backend/middleware"
	"sigo-dental-backend/models"
)

// OdontogramaRepository es el contrato de la capa de datos del odontograma.
type OdontogramaRepository interface {
	ListarHallazgos(ctx context.Context, pacienteID int, al *time.Time) ([]models.Hallazgo, error)
	HistorialHallazgos(ctx context.Context, pacienteID int) ([]models.Hallazgo, error)
	RegistrarHallazgo(ctx context.Context, n models.NuevoHallazgo, usuarioID int) (models.Hallazgo, error)
	CorregirHallazgo(ctx context.Context, id int, n models.NuevoHallazgo, usuarioID int) (models.Hallazgo, error)
	RetirarHallazgo(ctx context.Context, pacienteID, id, usuarioID int) error
}

// OdontogramaController atiende las rutas del odontograma.
type OdontogramaController struct {
	repo OdontogramaRepository
}

// NuevoOdontogramaController construye el controlador del odontograma.
func NuevoOdontogramaController(repo OdontogramaRepository) *OdontogramaController {
	return &OdontogramaController{repo: repo}
}

const maxObservacionHallazgo = 200

type cuerpoHallazgo struct {
	Pieza       int    `json:"pieza"`
	Superficie  string `json:"superficie"`
	Tipo        string `json:"tipo"`
	Estado      string `json:"estado"`
	Observacion string `json:"observacion"`
}

// validar comprueba el hallazgo y lo convierte al modelo; el segundo valor es el motivo si es inválido.
func (cuerpo cuerpoHallazgo) validar(pacienteID int) (models.NuevoHallazgo, string) {
	if motivo := models.ValidarHallazgo(cuerpo.Pieza, cuerpo.Superficie, cuerpo.Tipo, cuerpo.Estado); motivo != "" {
		return models.NuevoHallazgo{}, motivo
	}
	observacion, ok := textoOpcional(cuerpo.Observacion, maxObservacionHallazgo)
	if !ok {
		return models.NuevoHallazgo{}, "La observación admite hasta 200 caracteres"
	}
	return models.NuevoHallazgo{
		PacienteID: pacienteID, Pieza: cuerpo.Pieza, Superficie: cuerpo.Superficie, Tipo: cuerpo.Tipo,
		Estado: cuerpo.Estado, Observacion: observacion,
	}, ""
}

// errorHallazgo traduce los errores propios del odontograma; el resto sigue el camino común.
func errorHallazgo(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, models.ErrInmutable):
		return errorCliente(c, fiber.StatusConflict, "Un hallazgo realizado no se corrige ni se retira (RN-01): registra una nota de adenda")
	case errors.Is(err, models.ErrHallazgoRetirado):
		return errorCliente(c, fiber.StatusConflict, "El hallazgo ya fue retirado")
	case errors.Is(err, models.ErrDuplicado):
		return errorCliente(c, fiber.StatusConflict, "Ese hallazgo ya está registrado en la pieza")
	}
	return errorDominio(c, err)
}

// idsDeHallazgo lee :id (paciente) y :hid (hallazgo) de la ruta.
func idsDeHallazgo(c *fiber.Ctx) (pacienteID, hallazgoID int, ok bool) {
	pacienteID, ok = idDeRuta(c)
	if !ok {
		return 0, 0, false
	}
	hallazgoID, err := c.ParamsInt("hid")
	if err != nil || hallazgoID <= 0 {
		return 0, 0, false
	}
	return pacienteID, hallazgoID, true
}

// ObtenerOdontograma responde GET /api/pacientes/:id/odontograma. Con ?al=<RFC 3339> devuelve el estado que tenía
// en ese instante, para revisar visitas anteriores.
func (ctrl *OdontogramaController) ObtenerOdontograma(c *fiber.Ctx) error {
	id, ok := idDeRuta(c)
	if !ok {
		return errorCliente(c, fiber.StatusBadRequest, "Identificador inválido")
	}
	var al *time.Time
	if texto := c.Query("al"); texto != "" {
		instante, err := time.Parse(time.RFC3339, texto)
		if err != nil {
			return errorCliente(c, fiber.StatusBadRequest, "El instante \"al\" debe tener el formato ISO 8601 (RFC 3339)")
		}
		al = &instante
	}
	lista, err := ctrl.repo.ListarHallazgos(c.Context(), id, al)
	if err != nil {
		return errorDominio(c, err)
	}
	return c.JSON(lista)
}

// HistorialOdontograma responde GET /api/pacientes/:id/odontograma/historial con todos los hallazgos, retirados incluidos.
func (ctrl *OdontogramaController) HistorialOdontograma(c *fiber.Ctx) error {
	id, ok := idDeRuta(c)
	if !ok {
		return errorCliente(c, fiber.StatusBadRequest, "Identificador inválido")
	}
	lista, err := ctrl.repo.HistorialHallazgos(c.Context(), id)
	if err != nil {
		return errorDominio(c, err)
	}
	return c.JSON(lista)
}

// RegistrarHallazgo responde POST /api/pacientes/:id/odontograma/hallazgos.
func (ctrl *OdontogramaController) RegistrarHallazgo(c *fiber.Ctx) error {
	id, ok := idDeRuta(c)
	if !ok {
		return errorCliente(c, fiber.StatusBadRequest, "Identificador inválido")
	}
	var cuerpo cuerpoHallazgo
	if err := c.BodyParser(&cuerpo); err != nil {
		return cuerpoInvalido(c)
	}
	nuevo, motivo := cuerpo.validar(id)
	if motivo != "" {
		return errorCliente(c, fiber.StatusBadRequest, motivo)
	}
	hallazgo, err := ctrl.repo.RegistrarHallazgo(c.Context(), nuevo, usuarioDeLaPeticion(c))
	if err != nil {
		return errorHallazgo(c, err)
	}
	middleware.MarcarAuditoria(c, id, hallazgo.ID)
	return c.Status(fiber.StatusCreated).JSON(hallazgo)
}

// CorregirHallazgo responde PUT /api/pacientes/:id/odontograma/hallazgos/:hid: retira el hallazgo y registra el
// corregido; el anterior queda en el historial. Solo los hallazgos requeridos o existentes se corrigen.
func (ctrl *OdontogramaController) CorregirHallazgo(c *fiber.Ctx) error {
	pacienteID, hallazgoID, ok := idsDeHallazgo(c)
	if !ok {
		return errorCliente(c, fiber.StatusBadRequest, "Identificador inválido")
	}
	var cuerpo cuerpoHallazgo
	if err := c.BodyParser(&cuerpo); err != nil {
		return cuerpoInvalido(c)
	}
	nuevo, motivo := cuerpo.validar(pacienteID)
	if motivo != "" {
		return errorCliente(c, fiber.StatusBadRequest, motivo)
	}
	hallazgo, err := ctrl.repo.CorregirHallazgo(c.Context(), hallazgoID, nuevo, usuarioDeLaPeticion(c))
	if err != nil {
		return errorHallazgo(c, err)
	}
	middleware.MarcarAuditoria(c, pacienteID, hallazgo.ID)
	return c.JSON(hallazgo)
}

// RetirarHallazgo responde POST /api/pacientes/:id/odontograma/hallazgos/:hid/retirar. No borra: el hallazgo
// deja de estar vigente y se conserva en el historial.
func (ctrl *OdontogramaController) RetirarHallazgo(c *fiber.Ctx) error {
	pacienteID, hallazgoID, ok := idsDeHallazgo(c)
	if !ok {
		return errorCliente(c, fiber.StatusBadRequest, "Identificador inválido")
	}
	if err := ctrl.repo.RetirarHallazgo(c.Context(), pacienteID, hallazgoID, usuarioDeLaPeticion(c)); err != nil {
		return errorHallazgo(c, err)
	}
	middleware.MarcarAuditoria(c, pacienteID, hallazgoID)
	return c.SendStatus(fiber.StatusNoContent)
}
