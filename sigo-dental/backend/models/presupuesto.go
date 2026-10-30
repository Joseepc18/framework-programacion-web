package models

import (
	"errors"
	"slices"
	"time"
)

// Errores de los presupuestos.
var (
	// ErrDescuentoSinAutorizar: hay un descuento nuevo o modificado y falta la autorización de un rol habilitado.
	ErrDescuentoSinAutorizar = errors.New("el descuento requiere autorización")
	// ErrAutorizacionInvalida: las credenciales de quien autoriza no son válidas o su rol no está habilitado.
	ErrAutorizacionInvalida = errors.New("autorización inválida")
	// ErrPresupuestoNoEditable: solo un borrador se edita; un cambio posterior genera una versión nueva.
	ErrPresupuestoNoEditable = errors.New("el presupuesto ya no es un borrador")
	// ErrPresupuestoNoPresentado: solo un presupuesto presentado se aprueba o rechaza.
	ErrPresupuestoNoPresentado = errors.New("el presupuesto no está presentado")
	// ErrPresupuestoVacio: un presupuesto necesita al menos un ítem para presentarse.
	ErrPresupuestoVacio = errors.New("el presupuesto no tiene ítems")
	// ErrImporteInvalido: un importe del presupuesto no es válido (negativo, excesivo o descuento mayor que el ítem).
	ErrImporteInvalido = errors.New("importe inválido")
)

// Estados de un presupuesto (RF-07).
const (
	PresupBorrador        = "borrador"
	PresupPresentado      = "presentado"
	PresupAprobadoTotal   = "aprobado_total"
	PresupAprobadoParcial = "aprobado_parcial"
	PresupRechazado       = "rechazado"
)

// Estados derivados de un ítem.
const (
	ItemPendiente = "pendiente"
	ItemAprobado  = "aprobado"
	ItemRechazado = "rechazado"
	ItemAgendado  = "agendado"
	ItemCumplido  = "cumplido"
)

// ItemPedido es un ítem tal como lo envía el cliente: los totales nunca viajan, los calcula el servidor.
type ItemPedido struct {
	// ID del ítem existente que se conserva (para mantener la autorización de su descuento); nil si es nuevo.
	ID              *int   `json:"id"`
	Fase            int    `json:"fase"`
	Sesion          int    `json:"sesion"`
	ProcedimientoID int    `json:"procedimiento_id"`
	Pieza           *int   `json:"pieza"`
	Superficie      string `json:"superficie"`
	HallazgoID      *int   `json:"hallazgo_id"`
	// HonorarioCentavos es opcional: sin valor se usa el precio base del catálogo.
	HonorarioCentavos   *int `json:"honorario_centavos"`
	LaboratorioCentavos int  `json:"laboratorio_centavos"`
	DescuentoCentavos   int  `json:"descuento_centavos"`
}

// ItemPresupuesto es un ítem guardado, con sus totales.
type ItemPresupuesto struct {
	ID                       int    `json:"id"`
	Fase                     int    `json:"fase"`
	Sesion                   int    `json:"sesion"`
	ProcedimientoID          int    `json:"procedimiento_id"`
	Descripcion              string `json:"descripcion"`
	Pieza                    *int   `json:"pieza"`
	Superficie               string `json:"superficie"`
	HallazgoID               *int   `json:"hallazgo_id"`
	DuracionMin              int    `json:"duracion_min"`
	PrecioBaseCentavos       int    `json:"precio_base_centavos"`
	HonorarioCentavos        int    `json:"honorario_centavos"`
	LaboratorioCentavos      int    `json:"laboratorio_centavos"`
	DescuentoCentavos        int    `json:"descuento_centavos"`
	DescuentoAutorizadoPorID *int   `json:"descuento_autorizado_por_id"`
	DescuentoAutorizadoPor   string `json:"descuento_autorizado_por"`
	TotalCentavos            int    `json:"total_centavos"`
	Aprobado                 *bool  `json:"aprobado"`
	Estado                   string `json:"estado"`
	CitaID                   *int   `json:"cita_id"`
}

// Presupuesto es un presupuesto por fases y sesiones, con totales calculados en el servidor.
type Presupuesto struct {
	ID                int        `json:"id"`
	Numero            string     `json:"numero"`
	Version           int        `json:"version"`
	OrigenID          *int       `json:"origen_id"`
	PacienteID        int        `json:"paciente_id"`
	ProfesionalNombre string     `json:"profesional_nombre"`
	Estado            string     `json:"estado"`
	Notas             string     `json:"notas"`
	CreadoEn          time.Time  `json:"creado_en"`
	PresentadoEn      *time.Time `json:"presentado_en"`
	PresentadoPor     string     `json:"presentado_por"`
	ResueltoEn        *time.Time `json:"resuelto_en"`
	ResueltoPor       string     `json:"resuelto_por"`
	// Reemplazado indica que ya existe una versión más nueva.
	Reemplazado         bool              `json:"reemplazado"`
	Moneda              string            `json:"moneda"`
	HonorariosCentavos  int               `json:"honorarios_centavos"`
	LaboratorioCentavos int               `json:"laboratorio_centavos"`
	DescuentosCentavos  int               `json:"descuentos_centavos"`
	TotalCentavos       int               `json:"total_centavos"`
	AprobadoCentavos    int               `json:"aprobado_centavos"`
	Items               []ItemPresupuesto `json:"items"`
}

// TotalItem calcula el total de un ítem: honorario + laboratorio - descuento. Devuelve el motivo si no es válido.
func TotalItem(honorario, laboratorio, descuento int) (int, string) {
	switch {
	case honorario < 0 || laboratorio < 0 || descuento < 0:
		return 0, "Los importes no pueden ser negativos"
	case honorario > 100_000_000 || laboratorio > 100_000_000:
		return 0, "Un importe supera el máximo permitido"
	case descuento > honorario+laboratorio:
		return 0, "El descuento no puede superar el importe del ítem"
	}
	return honorario + laboratorio - descuento, ""
}

// SumarTotales suma los ítems del presupuesto: honorarios, laboratorio, descuentos y total. Los importes se
// calculan solo en el servidor; el total del presupuesto es siempre la suma de los totales de sus ítems.
func SumarTotales(items []ItemPresupuesto) (honorarios, laboratorio, descuentos, total int) {
	for _, it := range items {
		honorarios += it.HonorarioCentavos
		laboratorio += it.LaboratorioCentavos
		descuentos += it.DescuentoCentavos
		total += it.TotalCentavos
	}
	return honorarios, laboratorio, descuentos, total
}

// EstadoDeItem deriva el estado de un ítem a partir de la decisión y de lo que se hizo con él.
func EstadoDeItem(aprobado *bool, agendado, cumplido bool) string {
	switch {
	case aprobado == nil:
		return ItemPendiente
	case !*aprobado:
		return ItemRechazado
	case cumplido:
		return ItemCumplido
	case agendado:
		return ItemAgendado
	}
	return ItemAprobado
}

// EstadoTrasResolver decide el estado del presupuesto según los ítems aprobados: todos (total), algunos (parcial) o
// ninguno (rechazado).
func EstadoTrasResolver(aprobados, total int) string {
	switch {
	case total > 0 && aprobados == total:
		return PresupAprobadoTotal
	case aprobados > 0:
		return PresupAprobadoParcial
	}
	return PresupRechazado
}

// RolPuedeAutorizar indica si el código de rol figura en la lista de roles que autorizan descuentos (parámetro
// descuento_roles, separado por comas).
func RolPuedeAutorizar(rol, lista string) bool {
	return slices.Contains(partirLista(lista), rol)
}

func partirLista(lista string) []string {
	var partes []string
	actual := ""
	for _, r := range lista + "," {
		if r == ',' {
			if actual != "" {
				partes = append(partes, actual)
			}
			actual = ""
			continue
		}
		if r != ' ' {
			actual += string(r)
		}
	}
	return partes
}

// Tarea es un ítem aprobado que el paciente puede agendar (RF-08).
type Tarea struct {
	ItemID            int        `json:"item_id"`
	PresupuestoID     int        `json:"presupuesto_id"`
	PresupuestoNumero string     `json:"presupuesto_numero"`
	Version           int        `json:"version"`
	Fase              int        `json:"fase"`
	Sesion            int        `json:"sesion"`
	ProcedimientoID   int        `json:"procedimiento_id"`
	Descripcion       string     `json:"descripcion"`
	Pieza             *int       `json:"pieza"`
	Superficie        string     `json:"superficie"`
	DuracionMin       int        `json:"duracion_min"`
	TotalCentavos     int        `json:"total_centavos"`
	Estado            string     `json:"estado"`
	CitaID            *int       `json:"cita_id"`
	CitaInicio        *time.Time `json:"cita_inicio"`
}

// ItemParaGuardar es un ítem validado, con la autorización de su descuento ya resuelta por el controlador.
type ItemParaGuardar struct {
	ItemPedido
	DescuentoAutorizadoPorID *int
	Aprobado                 *bool
}
