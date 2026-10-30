// Este archivo atiende la evolución clínica (F2-4: RF-05 y RN-01): notas SOEP con plazo de edición, firma digital del
// profesional, adendas y procedimientos realizados. El odontólogo y el asistente redactan; solo el profesional
// responsable firma. El router fija los permisos.
package controllers

import (
	"context"
	"errors"
	"time"

	"github.com/gofiber/fiber/v2"
	"sigo-dental-backend/firma"
	"sigo-dental-backend/middleware"
	"sigo-dental-backend/models"
)

// EvolucionRepository es el contrato de la capa de datos de la evolución clínica.
type EvolucionRepository interface {
	ProfesionalDeUsuario(ctx context.Context, usuarioID int) (models.ProfesionalDeUsuario, bool, error)
	ProfesionalActivo(ctx context.Context, profesionalID int) (bool, error)
	ListarNotas(ctx context.Context, pacienteID int) ([]models.NotaCompleta, error)
	ObtenerNota(ctx context.Context, pacienteID, notaID int) (models.NotaCompleta, error)
	CrearNota(ctx context.Context, n models.Nota) (int, error)
	ActualizarNota(ctx context.Context, pacienteID, notaID int, n models.Nota) error
	GuardarFirma(ctx context.Context, f models.Firma) (models.Firma, error)
	CrearAdendaFirmada(ctx context.Context, a models.Adenda, firmar func(models.Adenda) models.Firma) (models.Adenda, error)
	ListarProcedimientosRealizados(ctx context.Context, pacienteID int) ([]models.ProcedimientoRealizado, error)
	ObtenerProcedimientoRealizado(ctx context.Context, pacienteID, id int) (models.ProcedimientoRealizado, error)
	IniciarProcedimiento(ctx context.Context, n models.NuevoProcedimientoRealizado, profesionalID, usuarioID int) (int, error)
	CambiarEstadoProcedimiento(ctx context.Context, pacienteID, id int, estado string) error
	ObtenerParametro(ctx context.Context, clave string) (string, error)
}

// Firmador verifica la identidad de quien firma (reautenticación con su contraseña).
type Firmador interface {
	Autenticar(ctx context.Context, usuarioID int, password string) (firma.Firmante, error)
}

// EvolucionController atiende notas, firmas, adendas y procedimientos realizados.
type EvolucionController struct {
	repo     EvolucionRepository
	firmador Firmador
	// ahora permite fijar el reloj en las pruebas del plazo de edición.
	ahora func() time.Time
}

// NuevoEvolucionController construye el controlador de la evolución clínica.
func NuevoEvolucionController(repo EvolucionRepository, firmador Firmador) *EvolucionController {
	return &EvolucionController{repo: repo, firmador: firmador, ahora: time.Now}
}

const (
	maxTextoSOEP        = 4000
	plazoNotaPorDefecto = 24
)

// plazoDeNotas lee el parámetro bloqueo_nota_horas (24 por defecto).
func (ctrl *EvolucionController) plazoDeNotas(ctx context.Context) time.Duration {
	return time.Duration(parametroEntero(ctx, ctrl.repo, "bloqueo_nota_horas", plazoNotaPorDefecto)) * time.Hour
}

// presentarNota completa el estado derivado de la nota y comprueba la huella de su firma y de sus adendas.
func (ctrl *EvolucionController) presentarNota(n models.NotaCompleta, plazo time.Duration) models.NotaCompleta {
	n.Estado = models.EstadoDeNota(n.CreadaEn, n.Firma != nil, ctrl.ahora(), plazo)
	n.EditableHasta = n.CreadaEn.Add(plazo)
	if n.Firma != nil {
		n.Firma.Valida = firma.Verificar(*n.Firma, models.ContenidoCanonicoNota(n.Nota))
	}
	n.Adendas = presentarAdendas(n.Adendas)
	return n
}

func presentarAdendas(adendas []models.Adenda) []models.Adenda {
	for i := range adendas {
		if adendas[i].Firma != nil {
			adendas[i].Firma.Valida = firma.Verificar(*adendas[i].Firma, models.ContenidoCanonicoAdenda(adendas[i]))
		}
	}
	return adendas
}

type cuerpoNota struct {
	Subjetivo     string `json:"subjetivo"`
	Objetivo      string `json:"objetivo"`
	Evaluacion    string `json:"evaluacion"`
	Plan          string `json:"plan"`
	CitaID        *int   `json:"cita_id"`
	ProfesionalID int    `json:"profesional_id"`
}

// validar exige los cuatro apartados SOEP con contenido.
func (cuerpo cuerpoNota) validar() (models.Nota, string) {
	var n models.Nota
	campos := []struct {
		origen  string
		destino *string
		nombre  string
	}{
		{cuerpo.Subjetivo, &n.Subjetivo, "Subjetivo"}, {cuerpo.Objetivo, &n.Objetivo, "Objetivo"},
		{cuerpo.Evaluacion, &n.Evaluacion, "Evaluación"}, {cuerpo.Plan, &n.Plan, "Plan"},
	}
	for _, campo := range campos {
		texto, ok := textoObligatorio(campo.origen, maxTextoSOEP)
		if !ok {
			return n, "El apartado " + campo.nombre + " es obligatorio (hasta 4000 caracteres)"
		}
		*campo.destino = texto
	}
	if cuerpo.CitaID != nil && *cuerpo.CitaID <= 0 {
		return n, "La cita indicada no es válida"
	}
	n.CitaID = cuerpo.CitaID
	return n, ""
}

// ListarNotas responde GET /api/pacientes/:id/notas: la bitácora cronológica con el estado de cada nota.
func (ctrl *EvolucionController) ListarNotas(c *fiber.Ctx) error {
	id, ok := idDeRuta(c)
	if !ok {
		return errorCliente(c, fiber.StatusBadRequest, "Identificador inválido")
	}
	notas, err := ctrl.repo.ListarNotas(c.Context(), id)
	if err != nil {
		return errorDominio(c, err)
	}
	plazo := ctrl.plazoDeNotas(c.Context())
	usuarioID := usuarioDeLaPeticion(c)
	profesionalID := 0
	if rol, _ := c.Locals("rol").(string); rol == models.RolOdontologo {
		if profesional, ok, err := ctrl.repo.ProfesionalDeUsuario(c.Context(), usuarioID); err == nil && ok {
			profesionalID = profesional.ID
		}
	}
	for i := range notas {
		notas[i] = ctrl.presentarNota(notas[i], plazo)
		notas[i].PuedeEditar = notas[i].Estado == models.NotaEditable && notas[i].AutorID == usuarioID
		notas[i].PuedeFirmar = notas[i].Firma == nil && profesionalID != 0 && notas[i].ProfesionalID == profesionalID
	}
	return c.JSON(notas)
}

// CrearNota responde POST /api/pacientes/:id/notas. El profesional responsable es quien redacta si es odontólogo;
// si redacta el asistente, indica el profesional que la firmará.
func (ctrl *EvolucionController) CrearNota(c *fiber.Ctx) error {
	id, ok := idDeRuta(c)
	if !ok {
		return errorCliente(c, fiber.StatusBadRequest, "Identificador inválido")
	}
	var cuerpo cuerpoNota
	if err := c.BodyParser(&cuerpo); err != nil {
		return cuerpoInvalido(c)
	}
	nota, motivo := cuerpo.validar()
	if motivo != "" {
		return errorCliente(c, fiber.StatusBadRequest, motivo)
	}
	usuarioID := usuarioDeLaPeticion(c)
	if rol, _ := c.Locals("rol").(string); rol == models.RolOdontologo {
		profesional, ok, err := ctrl.repo.ProfesionalDeUsuario(c.Context(), usuarioID)
		if err != nil {
			return errorDominio(c, err)
		}
		if !ok {
			return errorEvolucion(c, models.ErrNoEsProfesional)
		}
		nota.ProfesionalID = profesional.ID
	} else {
		activo, err := ctrl.repo.ProfesionalActivo(c.Context(), cuerpo.ProfesionalID)
		if err != nil {
			return errorDominio(c, err)
		}
		if cuerpo.ProfesionalID <= 0 || !activo {
			return errorCliente(c, fiber.StatusBadRequest, "Indica el profesional responsable que firmará la nota")
		}
		nota.ProfesionalID = cuerpo.ProfesionalID
	}
	nota.PacienteID, nota.AutorID = id, usuarioID

	notaID, err := ctrl.repo.CrearNota(c.Context(), nota)
	if err != nil {
		return errorEvolucion(c, err)
	}
	middleware.MarcarAuditoria(c, id, notaID)
	creada, err := ctrl.repo.ObtenerNota(c.Context(), id, notaID)
	if err != nil {
		return errorDominio(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(ctrl.presentarNota(creada, ctrl.plazoDeNotas(c.Context())))
}

func idsDeNota(c *fiber.Ctx) (pacienteID, notaID int, ok bool) {
	pacienteID, ok = idDeRuta(c)
	if !ok {
		return 0, 0, false
	}
	notaID, err := c.ParamsInt("nid")
	return pacienteID, notaID, err == nil && notaID > 0
}

// ActualizarNota responde PUT /api/pacientes/:id/notas/:nid. Solo la edita su autor, y solo mientras no esté firmada
// ni haya vencido el plazo: después el servidor rechaza cualquier cambio (RF-05) y la corrección es una adenda.
func (ctrl *EvolucionController) ActualizarNota(c *fiber.Ctx) error {
	pacienteID, notaID, ok := idsDeNota(c)
	if !ok {
		return errorCliente(c, fiber.StatusBadRequest, "Identificador inválido")
	}
	var cuerpo cuerpoNota
	if err := c.BodyParser(&cuerpo); err != nil {
		return cuerpoInvalido(c)
	}
	nueva, motivo := cuerpo.validar()
	if motivo != "" {
		return errorCliente(c, fiber.StatusBadRequest, motivo)
	}
	plazo := ctrl.plazoDeNotas(c.Context())
	actual, err := ctrl.repo.ObtenerNota(c.Context(), pacienteID, notaID)
	if err != nil {
		return errorDominio(c, err)
	}
	if actual.AutorID != usuarioDeLaPeticion(c) {
		return errorCliente(c, fiber.StatusForbidden, "Solo el autor de la nota puede editarla")
	}
	switch ctrl.presentarNota(actual, plazo).Estado {
	case models.NotaFirmada:
		return errorEvolucion(c, models.ErrNotaFirmada)
	case models.NotaPendienteFirma:
		return errorEvolucion(c, models.ErrNotaBloqueada)
	}
	if err := ctrl.repo.ActualizarNota(c.Context(), pacienteID, notaID, nueva); err != nil {
		return errorEvolucion(c, err)
	}
	middleware.MarcarAuditoria(c, pacienteID, notaID)
	editada, err := ctrl.repo.ObtenerNota(c.Context(), pacienteID, notaID)
	if err != nil {
		return errorDominio(c, err)
	}
	return c.JSON(ctrl.presentarNota(editada, plazo))
}

type cuerpoFirma struct {
	Password string `json:"password"`
}

// FirmarNota responde POST /api/pacientes/:id/notas/:nid/firmar. Reautentica al profesional con su contraseña y
// guarda la huella del contenido. Una nota con el plazo vencido todavía se puede firmar; firmarla la bloquea al instante.
func (ctrl *EvolucionController) FirmarNota(c *fiber.Ctx) error {
	pacienteID, notaID, ok := idsDeNota(c)
	if !ok {
		return errorCliente(c, fiber.StatusBadRequest, "Identificador inválido")
	}
	var cuerpo cuerpoFirma
	if err := c.BodyParser(&cuerpo); err != nil || cuerpo.Password == "" {
		return errorCliente(c, fiber.StatusBadRequest, "Escribe tu contraseña para firmar")
	}
	nota, err := ctrl.repo.ObtenerNota(c.Context(), pacienteID, notaID)
	if err != nil {
		return errorDominio(c, err)
	}
	if nota.Firma != nil {
		return errorEvolucion(c, models.ErrYaFirmado)
	}
	firmante, err := ctrl.firmador.Autenticar(c.Context(), usuarioDeLaPeticion(c), cuerpo.Password)
	if err != nil {
		return errorEvolucion(c, err)
	}
	if firmante.ProfesionalID != nota.ProfesionalID {
		return errorEvolucion(c, models.ErrFirmanteNoCorresponde)
	}
	if _, err := ctrl.repo.GuardarFirma(c.Context(), firma.Construir(firmante, models.FirmaNota, nota.ID, models.ContenidoCanonicoNota(nota.Nota))); err != nil {
		if errors.Is(err, models.ErrDuplicado) {
			err = models.ErrYaFirmado
		}
		return errorEvolucion(c, err)
	}
	middleware.MarcarAuditoria(c, pacienteID, notaID)
	firmada, err := ctrl.repo.ObtenerNota(c.Context(), pacienteID, notaID)
	if err != nil {
		return errorDominio(c, err)
	}
	return c.JSON(ctrl.presentarNota(firmada, ctrl.plazoDeNotas(c.Context())))
}

type cuerpoAdenda struct {
	NotaID                   *int   `json:"nota_id"`
	ProcedimientoRealizadoID *int   `json:"procedimiento_realizado_id"`
	Texto                    string `json:"texto"`
	Password                 string `json:"password"`
}

// CrearAdenda responde POST /api/pacientes/:id/adendas. Es la única forma de corregir una nota firmada o un
// procedimiento realizado (RN-01): una nota nueva, enlazada a la original y firmada por su profesional responsable.
func (ctrl *EvolucionController) CrearAdenda(c *fiber.Ctx) error {
	pacienteID, ok := idDeRuta(c)
	if !ok {
		return errorCliente(c, fiber.StatusBadRequest, "Identificador inválido")
	}
	var cuerpo cuerpoAdenda
	if err := c.BodyParser(&cuerpo); err != nil {
		return cuerpoInvalido(c)
	}
	if (cuerpo.NotaID == nil) == (cuerpo.ProcedimientoRealizadoID == nil) {
		return errorCliente(c, fiber.StatusBadRequest, "La adenda corrige una nota o un procedimiento, uno de los dos")
	}
	texto, okTexto := textoObligatorio(cuerpo.Texto, maxTextoSOEP)
	if !okTexto {
		return errorCliente(c, fiber.StatusBadRequest, "El texto de la adenda es obligatorio (hasta 4000 caracteres)")
	}
	if cuerpo.Password == "" {
		return errorCliente(c, fiber.StatusBadRequest, "Escribe tu contraseña para firmar la adenda")
	}

	// Se comprueba que lo corregido exista, sea de este paciente y no siga siendo editable, y quién es su responsable.
	var responsable int
	if cuerpo.NotaID != nil {
		nota, err := ctrl.repo.ObtenerNota(c.Context(), pacienteID, *cuerpo.NotaID)
		if err != nil {
			return errorDominio(c, err)
		}
		if ctrl.presentarNota(nota, ctrl.plazoDeNotas(c.Context())).Estado == models.NotaEditable {
			return errorCliente(c, fiber.StatusConflict, "La nota aún se puede editar: corrígela directamente o fírmala")
		}
		responsable = nota.ProfesionalID
	} else {
		proc, err := ctrl.repo.ObtenerProcedimientoRealizado(c.Context(), pacienteID, *cuerpo.ProcedimientoRealizadoID)
		if err != nil {
			return errorDominio(c, err)
		}
		if proc.Estado != models.ProcRealizado {
			return errorCliente(c, fiber.StatusConflict, "Solo un procedimiento realizado se corrige con una adenda")
		}
		responsable = proc.ProfesionalID
	}
	firmante, err := ctrl.firmador.Autenticar(c.Context(), usuarioDeLaPeticion(c), cuerpo.Password)
	if err != nil {
		return errorEvolucion(c, err)
	}
	if firmante.ProfesionalID != responsable {
		return errorEvolucion(c, models.ErrFirmanteNoCorresponde)
	}
	adenda := models.Adenda{
		PacienteID: pacienteID, NotaID: cuerpo.NotaID, ProcedimientoRealizadoID: cuerpo.ProcedimientoRealizadoID,
		ProfesionalID: firmante.ProfesionalID, ProfesionalNombre: firmante.Nombre, AutorID: usuarioDeLaPeticion(c), Texto: texto,
	}
	creada, err := ctrl.repo.CrearAdendaFirmada(c.Context(), adenda, func(a models.Adenda) models.Firma {
		return firma.Construir(firmante, models.FirmaAdenda, a.ID, models.ContenidoCanonicoAdenda(a))
	})
	if err != nil {
		return errorEvolucion(c, err)
	}
	middleware.MarcarAuditoria(c, pacienteID, creada.ID)
	return c.Status(fiber.StatusCreated).JSON(presentarAdendas([]models.Adenda{creada})[0])
}

// ListarProcedimientosRealizados responde GET /api/pacientes/:id/procedimientos-realizados.
func (ctrl *EvolucionController) ListarProcedimientosRealizados(c *fiber.Ctx) error {
	id, ok := idDeRuta(c)
	if !ok {
		return errorCliente(c, fiber.StatusBadRequest, "Identificador inválido")
	}
	lista, err := ctrl.repo.ListarProcedimientosRealizados(c.Context(), id)
	if err != nil {
		return errorDominio(c, err)
	}
	for i := range lista {
		lista[i].Adendas = presentarAdendas(lista[i].Adendas)
	}
	return c.JSON(lista)
}

type cuerpoProcRealizado struct {
	ProcedimientoID   int    `json:"procedimiento_id"`
	Pieza             *int   `json:"pieza"`
	Superficie        string `json:"superficie"`
	CitaID            *int   `json:"cita_id"`
	PresupuestoItemID *int   `json:"presupuesto_item_id"`
	Notas             string `json:"notas"`
}

// IniciarProcedimiento responde POST /api/pacientes/:id/procedimientos-realizados. Lo registra "en curso".
func (ctrl *EvolucionController) IniciarProcedimiento(c *fiber.Ctx) error {
	pacienteID, ok := idDeRuta(c)
	if !ok {
		return errorCliente(c, fiber.StatusBadRequest, "Identificador inválido")
	}
	var cuerpo cuerpoProcRealizado
	if err := c.BodyParser(&cuerpo); err != nil {
		return cuerpoInvalido(c)
	}
	if cuerpo.ProcedimientoID <= 0 {
		return errorCliente(c, fiber.StatusBadRequest, "Elige el procedimiento del catálogo")
	}
	if cuerpo.Pieza != nil && !models.PiezaFDIValida(*cuerpo.Pieza) {
		return errorCliente(c, fiber.StatusBadRequest, "La pieza dental no es válida (numeración FDI)")
	}
	if cuerpo.Superficie != "" && !contiene(models.SuperficiesValidas, cuerpo.Superficie) {
		return errorCliente(c, fiber.StatusBadRequest, "La superficie no es válida")
	}
	if cuerpo.Superficie != "" && cuerpo.Pieza == nil {
		return errorCliente(c, fiber.StatusBadRequest, "Indica la pieza de la superficie")
	}
	notas, okNotas := textoOpcional(cuerpo.Notas, 300)
	if !okNotas {
		return errorCliente(c, fiber.StatusBadRequest, "Las notas admiten hasta 300 caracteres")
	}
	usuarioID := usuarioDeLaPeticion(c)
	profesional, ok, err := ctrl.repo.ProfesionalDeUsuario(c.Context(), usuarioID)
	if err != nil {
		return errorDominio(c, err)
	}
	if !ok {
		return errorEvolucion(c, models.ErrNoEsProfesional)
	}
	nuevo := models.NuevoProcedimientoRealizado{
		PacienteID: pacienteID, ProcedimientoID: cuerpo.ProcedimientoID, Pieza: cuerpo.Pieza, Superficie: cuerpo.Superficie,
		CitaID: cuerpo.CitaID, PresupuestoItemID: cuerpo.PresupuestoItemID, Notas: notas,
	}
	id, err := ctrl.repo.IniciarProcedimiento(c.Context(), nuevo, profesional.ID, usuarioID)
	if err != nil {
		return errorEvolucion(c, err)
	}
	middleware.MarcarAuditoria(c, pacienteID, id)
	creado, err := ctrl.repo.ObtenerProcedimientoRealizado(c.Context(), pacienteID, id)
	if err != nil {
		return errorDominio(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(creado)
}

func contiene(lista []string, valor string) bool {
	for _, v := range lista {
		if v == valor {
			return true
		}
	}
	return false
}

func (ctrl *EvolucionController) cerrarProcedimiento(c *fiber.Ctx, estado string) error {
	pacienteID, ok := idDeRuta(c)
	id, err := c.ParamsInt("pid")
	if !ok || err != nil || id <= 0 {
		return errorCliente(c, fiber.StatusBadRequest, "Identificador inválido")
	}
	if err := ctrl.repo.CambiarEstadoProcedimiento(c.Context(), pacienteID, id, estado); err != nil {
		return errorEvolucion(c, err)
	}
	middleware.MarcarAuditoria(c, pacienteID, id)
	actualizado, err := ctrl.repo.ObtenerProcedimientoRealizado(c.Context(), pacienteID, id)
	if err != nil {
		return errorDominio(c, err)
	}
	return c.JSON(actualizado)
}

// CompletarProcedimiento responde POST .../procedimientos-realizados/:pid/completar: pasa a "realizado" y desde
// entonces no se borra ni se cambia (RN-01). No existe un endpoint de borrado, y la base lo impide además.
func (ctrl *EvolucionController) CompletarProcedimiento(c *fiber.Ctx) error {
	return ctrl.cerrarProcedimiento(c, models.ProcRealizado)
}

// CancelarProcedimiento responde POST .../procedimientos-realizados/:pid/cancelar: solo un procedimiento en curso.
func (ctrl *EvolucionController) CancelarProcedimiento(c *fiber.Ctx) error {
	return ctrl.cerrarProcedimiento(c, models.ProcCancelado)
}

// errorEvolucion traduce los errores de la evolución clínica y de la firma; el resto sigue el camino común.
func errorEvolucion(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, models.ErrNotaFirmada):
		return errorCliente(c, fiber.StatusConflict, "La nota está firmada y no se edita: corrígela con una adenda")
	case errors.Is(err, models.ErrNotaBloqueada):
		return errorCliente(c, fiber.StatusConflict, "Venció el plazo de edición de la nota: fírmala o corrígela con una adenda")
	case errors.Is(err, models.ErrYaFirmado):
		return errorCliente(c, fiber.StatusConflict, "El documento ya está firmado")
	case errors.Is(err, models.ErrInmutable):
		return errorCliente(c, fiber.StatusConflict, "El registro es inmutable y no admite cambios: corrígelo con una adenda (RN-01)")
	case errors.Is(err, models.ErrPasswordIncorrecta):
		// 403 y no 401: un 401 haría que el cliente cerrara la sesión por una contraseña mal escrita.
		return errorCliente(c, fiber.StatusForbidden, "Contraseña incorrecta")
	case errors.Is(err, models.ErrNoEsProfesional):
		return errorCliente(c, fiber.StatusForbidden, "Tu cuenta no es la de un profesional activo: no puede firmar")
	case errors.Is(err, models.ErrFirmanteNoCorresponde):
		return errorCliente(c, fiber.StatusForbidden, "Solo el profesional responsable puede firmar este documento")
	case errors.Is(err, models.ErrDemasiadosIntentos):
		return errorCliente(c, fiber.StatusTooManyRequests, "Demasiados intentos con contraseña incorrecta: espera unos minutos")
	case errors.Is(err, models.ErrConsentimientoRequerido):
		return errorCliente(c, fiber.StatusConflict, "Este procedimiento requiere un consentimiento informado firmado y vigente: obtenlo en la pestaña Consentimientos antes de iniciarlo")
	case errors.Is(err, models.ErrConsentimientoFirmado):
		return errorCliente(c, fiber.StatusConflict, "El consentimiento ya está firmado y no admite cambios")
	case errors.Is(err, models.ErrProcedimientoNoEnCurso):
		return errorCliente(c, fiber.StatusConflict, "El procedimiento ya no está en curso: un procedimiento realizado no se cambia")
	}
	return errorDominio(c, err)
}
