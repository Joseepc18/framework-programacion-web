package models

import "testing"

// El total de un ítem es honorario + laboratorio - descuento, y el del presupuesto la suma de sus ítems.
func TestTotalesDelPresupuesto(t *testing.T) {
	casos := []struct {
		honorario, laboratorio, descuento, total int
		valido                                   bool
	}{
		{10000, 0, 0, 10000, true},
		{10000, 2500, 0, 12500, true},
		{10000, 2500, 1500, 11000, true},
		{10000, 0, 10000, 0, true},
		{10000, 0, 10001, 0, false},
		{-1, 0, 0, 0, false},
		{100, -5, 0, 0, false},
	}
	for _, c := range casos {
		total, motivo := TotalItem(c.honorario, c.laboratorio, c.descuento)
		if (motivo == "") != c.valido || (c.valido && total != c.total) {
			t.Errorf("TotalItem(%d, %d, %d) = %d, %q", c.honorario, c.laboratorio, c.descuento, total, motivo)
		}
	}
	items := []ItemPresupuesto{
		{HonorarioCentavos: 10000, LaboratorioCentavos: 2500, DescuentoCentavos: 1500, TotalCentavos: 11000},
		{HonorarioCentavos: 4500, TotalCentavos: 4500},
		{HonorarioCentavos: 8000, LaboratorioCentavos: 1000, TotalCentavos: 9000},
	}
	h, l, d, total := SumarTotales(items)
	if h != 22500 || l != 3500 || d != 1500 || total != 24500 || total != h+l-d {
		t.Errorf("honorarios = %d, laboratorio = %d, descuentos = %d, total = %d", h, l, d, total)
	}
}

func TestEstadoDelPresupuestoTrasResolver(t *testing.T) {
	casos := []struct {
		aprobados, total int
		estado           string
	}{{3, 3, PresupAprobadoTotal}, {2, 3, PresupAprobadoParcial}, {1, 3, PresupAprobadoParcial}, {0, 3, PresupRechazado}, {0, 0, PresupRechazado}}
	for _, c := range casos {
		if got := EstadoTrasResolver(c.aprobados, c.total); got != c.estado {
			t.Errorf("%d de %d: estado = %s, se esperaba %s", c.aprobados, c.total, got, c.estado)
		}
	}
}

func TestEstadoDeItemYRolesQueAutorizan(t *testing.T) {
	si, no := true, false
	casos := []struct {
		aprobado           *bool
		agendado, cumplido bool
		estado             string
	}{{nil, false, false, ItemPendiente}, {&no, false, false, ItemRechazado}, {&si, false, false, ItemAprobado}, {&si, true, false, ItemAgendado}, {&si, true, true, ItemCumplido}}
	for _, c := range casos {
		if got := EstadoDeItem(c.aprobado, c.agendado, c.cumplido); got != c.estado {
			t.Errorf("estado = %s, se esperaba %s", got, c.estado)
		}
	}
	if !RolPuedeAutorizar("odontologo", "odontologo, admin_ti") || !RolPuedeAutorizar("admin_ti", "odontologo,admin_ti") || RolPuedeAutorizar("recepcionista", "odontologo,admin_ti") || RolPuedeAutorizar("odontologo", "") {
		t.Error("la lista de roles que autorizan descuentos se interpreta mal")
	}
}
