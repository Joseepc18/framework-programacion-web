// Este archivo agrupa el acceso a datos de las recetas (F2-6).
package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"sigo-dental-backend/models"
)

const columnasReceta = `r.id, r.numero, r.paciente_id, r.profesional_id, u.nombre, r.cita_id, r.indicaciones, r.creada_en`

func escanearReceta(row pgx.Row, r *models.Receta) error {
	return row.Scan(&r.ID, &r.Numero, &r.PacienteID, &r.ProfesionalID, &r.ProfesionalNombre, &r.CitaID, &r.Indicaciones, &r.CreadaEn)
}

const origenReceta = `FROM recetas r JOIN profesionales pf ON pf.id = r.profesional_id JOIN usuarios u ON u.id = pf.usuario_id`

// completarReceta agrega los medicamentos, la firma y la anulación, y deriva el estado.
func (pg *PostgresRepository) completarReceta(ctx context.Context, r *models.Receta) error {
	rows, err := pg.db.Query(ctx, `SELECT nombre, dosis, frecuencia, duracion, indicaciones FROM receta_medicamentos WHERE receta_id = $1 ORDER BY orden`, r.ID)
	if err != nil {
		return err
	}
	r.Medicamentos = []models.Medicamento{}
	for rows.Next() {
		var m models.Medicamento
		if err := rows.Scan(&m.Nombre, &m.Dosis, &m.Frecuencia, &m.Duracion, &m.Indicaciones); err != nil {
			rows.Close()
			return err
		}
		r.Medicamentos = append(r.Medicamentos, m)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return err
	}
	if r.Firma, err = pg.firmaDe(ctx, models.FirmaReceta, r.ID); err != nil {
		return err
	}
	var a models.AnulacionReceta
	err = pg.db.QueryRow(ctx,
		`SELECT a.motivo, u.nombre, a.reemplazo_id, a.creada_en FROM receta_anulaciones a JOIN usuarios u ON u.id = a.usuario_id WHERE a.receta_id = $1`, r.ID,
	).Scan(&a.Motivo, &a.UsuarioNombre, &a.ReemplazoID, &a.CreadaEn)
	switch {
	case err == nil:
		r.Anulacion = &a
		r.Estado = models.RecetaAnulada
	case errors.Is(err, pgx.ErrNoRows):
		r.Estado = models.RecetaBorrador
		if r.Firma != nil {
			r.Estado = models.RecetaFirmada
		}
	default:
		return err
	}
	r.Avisos = []models.AvisoAlergia{}
	return nil
}

// ListarRecetas devuelve el historial de recetas del paciente, de la más reciente a la más antigua.
func (pg *PostgresRepository) ListarRecetas(ctx context.Context, pacienteID int) ([]models.Receta, error) {
	rows, err := pg.db.Query(ctx, `SELECT `+columnasReceta+` `+origenReceta+` WHERE r.paciente_id = $1 ORDER BY r.id DESC`, pacienteID)
	if err != nil {
		return nil, err
	}
	lista := []models.Receta{}
	for rows.Next() {
		var r models.Receta
		if err := escanearReceta(rows, &r); err != nil {
			rows.Close()
			return nil, err
		}
		lista = append(lista, r)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, err
	}
	for i := range lista {
		if err := pg.completarReceta(ctx, &lista[i]); err != nil {
			return nil, err
		}
	}
	return lista, nil
}

// ObtenerReceta devuelve la receta del paciente; ErrNoEncontrado si no existe.
func (pg *PostgresRepository) ObtenerReceta(ctx context.Context, pacienteID, id int) (models.Receta, error) {
	var r models.Receta
	err := escanearReceta(pg.db.QueryRow(ctx, `SELECT `+columnasReceta+` `+origenReceta+` WHERE r.id = $1 AND r.paciente_id = $2`, id, pacienteID), &r)
	if errors.Is(err, pgx.ErrNoRows) {
		return r, models.ErrNoEncontrado
	}
	if err != nil {
		return r, err
	}
	return r, pg.completarReceta(ctx, &r)
}

func guardarMedicamentos(ctx context.Context, tx pgx.Tx, recetaID int, meds []models.Medicamento) error {
	for i, m := range meds {
		if _, err := tx.Exec(ctx,
			`INSERT INTO receta_medicamentos (receta_id, orden, nombre, dosis, frecuencia, duracion, indicaciones) VALUES ($1, $2, $3, $4, $5, $6, $7)`,
			recetaID, i+1, m.Nombre, m.Dosis, m.Frecuencia, m.Duracion, m.Indicaciones); err != nil {
			return err
		}
	}
	return nil
}

// crearRecetaTx numera y guarda una receta. El número sale de una fila de contadores que se incrementa dentro de la
// misma transacción: las solicitudes simultáneas se serializan por el bloqueo de la fila y el número no se repite.
func crearRecetaTx(ctx context.Context, tx pgx.Tx, r models.Receta, usuarioID int) (int, error) {
	if err := exigirPaciente(ctx, tx, r.PacienteID); err != nil {
		return 0, err
	}
	if r.CitaID != nil {
		var coincide bool
		if err := tx.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM citas WHERE id = $1 AND paciente_id = $2)`, *r.CitaID, r.PacienteID).Scan(&coincide); err != nil {
			return 0, err
		}
		if !coincide {
			return 0, models.ErrReferencia
		}
	}
	var correlativo int64
	if err := tx.QueryRow(ctx, `UPDATE contadores SET valor = valor + 1 WHERE clave = 'receta' RETURNING valor`).Scan(&correlativo); err != nil {
		return 0, err
	}
	var id int
	if err := tx.QueryRow(ctx,
		`INSERT INTO recetas (numero, paciente_id, profesional_id, cita_id, indicaciones, creada_por) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`,
		fmt.Sprintf("R-%06d", correlativo), r.PacienteID, r.ProfesionalID, r.CitaID, r.Indicaciones, usuarioID).Scan(&id); err != nil {
		return 0, err
	}
	return id, guardarMedicamentos(ctx, tx, id, r.Medicamentos)
}

// CrearReceta guarda un borrador numerado y devuelve su ID.
func (pg *PostgresRepository) CrearReceta(ctx context.Context, r models.Receta, usuarioID int) (int, error) {
	var id int
	err := pgx.BeginFunc(ctx, pg.db, func(tx pgx.Tx) error {
		var err error
		id, err = crearRecetaTx(ctx, tx, r, usuarioID)
		return err
	})
	if err != nil {
		return 0, traducirError(err)
	}
	return id, nil
}

// ActualizarReceta reemplaza las indicaciones y los medicamentos de un borrador. Si la receta ya está firmada, el
// disparador de la base lo rechaza (models.ErrInmutable).
func (pg *PostgresRepository) ActualizarReceta(ctx context.Context, pacienteID, id int, indicaciones string, meds []models.Medicamento) error {
	err := pgx.BeginFunc(ctx, pg.db, func(tx pgx.Tx) error {
		tag, err := tx.Exec(ctx, `UPDATE recetas SET indicaciones = $3 WHERE id = $2 AND paciente_id = $1`, pacienteID, id, indicaciones)
		if err != nil {
			return err
		}
		if tag.RowsAffected() == 0 {
			return models.ErrNoEncontrado
		}
		if _, err := tx.Exec(ctx, `DELETE FROM receta_medicamentos WHERE receta_id = $1`, id); err != nil {
			return err
		}
		return guardarMedicamentos(ctx, tx, id, meds)
	})
	return traducirError(err)
}

// AnularReceta registra la anulación (quién, cuándo y por qué). Con copiar crea además un borrador nuevo con los mismos
// medicamentos que la reemplaza. Devuelve el ID de la receta nueva, o 0 si no se copió.
func (pg *PostgresRepository) AnularReceta(ctx context.Context, pacienteID, id int, motivo string, usuarioID int, copiar bool) (int, error) {
	var nuevoID int
	err := pgx.BeginFunc(ctx, pg.db, func(tx pgx.Tx) error {
		var r models.Receta
		if err := tx.QueryRow(ctx, `SELECT id, paciente_id, profesional_id, cita_id, indicaciones FROM recetas WHERE id = $1 AND paciente_id = $2 FOR UPDATE`, id, pacienteID).
			Scan(&r.ID, &r.PacienteID, &r.ProfesionalID, &r.CitaID, &r.Indicaciones); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return models.ErrNoEncontrado
			}
			return err
		}
		var anulada bool
		if err := tx.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM receta_anulaciones WHERE receta_id = $1)`, id).Scan(&anulada); err != nil {
			return err
		}
		if anulada {
			return models.ErrConflicto
		}
		var reemplazo *int
		if copiar {
			rows, err := tx.Query(ctx, `SELECT nombre, dosis, frecuencia, duracion, indicaciones FROM receta_medicamentos WHERE receta_id = $1 ORDER BY orden`, id)
			if err != nil {
				return err
			}
			for rows.Next() {
				var m models.Medicamento
				if err := rows.Scan(&m.Nombre, &m.Dosis, &m.Frecuencia, &m.Duracion, &m.Indicaciones); err != nil {
					rows.Close()
					return err
				}
				r.Medicamentos = append(r.Medicamentos, m)
			}
			rows.Close()
			if err := rows.Err(); err != nil {
				return err
			}
			if nuevoID, err = crearRecetaTx(ctx, tx, r, usuarioID); err != nil {
				return err
			}
			reemplazo = &nuevoID
		}
		_, err := tx.Exec(ctx, `INSERT INTO receta_anulaciones (receta_id, motivo, usuario_id, reemplazo_id) VALUES ($1, $2, $3, $4)`, id, motivo, usuarioID, reemplazo)
		return err
	})
	if err != nil {
		return 0, traducirError(err)
	}
	return nuevoID, nil
}
