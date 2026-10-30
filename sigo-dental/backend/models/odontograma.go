package models

import (
	"errors"
	"slices"
	"time"
)

// Tipos de hallazgo del odontograma (RF-03); coinciden con odontograma_hallazgos.tipo en la base de datos.
const (
	HallazgoCaries            = "caries"
	HallazgoRestauracion      = "restauracion"
	HallazgoEndodoncia        = "endodoncia"
	HallazgoImplante          = "implante"
	HallazgoProtesisFija      = "protesis_fija"
	HallazgoProtesisRemovible = "protesis_removible"
	HallazgoExodoncia         = "exodoncia"
	HallazgoSellante          = "sellante"
	HallazgoAusente           = "ausente"
)

// Estados de un hallazgo. Solo "realizado" (un tratamiento hecho en la clínica) es inmutable (RN-01); "requerido" es
// lo que falta por hacer y "existente" lo que el paciente ya traía, y ambos se pueden corregir o retirar.
const (
	HallazgoRequerido = "requerido"
	HallazgoExistente = "existente"
	HallazgoRealizado = "realizado"
)

// Superficies de un diente. "oclusal" es la cara de masticación (incisal en incisivos y caninos); "lingual" es la
// palatina en el maxilar superior. Un hallazgo sin superficie afecta a la pieza completa.
const (
	SuperficieMesial     = "mesial"
	SuperficieDistal     = "distal"
	SuperficieOclusal    = "oclusal"
	SuperficieVestibular = "vestibular"
	SuperficieLingual    = "lingual"
)

// SuperficiesValidas es la lista cerrada de superficies.
var SuperficiesValidas = []string{SuperficieMesial, SuperficieDistal, SuperficieOclusal, SuperficieVestibular, SuperficieLingual}

// estadosPermitidos indica qué estados admite cada tipo de hallazgo y si se marca por superficie.
var reglasHallazgo = map[string]struct {
	estados    []string
	superficie bool
}{
	// Una caries es siempre algo por tratar: cuando se trata, se retira y se registra la restauración.
	HallazgoCaries:            {[]string{HallazgoRequerido}, true},
	HallazgoRestauracion:      {[]string{HallazgoRequerido, HallazgoExistente, HallazgoRealizado}, true},
	HallazgoSellante:          {[]string{HallazgoRequerido, HallazgoExistente, HallazgoRealizado}, true},
	HallazgoEndodoncia:        {[]string{HallazgoRequerido, HallazgoExistente, HallazgoRealizado}, false},
	HallazgoImplante:          {[]string{HallazgoRequerido, HallazgoExistente, HallazgoRealizado}, false},
	HallazgoProtesisFija:      {[]string{HallazgoRequerido, HallazgoExistente, HallazgoRealizado}, false},
	HallazgoProtesisRemovible: {[]string{HallazgoRequerido, HallazgoExistente, HallazgoRealizado}, false},
	HallazgoExodoncia:         {[]string{HallazgoRequerido, HallazgoRealizado}, false},
	HallazgoAusente:           {[]string{HallazgoExistente}, false},
}

// TiposDeHallazgo devuelve los tipos válidos.
func TiposDeHallazgo() []string {
	tipos := make([]string, 0, len(reglasHallazgo))
	for tipo := range reglasHallazgo {
		tipos = append(tipos, tipo)
	}
	slices.Sort(tipos)
	return tipos
}

// ValidarHallazgo comprueba la combinación de pieza, superficie, tipo y estado; devuelve el motivo si no es válida.
func ValidarHallazgo(pieza int, superficie, tipo, estado string) string {
	if !PiezaFDIValida(pieza) {
		return "La pieza dental no es válida: usa numeración FDI (11 a 48 permanentes, 51 a 85 temporales)"
	}
	regla, ok := reglasHallazgo[tipo]
	if !ok {
		return "El tipo de hallazgo no es válido"
	}
	if !slices.Contains(regla.estados, estado) {
		return "Ese hallazgo no admite el estado indicado"
	}
	if regla.superficie {
		if !slices.Contains(SuperficiesValidas, superficie) {
			return "Este hallazgo se marca en una superficie: mesial, distal, oclusal, vestibular o lingual"
		}
	} else if superficie != "" {
		return "Este hallazgo afecta a la pieza completa: no lleva superficie"
	}
	return ""
}

// PiezaFDIValida indica si el número es un diente de la numeración FDI: permanentes 11-18, 21-28, 31-38 y 41-48
// (32 dientes) y temporales 51-55, 61-65, 71-75 y 81-85 (20 dientes).
func PiezaFDIValida(pieza int) bool {
	cuadrante, posicion := pieza/10, pieza%10
	switch {
	case cuadrante >= 1 && cuadrante <= 4:
		return posicion >= 1 && posicion <= 8
	case cuadrante >= 5 && cuadrante <= 8:
		return posicion >= 1 && posicion <= 5
	}
	return false
}

// PiezaPermanente indica si la pieza (FDI válida) es de la dentición permanente.
func PiezaPermanente(pieza int) bool { return pieza/10 >= 1 && pieza/10 <= 4 }

// NumeroUniversal convierte una pieza FDI a la numeración Universal (1 a 32 permanentes, A a T temporales).
// Devuelve cadena vacía si la pieza no es válida. Internamente el sistema guarda siempre FDI.
func NumeroUniversal(pieza int) string {
	if !PiezaFDIValida(pieza) {
		return ""
	}
	cuadrante, posicion := pieza/10, pieza%10
	if PiezaPermanente(pieza) {
		var n int
		switch cuadrante {
		case 1: // 18..11 -> 1..8
			n = 9 - posicion
		case 2: // 21..28 -> 9..16
			n = 8 + posicion
		case 3: // 31..38 -> 24..17 (de derecha a izquierda de la arcada inferior)
			n = 25 - posicion
		case 4: // 41..48 -> 25..32
			n = 24 + posicion
		}
		return itoa(n)
	}
	var n int
	switch cuadrante {
	case 5: // 55..51 -> A..E
		n = 5 - posicion
	case 6: // 61..65 -> F..J
		n = 5 + posicion - 1
	case 7: // 75..71 -> K..O (lado izquierdo de la arcada inferior)
		n = 15 - posicion
	case 8: // 81..85 -> P..T
		n = 14 + posicion
	}
	return string(rune('A' + n))
}

func itoa(n int) string {
	if n < 10 {
		return string(rune('0' + n))
	}
	return string(rune('0'+n/10)) + string(rune('0'+n%10))
}

// Hallazgo es una marca del odontograma. Nunca se sobrescribe: corregir un hallazgo retira el anterior y crea otro,
// de modo que el historial por visita se conserva.
type Hallazgo struct {
	ID            int        `json:"id"`
	PacienteID    int        `json:"paciente_id"`
	Pieza         int        `json:"pieza"`
	Superficie    string     `json:"superficie"`
	Tipo          string     `json:"tipo"`
	Estado        string     `json:"estado"`
	Observacion   string     `json:"observacion"`
	CitaID        *int       `json:"cita_id"`
	RegistradoPor string     `json:"registrado_por"`
	RegistradoEn  time.Time  `json:"registrado_en"`
	RetiradoPor   string     `json:"retirado_por"`
	RetiradoEn    *time.Time `json:"retirado_en"`
	// ReemplazaA es el hallazgo que este corrigió (y retiró), si lo hay.
	ReemplazaA *int `json:"reemplaza_a"`
}

// NuevoHallazgo son los datos de un hallazgo ya validado.
type NuevoHallazgo struct {
	PacienteID  int
	Pieza       int
	Superficie  string
	Tipo        string
	Estado      string
	Observacion string
}

// ErrHallazgoRetirado: el hallazgo ya no está vigente, no se puede corregir ni retirar otra vez.
var ErrHallazgoRetirado = errors.New("el hallazgo ya fue retirado")
