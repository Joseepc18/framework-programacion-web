// Este archivo agrupa el acceso a datos de la evolución clínica: notas SOEP, adendas, firmas y procedimientos
// realizados (F2-4: RF-05 y RN-01).
package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"sigo-dental-backend/models"
)

// CredencialesDeUsuario devuelve el hash de la contraseña de la cuenta y si está activa (reautenticación de la firma).
func (pg *PostgresRepository) CredencialesDeUsuario(ctx context.Context, usuarioID int) (string, bool, bool, error) {
	var hash string
	var activo bool
	err := pg.db.QueryRow(ctx, `SELECT password_hash, activo FROM usuarios WHERE id = $1`, usuarioID).Scan(&hash, &activo)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", false, false, nil
	}
	return hash, activo, err == nil, err
}

// ProfesionalDeUsuario devuelve el profesional activo de la cuenta; ok es false si la cuenta no lo es.
func (pg *PostgresRepository) ProfesionalDeUsuario(ctx context.Context, usuarioID int) (models.ProfesionalDeUsuario, bool, error) {
	var p models.ProfesionalDeUsuario
	err := pg.db.QueryRow(ctx,
		`SELECT pf.id, u.nombre, pf.registro_profesional FROM profesionales pf JOIN usuarios u ON u.id = pf.usuario_id
		  WHERE pf.usuario_id = $1 AND pf.activo AND u.activo`, usuarioID).Scan(&p.ID, &p.Nombre, &p.RegistroProfesional)
	if errors.Is(err, pgx.ErrNoRows) {
		return p, false, nil
	}
	return p, err == nil, err
}

// ProfesionalActivo indica si el profesional existe y está activo (el asistente elige al responsable de su nota).
func (pg *PostgresRepository) ProfesionalActivo(ctx context.Context, profesionalID int) (bool, error) {
	var ok bool
	err := pg.db.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM profesionales WHERE id = $1 AND activo)`, profesionalID).Scan(&ok)
	return ok, err
}

const columnasNota = `n.id, n.paciente_id, n.cita_id, n.profesional_id, up.nombre, n.autor_id, ua.nombre, n.subjetivo, n.objetivo,
	n.evaluacion, n.plan, n.creada_en, n.actualizada_en`

const origenNota = `FROM notas_clinicas n
	JOIN profesionales pf ON pf.id = n.profesional_id JOIN usuarios up ON up.id = pf.usuario_id
	JOIN usuarios ua ON ua.id = n.autor_id`

func escanearNota(row pgx.Row, n *models.Nota) error {
	return row.Scan(&n.ID, &n.PacienteID, &n.CitaID, &n.ProfesionalID, &n.ProfesionalNombre, &n.AutorID, &n.AutorNombre,
		&n.Subjetivo, &n.Objetivo, &n.Evaluacion, &n.Plan, &n.CreadaEn, &n.ActualizadaEn)
}

const columnasFirma = `f.id, f.entidad, f.entidad_id, f.usuario_id, f.profesional_id, f.nombre_firmante, f.registro_profesional,
	f.huella_sha256, f.contenido_canonico, f.firmado_en`

func escanearFirma(row pgx.Row, f *models.Firma) error {
	return row.Scan(&f.ID, &f.Entidad, &f.EntidadID, &f.UsuarioID, &f.ProfesionalID, &f.NombreFirmante, &f.RegistroProfesional,
		&f.HuellaSHA256, &f.ContenidoCanonico, &f.FirmadoEn)
}

// firmaDe devuelve la firma de la entidad, o nil si no está firmada.
func (pg *PostgresRepository) firmaDe(ctx context.Context, entidad string, id int) (*models.Firma, error) {
	var f models.Firma
	err := escanearFirma(pg.db.QueryRow(ctx, `SELECT `+columnasFirma+` FROM firmas f WHERE f.entidad = $1 AND f.entidad_id = $2`, entidad, id), &f)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return &f, err
}

// adendasDe devuelve las adendas del objetivo (nota o procedimiento) con su firma, de la más antigua a la más reciente.
func (pg *PostgresRepository) adendasDe(ctx context.Context, campo string, id int) ([]models.Adenda, error) {
	// campo es siempre una constante interna (nota_id o procedimiento_realizado_id), nunca texto de la petición.
	rows, err := pg.db.Query(ctx,
		`SELECT a.id, a.paciente_id, a.nota_id, a.procedimiento_realizado_id, a.profesional_id, u.nombre, a.autor_id, a.texto, a.creada_en
		   FROM adendas a JOIN profesionales pf ON pf.id = a.profesional_id JOIN usuarios u ON u.id = pf.usuario_id
		  WHERE a.`+campo+` = $1 ORDER BY a.id`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	adendas := []models.Adenda{}
	for rows.Next() {
		var a models.Adenda
		if err := rows.Scan(&a.ID, &a.PacienteID, &a.NotaID, &a.ProcedimientoRealizadoID, &a.ProfesionalID, &a.ProfesionalNombre,
			&a.AutorID, &a.Texto, &a.CreadaEn); err != nil {
			return nil, err
		}
		adendas = append(adendas, a)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	for i := range adendas {
		if adendas[i].Firma, err = pg.firmaDe(ctx, models.FirmaAdenda, adendas[i].ID); err != nil {
			return nil, err
		}
	}
	return adendas, nil
}

func (pg *PostgresRepository) completarNota(ctx context.Context, n models.Nota) (models.NotaCompleta, error) {
	completa := models.NotaCompleta{Nota: n}
	var err error
	if completa.Firma, err = pg.firmaDe(ctx, models.FirmaNota, n.ID); err != nil {
		return completa, err
	}
	completa.Adendas, err = pg.adendasDe(ctx, "nota_id", n.ID)
	return completa, err
}

// ListarNotas devuelve la bitácora del paciente, de la nota más reciente a la más antigua, con firma y adendas.
func (pg *PostgresRepository) ListarNotas(ctx context.Context, pacienteID int) ([]models.NotaCompleta, error) {
	rows, err := pg.db.Query(ctx, `SELECT `+columnasNota+` `+origenNota+` WHERE n.paciente_id = $1 ORDER BY n.creada_en DESC, n.id DESC`, pacienteID)
	if err != nil {
		return nil, err
	}
	notas := []models.Nota{}
	for rows.Next() {
		var n models.Nota
		if err := escanearNota(rows, &n); err != nil {
			rows.Close()
			return nil, err
		}
		notas = append(notas, n)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, err
	}
	lista := make([]models.NotaCompleta, 0, len(notas))
	for _, n := range notas {
		completa, err := pg.completarNota(ctx, n)
		if err != nil {
			return nil, err
		}
		lista = append(lista, completa)
	}
	return lista, nil
}

// ObtenerNota devuelve la nota del paciente con su firma y adendas; ErrNoEncontrado si no existe o es de otro paciente.
func (pg *PostgresRepository) ObtenerNota(ctx context.Context, pacienteID, notaID int) (models.NotaCompleta, error) {
	var n models.Nota
	err := escanearNota(pg.db.QueryRow(ctx, `SELECT `+columnasNota+` `+origenNota+` WHERE n.id = $1 AND n.paciente_id = $2`, notaID, pacienteID), &n)
	if errors.Is(err, pgx.ErrNoRows) {
		return models.NotaCompleta{}, models.ErrNoEncontrado
	}
	if err != nil {
		return models.NotaCompleta{}, err
	}
	return pg.completarNota(ctx, n)
}

// CrearNota guarda una nota nueva y devuelve su ID. La cita, si se indica, debe ser del mismo paciente.
func (pg *PostgresRepository) CrearNota(ctx context.Context, n models.Nota) (int, error) {
	var id int
	err := pgx.BeginFunc(ctx, pg.db, func(tx pgx.Tx) error {
		if err := exigirPaciente(ctx, tx, n.PacienteID); err != nil {
			return err
		}
		if n.CitaID != nil {
			var coincide bool
			if err := tx.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM citas WHERE id = $1 AND paciente_id = $2)`, *n.CitaID, n.PacienteID).Scan(&coincide); err != nil {
				return err
			}
			if !coincide {
				return models.ErrReferencia
			}
		}
		return tx.QueryRow(ctx,
			`INSERT INTO notas_clinicas (paciente_id, cita_id, profesional_id, autor_id, subjetivo, objetivo, evaluacion, plan)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id`,
			n.PacienteID, n.CitaID, n.ProfesionalID, n.AutorID, n.Subjetivo, n.Objetivo, n.Evaluacion, n.Plan).Scan(&id)
	})
	if err != nil {
		return 0, traducirError(err)
	}
	return id, nil
}

// ActualizarNota reemplaza los cuatro campos SOEP. Si la nota ya está firmada, el disparador de la base lo rechaza
// (models.ErrInmutable) aunque otra petición la hubiera firmado un instante antes.
func (pg *PostgresRepository) ActualizarNota(ctx context.Context, pacienteID, notaID int, n models.Nota) error {
	tag, err := pg.db.Exec(ctx,
		`UPDATE notas_clinicas SET subjetivo = $3, objetivo = $4, evaluacion = $5, plan = $6, actualizada_en = NOW()
		  WHERE id = $2 AND paciente_id = $1`, pacienteID, notaID, n.Subjetivo, n.Objetivo, n.Evaluacion, n.Plan)
	if err != nil {
		return traducirError(err)
	}
	if tag.RowsAffected() == 0 {
		return models.ErrNoEncontrado
	}
	return nil
}

// GuardarFirma registra la firma de un documento; ErrDuplicado si ya estaba firmado. La fecha la pone la base.
func (pg *PostgresRepository) GuardarFirma(ctx context.Context, f models.Firma) (models.Firma, error) {
	err := pg.db.QueryRow(ctx,
		`INSERT INTO firmas (entidad, entidad_id, usuario_id, profesional_id, nombre_firmante, registro_profesional, huella_sha256, contenido_canonico)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id, firmado_en`,
		f.Entidad, f.EntidadID, f.UsuarioID, f.ProfesionalID, f.NombreFirmante, f.RegistroProfesional, f.HuellaSHA256, f.ContenidoCanonico,
	).Scan(&f.ID, &f.FirmadoEn)
	if err != nil {
		return models.Firma{}, traducirError(err)
	}
	f.Valida = true
	return f, nil
}

// CrearAdendaFirmada guarda la adenda y su firma en una sola transacción: una adenda sin firma no existe.
// firmar recibe la adenda ya guardada para construir su firma.
func (pg *PostgresRepository) CrearAdendaFirmada(ctx context.Context, a models.Adenda, firmar func(models.Adenda) models.Firma) (models.Adenda, error) {
	err := pgx.BeginFunc(ctx, pg.db, func(tx pgx.Tx) error {
		var duenio int
		var err error
		if a.NotaID != nil {
			err = tx.QueryRow(ctx, `SELECT paciente_id FROM notas_clinicas WHERE id = $1`, *a.NotaID).Scan(&duenio)
		} else {
			err = tx.QueryRow(ctx, `SELECT paciente_id FROM procedimientos_realizados WHERE id = $1`, *a.ProcedimientoRealizadoID).Scan(&duenio)
		}
		if errors.Is(err, pgx.ErrNoRows) || (err == nil && duenio != a.PacienteID) {
			return models.ErrNoEncontrado
		}
		if err != nil {
			return err
		}
		if err := tx.QueryRow(ctx,
			`INSERT INTO adendas (paciente_id, nota_id, procedimiento_realizado_id, profesional_id, autor_id, texto)
			 VALUES ($1, $2, $3, $4, $5, $6) RETURNING id, creada_en`,
			a.PacienteID, a.NotaID, a.ProcedimientoRealizadoID, a.ProfesionalID, a.AutorID, a.Texto).Scan(&a.ID, &a.CreadaEn); err != nil {
			return err
		}
		f := firmar(a)
		f.EntidadID = a.ID
		firmada := f
		if err := tx.QueryRow(ctx,
			`INSERT INTO firmas (entidad, entidad_id, usuario_id, profesional_id, nombre_firmante, registro_profesional, huella_sha256, contenido_canonico)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id, firmado_en`,
			f.Entidad, f.EntidadID, f.UsuarioID, f.ProfesionalID, f.NombreFirmante, f.RegistroProfesional, f.HuellaSHA256, f.ContenidoCanonico,
		).Scan(&firmada.ID, &firmada.FirmadoEn); err != nil {
			return err
		}
		firmada.Valida = true
		a.Firma = &firmada
		return nil
	})
	if err != nil {
		return models.Adenda{}, traducirError(err)
	}
	return a, nil
}

const columnasProcRealizado = `r.id, r.paciente_id, r.procedimiento_id, p.nombre, r.pieza, COALESCE(r.superficie, ''), r.profesional_id, u.nombre,
	r.cita_id, r.estado, r.notas, r.iniciado_en, r.realizado_en`

const origenProcRealizado = `FROM procedimientos_realizados r JOIN procedimientos p ON p.id = r.procedimiento_id
	JOIN profesionales pf ON pf.id = r.profesional_id JOIN usuarios u ON u.id = pf.usuario_id`

func escanearProcRealizado(row pgx.Row, r *models.ProcedimientoRealizado) error {
	return row.Scan(&r.ID, &r.PacienteID, &r.ProcedimientoID, &r.ProcedimientoNombre, &r.Pieza, &r.Superficie, &r.ProfesionalID,
		&r.ProfesionalNombre, &r.CitaID, &r.Estado, &r.Notas, &r.IniciadoEn, &r.RealizadoEn)
}

// ListarProcedimientosRealizados devuelve los procedimientos del paciente con sus adendas, del más reciente al más antiguo.
func (pg *PostgresRepository) ListarProcedimientosRealizados(ctx context.Context, pacienteID int) ([]models.ProcedimientoRealizado, error) {
	rows, err := pg.db.Query(ctx, `SELECT `+columnasProcRealizado+` `+origenProcRealizado+` WHERE r.paciente_id = $1 ORDER BY r.iniciado_en DESC, r.id DESC`, pacienteID)
	if err != nil {
		return nil, err
	}
	lista := []models.ProcedimientoRealizado{}
	for rows.Next() {
		var r models.ProcedimientoRealizado
		if err := escanearProcRealizado(rows, &r); err != nil {
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
		if lista[i].Adendas, err = pg.adendasDe(ctx, "procedimiento_realizado_id", lista[i].ID); err != nil {
			return nil, err
		}
	}
	return lista, nil
}

// ObtenerProcedimientoRealizado devuelve el procedimiento del paciente; ErrNoEncontrado si no existe.
func (pg *PostgresRepository) ObtenerProcedimientoRealizado(ctx context.Context, pacienteID, id int) (models.ProcedimientoRealizado, error) {
	var r models.ProcedimientoRealizado
	err := escanearProcRealizado(pg.db.QueryRow(ctx, `SELECT `+columnasProcRealizado+` `+origenProcRealizado+` WHERE r.id = $1 AND r.paciente_id = $2`, id, pacienteID), &r)
	if errors.Is(err, pgx.ErrNoRows) {
		return r, models.ErrNoEncontrado
	}
	if err != nil {
		return r, err
	}
	r.Adendas, err = pg.adendasDe(ctx, "procedimiento_realizado_id", r.ID)
	return r, err
}

// IniciarProcedimiento registra el procedimiento como "en curso". El procedimiento del catálogo debe estar activo.
func (pg *PostgresRepository) IniciarProcedimiento(ctx context.Context, n models.NuevoProcedimientoRealizado, profesionalID, usuarioID int) (int, error) {
	var id int
	err := pgx.BeginFunc(ctx, pg.db, func(tx pgx.Tx) error {
		if err := exigirPaciente(ctx, tx, n.PacienteID); err != nil {
			return err
		}
		var activo, exigeConsentimiento bool
		err := tx.QueryRow(ctx, `SELECT activo, requiere_consentimiento FROM procedimientos WHERE id = $1`, n.ProcedimientoID).Scan(&activo, &exigeConsentimiento)
		if errors.Is(err, pgx.ErrNoRows) || (err == nil && !activo) {
			return models.ErrReferencia
		}
		if err != nil {
			return err
		}
		// RF-06: un procedimiento que exige consentimiento no se inicia sin uno firmado, vigente y no revocado.
		if exigeConsentimiento {
			var hayConsentimiento bool
			if err := tx.QueryRow(ctx,
				`SELECT EXISTS (SELECT 1 FROM consentimientos c WHERE c.paciente_id = $1 AND c.procedimiento_id = $2
				        AND c.tipo = 'consentimiento' AND c.estado = 'firmado'
				        AND c.firmado_en > NOW() - make_interval(days => `+vigenciaDias+`) AND NOT `+hayRevocacion+`)`,
				n.PacienteID, n.ProcedimientoID).Scan(&hayConsentimiento); err != nil {
				return err
			}
			if !hayConsentimiento {
				return models.ErrConsentimientoRequerido
			}
		}
		return tx.QueryRow(ctx,
			`INSERT INTO procedimientos_realizados (paciente_id, procedimiento_id, pieza, superficie, profesional_id, cita_id, notas, registrado_por)
			 VALUES ($1, $2, $3, NULLIF($4, ''), $5, $6, $7, $8) RETURNING id`,
			n.PacienteID, n.ProcedimientoID, n.Pieza, n.Superficie, profesionalID, n.CitaID, n.Notas, usuarioID).Scan(&id)
	})
	if err != nil {
		return 0, traducirError(err)
	}
	return id, nil
}

// CambiarEstadoProcedimiento pasa un procedimiento "en curso" a "realizado" o "cancelado". Desde entonces ya no
// cambia: lo impide el disparador de la base. ErrProcedimientoNoEnCurso si ya estaba cerrado.
func (pg *PostgresRepository) CambiarEstadoProcedimiento(ctx context.Context, pacienteID, id int, estado string) error {
	tag, err := pg.db.Exec(ctx,
		`UPDATE procedimientos_realizados SET estado = $3::varchar, realizado_en = CASE WHEN $3::varchar = 'realizado' THEN NOW() END
		  WHERE id = $2 AND paciente_id = $1 AND estado = 'en_curso'`, pacienteID, id, estado)
	if err != nil {
		return traducirError(err)
	}
	if tag.RowsAffected() == 0 {
		if _, err := pg.ObtenerProcedimientoRealizado(ctx, pacienteID, id); err != nil {
			return err
		}
		return models.ErrProcedimientoNoEnCurso
	}
	return nil
}
