// Este archivo agrupa el acceso a datos de pacientes y anamnesis (F1-2).
package repository

import (
	"context"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5"
	"sigo-dental-backend/models"
)

// limiteListado acota el número de filas que devuelve la búsqueda de pacientes.
const limiteListado = 200

const columnasPaciente = `p.id, p.numero_historia, p.documento, p.nombres, p.apellidos,
	to_char(p.fecha_nacimiento, 'YYYY-MM-DD'), date_part('year', age(p.fecha_nacimiento))::int,
	p.sexo, p.telefono, p.email, p.direccion,
	p.emergencia_nombre, p.emergencia_telefono, p.emergencia_parentesco,
	p.tutor_nombre, p.tutor_documento, p.tutor_telefono,
	p.convenio_id, COALESCE(cv.nombre, ''), p.canal_preferido, p.activo`

// alertasDe devuelve la subconsulta con los nombres de las condiciones críticas del paciente.
const alertasDe = `ARRAY(SELECT cc.nombre FROM paciente_condiciones_criticas pc
	JOIN condiciones_criticas cc ON cc.id = pc.condicion_id
	WHERE pc.paciente_id = p.id ORDER BY cc.nombre)`

// escaparLike evita que los comodines de LIKE escritos por la persona usuaria alteren la búsqueda.
var escaparLike = strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)

// ListarPacientes busca por nombre, apellido, documento o número de historia (vacío = todos).
func (pg *PostgresRepository) ListarPacientes(ctx context.Context, busqueda string, soloActivos bool) ([]models.PacienteResumen, error) {
	patron := "%" + escaparLike.Replace(strings.TrimSpace(busqueda)) + "%"
	rows, err := pg.db.Query(ctx,
		`SELECT p.id, p.numero_historia, p.documento, p.nombres, p.apellidos,
		        date_part('year', age(p.fecha_nacimiento))::int,
		        p.telefono, COALESCE(cv.nombre, ''), p.activo, `+alertasDe+`
		   FROM pacientes p LEFT JOIN convenios cv ON cv.id = p.convenio_id
		  WHERE ($2 OR p.activo)
		    AND (p.documento ILIKE $1 OR p.numero_historia ILIKE $1
		         OR (p.nombres || ' ' || p.apellidos) ILIKE $1 OR (p.apellidos || ' ' || p.nombres) ILIKE $1)
		  ORDER BY lower(p.apellidos), lower(p.nombres)
		  LIMIT $3`, patron, !soloActivos, limiteListado)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	lista := []models.PacienteResumen{}
	for rows.Next() {
		var r models.PacienteResumen
		if err := rows.Scan(&r.ID, &r.NumeroHistoria, &r.Documento, &r.Nombres, &r.Apellidos, &r.Edad,
			&r.Telefono, &r.ConvenioNombre, &r.Activo, &r.Alertas); err != nil {
			return nil, err
		}
		lista = append(lista, r)
	}
	return lista, rows.Err()
}

// ObtenerPaciente devuelve la ficha completa con la anamnesis y las alertas; ErrNoEncontrado si no existe.
func (pg *PostgresRepository) ObtenerPaciente(ctx context.Context, id int) (models.PacienteDetalle, error) {
	var d models.PacienteDetalle
	p := &d.Paciente
	err := pg.db.QueryRow(ctx,
		`SELECT `+columnasPaciente+`, `+alertasDe+`
		   FROM pacientes p LEFT JOIN convenios cv ON cv.id = p.convenio_id
		  WHERE p.id = $1`, id,
	).Scan(&p.ID, &p.NumeroHistoria, &p.Documento, &p.Nombres, &p.Apellidos, &p.FechaNacimiento, &p.Edad,
		&p.Sexo, &p.Telefono, &p.Email, &p.Direccion,
		&p.EmergenciaNombre, &p.EmergenciaTelefono, &p.EmergenciaParentesco,
		&p.TutorNombre, &p.TutorDocumento, &p.TutorTelefono,
		&p.ConvenioID, &p.ConvenioNombre, &p.CanalPreferido, &p.Activo, &d.Alertas)
	if errors.Is(err, pgx.ErrNoRows) {
		return models.PacienteDetalle{}, models.ErrNoEncontrado
	}
	if err != nil {
		return models.PacienteDetalle{}, err
	}

	a := &d.Anamnesis
	err = pg.db.QueryRow(ctx,
		`SELECT enfermedades_sistemicas, alergias_farmacos, antecedentes_quirurgicos, coagulopatias
		   FROM anamnesis WHERE paciente_id = $1`, id,
	).Scan(&a.EnfermedadesSistemicas, &a.AlergiasFarmacos, &a.AntecedentesQuirurgicos, &a.Coagulopatias)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return models.PacienteDetalle{}, err
	}

	rows, err := pg.db.Query(ctx,
		`SELECT condicion_id FROM paciente_condiciones_criticas WHERE paciente_id = $1 ORDER BY condicion_id`, id)
	if err != nil {
		return models.PacienteDetalle{}, err
	}
	defer rows.Close()
	a.CondicionesCriticas = []int{}
	for rows.Next() {
		var condicion int
		if err := rows.Scan(&condicion); err != nil {
			return models.PacienteDetalle{}, err
		}
		a.CondicionesCriticas = append(a.CondicionesCriticas, condicion)
	}
	return d, rows.Err()
}

// CrearPaciente da de alta la ficha, la anamnesis y las condiciones críticas en una sola transacción,
// de modo que nunca queda un paciente sin su historial médico. Devuelve el ID creado.
func (pg *PostgresRepository) CrearPaciente(ctx context.Context, p models.Paciente, a models.Anamnesis, usuarioID int) (int, error) {
	var id int
	err := pgx.BeginFunc(ctx, pg.db, func(tx pgx.Tx) error {
		err := tx.QueryRow(ctx,
			`INSERT INTO pacientes (documento, nombres, apellidos, fecha_nacimiento, sexo, telefono, email, direccion,
			        emergencia_nombre, emergencia_telefono, emergencia_parentesco,
			        tutor_nombre, tutor_documento, tutor_telefono, convenio_id, canal_preferido)
			 VALUES ($1, $2, $3, $4::date, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
			 RETURNING id`,
			p.Documento, p.Nombres, p.Apellidos, p.FechaNacimiento, p.Sexo, p.Telefono, p.Email, p.Direccion,
			p.EmergenciaNombre, p.EmergenciaTelefono, p.EmergenciaParentesco,
			p.TutorNombre, p.TutorDocumento, p.TutorTelefono, p.ConvenioID, p.CanalPreferido,
		).Scan(&id)
		if err != nil {
			return err
		}
		return guardarAnamnesis(ctx, tx, id, a, usuarioID)
	})
	if err != nil {
		return 0, traducirError(err)
	}
	return id, nil
}

// ActualizarPaciente reemplaza los datos de la ficha (no toca la anamnesis).
func (pg *PostgresRepository) ActualizarPaciente(ctx context.Context, id int, p models.Paciente) error {
	tag, err := pg.db.Exec(ctx,
		`UPDATE pacientes
		    SET documento = $2, nombres = $3, apellidos = $4, fecha_nacimiento = $5::date, sexo = $6,
		        telefono = $7, email = $8, direccion = $9,
		        emergencia_nombre = $10, emergencia_telefono = $11, emergencia_parentesco = $12,
		        tutor_nombre = $13, tutor_documento = $14, tutor_telefono = $15,
		        convenio_id = $16, canal_preferido = $17, activo = $18, actualizado_en = NOW()
		  WHERE id = $1`,
		id, p.Documento, p.Nombres, p.Apellidos, p.FechaNacimiento, p.Sexo, p.Telefono, p.Email, p.Direccion,
		p.EmergenciaNombre, p.EmergenciaTelefono, p.EmergenciaParentesco,
		p.TutorNombre, p.TutorDocumento, p.TutorTelefono, p.ConvenioID, p.CanalPreferido, p.Activo)
	if err != nil {
		return traducirError(err)
	}
	if tag.RowsAffected() == 0 {
		return models.ErrNoEncontrado
	}
	return nil
}

// ActualizarAnamnesis reemplaza el historial médico y las condiciones críticas del paciente.
func (pg *PostgresRepository) ActualizarAnamnesis(ctx context.Context, id int, a models.Anamnesis, usuarioID int) error {
	err := pgx.BeginFunc(ctx, pg.db, func(tx pgx.Tx) error {
		var existe bool
		if err := tx.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM pacientes WHERE id = $1)`, id).Scan(&existe); err != nil {
			return err
		}
		if !existe {
			return models.ErrNoEncontrado
		}
		return guardarAnamnesis(ctx, tx, id, a, usuarioID)
	})
	return traducirError(err)
}

// guardarAnamnesis inserta o reemplaza el historial y deja exactamente las condiciones críticas indicadas.
func guardarAnamnesis(ctx context.Context, tx pgx.Tx, pacienteID int, a models.Anamnesis, usuarioID int) error {
	if _, err := tx.Exec(ctx,
		`INSERT INTO anamnesis (paciente_id, enfermedades_sistemicas, alergias_farmacos, antecedentes_quirurgicos,
		        coagulopatias, actualizado_por)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 ON CONFLICT (paciente_id) DO UPDATE
		    SET enfermedades_sistemicas = EXCLUDED.enfermedades_sistemicas,
		        alergias_farmacos = EXCLUDED.alergias_farmacos,
		        antecedentes_quirurgicos = EXCLUDED.antecedentes_quirurgicos,
		        coagulopatias = EXCLUDED.coagulopatias,
		        actualizado_por = EXCLUDED.actualizado_por, actualizado_en = NOW()`,
		pacienteID, a.EnfermedadesSistemicas, a.AlergiasFarmacos, a.AntecedentesQuirurgicos, a.Coagulopatias, usuarioID,
	); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `DELETE FROM paciente_condiciones_criticas WHERE paciente_id = $1`, pacienteID); err != nil {
		return err
	}
	_, err := tx.Exec(ctx,
		`INSERT INTO paciente_condiciones_criticas (paciente_id, condicion_id)
		 SELECT $1, unnest($2::int[])`, pacienteID, a.CondicionesCriticas)
	return err
}
