// Este archivo atiende los presupuestos y las tareas agendables (F2-7: RF-07 y RF-08). El odontólogo arma el plan; la
// recepción presenta y registra la aprobación. Los importes los calcula siempre el servidor. Un descuento exige la
// autorización de un rol habilitado (parámetro descuento_roles) y deja registrado quién lo autorizó.
package controllers

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
	"sigo-dental-backend/documentos"
	"sigo-dental-backend/middleware"
	"sigo-dental-backend/models"
)

// PresupuestoRepository es el contrato de la capa de datos de presupuestos.
type PresupuestoRepository interface {
	ObtenerPaciente(ctx context.Context, id int) (models.PacienteDetalle, error)
	ProfesionalDeUsuario(ctx context.Context, usuarioID int) (models.ProfesionalDeUsuario, bool, error)
	BuscarUsuarioPorEmail(ctx context.Context, email string) (models.Usuario, bool, error)
	ListarPresupuestos(ctx context.Context, pacienteID int) ([]models.Presupuesto, error)
	ObtenerPresupuesto(ctx context.Context, pacienteID, id int) (models.Presupuesto, error)
	GuardarPresupuesto(ctx context.Context, pacienteID, presupuestoID, profesionalID int, notas string, items []models.ItemParaGuardar, usuarioID int) (int, error)
	PresentarPresupuesto(ctx context.Context, pacienteID, id, usuarioID int) error
	ResolverPresupuesto(ctx context.Context, pacienteID, id int, aprobados []int, usuarioID int) error
	NuevaVersionPresupuesto(ctx context.Context, pacienteID, id, usuarioID int) (int, error)
	ListarTareas(ctx context.Context, pacienteID int) ([]models.Tarea, error)
	ObtenerParametro(ctx context.Context, clave string) (string, error)
}

// PresupuestosController atiende las rutas de presupuestos.
type PresupuestosController struct {
	repo  PresupuestoRepository
	ahora func() time.Time
}

// NuevoPresupuestosController construye el controlador de presupuestos.
func NuevoPresupuestosController(repo PresupuestoRepository) *PresupuestosController {
	return &PresupuestosController{repo: repo, ahora: time.Now}
}

const maxItemsPresupuesto = 60

type credencialesAutorizacion struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type cuerpoPresupuesto struct {
	Notas        string                    `json:"notas"`
	Items        []models.ItemPedido       `json:"items"`
	Autorizacion *credencialesAutorizacion `json:"autorizacion"`
}

// validarItems comprueba cada ítem pedido. El honorario, si se indica, es editable sobre el precio base.
func validarItems(items []models.ItemPedido) string {
	if len(items) > maxItemsPresupuesto {
		return "Un presupuesto admite hasta 60 ítems"
	}
	for i, it := range items {
		n := i + 1
		switch {
		case it.ProcedimientoID <= 0:
			return fmt.Sprintf("Ítem %d: elige el procedimiento del catálogo", n)
		case it.Fase < 1 || it.Fase > 20:
			return fmt.Sprintf("Ítem %d: la fase va de 1 a 20", n)
		case it.Sesion < 1 || it.Sesion > 50:
			return fmt.Sprintf("Ítem %d: la sesión va de 1 a 50", n)
		case it.Pieza != nil && !models.PiezaFDIValida(*it.Pieza):
			return fmt.Sprintf("Ítem %d: la pieza dental no es válida (numeración FDI)", n)
		case it.Superficie != "" && (it.Pieza == nil || !contiene(models.SuperficiesValidas, it.Superficie)):
			return fmt.Sprintf("Ítem %d: la superficie no es válida o falta la pieza", n)
		case it.HonorarioCentavos != nil:
			if _, motivo := models.TotalItem(*it.HonorarioCentavos, it.LaboratorioCentavos, it.DescuentoCentavos); motivo != "" {
				return fmt.Sprintf("Ítem %d: %s", n, motivo)
			}
		case it.LaboratorioCentavos < 0 || it.DescuentoCentavos < 0:
			return fmt.Sprintf("Ítem %d: los importes no pueden ser negativos", n)
		}
	}
	return ""
}

// autorizar verifica a quien autoriza el descuento: cuenta activa, contraseña correcta y rol habilitado.
func (ctrl *PresupuestosController) autorizar(ctx context.Context, cred *credencialesAutorizacion) (int, error) {
	if cred == nil || cred.Email == "" || cred.Password == "" {
		return 0, models.ErrDescuentoSinAutorizar
	}
	usuario, existe, err := ctrl.repo.BuscarUsuarioPorEmail(ctx, cred.Email)
	if err != nil {
		return 0, err
	}
	hash := usuario.PasswordHash
	if !existe {
		// Se compara contra un hash falso para no delatar por el tiempo de respuesta qué correos existen.
		hash = "$2a$10$7EqJtq98hPqEX7fNZaFWoOhi5BoYSe8IZ4gC1VQe6o2yEcdYpWYBG"
	}
	if bcrypt.CompareHashAndPassword([]byte(hash), []byte(cred.Password)) != nil || !existe || !usuario.Activo {
		return 0, models.ErrAutorizacionInvalida
	}
	if !models.RolPuedeAutorizar(usuario.Rol, parametroTexto(ctx, ctrl.repo, "descuento_roles", "odontologo,admin_ti")) {
		return 0, models.ErrAutorizacionInvalida
	}
	return usuario.ID, nil
}

// prepararItems resuelve la autorización de los descuentos: uno que ya estaba autorizado y no cambió la conserva;
// uno nuevo o modificado exige credenciales de un rol habilitado, y se registra quién lo autorizó.
func (ctrl *PresupuestosController) prepararItems(ctx context.Context, items []models.ItemPedido, previo *models.Presupuesto, cred *credencialesAutorizacion) ([]models.ItemParaGuardar, error) {
	anteriores := map[int]models.ItemPresupuesto{}
	if previo != nil {
		for _, it := range previo.Items {
			anteriores[it.ID] = it
		}
	}
	resultado := make([]models.ItemParaGuardar, 0, len(items))
	autorizador := 0
	for _, it := range items {
		guardar := models.ItemParaGuardar{ItemPedido: it}
		if it.DescuentoCentavos > 0 {
			if it.ID != nil {
				if antes, ok := anteriores[*it.ID]; ok && antes.DescuentoCentavos == it.DescuentoCentavos && antes.DescuentoAutorizadoPorID != nil {
					guardar.DescuentoAutorizadoPorID = antes.DescuentoAutorizadoPorID
				}
			}
			if guardar.DescuentoAutorizadoPorID == nil {
				if autorizador == 0 {
					id, err := ctrl.autorizar(ctx, cred)
					if err != nil {
						return nil, err
					}
					autorizador = id
				}
				guardar.DescuentoAutorizadoPorID = &autorizador
			}
		}
		resultado = append(resultado, guardar)
	}
	return resultado, nil
}

func idsDePresupuesto(c *fiber.Ctx) (pacienteID, presupuestoID int, ok bool) {
	pacienteID, ok = idDeRuta(c)
	if !ok {
		return 0, 0, false
	}
	presupuestoID, err := c.ParamsInt("pid")
	return pacienteID, presupuestoID, err == nil && presupuestoID > 0
}

// errorPresupuesto traduce los errores propios de presupuestos; el resto sigue el camino común.
func errorPresupuesto(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, models.ErrDescuentoSinAutorizar):
		return errorCliente(c, fiber.StatusUnprocessableEntity, "Los descuentos requieren la autorización de un rol habilitado (odontólogo o administrador): escribe su correo y contraseña")
	case errors.Is(err, models.ErrAutorizacionInvalida):
		return errorCliente(c, fiber.StatusForbidden, "La autorización del descuento no es válida: revisa el correo y la contraseña, y que el rol esté habilitado")
	case errors.Is(err, models.ErrPresupuestoNoEditable):
		return errorCliente(c, fiber.StatusConflict, "Solo un borrador se edita: un presupuesto presentado o resuelto se cambia con una versión nueva")
	case errors.Is(err, models.ErrPresupuestoNoPresentado):
		return errorCliente(c, fiber.StatusConflict, "Solo un presupuesto presentado se aprueba o se rechaza")
	case errors.Is(err, models.ErrPresupuestoVacio):
		return errorCliente(c, fiber.StatusBadRequest, "Agrega al menos un ítem antes de presentar el presupuesto")
	case errors.Is(err, models.ErrImporteInvalido):
		return errorCliente(c, fiber.StatusBadRequest, err.Error())
	case errors.Is(err, models.ErrConflicto):
		return errorCliente(c, fiber.StatusConflict, "Ese presupuesto ya tiene una versión nueva, o todavía es un borrador")
	case errors.Is(err, models.ErrDuplicado):
		return errorCliente(c, fiber.StatusConflict, "Ese presupuesto ya tiene una versión nueva")
	}
	return errorEvolucion(c, err)
}

// ListarPresupuestos responde GET /api/pacientes/:id/presupuestos.
func (ctrl *PresupuestosController) ListarPresupuestos(c *fiber.Ctx) error {
	id, ok := idDeRuta(c)
	if !ok {
		return errorCliente(c, fiber.StatusBadRequest, "Identificador inválido")
	}
	lista, err := ctrl.repo.ListarPresupuestos(c.Context(), id)
	if err != nil {
		return errorDominio(c, err)
	}
	return c.JSON(lista)
}

// ListarTareas responde GET /api/pacientes/:id/tareas: los ítems aprobados que se pueden agendar.
func (ctrl *PresupuestosController) ListarTareas(c *fiber.Ctx) error {
	id, ok := idDeRuta(c)
	if !ok {
		return errorCliente(c, fiber.StatusBadRequest, "Identificador inválido")
	}
	lista, err := ctrl.repo.ListarTareas(c.Context(), id)
	if err != nil {
		return errorDominio(c, err)
	}
	return c.JSON(lista)
}

// CrearPresupuesto responde POST /api/pacientes/:id/presupuestos: un borrador con sus ítems y totales calculados.
func (ctrl *PresupuestosController) CrearPresupuesto(c *fiber.Ctx) error {
	return ctrl.guardar(c, 0)
}

// ActualizarPresupuesto responde PUT /api/pacientes/:id/presupuestos/:pid: reemplaza los ítems de un borrador.
func (ctrl *PresupuestosController) ActualizarPresupuesto(c *fiber.Ctx) error {
	_, presupuestoID, ok := idsDePresupuesto(c)
	if !ok {
		return errorCliente(c, fiber.StatusBadRequest, "Identificador inválido")
	}
	return ctrl.guardar(c, presupuestoID)
}

func (ctrl *PresupuestosController) guardar(c *fiber.Ctx, presupuestoID int) error {
	pacienteID, ok := idDeRuta(c)
	if !ok {
		return errorCliente(c, fiber.StatusBadRequest, "Identificador inválido")
	}
	var cuerpo cuerpoPresupuesto
	if err := c.BodyParser(&cuerpo); err != nil {
		return cuerpoInvalido(c)
	}
	notas, okNotas := textoOpcional(cuerpo.Notas, 500)
	if !okNotas {
		return errorCliente(c, fiber.StatusBadRequest, "Las notas admiten hasta 500 caracteres")
	}
	if motivo := validarItems(cuerpo.Items); motivo != "" {
		return errorCliente(c, fiber.StatusBadRequest, motivo)
	}
	usuarioID := usuarioDeLaPeticion(c)
	profesional, ok, err := ctrl.repo.ProfesionalDeUsuario(c.Context(), usuarioID)
	if err != nil {
		return errorDominio(c, err)
	}
	if !ok {
		return errorEvolucion(c, models.ErrNoEsProfesional)
	}
	var previo *models.Presupuesto
	if presupuestoID != 0 {
		actual, err := ctrl.repo.ObtenerPresupuesto(c.Context(), pacienteID, presupuestoID)
		if err != nil {
			return errorDominio(c, err)
		}
		if actual.Estado != models.PresupBorrador {
			return errorPresupuesto(c, models.ErrPresupuestoNoEditable)
		}
		previo = &actual
	}
	items, err := ctrl.prepararItems(c.Context(), cuerpo.Items, previo, cuerpo.Autorizacion)
	if err != nil {
		return errorPresupuesto(c, err)
	}
	id, err := ctrl.repo.GuardarPresupuesto(c.Context(), pacienteID, presupuestoID, profesional.ID, notas, items, usuarioID)
	if err != nil {
		return errorPresupuesto(c, err)
	}
	middleware.MarcarAuditoria(c, pacienteID, id)
	guardado, err := ctrl.repo.ObtenerPresupuesto(c.Context(), pacienteID, id)
	if err != nil {
		return errorDominio(c, err)
	}
	estado := fiber.StatusOK
	if presupuestoID == 0 {
		estado = fiber.StatusCreated
	}
	return c.Status(estado).JSON(guardado)
}

// PresentarPresupuesto responde POST /api/pacientes/:id/presupuestos/:pid/presentar.
func (ctrl *PresupuestosController) PresentarPresupuesto(c *fiber.Ctx) error {
	pacienteID, presupuestoID, ok := idsDePresupuesto(c)
	if !ok {
		return errorCliente(c, fiber.StatusBadRequest, "Identificador inválido")
	}
	if err := ctrl.repo.PresentarPresupuesto(c.Context(), pacienteID, presupuestoID, usuarioDeLaPeticion(c)); err != nil {
		return errorPresupuesto(c, err)
	}
	middleware.MarcarAuditoria(c, pacienteID, presupuestoID)
	return ctrl.responder(c, pacienteID, presupuestoID)
}

type cuerpoResolucion struct {
	Aprobados []int `json:"aprobados"`
}

// ResolverPresupuesto responde POST /api/pacientes/:id/presupuestos/:pid/resolver. Los ítems de "aprobados" quedan
// autorizados (y pasan a ser tareas agendables); los demás, rechazados. Todos = aprobado total; algunos = parcial;
// ninguno = rechazado. Queda registrado quién y cuándo.
func (ctrl *PresupuestosController) ResolverPresupuesto(c *fiber.Ctx) error {
	pacienteID, presupuestoID, ok := idsDePresupuesto(c)
	if !ok {
		return errorCliente(c, fiber.StatusBadRequest, "Identificador inválido")
	}
	var cuerpo cuerpoResolucion
	if err := c.BodyParser(&cuerpo); err != nil {
		return cuerpoInvalido(c)
	}
	vistos := map[int]bool{}
	for _, id := range cuerpo.Aprobados {
		if id <= 0 || vistos[id] {
			return errorCliente(c, fiber.StatusBadRequest, "La lista de ítems aprobados no es válida")
		}
		vistos[id] = true
	}
	if cuerpo.Aprobados == nil {
		cuerpo.Aprobados = []int{}
	}
	if err := ctrl.repo.ResolverPresupuesto(c.Context(), pacienteID, presupuestoID, cuerpo.Aprobados, usuarioDeLaPeticion(c)); err != nil {
		return errorPresupuesto(c, err)
	}
	middleware.MarcarAuditoria(c, pacienteID, presupuestoID)
	return ctrl.responder(c, pacienteID, presupuestoID)
}

// NuevaVersion responde POST /api/pacientes/:id/presupuestos/:pid/nueva-version: copia el presupuesto en un borrador.
func (ctrl *PresupuestosController) NuevaVersion(c *fiber.Ctx) error {
	pacienteID, presupuestoID, ok := idsDePresupuesto(c)
	if !ok {
		return errorCliente(c, fiber.StatusBadRequest, "Identificador inválido")
	}
	nuevo, err := ctrl.repo.NuevaVersionPresupuesto(c.Context(), pacienteID, presupuestoID, usuarioDeLaPeticion(c))
	if err != nil {
		return errorPresupuesto(c, err)
	}
	middleware.MarcarAuditoria(c, pacienteID, nuevo)
	creado, err := ctrl.repo.ObtenerPresupuesto(c.Context(), pacienteID, nuevo)
	if err != nil {
		return errorDominio(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(creado)
}

func (ctrl *PresupuestosController) responder(c *fiber.Ctx, pacienteID, presupuestoID int) error {
	p, err := ctrl.repo.ObtenerPresupuesto(c.Context(), pacienteID, presupuestoID)
	if err != nil {
		return errorDominio(c, err)
	}
	return c.JSON(p)
}

// PDFPresupuesto responde GET /api/pacientes/:id/presupuestos/:pid/pdf.
func (ctrl *PresupuestosController) PDFPresupuesto(c *fiber.Ctx) error {
	pacienteID, presupuestoID, ok := idsDePresupuesto(c)
	if !ok {
		return errorCliente(c, fiber.StatusBadRequest, "Identificador inválido")
	}
	presupuesto, err := ctrl.repo.ObtenerPresupuesto(c.Context(), pacienteID, presupuestoID)
	if err != nil {
		return errorDominio(c, err)
	}
	paciente, err := ctrl.repo.ObtenerPaciente(c.Context(), pacienteID)
	if err != nil {
		return errorDominio(c, err)
	}
	zona, errZona := time.LoadLocation(parametroTexto(c.Context(), ctrl.repo, "zona_horaria", "America/Guayaquil"))
	if errZona != nil {
		zona = time.UTC
	}
	contenido, err := documentos.PresupuestoPDF(documentos.DatosPresupuestoPDF{
		Clinica: clinicaDe(c.Context(), ctrl.repo), Presupuesto: presupuesto,
		Paciente: nombreCompleto(paciente.Nombres, paciente.Apellidos), Documento: paciente.Documento,
		Fecha: presupuesto.CreadoEn.In(zona).Format("02/01/2006"),
	})
	if err != nil {
		return errorInterno(c, err)
	}
	middleware.MarcarAuditoria(c, pacienteID, presupuestoID)
	c.Set(fiber.HeaderContentType, "application/pdf")
	c.Set(fiber.HeaderContentDisposition, fmt.Sprintf(`inline; filename="presupuesto-%s-v%d.pdf"`, presupuesto.Numero, presupuesto.Version))
	c.Set(fiber.HeaderCacheControl, "private, no-store")
	return c.Send(contenido)
}
