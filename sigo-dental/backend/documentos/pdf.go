// Paquete documentos: genera en el servidor los PDF de los documentos clínicos (consentimientos, recetas y
// presupuestos) con el encabezado de la clínica.
package documentos

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/go-pdf/fpdf"
	"sigo-dental-backend/models"
)

// Clinica son los datos del encabezado de los documentos (parámetros del sistema).
type Clinica struct {
	Nombre    string
	Direccion string
	Telefono  string
}

// documento envuelve fpdf con el traductor a la codificación de las fuentes básicas (cubre el español).
type documento struct {
	pdf *fpdf.Fpdf
	tr  func(string) string
}

func nuevoDocumento(c Clinica, titulo string) *documento {
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(18, 18, 18)
	pdf.SetAutoPageBreak(true, 20)
	d := &documento{pdf: pdf, tr: pdf.UnicodeTranslatorFromDescriptor("")}
	pdf.SetTitle(d.tr(titulo), true)
	pdf.SetCreator("SIGO-Dental", true)
	pdf.SetFooterFunc(func() {
		pdf.SetY(-14)
		pdf.SetFont("Helvetica", "", 8)
		pdf.SetTextColor(110, 110, 110)
		pdf.CellFormat(0, 5, d.tr(fmt.Sprintf("%s · Página %d de {nb}", c.Nombre, pdf.PageNo())), "", 0, "C", false, 0, "")
	})
	pdf.AliasNbPages("{nb}")
	pdf.AddPage()
	d.encabezado(c)
	return d
}

func (d *documento) encabezado(c Clinica) {
	d.pdf.SetFont("Helvetica", "B", 15)
	d.pdf.SetTextColor(20, 20, 20)
	d.pdf.CellFormat(0, 7, d.tr(c.Nombre), "", 1, "L", false, 0, "")
	d.pdf.SetFont("Helvetica", "", 9)
	d.pdf.SetTextColor(90, 90, 90)
	if linea := strings.TrimSpace(strings.Join(nonEmpty(c.Direccion, c.Telefono), " · ")); linea != "" {
		d.pdf.CellFormat(0, 5, d.tr(linea), "", 1, "L", false, 0, "")
	}
	d.pdf.SetDrawColor(180, 180, 180)
	d.pdf.SetLineWidth(0.3)
	d.pdf.Line(18, d.pdf.GetY()+2, 192, d.pdf.GetY()+2)
	d.pdf.Ln(6)
	d.pdf.SetTextColor(20, 20, 20)
}

func nonEmpty(valores ...string) []string {
	var r []string
	for _, v := range valores {
		if strings.TrimSpace(v) != "" {
			r = append(r, v)
		}
	}
	return r
}

func (d *documento) titulo(texto string) {
	d.pdf.SetFont("Helvetica", "B", 13)
	d.pdf.MultiCell(0, 6.5, d.tr(texto), "", "L", false)
	d.pdf.Ln(2)
}

func (d *documento) subtitulo(texto string) {
	d.pdf.SetFont("Helvetica", "B", 10)
	d.pdf.CellFormat(0, 6, d.tr(texto), "", 1, "L", false, 0, "")
}

func (d *documento) parrafo(texto string) {
	d.pdf.SetFont("Helvetica", "", 10.5)
	d.pdf.MultiCell(0, 5.2, d.tr(texto), "", "L", false)
	d.pdf.Ln(1.5)
}

func (d *documento) dato(etiqueta, valor string) {
	d.pdf.SetFont("Helvetica", "B", 9.5)
	d.pdf.CellFormat(42, 5.2, d.tr(etiqueta), "", 0, "L", false, 0, "")
	d.pdf.SetFont("Helvetica", "", 9.5)
	d.pdf.MultiCell(0, 5.2, d.tr(valor), "", "L", false)
}

func (d *documento) bytes() ([]byte, error) {
	var buf bytes.Buffer
	if err := d.pdf.Output(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// DatosConsentimientoPDF es lo necesario para dibujar el consentimiento firmado.
type DatosConsentimientoPDF struct {
	Clinica           Clinica
	Consentimiento    models.Consentimiento
	Paciente          string
	FirmadoEn         string
	Trazo             *models.FirmaManuscrita
	CodigoOTPEnviadoA string
}

// ConsentimientoPDF dibuja el consentimiento (o la revocación) firmado: el texto tal cual se mostró, quién firmó, la
// modalidad (trazo manuscrito o código OTP), la fecha, el dispositivo y la huella del texto.
func ConsentimientoPDF(datos DatosConsentimientoPDF) ([]byte, error) {
	c := datos.Consentimiento
	d := nuevoDocumento(datos.Clinica, c.Titulo)
	d.titulo(c.Titulo)
	d.dato("Paciente:", datos.Paciente)
	d.dato("Procedimiento:", c.ProcedimientoNombre)
	d.dato("Profesional:", c.ProfesionalNombre)
	if c.Tipo == models.ConsentTipoRevocacion && c.Motivo != "" {
		d.dato("Motivo:", c.Motivo)
	}
	d.pdf.Ln(3)
	d.parrafo(c.Texto)
	d.pdf.Ln(4)

	quien := "el paciente"
	if c.FirmanteTipo == models.FirmanteTutor {
		quien = "su tutor legal"
	}
	d.subtitulo(fmt.Sprintf("Firma de %s", quien))
	d.dato("Nombre:", c.FirmanteNombre)
	d.dato("Documento:", c.FirmanteDocumento)
	d.dato("Fecha y hora:", datos.FirmadoEn)

	if c.Modalidad == models.ModalidadManuscrita && datos.Trazo != nil {
		d.dato("Modalidad:", "Firma manuscrita en pantalla táctil")
		d.dibujarTrazo(*datos.Trazo)
		d.dato("Dispositivo:", c.Dispositivo)
	} else {
		d.dato("Modalidad:", "Validación con código de un solo uso (OTP) enviado al "+quien)
		if datos.CodigoOTPEnviadoA != "" {
			d.dato("Enviado a:", datos.CodigoOTPEnviadoA)
		}
		d.dato("Dispositivo:", c.Dispositivo)
	}
	d.pdf.Ln(2)
	d.pdf.SetFont("Courier", "", 8)
	d.pdf.SetTextColor(100, 100, 100)
	d.pdf.MultiCell(0, 4, d.tr("Huella SHA-256 del texto firmado: "+c.HuellaTexto), "", "L", false)
	return d.bytes()
}

// dibujarTrazo dibuja la firma manuscrita escalada a un recuadro de 90 x 32 mm, conservando la proporción.
func (d *documento) dibujarTrazo(f models.FirmaManuscrita) {
	const anchoMax, altoMax = 90.0, 32.0
	x0, y0 := d.pdf.GetX()+2, d.pdf.GetY()+2
	if y0+altoMax+6 > 275 {
		d.pdf.AddPage()
		y0 = d.pdf.GetY() + 2
	}
	escala := min(anchoMax/f.Ancho, altoMax/f.Alto)
	d.pdf.SetDrawColor(30, 30, 30)
	d.pdf.SetLineWidth(0.45)
	d.pdf.SetLineCapStyle("round")
	for _, trazo := range f.Trazos {
		for i := 1; i < len(trazo); i++ {
			d.pdf.Line(x0+trazo[i-1].X*escala, y0+trazo[i-1].Y*escala, x0+trazo[i].X*escala, y0+trazo[i].Y*escala)
		}
		if len(trazo) == 1 {
			d.pdf.Line(x0+trazo[0].X*escala, y0+trazo[0].Y*escala, x0+trazo[0].X*escala+0.1, y0+trazo[0].Y*escala)
		}
	}
	d.pdf.SetDrawColor(180, 180, 180)
	d.pdf.SetLineWidth(0.2)
	d.pdf.Rect(x0-1, y0-1, anchoMax+2, altoMax+2, "D")
	d.pdf.SetY(y0 + altoMax + 4)
	d.pdf.SetX(18)
}
