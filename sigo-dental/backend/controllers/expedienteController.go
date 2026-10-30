// Este archivo atiende los signos vitales y los archivos del paciente (F2-1). El odontólogo y el asistente los
// registran y los consultan; el router fija quién entra a cada ruta.
package controllers

import (
	"context"
	"errors"
	"io"
	"mime"
	"net/http"
	"slices"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"sigo-dental-backend/almacenamiento"
	"sigo-dental-backend/middleware"
	"sigo-dental-backend/models"
)

// ExpedienteRepository es el contrato de la capa de datos de signos vitales y archivos.
type ExpedienteRepository interface {
	ListarSignosVitales(ctx context.Context, pacienteID int) ([]models.SignosVitales, error)
	RegistrarSignosVitales(ctx context.Context, s models.SignosVitales, usuarioID int) (models.SignosVitales, error)
	ListarArchivos(ctx context.Context, pacienteID int) ([]models.Archivo, error)
	ObtenerArchivo(ctx context.Context, id int) (models.Archivo, error)
	ObtenerParametro(ctx context.Context, clave string) (string, error)
}

// ExpedienteController atiende signos vitales y archivos con el repositorio y el almacén inyectados.
type ExpedienteController struct {
	repo     ExpedienteRepository
	archivos *almacenamiento.Servicio
}

// NuevoExpedienteController construye el controlador del expediente.
func NuevoExpedienteController(repo ExpedienteRepository, archivos *almacenamiento.Servicio) *ExpedienteController {
	return &ExpedienteController{repo: repo, archivos: archivos}
}

const (
	maxObservaciones    = 300
	maxDescripcionArch  = 200
	maxNombreArchivo    = 200
	archivoMaxMBDefecto = 15
)

// tiposDeArchivo son los únicos contenidos que se aceptan; se detectan por el contenido, no por lo que declare el cliente.
var tiposDeArchivo = []string{"image/jpeg", "image/png", "image/webp", "application/pdf"}

type cuerpoSignos struct {
	CitaID                 *int     `json:"cita_id"`
	Sistolica              *int     `json:"sistolica"`
	Diastolica             *int     `json:"diastolica"`
	FrecuenciaCardiaca     *int     `json:"frecuencia_cardiaca"`
	FrecuenciaRespiratoria *int     `json:"frecuencia_respiratoria"`
	Temperatura            *float64 `json:"temperatura"`
	Saturacion             *int     `json:"saturacion"`
	PesoKg                 *float64 `json:"peso_kg"`
	TallaCm                *float64 `json:"talla_cm"`
	Observaciones          string   `json:"observaciones"`
}

func entreInt(v *int, min, max int) bool           { return v == nil || (*v >= min && *v <= max) }
func entreFloat(v *float64, min, max float64) bool { return v == nil || (*v >= min && *v <= max) }

// validar comprueba los rangos fisiológicos (los mismos que las restricciones de la base); devuelve el motivo si falla.
func (cuerpo cuerpoSignos) validar() (models.SignosVitales, string) {
	var s models.SignosVitales
	if (cuerpo.Sistolica == nil) != (cuerpo.Diastolica == nil) {
		return s, "La presión arterial se registra completa: sistólica y diastólica"
	}
	if cuerpo.Sistolica == nil && cuerpo.FrecuenciaCardiaca == nil && cuerpo.FrecuenciaRespiratoria == nil &&
		cuerpo.Temperatura == nil && cuerpo.Saturacion == nil && cuerpo.PesoKg == nil && cuerpo.TallaCm == nil {
		return s, "Registra al menos un signo vital"
	}
	switch {
	case !entreInt(cuerpo.Sistolica, 50, 260):
		return s, "La presión sistólica debe estar entre 50 y 260 mmHg"
	case !entreInt(cuerpo.Diastolica, 20, 160):
		return s, "La presión diastólica debe estar entre 20 y 160 mmHg"
	case cuerpo.Sistolica != nil && *cuerpo.Diastolica >= *cuerpo.Sistolica:
		return s, "La presión diastólica debe ser menor que la sistólica"
	case !entreInt(cuerpo.FrecuenciaCardiaca, 20, 250):
		return s, "La frecuencia cardiaca debe estar entre 20 y 250 latidos por minuto"
	case !entreInt(cuerpo.FrecuenciaRespiratoria, 4, 80):
		return s, "La frecuencia respiratoria debe estar entre 4 y 80 respiraciones por minuto"
	case !entreFloat(cuerpo.Temperatura, 30, 45):
		return s, "La temperatura debe estar entre 30 y 45 °C"
	case !entreInt(cuerpo.Saturacion, 50, 100):
		return s, "La saturación de oxígeno debe estar entre 50 y 100 %"
	case !entreFloat(cuerpo.PesoKg, 0.5, 600):
		return s, "El peso debe estar entre 0,5 y 600 kg"
	case !entreFloat(cuerpo.TallaCm, 30, 250):
		return s, "La talla debe estar entre 30 y 250 cm"
	}
	observaciones, ok := textoOpcional(cuerpo.Observaciones, maxObservaciones)
	if !ok {
		return s, "Las observaciones admiten hasta 300 caracteres"
	}
	if cuerpo.CitaID != nil && *cuerpo.CitaID <= 0 {
		return s, "La cita indicada no es válida"
	}
	return models.SignosVitales{
		CitaID: cuerpo.CitaID, Sistolica: cuerpo.Sistolica, Diastolica: cuerpo.Diastolica,
		FrecuenciaCardiaca: cuerpo.FrecuenciaCardiaca, FrecuenciaRespiratoria: cuerpo.FrecuenciaRespiratoria,
		Temperatura: redondear1(cuerpo.Temperatura), Saturacion: cuerpo.Saturacion, PesoKg: redondear1(cuerpo.PesoKg),
		TallaCm: redondear1(cuerpo.TallaCm), Observaciones: observaciones,
	}, ""
}

// redondear1 deja un decimal, que es la precisión de las columnas NUMERIC.
func redondear1(v *float64) *float64 {
	if v == nil {
		return nil
	}
	r := float64(int(*v*10+0.5)) / 10
	return &r
}

// ListarSignosVitales responde GET /api/pacientes/:id/signos-vitales.
func (ctrl *ExpedienteController) ListarSignosVitales(c *fiber.Ctx) error {
	id, ok := idDeRuta(c)
	if !ok {
		return errorCliente(c, fiber.StatusBadRequest, "Identificador inválido")
	}
	lista, err := ctrl.repo.ListarSignosVitales(c.Context(), id)
	if err != nil {
		return errorDominio(c, err)
	}
	return c.JSON(lista)
}

// RegistrarSignosVitales responde POST /api/pacientes/:id/signos-vitales. Cada toma es un registro nuevo.
func (ctrl *ExpedienteController) RegistrarSignosVitales(c *fiber.Ctx) error {
	id, ok := idDeRuta(c)
	if !ok {
		return errorCliente(c, fiber.StatusBadRequest, "Identificador inválido")
	}
	var cuerpo cuerpoSignos
	if err := c.BodyParser(&cuerpo); err != nil {
		return cuerpoInvalido(c)
	}
	signos, motivo := cuerpo.validar()
	if motivo != "" {
		return errorCliente(c, fiber.StatusBadRequest, motivo)
	}
	signos.PacienteID = id
	guardado, err := ctrl.repo.RegistrarSignosVitales(c.Context(), signos, usuarioDeLaPeticion(c))
	if err != nil {
		return errorDominio(c, err)
	}
	middleware.MarcarAuditoria(c, id, guardado.ID)
	return c.Status(fiber.StatusCreated).JSON(guardado)
}

// ListarArchivos responde GET /api/pacientes/:id/archivos.
func (ctrl *ExpedienteController) ListarArchivos(c *fiber.Ctx) error {
	id, ok := idDeRuta(c)
	if !ok {
		return errorCliente(c, fiber.StatusBadRequest, "Identificador inválido")
	}
	lista, err := ctrl.repo.ListarArchivos(c.Context(), id)
	if err != nil {
		return errorDominio(c, err)
	}
	return c.JSON(lista)
}

// SubirArchivo responde POST /api/pacientes/:id/archivos (multipart: archivo, categoria, descripcion). Valida el
// tamaño contra el parámetro archivos_max_mb y el tipo por el contenido real del archivo.
func (ctrl *ExpedienteController) SubirArchivo(c *fiber.Ctx) error {
	id, ok := idDeRuta(c)
	if !ok {
		return errorCliente(c, fiber.StatusBadRequest, "Identificador inválido")
	}
	categoria := c.FormValue("categoria")
	if !slices.Contains(models.CategoriasSubibles, categoria) {
		return errorCliente(c, fiber.StatusBadRequest, "La categoría debe ser radiografía, fotografía o documento")
	}
	descripcion, okDescripcion := textoOpcional(c.FormValue("descripcion"), maxDescripcionArch)
	if !okDescripcion {
		return errorCliente(c, fiber.StatusBadRequest, "La descripción admite hasta 200 caracteres")
	}
	cabecera, err := c.FormFile("archivo")
	if err != nil {
		return errorCliente(c, fiber.StatusBadRequest, "Adjunta el archivo en el campo \"archivo\"")
	}
	limiteMB := parametroEntero(c.Context(), ctrl.repo, "archivos_max_mb", archivoMaxMBDefecto)
	if cabecera.Size <= 0 {
		return errorCliente(c, fiber.StatusBadRequest, "El archivo está vacío")
	}
	if cabecera.Size > int64(limiteMB)*1024*1024 {
		return errorCliente(c, fiber.StatusRequestEntityTooLarge, "El archivo supera el tamaño máximo de "+strconv.Itoa(limiteMB)+" MB")
	}
	origen, err := cabecera.Open()
	if err != nil {
		return errorInterno(c, err)
	}
	defer func() { _ = origen.Close() }()
	// Se lee un byte más del límite: si llega, el tamaño declarado mentía.
	contenido, err := io.ReadAll(io.LimitReader(origen, int64(limiteMB)*1024*1024+1))
	if err != nil {
		return errorInterno(c, err)
	}
	if len(contenido) > limiteMB*1024*1024 {
		return errorCliente(c, fiber.StatusRequestEntityTooLarge, "El archivo supera el tamaño máximo de "+strconv.Itoa(limiteMB)+" MB")
	}
	tipo := http.DetectContentType(contenido)
	if !slices.Contains(tiposDeArchivo, tipo) {
		return errorCliente(c, fiber.StatusUnsupportedMediaType, "Solo se aceptan imágenes JPG, PNG o WebP y documentos PDF")
	}
	nombre := nombreSeguro(cabecera.Filename)

	archivoID, err := ctrl.archivos.Guardar(c.Context(), models.NuevoArchivo{
		PacienteID: id, Categoria: categoria, Descripcion: descripcion, NombreOriginal: nombre, TipoMIME: tipo,
		UsuarioID: usuarioDeLaPeticion(c),
	}, contenido)
	if err != nil {
		return errorDominio(c, err)
	}
	middleware.MarcarAuditoria(c, id, archivoID)
	archivo, err := ctrl.repo.ObtenerArchivo(c.Context(), archivoID)
	if err != nil {
		return errorDominio(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(archivo)
}

// DescargarArchivo responde GET /api/archivos/:id/descargar. Solo con sesión y rol permitido (lo exige el router);
// la respuesta nunca se guarda en cachés compartidas.
func (ctrl *ExpedienteController) DescargarArchivo(c *fiber.Ctx) error {
	id, ok := idDeRuta(c)
	if !ok {
		return errorCliente(c, fiber.StatusBadRequest, "Identificador inválido")
	}
	archivo, err := ctrl.repo.ObtenerArchivo(c.Context(), id)
	if err != nil {
		return errorDominio(c, err)
	}
	return enviarArchivo(c, ctrl.archivos.Almacen(), archivo)
}

// enviarArchivo entrega el contenido de un archivo con cabeceras que impiden que el navegador lo reinterprete.
func enviarArchivo(c *fiber.Ctx, almacen almacenamiento.AlmacenArchivos, archivo models.Archivo) error {
	middleware.MarcarAuditoria(c, archivo.PacienteID, archivo.ID)
	contenido, err := almacen.Abrir(c.Context(), archivo.ClaveAlmacen)
	if errors.Is(err, almacenamiento.ErrNoExiste) {
		return errorCliente(c, fiber.StatusNotFound, "El contenido del archivo no está disponible")
	}
	if err != nil {
		return errorInterno(c, err)
	}
	c.Set(fiber.HeaderContentType, archivo.TipoMIME)
	c.Set(fiber.HeaderContentDisposition, mime.FormatMediaType("attachment", map[string]string{"filename": archivo.NombreOriginal}))
	c.Set("X-Content-Type-Options", "nosniff")
	c.Set(fiber.HeaderCacheControl, "private, no-store")
	return c.SendStream(contenido, archivo.TamanoBytes)
}

// nombreSeguro deja solo el nombre del archivo, sin ruta ni caracteres de control, con un máximo de 200 caracteres.
func nombreSeguro(nombre string) string {
	nombre = strings.ReplaceAll(nombre, "\\", "/")
	if i := strings.LastIndex(nombre, "/"); i >= 0 {
		nombre = nombre[i+1:]
	}
	nombre = strings.Map(func(r rune) rune {
		if r < 32 || r == 127 {
			return -1
		}
		return r
	}, nombre)
	nombre = strings.TrimSpace(nombre)
	if nombre == "" {
		return "archivo"
	}
	if r := []rune(nombre); len(r) > maxNombreArchivo {
		nombre = string(r[:maxNombreArchivo])
	}
	return nombre
}
