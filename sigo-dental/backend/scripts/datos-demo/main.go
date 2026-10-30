// Comando datos-demo: carga datos de demostración de las Fases 1 y 2 llamando a la API real (no con INSERT directos),
// para que los disparadores, las firmas con su huella SHA-256, la numeración y la auditoría funcionen de verdad.
//
// Uso (desde backend/): API_URL=http://localhost:3001 go run ./scripts/datos-demo
// Variables: API_URL (por defecto http://localhost:3001), DEMO_PASSWORD (la de las cuentas de 900_datos_prueba.sql) y
// DEMO_FORZAR=1 para cargar aunque ya haya datos de demostración. Solo añade datos: no borra ni cambia lo existente.
// Requiere los pacientes y el catálogo de 901_ y 902_ y el adaptador de mensajes simulado (el código OTP se lee de la
// bitácora de notificaciones).
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"os"
	"regexp"
	"strings"
	"time"
	// Zonas horarias incrustadas: la clínica trabaja en America/Guayaquil aunque el sistema no las tenga (Windows).
	_ "time/tzdata"
)

const marca = "[Demo]"

type cliente struct {
	base     string
	password string
	tokens   map[string]string
	http     *http.Client
}

type errorAPI struct {
	Estado  int
	Mensaje string
}

func (e errorAPI) Error() string { return fmt.Sprintf("HTTP %d: %s", e.Estado, e.Mensaje) }

// cuentas por rol: el correo de 900_datos_prueba.sql.
var cuentas = map[string]string{
	"recepcion": "recepcion@sigo-dental.test", "odontologo": "odontologo@sigo-dental.test",
	"asistente": "asistente@sigo-dental.test", "auditor": "auditor@sigo-dental.test", "admin": "admin@sigo-dental.test",
}

type mapa = map[string]any

func (c *cliente) token(rol string) (string, error) {
	if t, ok := c.tokens[rol]; ok {
		return t, nil
	}
	var r struct {
		Token string `json:"token"`
	}
	if err := c.enviar("", "POST", "/api/login", mapa{"email": cuentas[rol], "password": c.password}, &r); err != nil {
		return "", fmt.Errorf("no se pudo iniciar sesión como %s: %w", rol, err)
	}
	c.tokens[rol] = r.Token
	return r.Token, nil
}

// enviar hace una petición JSON; con rol vacío va sin sesión. Un 4xx o 5xx se devuelve como errorAPI.
func (c *cliente) enviar(rol, metodo, ruta string, cuerpo any, salida any) error {
	var lector io.Reader
	if cuerpo != nil {
		datos, _ := json.Marshal(cuerpo)
		lector = bytes.NewReader(datos)
	}
	return c.peticion(rol, metodo, ruta, lector, "application/json", salida)
}

func (c *cliente) peticion(rol, metodo, ruta string, cuerpo io.Reader, tipo string, salida any) error {
	req, err := http.NewRequest(metodo, c.base+ruta, cuerpo)
	if err != nil {
		return err
	}
	if cuerpo != nil {
		req.Header.Set("Content-Type", tipo)
	}
	if rol != "" {
		t, err := c.token(rol)
		if err != nil {
			return err
		}
		req.Header.Set("Authorization", "Bearer "+t)
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	datos, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		var e struct {
			Message string `json:"message"`
		}
		_ = json.Unmarshal(datos, &e)
		if e.Message == "" {
			e.Message = strings.TrimSpace(string(datos))
		}
		return errorAPI{resp.StatusCode, e.Message}
	}
	if salida != nil && len(datos) > 0 {
		return json.Unmarshal(datos, salida)
	}
	return nil
}

// --- Seguimiento de la carga ---

var (
	fallos  []string
	cargado []string
)

// paso ejecuta un bloque; si falla lo anota y la carga sigue con lo demás.
func paso(nombre string, f func() (string, error)) {
	resumen, err := f()
	if err != nil {
		fmt.Printf("  ✗ %s: %v\n", nombre, err)
		fallos = append(fallos, nombre+": "+err.Error())
		return
	}
	fmt.Printf("  ✓ %s: %s\n", nombre, resumen)
	cargado = append(cargado, nombre+": "+resumen)
}

type id struct {
	ID int `json:"id"`
}

func main() {
	api := strings.TrimRight(os.Getenv("API_URL"), "/")
	if api == "" {
		api = "http://localhost:3001"
	}
	password := os.Getenv("DEMO_PASSWORD")
	if password == "" {
		password = "Sigo-Dental-2026"
	}
	c := &cliente{base: api, password: password, tokens: map[string]string{}, http: &http.Client{Timeout: 30 * time.Second}}
	fmt.Println("Datos de demostración de SIGO-Dental →", api)

	// Un servidor gratuito que duerme tarda cerca de un minuto en despertar: se insiste antes de rendirse.
	var errSalud error
	for intento := 1; intento <= 8; intento++ {
		if errSalud = c.enviar("", "GET", "/api/health", nil, nil); errSalud == nil {
			break
		}
		fmt.Printf("  esperando a la API (intento %d de 8): %v\n", intento, errSalud)
		time.Sleep(10 * time.Second)
	}
	if errSalud != nil {
		fmt.Println("La API no responde:", errSalud)
		os.Exit(1)
	}
	d, err := preparar(c)
	if err != nil {
		fmt.Println("No se pudo preparar la carga:", err)
		os.Exit(1)
	}
	forzar := os.Getenv("DEMO_FORZAR") == "1"
	// Cada bloque tiene su propia marca: el de Carlos y el de Ana se cargan por separado y no se duplican.
	if forzar || !d.yaCargado(c, d.p(2)) {
		d.cargar(c)
	} else {
		fmt.Println("Ya están los datos de demostración de Carlos (nota con la marca " + marca + "): se omiten.")
	}
	if forzar || !d.yaCargado(c, d.p(1)) {
		d.cargarAna(c)
	} else {
		fmt.Println("Ya están los datos de demostración de Ana (nota con la marca " + marca + "): se omiten.")
	}

	var aud struct {
		Total int `json:"total"`
	}
	if err := c.enviar("auditor", "GET", "/api/auditoria?limite=1", nil, &aud); err == nil {
		fmt.Printf("\nAuditoría: %d registros en la bitácora (se llenó sola).\n", aud.Total)
	}
	fmt.Printf("Carga terminada: %d bloques cargados, %d con problemas.\n", len(cargado), len(fallos))
	if len(fallos) > 0 {
		os.Exit(1)
	}
}

// --- Datos de apoyo ---

type demo struct {
	pacientes map[string]int // por número de historia 1..4 → id
	sillon    int
	prof      int
	proc      map[int]int // id del catálogo (1..5) → id real por orden
	hoy       time.Time
	zona      *time.Location
	hora      int // contador de huecos de 30 minutos en la agenda de la demo
}

func preparar(c *cliente) (*demo, error) {
	d := &demo{pacientes: map[string]int{}, proc: map[int]int{}}
	var err error
	if d.zona, err = time.LoadLocation("America/Guayaquil"); err != nil {
		return nil, err
	}
	d.hoy = time.Now().In(d.zona)
	var lista []struct {
		ID        int    `json:"id"`
		Documento string `json:"documento"`
	}
	if err := c.enviar("odontologo", "GET", "/api/pacientes", nil, &lista); err != nil {
		return nil, err
	}
	for _, p := range lista {
		d.pacientes[p.Documento] = p.ID
	}
	for _, doc := range []string{"0900000001", "0900000002", "0900000003", "0900000004"} {
		if d.pacientes[doc] == 0 {
			return nil, fmt.Errorf("falta el paciente con documento %s (902_pacientes_prueba.sql)", doc)
		}
	}
	var sillones []id
	if err := c.enviar("recepcion", "GET", "/api/sillones", nil, &sillones); err != nil || len(sillones) == 0 {
		return nil, fmt.Errorf("no hay sillones: %v", err)
	}
	d.sillon = sillones[len(sillones)-1].ID
	var profs []id
	if err := c.enviar("recepcion", "GET", "/api/profesionales", nil, &profs); err != nil || len(profs) == 0 {
		return nil, fmt.Errorf("no hay profesionales: %v", err)
	}
	d.prof = profs[0].ID
	var procs []struct {
		ID     int    `json:"id"`
		Nombre string `json:"nombre"`
	}
	if err := c.enviar("recepcion", "GET", "/api/procedimientos", nil, &procs); err != nil {
		return nil, err
	}
	for _, p := range procs {
		switch p.Nombre {
		case "Consulta y valoración":
			d.proc[1] = p.ID
		case "Limpieza dental":
			d.proc[2] = p.ID
		case "Restauración con resina":
			d.proc[3] = p.ID
		case "Extracción simple":
			d.proc[4] = p.ID
		case "Endodoncia unirradicular":
			d.proc[5] = p.ID
		}
	}
	if len(d.proc) < 5 {
		return nil, fmt.Errorf("falta parte del catálogo de procedimientos (901_catalogo_prueba.sql)")
	}
	return d, nil
}

func (d *demo) p(n int) int { return d.pacientes[fmt.Sprintf("090000000%d", n)] }

// yaCargado indica si el paciente ya tiene una nota de la demostración (la marca del bloque).
func (d *demo) yaCargado(c *cliente, paciente int) bool {
	var notas []struct {
		Subjetivo string `json:"subjetivo"`
	}
	if err := c.enviar("odontologo", "GET", fmt.Sprintf("/api/pacientes/%d/notas", paciente), nil, &notas); err != nil {
		return false
	}
	for _, n := range notas {
		if strings.HasPrefix(n.Subjetivo, marca) {
			return true
		}
	}
	return false
}

// hueco devuelve el inicio y el fin de la siguiente cita de la demo: 30 minutos, a partir de las 08:00 de dentro de
// 5 días en la zona de la clínica, para no chocar con las citas que ya existen.
func (d *demo) hueco() (string, string) {
	dia := d.hoy.AddDate(0, 0, 5)
	inicio := time.Date(dia.Year(), dia.Month(), dia.Day(), 8, 0, 0, 0, d.zona).Add(time.Duration(d.hora) * 30 * time.Minute)
	d.hora++
	return inicio.UTC().Format(time.RFC3339), inicio.Add(30 * time.Minute).UTC().Format(time.RFC3339)
}

// cita agenda una cita y la lleva por los estados indicados.
func (d *demo) cita(c *cliente, paciente, procedimiento int, motivo string, estados ...[2]string) (int, error) {
	inicio, fin := d.hueco()
	cuerpo := mapa{"paciente_id": paciente, "profesional_id": d.prof, "sillon_id": d.sillon, "motivo": motivo, "inicio": inicio, "fin": fin}
	if procedimiento != 0 {
		cuerpo["procedimiento_id"] = procedimiento
	}
	var creada id
	if err := c.enviar("recepcion", "POST", "/api/citas", cuerpo, &creada); err != nil {
		return 0, err
	}
	for _, e := range estados {
		if err := c.enviar("recepcion", "POST", fmt.Sprintf("/api/citas/%d/estado", creada.ID), mapa{"estado": e[0], "motivo": e[1]}, nil); err != nil {
			return creada.ID, fmt.Errorf("cita %d → %s: %w", creada.ID, e[0], err)
		}
	}
	return creada.ID, nil
}

func est(estado string) [2]string { return [2]string{estado, ""} }

var reCodigo = regexp.MustCompile(`es (\d{6})\.`)

// codigoDe lee de la bitácora de notificaciones el último código enviado del tipo dado (solo con el adaptador simulado).
func codigoDe(c *cliente, paciente int, tipo string) (string, error) {
	var lista []struct {
		Tipo    string `json:"tipo"`
		Mensaje string `json:"mensaje"`
	}
	if err := c.enviar("odontologo", "GET", fmt.Sprintf("/api/pacientes/%d/notificaciones", paciente), nil, &lista); err != nil {
		return "", err
	}
	for _, n := range lista { // la bitácora viene del más reciente al más antiguo
		if n.Tipo == tipo {
			if m := reCodigo.FindStringSubmatch(n.Mensaje); m != nil {
				return m[1], nil
			}
			return "", fmt.Errorf("el código está enmascarado: el adaptador de mensajes no es el simulado")
		}
	}
	return "", fmt.Errorf("no hay un código en la bitácora")
}

func firmaDePrueba() mapa {
	puntos := []mapa{}
	for i := 0; i < 45; i++ {
		puntos = append(puntos, mapa{"x": 30 + float64(i)*11, "y": 110 + math.Sin(float64(i)/3)*45, "t": i * 18})
	}
	return mapa{"ancho": 600, "alto": 220, "trazos": []any{puntos}}
}

const pdfMinimo = "%PDF-1.4\n1 0 obj\n<< /Type /Catalog >>\nendobj\ntrailer\n<< /Root 1 0 R >>\n%%EOF\n"

var pngMinimo = []byte{0x89, 'P', 'N', 'G', 0x0d, 0x0a, 0x1a, 0x0a, 0, 0, 0, 0x0d, 'I', 'H', 'D', 'R', 0, 0, 0, 1, 0, 0, 0, 1, 8, 6, 0, 0, 0, 0x1f, 0x15, 0xc4, 0x89}

func subir(c *cliente, paciente int, categoria, descripcion, nombre string, contenido []byte) error {
	var cuerpo bytes.Buffer
	escritor := multipart.NewWriter(&cuerpo)
	_ = escritor.WriteField("categoria", categoria)
	_ = escritor.WriteField("descripcion", descripcion)
	cabecera := make(textproto.MIMEHeader)
	cabecera.Set("Content-Disposition", fmt.Sprintf(`form-data; name="archivo"; filename="%s"`, nombre))
	parte, _ := escritor.CreatePart(cabecera)
	_, _ = parte.Write(contenido)
	_ = escritor.Close()
	return c.peticion("odontologo", "POST", fmt.Sprintf("/api/pacientes/%d/archivos", paciente), &cuerpo, escritor.FormDataContentType(), nil)
}

func sitios(prof, margen [6]int, sangra [6]bool) []mapa {
	nombres := []string{"DV", "V", "MV", "DL", "L", "ML"}
	r := []mapa{}
	for i, n := range nombres {
		r = append(r, mapa{"sitio": n, "profundidad": prof[i], "margen": margen[i], "sangrado": sangra[i]})
	}
	return r
}

func ruta(paciente int, resto string) string {
	return fmt.Sprintf("/api/pacientes/%d%s", paciente, resto)
}
