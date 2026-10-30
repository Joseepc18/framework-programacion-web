// Este archivo agrupa el acceso a datos de presupuestos y tareas agendables (F2-7: RF-07 y RF-08).
package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"sigo-dental-backend/models"
)

const columnasPresupuesto = `p.id, p.numero, p.version, p.origen_id, p.paciente_id, up.nombre, p.estado, p.notas, p.creado_en,
	p.presentado_en, COALESCE(upr.nombre, ''), p.resuelto_en, COALESCE(ure.nombre, ''),
	EXISTS (SELECT 1 FROM presupuestos n WHERE n.origen_id = p.id),
	COALESCE((SELECT valor FROM parametros WHERE clave = 'moneda'), 'USD')`

const origenPresupuesto = `FROM presupuestos p
	JOIN profesionales pf ON pf.id = p.profesional_id JOIN usuarios up ON up.id = pf.usuario_id
	LEFT JOIN usuarios upr ON upr.id = p.presentado_por LEFT JOIN usuarios ure ON ure.id = p.resuelto_por`

func escanearPresupuesto(row pgx.Row, p *models.Presupuesto) error {
	return row.Scan(&p.ID, &p.Numero, &p.Version, &p.OrigenID, &p.PacienteID, &p.ProfesionalNombre, &p.Estado, &p.Notas, &p.CreadoEn,
		&p.PresentadoEn, &p.PresentadoPor, &p.ResueltoEn, &p.ResueltoPor, &p.Reemplazado, &p.Moneda)
}

// itemsDe devuelve los ítems con su estado derivado: aprobado, agendado (hay una cita vigente) o cumplido (hay un
// procedimiento realizado ligado).
func (pg *PostgresRepository) itemsDe(ctx context.Context, presupuestoID int) ([]models.ItemPresupuesto, error) {
	rows, err := pg.db.Query(ctx,
		`SELECT i.id, i.fase, i.sesion, i.procedimiento_id, i.descripcion, i.pieza, COALESCE(i.superficie, ''), i.hallazgo_id,
		        pr.duracion_min, pr.precio_centavos, i.honorario_centavos, i.laboratorio_centavos, i.descuento_centavos,
		        i.descuento_autorizado_por, COALESCE(ua.nombre, ''), i.total_centavos, i.aprobado,
		        (SELECT c.id FROM citas c WHERE c.presupuesto_item_id = i.id AND c.estado NOT IN ('cancelado', 'no_asistio') ORDER BY c.id DESC LIMIT 1),
		        EXISTS (SELECT 1 FROM procedimientos_realizados r WHERE r.presupuesto_item_id = i.id AND r.estado = 'realizado')
		   FROM presupuesto_items i JOIN procedimientos pr ON pr.id = i.procedimiento_id
		   LEFT JOIN usuarios ua ON ua.id = i.descuento_autorizado_por
		  WHERE i.presupuesto_id = $1 ORDER BY i.fase, i.sesion, i.orden, i.id`, presupuestoID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []models.ItemPresupuesto{}
	for rows.Next() {
		var it models.ItemPresupuesto
		var cumplido bool
		if err := rows.Scan(&it.ID, &it.Fase, &it.Sesion, &it.ProcedimientoID, &it.Descripcion, &it.Pieza, &it.Superficie, &it.HallazgoID,
			&it.DuracionMin, &it.PrecioBaseCentavos, &it.HonorarioCentavos, &it.LaboratorioCentavos, &it.DescuentoCentavos,
			&it.DescuentoAutorizadoPorID, &it.DescuentoAutorizadoPor, &it.TotalCentavos, &it.Aprobado, &it.CitaID, &cumplido); err != nil {
			return nil, err
		}
		it.Estado = models.EstadoDeItem(it.Aprobado, it.CitaID != nil, cumplido)
		items = append(items, it)
	}
	return items, rows.Err()
}

func (pg *PostgresRepository) completarPresupuesto(ctx context.Context, p *models.Presupuesto) error {
	var err error
	if p.Items, err = pg.itemsDe(ctx, p.ID); err != nil {
		return err
	}
	p.HonorariosCentavos, p.LaboratorioCentavos, p.DescuentosCentavos, p.TotalCentavos = models.SumarTotales(p.Items)
	for _, it := range p.Items {
		if it.Aprobado != nil && *it.Aprobado {
			p.AprobadoCentavos += it.TotalCentavos
		}
	}
	return nil
}

// ListarPresupuestos devuelve los presupuestos del paciente, el más reciente primero, con sus ítems y totales.
func (pg *PostgresRepository) ListarPresupuestos(ctx context.Context, pacienteID int) ([]models.Presupuesto, error) {
	rows, err := pg.db.Query(ctx, `SELECT `+columnasPresupuesto+` `+origenPresupuesto+` WHERE p.paciente_id = $1 ORDER BY p.id DESC`, pacienteID)
	if err != nil {
		return nil, err
	}
	lista := []models.Presupuesto{}
	for rows.Next() {
		var p models.Presupuesto
		if err := escanearPresupuesto(rows, &p); err != nil {
			rows.Close()
			return nil, err
		}
		lista = append(lista, p)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, err
	}
	for i := range lista {
		if err := pg.completarPresupuesto(ctx, &lista[i]); err != nil {
			return nil, err
		}
	}
	return lista, nil
}

// ObtenerPresupuesto devuelve el presupuesto del paciente; ErrNoEncontrado si no existe.
func (pg *PostgresRepository) ObtenerPresupuesto(ctx context.Context, pacienteID, id int) (models.Presupuesto, error) {
	var p models.Presupuesto
	err := escanearPresupuesto(pg.db.QueryRow(ctx, `SELECT `+columnasPresupuesto+` `+origenPresupuesto+` WHERE p.id = $1 AND p.paciente_id = $2`, id, pacienteID), &p)
	if errors.Is(err, pgx.ErrNoRows) {
		return p, models.ErrNoEncontrado
	}
	if err != nil {
		return p, err
	}
	return p, pg.completarPresupuesto(ctx, &p)
}

// bloquearPresupuesto toma el presupuesto con bloqueo de fila y devuelve su estado.
func bloquearPresupuesto(ctx context.Context, tx pgx.Tx, pacienteID, id int) (estado string, err error) {
	err = tx.QueryRow(ctx, `SELECT estado FROM presupuestos WHERE id = $1 AND paciente_id = $2 FOR UPDATE`, id, pacienteID).Scan(&estado)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", models.ErrNoEncontrado
	}
	return estado, err
}

func insertarItems(ctx context.Context, tx pgx.Tx, presupuestoID, pacienteID int, items []models.ItemParaGuardar) error {
	for orden, it := range items {
		var descripcion string
		var precio int
		var activo bool
		err := tx.QueryRow(ctx, `SELECT nombre, precio_centavos, activo FROM procedimientos WHERE id = $1`, it.ProcedimientoID).Scan(&descripcion, &precio, &activo)
		if errors.Is(err, pgx.ErrNoRows) || (err == nil && !activo) {
			return fmt.Errorf("ítem %d: %w", orden+1, models.ErrReferencia)
		}
		if err != nil {
			return err
		}
		honorario := precio
		if it.HonorarioCentavos != nil {
			honorario = *it.HonorarioCentavos
		}
		total, motivo := models.TotalItem(honorario, it.LaboratorioCentavos, it.DescuentoCentavos)
		if motivo != "" {
			return fmt.Errorf("ítem %d: %s: %w", orden+1, motivo, models.ErrImporteInvalido)
		}
		if it.HallazgoID != nil {
			var propio bool
			if err := tx.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM odontograma_hallazgos WHERE id = $1 AND paciente_id = $2)`, *it.HallazgoID, pacienteID).Scan(&propio); err != nil {
				return err
			}
			if !propio {
				return fmt.Errorf("ítem %d: el hallazgo no es de este paciente: %w", orden+1, models.ErrReferencia)
			}
		}
		if _, err := tx.Exec(ctx,
			`INSERT INTO presupuesto_items (presupuesto_id, fase, sesion, procedimiento_id, descripcion, pieza, superficie, hallazgo_id,
			        honorario_centavos, laboratorio_centavos, descuento_centavos, descuento_autorizado_por, total_centavos, aprobado, orden)
			 VALUES ($1, $2, $3, $4, $5, $6, NULLIF($7, ''), $8, $9, $10, $11, $12, $13, $14, $15)`,
			presupuestoID, it.Fase, it.Sesion, it.ProcedimientoID, descripcion, it.Pieza, it.Superficie, it.HallazgoID,
			honorario, it.LaboratorioCentavos, it.DescuentoCentavos, it.DescuentoAutorizadoPorID, total, it.Aprobado, orden); err != nil {
			return err
		}
	}
	return nil
}

// GuardarPresupuesto crea un borrador (presupuestoID = 0) o reemplaza los ítems de uno existente. Los importes los
// calcula el servidor; el honorario sin indicar toma el precio base del catálogo. Solo un borrador se edita.
func (pg *PostgresRepository) GuardarPresupuesto(ctx context.Context, pacienteID, presupuestoID, profesionalID int, notas string, items []models.ItemParaGuardar, usuarioID int) (int, error) {
	id := presupuestoID
	err := pgx.BeginFunc(ctx, pg.db, func(tx pgx.Tx) error {
		if err := exigirPaciente(ctx, tx, pacienteID); err != nil {
			return err
		}
		if id == 0 {
			var correlativo int64
			if err := tx.QueryRow(ctx, `UPDATE contadores SET valor = valor + 1 WHERE clave = 'presupuesto' RETURNING valor`).Scan(&correlativo); err != nil {
				return err
			}
			if err := tx.QueryRow(ctx,
				`INSERT INTO presupuestos (numero, paciente_id, profesional_id, notas, creado_por) VALUES ($1, $2, $3, $4, $5) RETURNING id`,
				fmt.Sprintf("P-%06d", correlativo), pacienteID, profesionalID, notas, usuarioID).Scan(&id); err != nil {
				return err
			}
		} else {
			estado, err := bloquearPresupuesto(ctx, tx, pacienteID, id)
			if err != nil {
				return err
			}
			if estado != models.PresupBorrador {
				return models.ErrPresupuestoNoEditable
			}
			if _, err := tx.Exec(ctx, `UPDATE presupuestos SET notas = $2 WHERE id = $1`, id, notas); err != nil {
				return err
			}
			if _, err := tx.Exec(ctx, `DELETE FROM presupuesto_items WHERE presupuesto_id = $1`, id); err != nil {
				return err
			}
		}
		return insertarItems(ctx, tx, id, pacienteID, items)
	})
	if err != nil {
		return 0, traducirError(err)
	}
	return id, nil
}

// PresentarPresupuesto pasa el borrador a presentado: desde entonces no se edita.
func (pg *PostgresRepository) PresentarPresupuesto(ctx context.Context, pacienteID, id, usuarioID int) error {
	err := pgx.BeginFunc(ctx, pg.db, func(tx pgx.Tx) error {
		estado, err := bloquearPresupuesto(ctx, tx, pacienteID, id)
		if err != nil {
			return err
		}
		if estado != models.PresupBorrador {
			return models.ErrPresupuestoNoEditable
		}
		var n int
		if err := tx.QueryRow(ctx, `SELECT COUNT(*) FROM presupuesto_items WHERE presupuesto_id = $1`, id).Scan(&n); err != nil {
			return err
		}
		if n == 0 {
			return models.ErrPresupuestoVacio
		}
		_, err = tx.Exec(ctx, `UPDATE presupuestos SET estado = 'presentado', presentado_por = $2, presentado_en = NOW() WHERE id = $1`, id, usuarioID)
		return err
	})
	return traducirError(err)
}

// ResolverPresupuesto registra la aprobación (total o parcial) o el rechazo: los ítems de la lista quedan aprobados y
// los demás rechazados. Guarda quién y cuándo. Solo un presupuesto presentado se resuelve.
func (pg *PostgresRepository) ResolverPresupuesto(ctx context.Context, pacienteID, id int, aprobados []int, usuarioID int) error {
	err := pgx.BeginFunc(ctx, pg.db, func(tx pgx.Tx) error {
		estado, err := bloquearPresupuesto(ctx, tx, pacienteID, id)
		if err != nil {
			return err
		}
		if estado != models.PresupPresentado {
			return models.ErrPresupuestoNoPresentado
		}
		var total, validos int
		if err := tx.QueryRow(ctx,
			`SELECT COUNT(*), COUNT(*) FILTER (WHERE id = ANY($2::int[])) FROM presupuesto_items WHERE presupuesto_id = $1`, id, aprobados).Scan(&total, &validos); err != nil {
			return err
		}
		if validos != len(aprobados) {
			return models.ErrReferencia
		}
		if _, err := tx.Exec(ctx, `UPDATE presupuesto_items SET aprobado = (id = ANY($2::int[])) WHERE presupuesto_id = $1`, id, aprobados); err != nil {
			return err
		}
		_, err = tx.Exec(ctx, `UPDATE presupuestos SET estado = $2, resuelto_por = $3, resuelto_en = NOW() WHERE id = $1`,
			id, models.EstadoTrasResolver(len(aprobados), total), usuarioID)
		return err
	})
	return traducirError(err)
}

// NuevaVersionPresupuesto copia un presupuesto presentado o resuelto en un borrador nuevo (versión + 1) para poder
// cambiarlo. Cada presupuesto tiene a lo sumo una versión siguiente. Devuelve el ID del borrador.
func (pg *PostgresRepository) NuevaVersionPresupuesto(ctx context.Context, pacienteID, id, usuarioID int) (int, error) {
	var nuevo int
	err := pgx.BeginFunc(ctx, pg.db, func(tx pgx.Tx) error {
		estado, err := bloquearPresupuesto(ctx, tx, pacienteID, id)
		if err != nil {
			return err
		}
		if estado == models.PresupBorrador {
			return models.ErrConflicto
		}
		if err := tx.QueryRow(ctx,
			`INSERT INTO presupuestos (numero, version, origen_id, paciente_id, profesional_id, notas, creado_por)
			 SELECT numero, version + 1, id, paciente_id, profesional_id, notas, $2 FROM presupuestos WHERE id = $1 RETURNING id`, id, usuarioID).Scan(&nuevo); err != nil {
			return err
		}
		_, err = tx.Exec(ctx,
			`INSERT INTO presupuesto_items (presupuesto_id, fase, sesion, procedimiento_id, descripcion, pieza, superficie, hallazgo_id,
			        honorario_centavos, laboratorio_centavos, descuento_centavos, descuento_autorizado_por, total_centavos, orden)
			 SELECT $2, fase, sesion, procedimiento_id, descripcion, pieza, superficie, hallazgo_id, honorario_centavos,
			        laboratorio_centavos, descuento_centavos, descuento_autorizado_por, total_centavos, orden
			   FROM presupuesto_items WHERE presupuesto_id = $1`, id, nuevo)
		return err
	})
	if err != nil {
		return 0, traducirError(err)
	}
	return nuevo, nil
}

// ListarTareas devuelve los ítems aprobados de los presupuestos vigentes (sin una versión más nueva) del paciente: lo
// que se puede agendar, lo ya agendado y lo cumplido.
func (pg *PostgresRepository) ListarTareas(ctx context.Context, pacienteID int) ([]models.Tarea, error) {
	rows, err := pg.db.Query(ctx,
		`SELECT i.id, p.id, p.numero, p.version, i.fase, i.sesion, i.procedimiento_id, i.descripcion, i.pieza, COALESCE(i.superficie, ''),
		        pr.duracion_min, i.total_centavos,
		        (SELECT c.id FROM citas c WHERE c.presupuesto_item_id = i.id AND c.estado NOT IN ('cancelado', 'no_asistio') ORDER BY c.id DESC LIMIT 1),
		        (SELECT c.inicio FROM citas c WHERE c.presupuesto_item_id = i.id AND c.estado NOT IN ('cancelado', 'no_asistio') ORDER BY c.id DESC LIMIT 1),
		        EXISTS (SELECT 1 FROM procedimientos_realizados r WHERE r.presupuesto_item_id = i.id AND r.estado = 'realizado')
		   FROM presupuesto_items i JOIN presupuestos p ON p.id = i.presupuesto_id JOIN procedimientos pr ON pr.id = i.procedimiento_id
		  WHERE p.paciente_id = $1 AND i.aprobado AND NOT EXISTS (SELECT 1 FROM presupuestos n WHERE n.origen_id = p.id)
		  ORDER BY p.id, i.fase, i.sesion, i.orden`, pacienteID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	lista := []models.Tarea{}
	aprobado := true
	for rows.Next() {
		var t models.Tarea
		var cumplido bool
		if err := rows.Scan(&t.ItemID, &t.PresupuestoID, &t.PresupuestoNumero, &t.Version, &t.Fase, &t.Sesion, &t.ProcedimientoID, &t.Descripcion,
			&t.Pieza, &t.Superficie, &t.DuracionMin, &t.TotalCentavos, &t.CitaID, &t.CitaInicio, &cumplido); err != nil {
			return nil, err
		}
		t.Estado = models.EstadoDeItem(&aprobado, t.CitaID != nil, cumplido)
		lista = append(lista, t)
	}
	return lista, rows.Err()
}

// itemAgendable comprueba, dentro de la transacción, que el ítem sea del paciente, esté aprobado, en un presupuesto
// vigente y sin cumplir. Se usa al agendar una cita o iniciar un procedimiento desde una tarea.
func itemAgendable(ctx context.Context, tx pgx.Tx, pacienteID, itemID int) error {
	var ok bool
	err := tx.QueryRow(ctx,
		`SELECT EXISTS (
		    SELECT 1 FROM presupuesto_items i JOIN presupuestos p ON p.id = i.presupuesto_id
		     WHERE i.id = $1 AND p.paciente_id = $2 AND i.aprobado
		       AND NOT EXISTS (SELECT 1 FROM presupuestos n WHERE n.origen_id = p.id)
		       AND NOT EXISTS (SELECT 1 FROM procedimientos_realizados r WHERE r.presupuesto_item_id = i.id AND r.estado = 'realizado'))`,
		itemID, pacienteID).Scan(&ok)
	if err != nil {
		return err
	}
	if !ok {
		return models.ErrReferencia
	}
	return nil
}
