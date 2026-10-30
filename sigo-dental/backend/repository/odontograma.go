// Este archivo agrupa el acceso a datos del odontograma (F2-2: RF-03).
package repository

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"sigo-dental-backend/models"
)

const columnasHallazgo = `h.id, h.paciente_id, h.pieza, COALESCE(h.superficie, ''), h.tipo, h.estado, h.observacion, h.cita_id,
	u.nombre, h.registrado_en, COALESCE(ur.nombre, ''), h.retirado_en, h.reemplaza_a`

const origenHallazgo = `FROM odontograma_hallazgos h
	JOIN usuarios u ON u.id = h.registrado_por
	LEFT JOIN usuarios ur ON ur.id = h.retirado_por`

func escanearHallazgo(row pgx.Row, h *models.Hallazgo) error {
	return row.Scan(&h.ID, &h.PacienteID, &h.Pieza, &h.Superficie, &h.Tipo, &h.Estado, &h.Observacion, &h.CitaID,
		&h.RegistradoPor, &h.RegistradoEn, &h.RetiradoPor, &h.RetiradoEn, &h.ReemplazaA)
}

func leerHallazgos(rows pgx.Rows) ([]models.Hallazgo, error) {
	defer rows.Close()
	lista := []models.Hallazgo{}
	for rows.Next() {
		var h models.Hallazgo
		if err := escanearHallazgo(rows, &h); err != nil {
			return nil, err
		}
		lista = append(lista, h)
	}
	return lista, rows.Err()
}

// ListarHallazgos devuelve los hallazgos vigentes del paciente. Con "al" devuelve el estado que tenía el odontograma
// en ese instante (los registrados hasta entonces y aún no retirados), para revisar visitas anteriores.
func (pg *PostgresRepository) ListarHallazgos(ctx context.Context, pacienteID int, al *time.Time) ([]models.Hallazgo, error) {
	rows, err := pg.db.Query(ctx,
		`SELECT `+columnasHallazgo+` `+origenHallazgo+`
		  WHERE h.paciente_id = $1
		    AND (($2::timestamptz IS NULL AND h.retirado_en IS NULL)
		         OR ($2::timestamptz IS NOT NULL AND h.registrado_en <= $2 AND (h.retirado_en IS NULL OR h.retirado_en > $2)))
		  ORDER BY h.pieza, h.id`, pacienteID, al)
	if err != nil {
		return nil, err
	}
	return leerHallazgos(rows)
}

// HistorialHallazgos devuelve todos los hallazgos del paciente, vigentes y retirados, del más reciente al más antiguo.
func (pg *PostgresRepository) HistorialHallazgos(ctx context.Context, pacienteID int) ([]models.Hallazgo, error) {
	rows, err := pg.db.Query(ctx,
		`SELECT `+columnasHallazgo+` `+origenHallazgo+` WHERE h.paciente_id = $1 ORDER BY h.registrado_en DESC, h.id DESC`, pacienteID)
	if err != nil {
		return nil, err
	}
	return leerHallazgos(rows)
}

// insertarHallazgo guarda el hallazgo ligado a la cita En atención del paciente, si la hay, y lo devuelve completo.
func insertarHallazgo(ctx context.Context, tx pgx.Tx, n models.NuevoHallazgo, reemplazaA *int, usuarioID int) (models.Hallazgo, error) {
	var id int
	err := tx.QueryRow(ctx,
		`INSERT INTO odontograma_hallazgos (paciente_id, pieza, superficie, tipo, estado, observacion, cita_id, registrado_por, reemplaza_a)
		 VALUES ($1, $2, NULLIF($3, ''), $4, $5, $6,
		         (SELECT c.id FROM citas c WHERE c.paciente_id = $1 AND c.estado = 'en_atencion' ORDER BY c.inicio DESC LIMIT 1),
		         $7, $8)
		 RETURNING id`,
		n.PacienteID, n.Pieza, n.Superficie, n.Tipo, n.Estado, n.Observacion, usuarioID, reemplazaA).Scan(&id)
	if err != nil {
		return models.Hallazgo{}, err
	}
	var h models.Hallazgo
	err = escanearHallazgo(tx.QueryRow(ctx, `SELECT `+columnasHallazgo+` `+origenHallazgo+` WHERE h.id = $1`, id), &h)
	return h, err
}

// RegistrarHallazgo agrega una marca al odontograma; ErrDuplicado si ya hay una vigente igual.
func (pg *PostgresRepository) RegistrarHallazgo(ctx context.Context, n models.NuevoHallazgo, usuarioID int) (models.Hallazgo, error) {
	var h models.Hallazgo
	err := pgx.BeginFunc(ctx, pg.db, func(tx pgx.Tx) error {
		if err := exigirPaciente(ctx, tx, n.PacienteID); err != nil {
			return err
		}
		var err error
		h, err = insertarHallazgo(ctx, tx, n, nil, usuarioID)
		return err
	})
	if err != nil {
		return models.Hallazgo{}, traducirError(err)
	}
	return h, nil
}

// bloquearHallazgoVigente toma el hallazgo con bloqueo de fila y comprueba que se pueda corregir o retirar:
// existe, es del paciente, sigue vigente y no es un tratamiento realizado (RN-01).
func bloquearHallazgoVigente(ctx context.Context, tx pgx.Tx, pacienteID, id int) error {
	var estado string
	var retirado bool
	err := tx.QueryRow(ctx,
		`SELECT estado, retirado_en IS NOT NULL FROM odontograma_hallazgos WHERE id = $1 AND paciente_id = $2 FOR UPDATE`,
		id, pacienteID).Scan(&estado, &retirado)
	if errors.Is(err, pgx.ErrNoRows) {
		return models.ErrNoEncontrado
	}
	if err != nil {
		return err
	}
	if retirado {
		return models.ErrHallazgoRetirado
	}
	if estado == models.HallazgoRealizado {
		return models.ErrInmutable
	}
	return nil
}

const retirarHallazgo = `UPDATE odontograma_hallazgos SET retirado_en = NOW(), retirado_por = $2 WHERE id = $1`

// CorregirHallazgo retira el hallazgo y registra el corregido en su lugar, en una sola transacción: el anterior
// queda en el historial. Un hallazgo realizado devuelve ErrInmutable.
func (pg *PostgresRepository) CorregirHallazgo(ctx context.Context, id int, n models.NuevoHallazgo, usuarioID int) (models.Hallazgo, error) {
	var h models.Hallazgo
	err := pgx.BeginFunc(ctx, pg.db, func(tx pgx.Tx) error {
		if err := bloquearHallazgoVigente(ctx, tx, n.PacienteID, id); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, retirarHallazgo, id, usuarioID); err != nil {
			return err
		}
		var err error
		h, err = insertarHallazgo(ctx, tx, n, &id, usuarioID)
		return err
	})
	if err != nil {
		return models.Hallazgo{}, traducirError(err)
	}
	return h, nil
}

// RetirarHallazgo deja el hallazgo fuera del odontograma vigente; sigue en el historial. Un hallazgo realizado
// devuelve ErrInmutable: se corrige con una nota de adenda (RN-01).
func (pg *PostgresRepository) RetirarHallazgo(ctx context.Context, pacienteID, id, usuarioID int) error {
	err := pgx.BeginFunc(ctx, pg.db, func(tx pgx.Tx) error {
		if err := bloquearHallazgoVigente(ctx, tx, pacienteID, id); err != nil {
			return err
		}
		_, err := tx.Exec(ctx, retirarHallazgo, id, usuarioID)
		return err
	})
	return traducirError(err)
}

// PiezasAusentes devuelve las piezas que el odontograma vigente marca como ausentes o con exodoncia realizada
// (el periodontograma no las ofrece).
func (pg *PostgresRepository) PiezasAusentes(ctx context.Context, pacienteID int) ([]int, error) {
	rows, err := pg.db.Query(ctx,
		`SELECT DISTINCT pieza FROM odontograma_hallazgos
		  WHERE paciente_id = $1 AND retirado_en IS NULL AND (tipo = 'ausente' OR (tipo = 'exodoncia' AND estado = 'realizado'))
		  ORDER BY pieza`, pacienteID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	piezas := []int{}
	for rows.Next() {
		var p int
		if err := rows.Scan(&p); err != nil {
			return nil, err
		}
		piezas = append(piezas, p)
	}
	return piezas, rows.Err()
}
