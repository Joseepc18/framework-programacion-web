// Este archivo agrupa el acceso a datos de la agenda de citas (F1-3).
package repository

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"sigo-dental-backend/models"
)

// limiteCitas acota las filas que devuelve el calendario para un rango.
const limiteCitas = 2000

const columnasCita = `c.id, c.estado, c.paciente_id, p.nombres || ' ' || p.apellidos, p.numero_historia, ` + alertasDe + `,
	c.profesional_id, u.nombre, c.sillon_id, s.nombre, s.consultorio_id, co.nombre,
	c.procedimiento_id, COALESCE(pr.nombre, ''), c.motivo, c.notas, c.inicio, c.fin`

const origenCita = `FROM citas c
	JOIN pacientes p ON p.id = c.paciente_id
	JOIN profesionales pf ON pf.id = c.profesional_id
	JOIN usuarios u ON u.id = pf.usuario_id
	JOIN sillones s ON s.id = c.sillon_id
	JOIN consultorios co ON co.id = s.consultorio_id
	LEFT JOIN procedimientos pr ON pr.id = c.procedimiento_id`

func escanearCita(row pgx.Row, c *models.Cita) error {
	return row.Scan(&c.ID, &c.Estado, &c.PacienteID, &c.PacienteNombre, &c.NumeroHistoria, &c.Alertas,
		&c.ProfesionalID, &c.ProfesionalNombre, &c.SillonID, &c.SillonNombre, &c.ConsultorioID, &c.ConsultorioNombre,
		&c.ProcedimientoID, &c.ProcedimientoNombre, &c.Motivo, &c.Notas, &c.Inicio, &c.Fin)
}

// ListarCitas devuelve las citas que tocan el rango pedido, con los filtros opcionales aplicados.
func (pg *PostgresRepository) ListarCitas(ctx context.Context, f models.FiltroCitas) ([]models.Cita, error) {
	rows, err := pg.db.Query(ctx,
		`SELECT `+columnasCita+` `+origenCita+`
		  WHERE c.inicio < $2 AND c.fin > $1
		    AND ($3 = 0 OR s.consultorio_id = $3)
		    AND ($4 = 0 OR c.sillon_id = $4)
		    AND ($5 = 0 OR c.profesional_id = $5)
		    AND ($6 = 0 OR c.paciente_id = $6)
		  ORDER BY c.inicio, c.id
		  LIMIT $7`,
		f.Desde, f.Hasta, f.ConsultorioID, f.SillonID, f.ProfesionalID, f.PacienteID, limiteCitas)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	lista := []models.Cita{}
	for rows.Next() {
		var c models.Cita
		if err := escanearCita(rows, &c); err != nil {
			return nil, err
		}
		lista = append(lista, c)
	}
	return lista, rows.Err()
}

// ObtenerCita devuelve la cita con su historial de estados; ErrNoEncontrado si no existe.
func (pg *PostgresRepository) ObtenerCita(ctx context.Context, id int) (models.CitaDetalle, error) {
	var d models.CitaDetalle
	err := escanearCita(pg.db.QueryRow(ctx, `SELECT `+columnasCita+` `+origenCita+` WHERE c.id = $1`, id), &d.Cita)
	if errors.Is(err, pgx.ErrNoRows) {
		return models.CitaDetalle{}, models.ErrNoEncontrado
	}
	if err != nil {
		return models.CitaDetalle{}, err
	}

	rows, err := pg.db.Query(ctx,
		`SELECT COALESCE(h.estado_anterior, ''), h.estado_nuevo, h.motivo, COALESCE(u.nombre, 'Paciente (enlace de confirmación)'), h.creado_en
		   FROM cita_historial h LEFT JOIN usuarios u ON u.id = h.usuario_id
		  WHERE h.cita_id = $1 ORDER BY h.id`, id)
	if err != nil {
		return models.CitaDetalle{}, err
	}
	defer rows.Close()
	d.Historial = []models.CambioEstado{}
	for rows.Next() {
		var h models.CambioEstado
		if err := rows.Scan(&h.EstadoAnterior, &h.EstadoNuevo, &h.Motivo, &h.UsuarioNombre, &h.Fecha); err != nil {
			return models.CitaDetalle{}, err
		}
		d.Historial = append(d.Historial, h)
	}
	return d, rows.Err()
}

// referenciasActivas comprueba que el paciente, el profesional, el sillón y el procedimiento existan y estén activos.
func referenciasActivas(ctx context.Context, tx pgx.Tx, pacienteID int, d models.DatosCita) (bool, error) {
	var ok bool
	err := tx.QueryRow(ctx,
		`SELECT EXISTS (SELECT 1 FROM pacientes WHERE id = $1 AND activo)
		    AND EXISTS (SELECT 1 FROM profesionales WHERE id = $2 AND activo)
		    AND EXISTS (SELECT 1 FROM sillones s JOIN consultorios co ON co.id = s.consultorio_id
		                 WHERE s.id = $3 AND s.activo AND co.activo)
		    AND ($4::int IS NULL OR EXISTS (SELECT 1 FROM procedimientos WHERE id = $4 AND activo))`,
		pacienteID, d.ProfesionalID, d.SillonID, d.ProcedimientoID).Scan(&ok)
	return ok, err
}

// CrearCita agenda la cita en estado Programado. Si el sillón ya está ocupado en ese rango (RN-03),
// la restricción de exclusión de la base responde y se devuelve ErrConflicto.
func (pg *PostgresRepository) CrearCita(ctx context.Context, d models.DatosCita, usuarioID int) (int, error) {
	var id int
	err := pgx.BeginFunc(ctx, pg.db, func(tx pgx.Tx) error {
		ok, err := referenciasActivas(ctx, tx, d.PacienteID, d)
		if err != nil {
			return err
		}
		if !ok {
			return models.ErrReferencia
		}
		if d.PresupuestoItemID != nil {
			if err := itemAgendable(ctx, tx, d.PacienteID, *d.PresupuestoItemID); err != nil {
				return err
			}
		}
		if err := tx.QueryRow(ctx,
			`INSERT INTO citas (paciente_id, profesional_id, sillon_id, procedimiento_id, motivo, notas, inicio, fin, creado_por, presupuesto_item_id)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) RETURNING id`,
			d.PacienteID, d.ProfesionalID, d.SillonID, d.ProcedimientoID, d.Motivo, d.Notas, d.Inicio, d.Fin, usuarioID, d.PresupuestoItemID,
		).Scan(&id); err != nil {
			return err
		}
		_, err = tx.Exec(ctx,
			`INSERT INTO cita_historial (cita_id, estado_nuevo, motivo, usuario_id) VALUES ($1, $2, 'Cita agendada', $3)`,
			id, models.EstadoProgramado, usuarioID)
		return err
	})
	if err != nil {
		return 0, traducirError(err)
	}
	return id, nil
}

// ReprogramarCita cambia horario, sillón, profesional o detalle de una cita que aún no empezó
// (Programado o Confirmado). Vuelve a Programado: el paciente debe confirmar de nuevo.
// ErrConflicto si la cita ya no se puede mover o el sillón está ocupado.
func (pg *PostgresRepository) ReprogramarCita(ctx context.Context, id int, d models.DatosCita, usuarioID int) error {
	err := pgx.BeginFunc(ctx, pg.db, func(tx pgx.Tx) error {
		var estado string
		var pacienteID int
		err := tx.QueryRow(ctx, `SELECT estado, paciente_id FROM citas WHERE id = $1 FOR UPDATE`, id).Scan(&estado, &pacienteID)
		if errors.Is(err, pgx.ErrNoRows) {
			return models.ErrNoEncontrado
		}
		if err != nil {
			return err
		}
		if estado != models.EstadoProgramado && estado != models.EstadoConfirmado {
			return models.ErrEstadoCambiado
		}
		ok, err := referenciasActivas(ctx, tx, pacienteID, d)
		if err != nil {
			return err
		}
		if !ok {
			return models.ErrReferencia
		}
		if _, err := tx.Exec(ctx,
			`UPDATE citas SET profesional_id = $2, sillon_id = $3, procedimiento_id = $4, motivo = $5, notas = $6,
			        inicio = $7, fin = $8, estado = $9, actualizado_en = NOW()
			  WHERE id = $1`,
			id, d.ProfesionalID, d.SillonID, d.ProcedimientoID, d.Motivo, d.Notas, d.Inicio, d.Fin, models.EstadoProgramado); err != nil {
			return err
		}
		_, err = tx.Exec(ctx,
			`INSERT INTO cita_historial (cita_id, estado_anterior, estado_nuevo, motivo, usuario_id)
			 VALUES ($1, $2, $3, 'Cita reprogramada', $4)`, id, estado, models.EstadoProgramado, usuarioID)
		return err
	})
	return traducirError(err)
}

// CambiarEstado (usuarioID 0 = el propio paciente, desde su enlace) mueve la cita de un estado a otro y lo deja en el historial. La actualización exige que
// la cita siga en el estado de origen: si otra persona la cambió antes, devuelve ErrConflicto.
func (pg *PostgresRepository) CambiarEstado(ctx context.Context, id int, desde, hasta, motivo string, usuarioID int) error {
	err := pgx.BeginFunc(ctx, pg.db, func(tx pgx.Tx) error {
		tag, err := tx.Exec(ctx,
			`UPDATE citas SET estado = $3, actualizado_en = NOW() WHERE id = $1 AND estado = $2`, id, desde, hasta)
		if err != nil {
			return err
		}
		if tag.RowsAffected() == 0 {
			return models.ErrEstadoCambiado
		}
		_, err = tx.Exec(ctx,
			`INSERT INTO cita_historial (cita_id, estado_anterior, estado_nuevo, motivo, usuario_id)
			 VALUES ($1, $2, $3, $4, NULLIF($5, 0))`, id, desde, hasta, motivo, usuarioID)
		return err
	})
	return traducirError(err)
}

// CitasSimultaneasDelProfesional cuenta las citas vigentes del profesional que se cruzan con el rango,
// sin contar la cita excluida (cero si no se excluye ninguna). Sirve para avisar, no para bloquear.
func (pg *PostgresRepository) CitasSimultaneasDelProfesional(ctx context.Context, profesionalID int, inicio, fin time.Time, excluirID int) (int, error) {
	var n int
	err := pg.db.QueryRow(ctx,
		`SELECT count(*) FROM citas
		  WHERE profesional_id = $1 AND id <> $4 AND inicio < $3 AND fin > $2
		    AND estado NOT IN ('cancelado', 'no_asistio')`,
		profesionalID, inicio, fin, excluirID).Scan(&n)
	return n, err
}

// HorarioAgenda lee del catálogo de parámetros el horario operativo y la zona horaria de la clínica.
func (pg *PostgresRepository) HorarioAgenda(ctx context.Context) (models.HorarioAgenda, error) {
	var h models.HorarioAgenda
	err := pg.db.QueryRow(ctx,
		`SELECT
		    COALESCE((SELECT valor FROM parametros WHERE clave = 'agenda_hora_inicio'), '07:00'),
		    COALESCE((SELECT valor FROM parametros WHERE clave = 'agenda_hora_fin'), '21:00'),
		    COALESCE((SELECT valor FROM parametros WHERE clave = 'zona_horaria'), 'America/Guayaquil')`).Scan(&h.Inicio, &h.Fin, &h.Zona)
	return h, err
}
