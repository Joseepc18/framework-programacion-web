// Pruebas de signos vitales, archivos y consulta de la auditoría con repositorios falsos.
package controllers

import (
	"bytes"
	"context"
	"errors"
	"io"
	"mime/multipart"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"sigo-dental-backend/almacenamiento"
	"sigo-dental-backend/models"
)

type repoExpediente struct {
	ExpedienteRepository
	err      error
	signos   models.SignosVitales
	archivo  models.Archivo
	limiteMB string
}

func (r *repoExpediente) RegistrarSignosVitales(_ context.Context, s models.SignosVitales, _ int) (models.SignosVitales, error) {
	r.signos = s
	s.ID = 3
	return s, r.err
}

func (r *repoExpediente) ObtenerParametro(_ context.Context, _ string) (string, error) {
	if r.limiteMB == "" {
		return "", models.ErrNoEncontrado
	}
	return r.limiteMB, nil
}

func (r *repoExpediente) RegistrarArchivo(_ context.Context, a models.NuevoArchivo) (int, error) {
	r.archivo = models.Archivo{ID: 8, PacienteID: a.PacienteID, Categoria: a.Categoria, NombreOriginal: a.NombreOriginal, TipoMIME: a.TipoMIME}
	return 8, r.err
}

func (r *repoExpediente) ObtenerArchivo(_ context.Context, id int) (models.Archivo, error) {
	if r.err != nil {
		return models.Archivo{}, r.err
	}
	a := r.archivo
	a.ID = id
	return a, nil
}

type almacenEnMemoria struct{ datos map[string][]byte }

func (a *almacenEnMemoria) Guardar(_ context.Context, clave string, c io.Reader, _ int64, _ string) error {
	b, err := io.ReadAll(c)
	a.datos[clave] = b
	return err
}

func (a *almacenEnMemoria) Abrir(_ context.Context, clave string) (io.ReadCloser, error) {
	b, ok := a.datos[clave]
	if !ok {
		return nil, almacenamiento.ErrNoExiste
	}
	return io.NopCloser(bytes.NewReader(b)), nil
}

func appExpediente(repo *repoExpediente, almacen *almacenEnMemoria) *fiber.App {
	ctrl := NuevoExpedienteController(repo, almacenamiento.NuevoServicio(almacen, repo))
	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("rol", models.RolAsistente)
		c.Locals("usuario_id", 1)
		return c.Next()
	})
	app.Post("/api/pacientes/:id/signos-vitales", ctrl.RegistrarSignosVitales)
	app.Post("/api/pacientes/:id/archivos", ctrl.SubirArchivo)
	app.Get("/api/archivos/:id/descargar", ctrl.DescargarArchivo)
	return app
}

func TestRegistrarSignosVitales(t *testing.T) {
	casos := []struct {
		nombre, cuerpo string
		err            error
		estado         int
	}{
		{"toma completa", `{"sistolica":120,"diastolica":80,"frecuencia_cardiaca":72,"frecuencia_respiratoria":16,"temperatura":36.6,"saturacion":98,"peso_kg":70.5,"talla_cm":172}`, nil, 201},
		{"solo la temperatura", `{"temperatura":37.2}`, nil, 201},
		{"sin ningún signo", `{"observaciones":"nada"}`, nil, 400},
		{"presión incompleta", `{"sistolica":120}`, nil, 400},
		{"diastólica mayor que la sistólica", `{"sistolica":80,"diastolica":120}`, nil, 400},
		{"sistólica fuera de rango", `{"sistolica":400,"diastolica":80}`, nil, 400},
		{"frecuencia cardiaca imposible", `{"frecuencia_cardiaca":3}`, nil, 400},
		{"temperatura fuera de rango", `{"temperatura":50}`, nil, 400},
		{"saturación sobre 100", `{"saturacion":101}`, nil, 400},
		{"peso negativo", `{"peso_kg":-4}`, nil, 400},
		{"observaciones demasiado largas", `{"temperatura":36,"observaciones":"` + strings.Repeat("a", 301) + `"}`, nil, 400},
		{"cita inválida", `{"temperatura":36,"cita_id":0}`, nil, 400},
		{"cita de otro paciente", `{"temperatura":36,"cita_id":5}`, models.ErrReferencia, 400},
		{"paciente inexistente", `{"temperatura":36}`, models.ErrNoEncontrado, 404},
		{"cuerpo que no es JSON", `no-es-json`, nil, 400},
	}
	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			app := appExpediente(&repoExpediente{err: caso.err}, &almacenEnMemoria{datos: map[string][]byte{}})
			if estado := enviar(t, app, "POST", "/api/pacientes/4/signos-vitales", caso.cuerpo); estado != caso.estado {
				t.Errorf("estado = %d, se esperaba %d", estado, caso.estado)
			}
		})
	}
}

func TestSignosVitalesRedondeaAUnDecimal(t *testing.T) {
	repo := &repoExpediente{}
	app := appExpediente(repo, &almacenEnMemoria{datos: map[string][]byte{}})
	if estado := enviar(t, app, "POST", "/api/pacientes/4/signos-vitales", `{"temperatura":36.64}`); estado != 201 {
		t.Fatalf("estado = %d", estado)
	}
	if repo.signos.Temperatura == nil || *repo.signos.Temperatura != 36.6 {
		t.Errorf("temperatura guardada = %v", repo.signos.Temperatura)
	}
}

// pngMinimo es una imagen PNG válida de 1x1 píxel.
var pngMinimo = []byte{0x89, 'P', 'N', 'G', 0x0d, 0x0a, 0x1a, 0x0a, 0, 0, 0, 0x0d, 'I', 'H', 'D', 'R', 0, 0, 0, 1, 0, 0, 0, 1, 8, 6, 0, 0, 0, 0x1f, 0x15, 0xc4, 0x89}

func subir(t *testing.T, app *fiber.App, categoria, nombre string, contenido []byte) int {
	t.Helper()
	var cuerpo bytes.Buffer
	escritor := multipart.NewWriter(&cuerpo)
	_ = escritor.WriteField("categoria", categoria)
	_ = escritor.WriteField("descripcion", "prueba")
	parte, _ := escritor.CreateFormFile("archivo", nombre)
	_, _ = parte.Write(contenido)
	_ = escritor.Close()
	req := httptest.NewRequest("POST", "/api/pacientes/4/archivos", &cuerpo)
	req.Header.Set("Content-Type", escritor.FormDataContentType())
	resp, err := app.Test(req, 5000)
	if err != nil {
		t.Fatal(err)
	}
	return resp.StatusCode
}

func TestSubirArchivoValidaTipoCategoriaYTamano(t *testing.T) {
	pdf := []byte("%PDF-1.4\n1 0 obj\n<<>>\nendobj\n")
	casos := []struct {
		nombre, categoria, archivo string
		contenido                  []byte
		limiteMB                   string
		estado                     int
	}{
		{"imagen PNG", "radiografia", "rx.png", pngMinimo, "", 201},
		{"documento PDF", "documento", "informe.pdf", pdf, "", 201},
		{"el nombre con ruta se limpia", "fotografia", "../../etc/passwd.png", pngMinimo, "", 201},
		{"un ejecutable con extensión de imagen se rechaza por su contenido", "fotografia", "virus.png", []byte("MZ\x90\x00\x03\x00\x00\x00"), "", 415},
		{"texto plano", "documento", "notas.pdf", []byte("hola"), "", 415},
		{"categoría que genera el sistema", "consentimiento", "x.png", pngMinimo, "", 400},
		{"categoría inexistente", "otra", "x.png", pngMinimo, "", 400},
		{"archivo vacío", "documento", "vacio.pdf", nil, "", 400},
		{"supera el límite configurado", "documento", "grande.pdf", append(pdf, make([]byte, 1024*1024+1)...), "1", 413},
	}
	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			almacen := &almacenEnMemoria{datos: map[string][]byte{}}
			repo := &repoExpediente{limiteMB: caso.limiteMB}
			app := appExpediente(repo, almacen)
			if estado := subir(t, app, caso.categoria, caso.archivo, caso.contenido); estado != caso.estado {
				t.Fatalf("estado = %d, se esperaba %d", estado, caso.estado)
			}
			if caso.estado == 201 && len(almacen.datos) != 1 {
				t.Errorf("el almacén tiene %d archivos, se esperaba 1", len(almacen.datos))
			}
			if caso.estado != 201 && len(almacen.datos) != 0 {
				t.Error("se guardó un archivo que debía rechazarse")
			}
			if strings.Contains(repo.archivo.NombreOriginal, "/") {
				t.Errorf("el nombre conserva la ruta: %q", repo.archivo.NombreOriginal)
			}
		})
	}
}

func TestDescargarArchivoEntregaElContenidoSinCache(t *testing.T) {
	almacen := &almacenEnMemoria{datos: map[string][]byte{"pacientes/4/a.png": pngMinimo}}
	repo := &repoExpediente{archivo: models.Archivo{PacienteID: 4, NombreOriginal: "rx.png", TipoMIME: "image/png", TamanoBytes: len(pngMinimo), ClaveAlmacen: "pacientes/4/a.png"}}
	app := appExpediente(repo, almacen)
	resp, err := app.Test(httptest.NewRequest("GET", "/api/archivos/8/descargar", nil))
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 200 || resp.Header.Get("Content-Type") != "image/png" {
		t.Fatalf("respuesta = %d %q", resp.StatusCode, resp.Header.Get("Content-Type"))
	}
	if resp.Header.Get("Cache-Control") != "private, no-store" || resp.Header.Get("X-Content-Type-Options") != "nosniff" {
		t.Errorf("faltan cabeceras de seguridad: %v", resp.Header)
	}
	cuerpo, _ := io.ReadAll(resp.Body)
	if !bytes.Equal(cuerpo, pngMinimo) {
		t.Error("el contenido descargado no coincide")
	}
}

func TestDescargarArchivoInexistente(t *testing.T) {
	app := appExpediente(&repoExpediente{err: models.ErrNoEncontrado}, &almacenEnMemoria{datos: map[string][]byte{}})
	resp, _ := app.Test(httptest.NewRequest("GET", "/api/archivos/99/descargar", nil))
	if resp.StatusCode != 404 {
		t.Errorf("estado = %d", resp.StatusCode)
	}
}

type repoAuditoria struct {
	AuditoriaRepository
	filtro models.FiltroAuditoria
	err    error
}

func (r *repoAuditoria) ListarAuditoria(_ context.Context, f models.FiltroAuditoria) (models.PaginaAuditoria, error) {
	r.filtro = f
	return models.PaginaAuditoria{Registros: []models.RegistroAuditoria{}}, r.err
}

func TestListarAuditoriaValidaLosFiltros(t *testing.T) {
	casos := []struct {
		nombre, consulta string
		estado           int
	}{
		{"sin filtros", "", 200},
		{"con todos los filtros", "?usuario=ana&paciente=perez&accion=consulta&desde=2026-10-01T00:00:00Z&hasta=2026-10-31T00:00:00Z&limite=20&pagina=2", 200},
		{"acción inexistente", "?accion=borrar", 400},
		{"fecha mal escrita", "?desde=ayer", 400},
		{"rango invertido", "?desde=2026-10-31T00:00:00Z&hasta=2026-10-01T00:00:00Z", 400},
		{"límite excesivo", "?limite=5000", 400},
		{"página cero", "?pagina=0", 400},
		{"filtro de texto demasiado largo", "?usuario=" + strings.Repeat("a", 101), 400},
	}
	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			app := fiber.New()
			app.Get("/api/auditoria", NuevoAuditoriaController(&repoAuditoria{}).ListarAuditoria)
			resp, err := app.Test(httptest.NewRequest("GET", "/api/auditoria"+caso.consulta, nil))
			if err != nil {
				t.Fatal(err)
			}
			if resp.StatusCode != caso.estado {
				t.Errorf("estado = %d, se esperaba %d", resp.StatusCode, caso.estado)
			}
		})
	}
}

func TestListarAuditoriaPasaLosFiltrosAlRepositorio(t *testing.T) {
	repo := &repoAuditoria{}
	app := fiber.New()
	app.Get("/api/auditoria", NuevoAuditoriaController(repo).ListarAuditoria)
	if _, err := app.Test(httptest.NewRequest("GET", "/api/auditoria?usuario=ana&accion=exportacion&desde=2026-10-01T05:00:00Z&limite=20&pagina=3", nil)); err != nil {
		t.Fatal(err)
	}
	f := repo.filtro
	if f.Usuario != "ana" || f.Accion != "exportacion" || f.Limite != 20 || f.Pagina != 3 {
		t.Errorf("filtro = %+v", f)
	}
	if f.Desde == nil || !f.Desde.Equal(time.Date(2026, 10, 1, 5, 0, 0, 0, time.UTC)) || f.Hasta != nil {
		t.Errorf("fechas = %v %v", f.Desde, f.Hasta)
	}
}

func TestListarAuditoriaFalloInterno(t *testing.T) {
	app := fiber.New()
	app.Get("/api/auditoria", NuevoAuditoriaController(&repoAuditoria{err: errors.New("caída")}).ListarAuditoria)
	resp, _ := app.Test(httptest.NewRequest("GET", "/api/auditoria", nil))
	if resp.StatusCode != 500 {
		t.Errorf("estado = %d", resp.StatusCode)
	}
}
