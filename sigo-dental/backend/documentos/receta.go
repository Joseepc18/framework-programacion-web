package documentos

import (
	"fmt"

	"sigo-dental-backend/models"
)

// DatosRecetaPDF es lo necesario para dibujar una receta.
type DatosRecetaPDF struct {
	Clinica   Clinica
	Receta    models.Receta
	Paciente  string
	Documento string
	Fecha     string
	Firmada   string
}

// RecetaPDF dibuja la receta con el encabezado de la clínica, los medicamentos y, si está firmada, los datos de la
// firma digital. Un borrador o una receta anulada lo dicen de forma visible para que no se use como válida.
func RecetaPDF(d DatosRecetaPDF) ([]byte, error) {
	r := d.Receta
	doc := nuevoDocumento(d.Clinica, "Receta "+r.Numero)
	doc.titulo("Receta médica " + r.Numero)
	switch r.Estado {
	case models.RecetaAnulada:
		doc.pdf.SetTextColor(190, 30, 50)
		doc.subtitulo("ANULADA: esta receta no es válida")
		doc.pdf.SetTextColor(20, 20, 20)
	case models.RecetaBorrador:
		doc.pdf.SetTextColor(150, 100, 0)
		doc.subtitulo("BORRADOR: sin firmar, no es válida")
		doc.pdf.SetTextColor(20, 20, 20)
	}
	doc.dato("Paciente:", d.Paciente)
	doc.dato("Documento:", d.Documento)
	doc.dato("Fecha:", d.Fecha)
	doc.dato("Profesional:", r.ProfesionalNombre)
	doc.pdf.Ln(3)

	doc.subtitulo("Prescripción")
	for i, m := range r.Medicamentos {
		doc.pdf.SetFont("Helvetica", "B", 10.5)
		doc.pdf.MultiCell(0, 5.4, doc.tr(fmt.Sprintf("%d. %s", i+1, m.Nombre)), "", "L", false)
		doc.pdf.SetFont("Helvetica", "", 10)
		doc.pdf.MultiCell(0, 5, doc.tr(fmt.Sprintf("Dosis: %s · Frecuencia: %s · Duración: %s", m.Dosis, m.Frecuencia, m.Duracion)), "", "L", false)
		if m.Indicaciones != "" {
			doc.pdf.MultiCell(0, 5, doc.tr("Indicaciones: "+m.Indicaciones), "", "L", false)
		}
		doc.pdf.Ln(2)
	}
	if r.Indicaciones != "" {
		doc.pdf.Ln(2)
		doc.subtitulo("Indicaciones generales")
		doc.parrafo(r.Indicaciones)
	}
	doc.pdf.Ln(6)
	if r.Firma != nil {
		doc.subtitulo("Firma digital del profesional")
		doc.dato("Firmante:", fmt.Sprintf("%s (reg. %s)", r.Firma.NombreFirmante, r.Firma.RegistroProfesional))
		doc.dato("Fecha de firma:", d.Firmada)
		doc.pdf.SetFont("Courier", "", 8)
		doc.pdf.SetTextColor(100, 100, 100)
		doc.pdf.MultiCell(0, 4, doc.tr("Huella SHA-256: "+r.Firma.HuellaSHA256), "", "L", false)
	}
	if r.Anulacion != nil {
		doc.pdf.SetTextColor(190, 30, 50)
		doc.pdf.SetFont("Helvetica", "", 9.5)
		doc.pdf.MultiCell(0, 5, doc.tr("Anulada por "+r.Anulacion.UsuarioNombre+". Motivo: "+r.Anulacion.Motivo), "", "L", false)
	}
	return doc.bytes()
}
