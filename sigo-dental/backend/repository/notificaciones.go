// Este archivo agrupa el acceso a datos de las notificaciones de citas (F1-4).
package repository

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"sigo-dental-backend/models"
)

const columnasParaNotificar = `c.id, c.estado, c.inicio, c.actualizado_en, p.nombres || ' ' || p.apellidos, p.telefono, p.email,
	p.canal_preferido, u.nombre,
	COALESCE((SELECT valor FROM parametros WHERE clave = 'zona_horaria'), 'America/Guayaquil')`

const origenParaNotificar = `FROM citas c
	JOIN pacientes p ON p.id = c.paciente_id
	JOIN profesionales pf ON pf.id = c.profesional_id
	JOIN usuarios u ON u.id = pf.usuario_id`

func escanearParaNotificar(row pgx.Row, d *models.DatosParaNotificar) error {
	return row.Scan(&d.CitaID, &d.Estado, &d.Inicio, &d.Actualizada, &d.PacienteNombre, &d.Telefono, &d.Email,
		&d.Canal, &d.ProfesionalNombre, &d.Zona)
}

// DatosParaNotificar devuelve lo necesario para redactar los mensajes de la cita; ErrNoEncontrado si no existe.
func (pg *PostgresRepository) DatosParaNotificar(ctx context.Context, citaID int) (models.DatosParaNotificar, error) {
	var d models.DatosParaNotificar
	err := escanearParaNotificar(pg.db.QueryRow(ctx, `SELECT `+columnasParaNotificar+` `+origenParaNotificar+` WHERE c.id = $1`, citaID), &d)
	if errors.Is(err, pgx.ErrNoRows) {
		return models.DatosParaNotificar{}, models.ErrNoEncontrado
	}
	return d, err
}

// CitasParaRecordar devuelve las citas programadas o confirmadas que empiezan dentro del rango.
func (pg *PostgresRepository) CitasParaRecordar(ctx context.Context, desde, hasta time.Time) ([]models.DatosParaNotificar, error) {
	rows, err := pg.db.Query(ctx,
		`SELECT `+columnasParaNotificar+` `+origenParaNotificar+`
		  WHERE c.estado IN ('programado', 'confirmado') AND c.inicio > $1 AND c.inicio <= $2
		  ORDER BY c.inicio`, desde, hasta)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	lista := []models.DatosParaNotificar{}
	for rows.Next() {
		var d models.DatosParaNotificar
		if err := escanearParaNotificar(rows, &d); err != nil {
			return nil, err
		}
		lista = append(lista, d)
	}
	return lista, rows.Err()
}

// ReservarNotificacion inserta la notificación como pendiente. Si ya hay una vigente del mismo tipo para
// la cita, no inserta nada y devuelve creada = false: el índice único parcial lo decide de forma atómica.
func (pg *PostgresRepository) ReservarNotificacion(ctx context.Context, n models.Notificacion) (int, bool, error) {
	var id int
	err := pg.db.QueryRow(ctx,
		`INSERT INTO notificaciones (cita_id, tipo, canal, destino, mensaje)
		 VALUES ($1, $2, $3, $4, $5)
		 ON CONFLICT (cita_id, tipo) WHERE vigente DO NOTHING
		 RETURNING id`, n.CitaID, n.Tipo, n.Canal, n.Destino, n.Mensaje).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, false, nil
	}
	if err != nil {
		return 0, false, traducirError(err)
	}
	return id, true, nil
}

// RegistrarResultado deja la notificación como enviada o fallida, con el motivo del fallo.
func (pg *PostgresRepository) RegistrarResultado(ctx context.Context, id int, enviada bool, detalle string) error {
	estado := models.NotifFallida
	if enviada {
		estado = models.NotifEnviada
	}
	if len(detalle) > 300 {
		detalle = detalle[:300]
	}
	_, err := pg.db.Exec(ctx,
		`UPDATE notificaciones SET estado = $2, detalle = $3, enviado_en = CASE WHEN $4::boolean THEN NOW() END WHERE id = $1`,
		id, estado, detalle, enviada)
	return err
}

// InvalidarNotificaciones marca como no vigentes las notificaciones de la cita (siguen en la bitácora).
func (pg *PostgresRepository) InvalidarNotificaciones(ctx context.Context, citaID int) error {
	_, err := pg.db.Exec(ctx, `UPDATE notificaciones SET vigente = FALSE WHERE cita_id = $1 AND vigente`, citaID)
	return err
}

// ListarNotificaciones devuelve la bitácora de la cita, de la más antigua a la más reciente.
func (pg *PostgresRepository) ListarNotificaciones(ctx context.Context, citaID int) ([]models.Notificacion, error) {
	rows, err := pg.db.Query(ctx,
		`SELECT id, cita_id, tipo, canal, destino, mensaje, estado, detalle, vigente, creado_en, enviado_en
		   FROM notificaciones WHERE cita_id = $1 ORDER BY id`, citaID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	lista := []models.Notificacion{}
	for rows.Next() {
		var n models.Notificacion
		if err := rows.Scan(&n.ID, &n.CitaID, &n.Tipo, &n.Canal, &n.Destino, &n.Mensaje, &n.Estado, &n.Detalle,
			&n.Vigente, &n.CreadoEn, &n.EnviadoEn); err != nil {
			return nil, err
		}
		lista = append(lista, n)
	}
	return lista, rows.Err()
}
