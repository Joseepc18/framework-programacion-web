package models

import (
	"fmt"
	"slices"
	"time"
)

// Rangos del sondaje en milímetros (RF-04).
const (
	ProfundidadMaxMM = 15
	MargenMinMM      = -10
	MargenMaxMM      = 15
)

// SitiosPeriodontales son los seis puntos de sondaje de cada diente: tres vestibulares y tres linguales, de distal a mesial.
var SitiosPeriodontales = []string{"DV", "V", "MV", "DL", "L", "ML"}

// NIC calcula el nivel de inserción clínica en milímetros. Convención: el margen gingival se mide desde la unión
// cemento-esmalte, positivo si el margen está apical a ella (recesión) y negativo si está coronal (agrandamiento), así
// que NIC = profundidad al sondaje + margen. Ejemplos: profundidad 3 con recesión 2 da 5; profundidad 4 con margen
// 1 mm coronal (-1) da 3.
func NIC(profundidadMM, margenMM int) int { return profundidadMM + margenMM }

// Sitio es la medición de un punto de sondaje. El NIC lo calcula el servidor al leer; nunca lo envía el cliente.
type Sitio struct {
	Sitio       string `json:"sitio"`
	Profundidad int    `json:"profundidad"`
	Margen      int    `json:"margen"`
	Sangrado    bool   `json:"sangrado"`
	NIC         int    `json:"nic"`
}

// DientePerio son los datos periodontales de un diente: movilidad (0 a 3), furca (0 a 3) y sus seis sitios.
type DientePerio struct {
	Pieza     int     `json:"pieza"`
	Movilidad int     `json:"movilidad"`
	Furca     int     `json:"furca"`
	Sitios    []Sitio `json:"sitios"`
}

// ExamenPerio es un examen periodontal completo.
type ExamenPerio struct {
	ID         int           `json:"id"`
	PacienteID int           `json:"paciente_id"`
	Fecha      string        `json:"fecha"`
	Notas      string        `json:"notas"`
	CreadoPor  string        `json:"creado_por"`
	CreadoEn   time.Time     `json:"creado_en"`
	Dientes    []DientePerio `json:"dientes"`
}

// ResumenExamenPerio es una fila de la lista de exámenes, con indicadores para compararlos.
type ResumenExamenPerio struct {
	ID                 int     `json:"id"`
	Fecha              string  `json:"fecha"`
	CreadoPor          string  `json:"creado_por"`
	Dientes            int     `json:"dientes"`
	ProfundidadMedia   float64 `json:"profundidad_media"`
	NICMedio           float64 `json:"nic_medio"`
	SangradoPorcentaje float64 `json:"sangrado_porcentaje"`
}

// ListaPerio es la respuesta de la lista de exámenes: incluye las piezas ausentes del odontograma, que no se ofrecen.
type ListaPerio struct {
	Examenes []ResumenExamenPerio `json:"examenes"`
	Ausentes []int                `json:"ausentes"`
}

// ValidarDientePerio comprueba un diente del examen y devuelve el motivo si es inválido. Solo dentición permanente; la
// furca solo existe en premolares y molares (piezas terminadas en 4 a 8).
func ValidarDientePerio(d DientePerio) string {
	if !PiezaFDIValida(d.Pieza) || !PiezaPermanente(d.Pieza) {
		return fmt.Sprintf("La pieza %d no es un diente permanente válido", d.Pieza)
	}
	if d.Movilidad < 0 || d.Movilidad > 3 {
		return fmt.Sprintf("Pieza %d: la movilidad va de 0 a 3", d.Pieza)
	}
	if d.Furca < 0 || d.Furca > 3 {
		return fmt.Sprintf("Pieza %d: la furca va de 0 a 3 (grado I a III)", d.Pieza)
	}
	if d.Furca > 0 && d.Pieza%10 < 4 {
		return fmt.Sprintf("Pieza %d: los incisivos y caninos no tienen furca", d.Pieza)
	}
	if len(d.Sitios) != len(SitiosPeriodontales) {
		return fmt.Sprintf("Pieza %d: se necesitan los 6 puntos de sondaje", d.Pieza)
	}
	vistos := []string{}
	for _, s := range d.Sitios {
		if !slices.Contains(SitiosPeriodontales, s.Sitio) || slices.Contains(vistos, s.Sitio) {
			return fmt.Sprintf("Pieza %d: el punto %q no es válido o está repetido", d.Pieza, s.Sitio)
		}
		vistos = append(vistos, s.Sitio)
		if s.Profundidad < 0 || s.Profundidad > ProfundidadMaxMM {
			return fmt.Sprintf("Pieza %d, %s: la profundidad al sondaje debe estar entre 0 y %d mm", d.Pieza, s.Sitio, ProfundidadMaxMM)
		}
		if s.Margen < MargenMinMM || s.Margen > MargenMaxMM {
			return fmt.Sprintf("Pieza %d, %s: el margen gingival debe estar entre %d y %d mm", d.Pieza, s.Sitio, MargenMinMM, MargenMaxMM)
		}
		if NIC(s.Profundidad, s.Margen) < 0 {
			return fmt.Sprintf("Pieza %d, %s: el nivel de inserción no puede ser negativo (revisa profundidad y margen)", d.Pieza, s.Sitio)
		}
	}
	return ""
}
