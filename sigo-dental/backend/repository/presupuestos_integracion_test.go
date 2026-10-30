package repository

import (
	"context"
	"errors"
	"testing"

	"sigo-dental-backend/models"
)

// RF-07 y RF-08: aprobar parcialmente genera tareas solo para los ítems aprobados; un presupuesto presentado no se
// edita (un cambio es una versión nueva) y el total es la suma de los ítems.
func TestAprobacionParcialGeneraTareasSoloDeLosAprobados(t *testing.T) {
	pg, tx := repoTransaccional(t)
	ctx := context.Background()
	pacienteID, usuarioID := pacienteDePrueba(t, tx)
	var profesionalID int
	if err := tx.QueryRow(ctx, `SELECT id FROM profesionales WHERE usuario_id = $1`, usuarioID).Scan(&profesionalID); err != nil {
		t.Skipf("falta el profesional de prueba: %v", err)
	}
	var proc1, proc2 int
	if err := tx.QueryRow(ctx, `SELECT id FROM procedimientos WHERE activo AND NOT requiere_consentimiento ORDER BY id LIMIT 1`).Scan(&proc1); err != nil {
		t.Skipf("falta el catálogo de prueba: %v", err)
	}
	if err := tx.QueryRow(ctx, `SELECT id FROM procedimientos WHERE activo AND id <> $1 ORDER BY id LIMIT 1`, proc1).Scan(&proc2); err != nil {
		t.Skipf("falta un segundo procedimiento: %v", err)
	}

	honorario := 10000
	autorizador := usuarioID
	items := []models.ItemParaGuardar{
		{ItemPedido: models.ItemPedido{Fase: 1, Sesion: 1, ProcedimientoID: proc1, HonorarioCentavos: &honorario, LaboratorioCentavos: 2500, DescuentoCentavos: 1500}, DescuentoAutorizadoPorID: &autorizador},
		{ItemPedido: models.ItemPedido{Fase: 1, Sesion: 2, ProcedimientoID: proc2}},
		{ItemPedido: models.ItemPedido{Fase: 2, Sesion: 1, ProcedimientoID: proc1}},
	}
	id, err := pg.GuardarPresupuesto(ctx, pacienteID, 0, profesionalID, "plan de prueba", items, usuarioID)
	if err != nil {
		t.Fatal(err)
	}
	borrador, _ := pg.ObtenerPresupuesto(ctx, pacienteID, id)
	if borrador.Estado != models.PresupBorrador || len(borrador.Items) != 3 || borrador.Numero == "" || borrador.Version != 1 {
		t.Fatalf("borrador mal creado: %+v", borrador)
	}
	if borrador.Items[0].TotalCentavos != 11000 {
		t.Errorf("total del primer ítem = %d, se esperaba 11000 (10000 + 2500 - 1500)", borrador.Items[0].TotalCentavos)
	}
	suma := 0
	for _, it := range borrador.Items {
		suma += it.TotalCentavos
	}
	if borrador.TotalCentavos != suma {
		t.Errorf("el total del presupuesto (%d) no es la suma de los ítems (%d)", borrador.TotalCentavos, suma)
	}

	if err := pg.ResolverPresupuesto(ctx, pacienteID, id, []int{borrador.Items[0].ID}, usuarioID); !errors.Is(err, models.ErrPresupuestoNoPresentado) {
		t.Errorf("resolver un borrador: error = %v", err)
	}
	if err := pg.PresentarPresupuesto(ctx, pacienteID, id, usuarioID); err != nil {
		t.Fatal(err)
	}
	if _, err := pg.GuardarPresupuesto(ctx, pacienteID, id, profesionalID, "otro", items, usuarioID); !errors.Is(err, models.ErrPresupuestoNoEditable) {
		t.Errorf("editar uno presentado: error = %v, se esperaba ErrPresupuestoNoEditable", err)
	}
	debeFallarComoInmutable(t, tx, "cambiar el precio de un ítem presentado", `UPDATE presupuesto_items SET honorario_centavos = 1, total_centavos = 1 - descuento_centavos + laboratorio_centavos WHERE presupuesto_id = $1`, id)

	aprobados := []int{borrador.Items[0].ID, borrador.Items[2].ID}
	if err := pg.ResolverPresupuesto(ctx, pacienteID, id, aprobados, usuarioID); err != nil {
		t.Fatal(err)
	}
	resuelto, _ := pg.ObtenerPresupuesto(ctx, pacienteID, id)
	if resuelto.Estado != models.PresupAprobadoParcial || resuelto.ResueltoPor == "" || resuelto.ResueltoEn == nil {
		t.Errorf("estado = %s, resuelto por %q", resuelto.Estado, resuelto.ResueltoPor)
	}
	tareas, err := pg.ListarTareas(ctx, pacienteID)
	if err != nil || len(tareas) != 2 {
		t.Fatalf("tareas = %d (%v), se esperaban 2: solo los ítems aprobados", len(tareas), err)
	}
	for _, tarea := range tareas {
		if tarea.ItemID == borrador.Items[1].ID {
			t.Error("el ítem no aprobado se volvió tarea")
		}
		if tarea.Estado != models.ItemAprobado {
			t.Errorf("estado de la tarea = %s", tarea.Estado)
		}
	}
	if err := pg.ResolverPresupuesto(ctx, pacienteID, id, aprobados, usuarioID); !errors.Is(err, models.ErrPresupuestoNoPresentado) {
		t.Errorf("resolver dos veces: error = %v", err)
	}

	// Un cambio genera una versión nueva; las tareas siguen a la versión vigente (aún sin aprobar: ninguna).
	nuevo, err := pg.NuevaVersionPresupuesto(ctx, pacienteID, id, usuarioID)
	if err != nil {
		t.Fatal(err)
	}
	v2, _ := pg.ObtenerPresupuesto(ctx, pacienteID, nuevo)
	if v2.Version != 2 || v2.Numero != resuelto.Numero || v2.Estado != models.PresupBorrador || len(v2.Items) != 3 || v2.TotalCentavos != resuelto.TotalCentavos {
		t.Errorf("versión nueva mal copiada: %+v", v2)
	}
	if err := intentar(t, tx, func(p *PostgresRepository) error {
		_, err := p.NuevaVersionPresupuesto(ctx, pacienteID, id, usuarioID)
		return err
	}); err == nil {
		t.Error("un presupuesto solo tiene una versión siguiente")
	}
	if tareas, _ := pg.ListarTareas(ctx, pacienteID); len(tareas) != 0 {
		t.Errorf("tras crear la versión nueva, las tareas de la anterior (%d) ya no son vigentes", len(tareas))
	}

	// Ligar una cita o un procedimiento a un ítem no aprobado se rechaza.
	err = intentar(t, tx, func(p *PostgresRepository) error {
		_, err := p.IniciarProcedimiento(ctx, models.NuevoProcedimientoRealizado{PacienteID: pacienteID, ProcedimientoID: proc1, PresupuestoItemID: &borrador.Items[1].ID}, profesionalID, usuarioID)
		return err
	})
	if !errors.Is(err, models.ErrReferencia) {
		t.Errorf("iniciar un procedimiento desde un ítem no aprobado: error = %v", err)
	}
}
