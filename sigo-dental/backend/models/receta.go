package models

import (
	"strings"
	"time"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

// Estados derivados de una receta.
const (
	RecetaBorrador = "borrador"
	RecetaFirmada  = "firmada"
	RecetaAnulada  = "anulada"
)

// Medicamento es una línea de la receta.
type Medicamento struct {
	Nombre       string `json:"nombre"`
	Dosis        string `json:"dosis"`
	Frecuencia   string `json:"frecuencia"`
	Duracion     string `json:"duracion"`
	Indicaciones string `json:"indicaciones"`
}

// AvisoAlergia advierte que un medicamento coincide con una alergia registrada en la anamnesis.
type AvisoAlergia struct {
	Medicamento string `json:"medicamento"`
	Alergia     string `json:"alergia"`
}

// Anulacion de una receta (ver también Anulacion de cobros: esta guarda además la receta que la reemplaza).
type AnulacionReceta struct {
	Motivo        string    `json:"motivo"`
	UsuarioNombre string    `json:"usuario_nombre"`
	ReemplazoID   *int      `json:"reemplazo_id"`
	CreadaEn      time.Time `json:"creada_en"`
}

// Receta es una receta con sus medicamentos. El estado se deriva: borrador, firmada o anulada.
type Receta struct {
	ID                int              `json:"id"`
	Numero            string           `json:"numero"`
	PacienteID        int              `json:"paciente_id"`
	ProfesionalID     int              `json:"profesional_id"`
	ProfesionalNombre string           `json:"profesional_nombre"`
	CitaID            *int             `json:"cita_id"`
	Indicaciones      string           `json:"indicaciones"`
	CreadaEn          time.Time        `json:"creada_en"`
	Medicamentos      []Medicamento    `json:"medicamentos"`
	Estado            string           `json:"estado"`
	Firma             *Firma           `json:"firma"`
	Anulacion         *AnulacionReceta `json:"anulacion"`
	Avisos            []AvisoAlergia   `json:"avisos_alergia"`
}

// ContenidoCanonicoReceta es lo que firma el odontólogo al firmar una receta.
func ContenidoCanonicoReceta(r Receta) string {
	return canonico(struct {
		Version       string        `json:"version"`
		Numero        string        `json:"numero"`
		PacienteID    int           `json:"paciente_id"`
		ProfesionalID int           `json:"profesional_id"`
		CitaID        *int          `json:"cita_id"`
		Indicaciones  string        `json:"indicaciones"`
		CreadaEn      string        `json:"creada_en"`
		Medicamentos  []Medicamento `json:"medicamentos"`
	}{"receta-v1", r.Numero, r.PacienteID, r.ProfesionalID, r.CitaID, r.Indicaciones, r.CreadaEn.UTC().Format(time.RFC3339Nano), r.Medicamentos})
}

// normalizar pasa a minúsculas y quita acentos para comparar nombres de fármacos.
func normalizar(texto string) string {
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	resultado, _, err := transform.String(t, strings.ToLower(texto))
	if err != nil {
		return strings.ToLower(texto)
	}
	return resultado
}

func palabras(texto string) []string {
	return strings.FieldsFunc(normalizar(texto), func(r rune) bool { return !unicode.IsLetter(r) })
}

// familiasDeFarmacos agrupa fármacos que comparten riesgo de reacción: quien es alérgico a la familia (o a uno de sus
// miembros) lo es, a efectos del aviso, a los demás. Es una ayuda, no un reemplazo del criterio del profesional.
var familiasDeFarmacos = [][]string{
	{"penicilina", "amoxicilina", "ampicilina", "cloxacilina", "dicloxacilina", "clavulanico", "betalactamico", "betalactamicos"},
	{"cefalosporina", "cefalosporinas", "cefalexina", "cefazolina", "cefadroxilo", "ceftriaxona", "cefuroxima", "cefixima"},
	{"aine", "aines", "antiinflamatorio", "antiinflamatorios", "ibuprofeno", "naproxeno", "diclofenaco", "ketorolaco", "ketoprofeno", "meloxicam", "aspirina", "acetilsalicilico", "nimesulida"},
	{"sulfa", "sulfas", "sulfonamida", "sulfonamidas", "sulfametoxazol", "trimetoprima"},
	{"macrolido", "macrolidos", "azitromicina", "claritromicina", "eritromicina"},
	{"opioide", "opioides", "codeina", "tramadol", "morfina"},
	{"lidocaina", "articaina", "mepivacaina", "bupivacaina", "anestesico", "anestesicos"},
}

// palabrasSinAlergia son las que aparecen al declarar que no hay alergias ("Ninguna", "No refiere...").
var palabrasSinAlergia = map[string]bool{
	"ninguna": true, "ninguno": true, "no": true, "niega": true, "refiere": true, "conocida": true, "conocidas": true,
	"alergia": true, "alergias": true, "a": true, "de": true, "sin": true, "farmacos": true, "medicamentos": true, "y": true, "n": true, "ni": true,
}

func familiaDe(palabra string) int {
	for i, familia := range familiasDeFarmacos {
		for _, miembro := range familia {
			if palabra == miembro {
				return i
			}
		}
	}
	return -1
}

// AvisosDeAlergia compara los medicamentos con el texto de alergias de la anamnesis y devuelve un aviso por cada
// coincidencia: el mismo fármaco o uno de su misma familia (por ejemplo, alergia a penicilina y receta de amoxicilina).
func AvisosDeAlergia(alergias string, medicamentos []string) []AvisoAlergia {
	var declaradas []string
	for _, p := range palabras(alergias) {
		if len(p) >= 4 && !palabrasSinAlergia[p] {
			declaradas = append(declaradas, p)
		}
	}
	avisos := []AvisoAlergia{}
	if len(declaradas) == 0 {
		return avisos
	}
	for _, medicamento := range medicamentos {
	busqueda:
		for _, palabra := range palabras(medicamento) {
			if len(palabra) < 4 {
				continue
			}
			for _, alergia := range declaradas {
				if palabra == alergia || (familiaDe(palabra) >= 0 && familiaDe(palabra) == familiaDe(alergia)) {
					avisos = append(avisos, AvisoAlergia{Medicamento: medicamento, Alergia: alergia})
					break busqueda
				}
			}
		}
	}
	return avisos
}
