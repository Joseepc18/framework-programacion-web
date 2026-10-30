package documentos

import (
	"fmt"

	"sigo-dental-backend/models"
)

// DatosPresupuestoPDF es lo necesario para dibujar un presupuesto.
type DatosPresupuestoPDF struct {
	Clinica     Clinica
	Presupuesto models.Presupuesto
	Paciente    string
	Documento   string
	Fecha       string
}

// dinero da formato a centavos: 12345 -> "123.45".
func dinero(centavos int, moneda string) string {
	signo := ""
	if centavos < 0 {
		signo, centavos = "-", -centavos
	}
	return fmt.Sprintf("%s%d.%02d %s", signo, centavos/100, centavos%100, moneda)
}

var nombreEstadoPresupuesto = map[string]string{
	models.PresupBorrador: "Borrador", models.PresupPresentado: "Presentado", models.PresupAprobadoTotal: "Aprobado totalmente",
	models.PresupAprobadoParcial: "Aprobado parcialmente", models.PresupRechazado: "Rechazado",
}

// PresupuestoPDF dibuja el presupuesto por fases y sesiones con honorarios, laboratorio, descuentos y totales.
func PresupuestoPDF(d DatosPresupuestoPDF) ([]byte, error) {
	p := d.Presupuesto
	doc := nuevoDocumento(d.Clinica, "Presupuesto "+p.Numero)
	doc.titulo(fmt.Sprintf("Presupuesto %s (versión %d)", p.Numero, p.Version))
	doc.dato("Estado:", nombreEstadoPresupuesto[p.Estado])
	doc.dato("Paciente:", d.Paciente)
	doc.dato("Documento:", d.Documento)
	doc.dato("Fecha:", d.Fecha)
	doc.dato("Profesional:", p.ProfesionalNombre)
	doc.pdf.Ln(3)

	fase, sesion := 0, 0
	for _, it := range p.Items {
		if it.Fase != fase {
			fase, sesion = it.Fase, 0
			doc.pdf.Ln(2)
			doc.subtitulo(fmt.Sprintf("Fase %d", fase))
		}
		if it.Sesion != sesion {
			sesion = it.Sesion
			doc.pdf.SetFont("Helvetica", "I", 9.5)
			doc.pdf.CellFormat(0, 5.5, doc.tr(fmt.Sprintf("Sesión %d", sesion)), "", 1, "L", false, 0, "")
		}
		detalle := it.Descripcion
		if it.Pieza != nil {
			detalle += fmt.Sprintf(" · pieza %d", *it.Pieza)
			if it.Superficie != "" {
				detalle += " (" + it.Superficie + ")"
			}
		}
		if it.Aprobado != nil && !*it.Aprobado {
			detalle += " · RECHAZADO"
		}
		doc.pdf.SetFont("Helvetica", "", 9.5)
		doc.pdf.CellFormat(118, 5.5, doc.tr(detalle), "", 0, "L", false, 0, "")
		doc.pdf.CellFormat(56, 5.5, doc.tr(dinero(it.TotalCentavos, p.Moneda)), "", 1, "R", false, 0, "")
		doc.pdf.SetFont("Helvetica", "", 8)
		doc.pdf.SetTextColor(100, 100, 100)
		desglose := fmt.Sprintf("Honorarios %s · Laboratorio %s", dinero(it.HonorarioCentavos, p.Moneda), dinero(it.LaboratorioCentavos, p.Moneda))
		if it.DescuentoCentavos > 0 {
			desglose += fmt.Sprintf(" · Descuento %s (autorizó %s)", dinero(it.DescuentoCentavos, p.Moneda), it.DescuentoAutorizadoPor)
		}
		doc.pdf.CellFormat(0, 4.2, doc.tr(desglose), "", 1, "L", false, 0, "")
		doc.pdf.SetTextColor(20, 20, 20)
	}

	doc.pdf.Ln(5)
	linea := func(etiqueta string, centavos int, negrita bool) {
		estilo := ""
		if negrita {
			estilo = "B"
		}
		doc.pdf.SetFont("Helvetica", estilo, 10.5)
		doc.pdf.CellFormat(118, 6, doc.tr(etiqueta), "", 0, "R", false, 0, "")
		doc.pdf.CellFormat(56, 6, doc.tr(dinero(centavos, p.Moneda)), "", 1, "R", false, 0, "")
	}
	linea("Honorarios", p.HonorariosCentavos, false)
	linea("Laboratorio", p.LaboratorioCentavos, false)
	linea("Descuentos", -p.DescuentosCentavos, false)
	linea("Total", p.TotalCentavos, true)
	if p.Estado == models.PresupAprobadoParcial {
		linea("Total aprobado", p.AprobadoCentavos, true)
	}
	if p.Notas != "" {
		doc.pdf.Ln(4)
		doc.parrafo(p.Notas)
	}
	return doc.bytes()
}
