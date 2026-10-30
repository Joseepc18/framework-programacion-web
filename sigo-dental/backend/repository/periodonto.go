// Este archivo agrupa el acceso a datos del periodontograma (F2-3: RF-04).
package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"sigo-dental-backend/models"
)

// ListarPeriodontogramas devuelve los exámenes del paciente con indicadores para compararlos, del más reciente al
// más antiguo, y las piezas que el odontograma marca como ausentes.
func (pg *PostgresRepository) ListarPeriodontogramas(ctx context.Context, pacienteID int) (models.ListaPerio, error) {
	lista := models.ListaPerio{Examenes: []models.ResumenExamenPerio{}}
	rows, err := pg.db.Query(ctx,
		`SELECT e.id, to_char(e.fecha, 'YYYY-MM-DD'), u.nombre,
		        (SELECT COUNT(*) FROM periodonto_dientes d WHERE d.examen_id = e.id)::int,
		        COALESCE((SELECT AVG(s.profundidad_mm) FROM periodonto_sitios s WHERE s.examen_id = e.id), 0)::float8,
		        COALESCE((SELECT AVG(s.profundidad_mm + s.margen_mm) FROM periodonto_sitios s WHERE s.examen_id = e.id), 0)::float8,
		        COALESCE((SELECT 100.0 * COUNT(*) FILTER (WHERE s.sangrado) / NULLIF(COUNT(*), 0)
		                    FROM periodonto_sitios s WHERE s.examen_id = e.id), 0)::float8
		   FROM periodontogramas e JOIN usuarios u ON u.id = e.creado_por
		  WHERE e.paciente_id = $1 ORDER BY e.fecha DESC, e.id DESC`, pacienteID)
	if err != nil {
		return lista, err
	}
	defer rows.Close()
	for rows.Next() {
		var r models.ResumenExamenPerio
		if err := rows.Scan(&r.ID, &r.Fecha, &r.CreadoPor, &r.Dientes, &r.ProfundidadMedia, &r.NICMedio, &r.SangradoPorcentaje); err != nil {
			return lista, err
		}
		lista.Examenes = append(lista.Examenes, r)
	}
	if err := rows.Err(); err != nil {
		return lista, err
	}
	lista.Ausentes, err = pg.PiezasAusentes(ctx, pacienteID)
	return lista, err
}

// ObtenerPeriodontograma devuelve el examen completo, con el NIC de cada punto calculado aquí (no se guarda).
func (pg *PostgresRepository) ObtenerPeriodontograma(ctx context.Context, pacienteID, id int) (models.ExamenPerio, error) {
	var e models.ExamenPerio
	err := pg.db.QueryRow(ctx,
		`SELECT e.id, e.paciente_id, to_char(e.fecha, 'YYYY-MM-DD'), e.notas, u.nombre, e.creado_en
		   FROM periodontogramas e JOIN usuarios u ON u.id = e.creado_por WHERE e.id = $1 AND e.paciente_id = $2`, id, pacienteID,
	).Scan(&e.ID, &e.PacienteID, &e.Fecha, &e.Notas, &e.CreadoPor, &e.CreadoEn)
	if errors.Is(err, pgx.ErrNoRows) {
		return models.ExamenPerio{}, models.ErrNoEncontrado
	}
	if err != nil {
		return models.ExamenPerio{}, err
	}
	dientes, err := pg.db.Query(ctx, `SELECT pieza, movilidad, furca FROM periodonto_dientes WHERE examen_id = $1 ORDER BY pieza`, id)
	if err != nil {
		return models.ExamenPerio{}, err
	}
	defer dientes.Close()
	indice := map[int]int{}
	e.Dientes = []models.DientePerio{}
	for dientes.Next() {
		var d models.DientePerio
		if err := dientes.Scan(&d.Pieza, &d.Movilidad, &d.Furca); err != nil {
			return models.ExamenPerio{}, err
		}
		d.Sitios = []models.Sitio{}
		indice[d.Pieza] = len(e.Dientes)
		e.Dientes = append(e.Dientes, d)
	}
	if err := dientes.Err(); err != nil {
		return models.ExamenPerio{}, err
	}
	sitios, err := pg.db.Query(ctx,
		`SELECT pieza, sitio, profundidad_mm, margen_mm, sangrado FROM periodonto_sitios WHERE examen_id = $1
		  ORDER BY pieza, array_position(ARRAY['DV','V','MV','DL','L','ML'], sitio)`, id)
	if err != nil {
		return models.ExamenPerio{}, err
	}
	defer sitios.Close()
	for sitios.Next() {
		var pieza int
		var s models.Sitio
		if err := sitios.Scan(&pieza, &s.Sitio, &s.Profundidad, &s.Margen, &s.Sangrado); err != nil {
			return models.ExamenPerio{}, err
		}
		s.NIC = models.NIC(s.Profundidad, s.Margen)
		i := indice[pieza]
		e.Dientes[i].Sitios = append(e.Dientes[i].Sitios, s)
	}
	return e, sitios.Err()
}

// CrearPeriodontograma guarda el examen completo (dientes y puntos) en una sola transacción.
func (pg *PostgresRepository) CrearPeriodontograma(ctx context.Context, pacienteID int, fecha, notas string, dientes []models.DientePerio, usuarioID int) (int, error) {
	var id int
	err := pgx.BeginFunc(ctx, pg.db, func(tx pgx.Tx) error {
		if err := exigirPaciente(ctx, tx, pacienteID); err != nil {
			return err
		}
		if err := tx.QueryRow(ctx,
			`INSERT INTO periodontogramas (paciente_id, fecha, notas, creado_por) VALUES ($1, $2::date, $3, $4) RETURNING id`,
			pacienteID, fecha, notas, usuarioID).Scan(&id); err != nil {
			return err
		}
		var piezas, movilidades, furcas []int
		var sitioPieza, sitioProf, sitioMargen []int
		var sitioCodigo []string
		var sitioSangrado []bool
		for _, d := range dientes {
			piezas, movilidades, furcas = append(piezas, d.Pieza), append(movilidades, d.Movilidad), append(furcas, d.Furca)
			for _, s := range d.Sitios {
				sitioPieza, sitioCodigo = append(sitioPieza, d.Pieza), append(sitioCodigo, s.Sitio)
				sitioProf, sitioMargen, sitioSangrado = append(sitioProf, s.Profundidad), append(sitioMargen, s.Margen), append(sitioSangrado, s.Sangrado)
			}
		}
		if _, err := tx.Exec(ctx,
			`INSERT INTO periodonto_dientes (examen_id, pieza, movilidad, furca) SELECT $1, * FROM unnest($2::int[], $3::int[], $4::int[])`,
			id, piezas, movilidades, furcas); err != nil {
			return err
		}
		_, err := tx.Exec(ctx,
			`INSERT INTO periodonto_sitios (examen_id, pieza, sitio, profundidad_mm, margen_mm, sangrado)
			 SELECT $1, * FROM unnest($2::int[], $3::text[], $4::int[], $5::int[], $6::bool[])`,
			id, sitioPieza, sitioCodigo, sitioProf, sitioMargen, sitioSangrado)
		return err
	})
	if err != nil {
		return 0, traducirError(err)
	}
	return id, nil
}
