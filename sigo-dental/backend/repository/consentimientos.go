// Este archivo agrupa el acceso a datos de los consentimientos informados, sus plantillas y los códigos OTP (F2-5).
package repository

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"sigo-dental-backend/models"
	"sigo-dental-backend/otp"
)

// --- Códigos OTP (otp.Repositorio) ---

// GuardarCodigo registra el código nuevo e inutiliza los anteriores del mismo propósito y referencia.
func (pg *PostgresRepository) GuardarCodigo(ctx context.Context, proposito string, referenciaID int, hash string, expira time.Time) error {
	return pgx.BeginFunc(ctx, pg.db, func(tx pgx.Tx) error {
		if _, err := tx.Exec(ctx, `UPDATE codigos_otp SET usado_en = NOW() WHERE proposito = $1 AND referencia_id = $2 AND usado_en IS NULL`, proposito, referenciaID); err != nil {
			return err
		}
		_, err := tx.Exec(ctx, `INSERT INTO codigos_otp (proposito, referencia_id, codigo_hash, expira_en) VALUES ($1, $2, $3, $4)`, proposito, referenciaID, hash, expira)
		return err
	})
}

// UltimoCodigo devuelve el código más reciente del propósito y la referencia.
func (pg *PostgresRepository) UltimoCodigo(ctx context.Context, proposito string, referenciaID int) (otp.Codigo, bool, error) {
	var c otp.Codigo
	var usado *time.Time
	err := pg.db.QueryRow(ctx,
		`SELECT id, codigo_hash, expira_en, intentos, usado_en FROM codigos_otp
		  WHERE proposito = $1 AND referencia_id = $2 ORDER BY id DESC LIMIT 1`, proposito, referenciaID,
	).Scan(&c.ID, &c.Hash, &c.ExpiraEn, &c.Intentos, &usado)
	if errors.Is(err, pgx.ErrNoRows) {
		return c, false, nil
	}
	c.Usado = usado != nil
	return c, err == nil, err
}

// RegistrarIntentoFallido suma un intento de forma atómica y devuelve el total (nunca más de 3).
func (pg *PostgresRepository) RegistrarIntentoFallido(ctx context.Context, id int) (int, error) {
	var intentos int
	err := pg.db.QueryRow(ctx, `UPDATE codigos_otp SET intentos = intentos + 1 WHERE id = $1 AND intentos < $2 AND usado_en IS NULL RETURNING intentos`, id, otp.MaxIntentos).Scan(&intentos)
	if errors.Is(err, pgx.ErrNoRows) {
		return otp.MaxIntentos, nil
	}
	return intentos, err
}

// ConsumirCodigo marca el código como usado; false si otra petición lo usó antes.
func (pg *PostgresRepository) ConsumirCodigo(ctx context.Context, id int) (bool, error) {
	tag, err := pg.db.Exec(ctx, `UPDATE codigos_otp SET usado_en = NOW() WHERE id = $1 AND usado_en IS NULL`, id)
	return tag.RowsAffected() == 1, err
}

// --- Bitácora de códigos enviados ---

// RegistrarNotificacionCodigo guarda el mensaje de un código como pendiente (el resultado del envío se anota después).
func (pg *PostgresRepository) RegistrarNotificacionCodigo(ctx context.Context, n models.NotificacionCodigo) (int, error) {
	var id int
	err := pg.db.QueryRow(ctx,
		`INSERT INTO notificaciones (paciente_id, tipo, canal, destino, mensaje) VALUES ($1, $2, $3, $4, $5) RETURNING id`,
		n.PacienteID, n.Tipo, n.Canal, n.Destino, n.Mensaje).Scan(&id)
	return id, traducirError(err)
}

// ListarNotificacionesPaciente devuelve los avisos del paciente que no pertenecen a una cita (códigos), del más reciente al más antiguo.
func (pg *PostgresRepository) ListarNotificacionesPaciente(ctx context.Context, pacienteID int) ([]models.Notificacion, error) {
	rows, err := pg.db.Query(ctx,
		`SELECT id, 0, tipo, canal, destino, mensaje, estado, detalle, vigente, creado_en, enviado_en
		   FROM notificaciones WHERE paciente_id = $1 AND cita_id IS NULL ORDER BY id DESC LIMIT 50`, pacienteID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	lista := []models.Notificacion{}
	for rows.Next() {
		var n models.Notificacion
		if err := rows.Scan(&n.ID, &n.CitaID, &n.Tipo, &n.Canal, &n.Destino, &n.Mensaje, &n.Estado, &n.Detalle, &n.Vigente, &n.CreadoEn, &n.EnviadoEn); err != nil {
			return nil, err
		}
		lista = append(lista, n)
	}
	return lista, rows.Err()
}

// --- Plantillas ---

const columnasPlantilla = `pl.id, pl.procedimiento_id, COALESCE(p.nombre, 'General (todos los procedimientos)'), pl.titulo, pl.cuerpo, pl.activo`

func escanearPlantilla(row pgx.Row, pl *models.Plantilla) error {
	return row.Scan(&pl.ID, &pl.ProcedimientoID, &pl.ProcedimientoNombre, &pl.Titulo, &pl.Cuerpo, &pl.Activo)
}

// ListarPlantillas devuelve las plantillas de consentimiento (la general primero).
func (pg *PostgresRepository) ListarPlantillas(ctx context.Context) ([]models.Plantilla, error) {
	rows, err := pg.db.Query(ctx,
		`SELECT `+columnasPlantilla+` FROM consentimiento_plantillas pl LEFT JOIN procedimientos p ON p.id = pl.procedimiento_id
		  ORDER BY pl.procedimiento_id NULLS FIRST, pl.id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	lista := []models.Plantilla{}
	for rows.Next() {
		var pl models.Plantilla
		if err := escanearPlantilla(rows, &pl); err != nil {
			return nil, err
		}
		lista = append(lista, pl)
	}
	return lista, rows.Err()
}

// PlantillaDe devuelve la plantilla activa del procedimiento o, si no tiene, la general.
func (pg *PostgresRepository) PlantillaDe(ctx context.Context, procedimientoID int) (models.Plantilla, error) {
	var pl models.Plantilla
	err := escanearPlantilla(pg.db.QueryRow(ctx,
		`SELECT `+columnasPlantilla+` FROM consentimiento_plantillas pl LEFT JOIN procedimientos p ON p.id = pl.procedimiento_id
		  WHERE pl.activo AND (pl.procedimiento_id = $1 OR pl.procedimiento_id IS NULL)
		  ORDER BY pl.procedimiento_id NULLS LAST LIMIT 1`, procedimientoID), &pl)
	if errors.Is(err, pgx.ErrNoRows) {
		return pl, models.ErrNoEncontrado
	}
	return pl, err
}

// GuardarPlantilla crea (id = 0) o actualiza una plantilla. Dos plantillas activas para el mismo procedimiento dan ErrDuplicado.
func (pg *PostgresRepository) GuardarPlantilla(ctx context.Context, id int, pl models.Plantilla) (int, error) {
	if id == 0 {
		err := pg.db.QueryRow(ctx,
			`INSERT INTO consentimiento_plantillas (procedimiento_id, titulo, cuerpo, activo) VALUES ($1, $2, $3, $4) RETURNING id`,
			pl.ProcedimientoID, pl.Titulo, pl.Cuerpo, pl.Activo).Scan(&id)
		return id, traducirError(err)
	}
	tag, err := pg.db.Exec(ctx,
		`UPDATE consentimiento_plantillas SET procedimiento_id = $2, titulo = $3, cuerpo = $4, activo = $5 WHERE id = $1`,
		id, pl.ProcedimientoID, pl.Titulo, pl.Cuerpo, pl.Activo)
	if err != nil {
		return 0, traducirError(err)
	}
	if tag.RowsAffected() == 0 {
		return 0, models.ErrNoEncontrado
	}
	return id, nil
}

// --- Consentimientos ---

// vigenciaDias es la subconsulta del parámetro de vigencia (180 días por defecto).
const vigenciaDias = `COALESCE((SELECT NULLIF(valor, '')::int FROM parametros WHERE clave = 'consentimiento_vigencia_dias'), 180)`

// hayRevocacion indica que existe una revocación firmada del consentimiento c.
const hayRevocacion = `EXISTS (SELECT 1 FROM consentimientos r WHERE r.revoca_a = c.id AND r.estado = 'firmado')`

const columnasConsentimiento = `c.id, c.paciente_id, c.procedimiento_id, pr.nombre, c.profesional_id, up.nombre, c.plantilla_id, c.tipo, c.revoca_a,
	c.motivo, c.titulo, c.texto, c.huella_texto, c.estado, c.firmante_tipo, c.firmante_nombre, c.firmante_documento,
	COALESCE(c.modalidad, ''), c.trazo, c.dispositivo, c.firmado_en, c.archivo_id, uc.nombre, c.creado_en,
	(c.tipo = 'consentimiento' AND c.estado = 'firmado' AND c.firmado_en > NOW() - make_interval(days => ` + vigenciaDias + `) AND NOT ` + hayRevocacion + `),
	` + hayRevocacion

const origenConsentimiento = `FROM consentimientos c
	JOIN procedimientos pr ON pr.id = c.procedimiento_id
	JOIN profesionales pf ON pf.id = c.profesional_id JOIN usuarios up ON up.id = pf.usuario_id
	JOIN usuarios uc ON uc.id = c.creado_por`

func escanearConsentimiento(row pgx.Row, c *models.Consentimiento) error {
	return row.Scan(&c.ID, &c.PacienteID, &c.ProcedimientoID, &c.ProcedimientoNombre, &c.ProfesionalID, &c.ProfesionalNombre, &c.PlantillaID,
		&c.Tipo, &c.RevocaA, &c.Motivo, &c.Titulo, &c.Texto, &c.HuellaTexto, &c.Estado, &c.FirmanteTipo, &c.FirmanteNombre,
		&c.FirmanteDocumento, &c.Modalidad, &c.Trazo, &c.Dispositivo, &c.FirmadoEn, &c.ArchivoID, &c.CreadoPor, &c.CreadoEn,
		&c.Vigente, &c.Revocado)
}

// ListarConsentimientos devuelve los consentimientos y revocaciones del paciente, del más reciente al más antiguo.
func (pg *PostgresRepository) ListarConsentimientos(ctx context.Context, pacienteID int) ([]models.Consentimiento, error) {
	rows, err := pg.db.Query(ctx, `SELECT `+columnasConsentimiento+` `+origenConsentimiento+` WHERE c.paciente_id = $1 ORDER BY c.id DESC`, pacienteID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	lista := []models.Consentimiento{}
	for rows.Next() {
		var c models.Consentimiento
		if err := escanearConsentimiento(rows, &c); err != nil {
			return nil, err
		}
		lista = append(lista, c)
	}
	return lista, rows.Err()
}

// ObtenerConsentimiento devuelve un consentimiento del paciente; ErrNoEncontrado si no existe.
func (pg *PostgresRepository) ObtenerConsentimiento(ctx context.Context, pacienteID, id int) (models.Consentimiento, error) {
	var c models.Consentimiento
	err := escanearConsentimiento(pg.db.QueryRow(ctx, `SELECT `+columnasConsentimiento+` `+origenConsentimiento+` WHERE c.id = $1 AND c.paciente_id = $2`, id, pacienteID), &c)
	if errors.Is(err, pgx.ErrNoRows) {
		return c, models.ErrNoEncontrado
	}
	return c, err
}

// CrearConsentimiento guarda el consentimiento pendiente con el texto ya completado y su huella.
func (pg *PostgresRepository) CrearConsentimiento(ctx context.Context, n models.NuevoConsentimiento, usuarioID int) (int, error) {
	var id int
	err := pgx.BeginFunc(ctx, pg.db, func(tx pgx.Tx) error {
		if err := exigirPaciente(ctx, tx, n.PacienteID); err != nil {
			return err
		}
		if n.RevocaA != nil {
			// Solo se revoca un consentimiento firmado del mismo paciente y procedimiento que no esté ya revocado.
			var ok bool
			if err := tx.QueryRow(ctx,
				`SELECT EXISTS (SELECT 1 FROM consentimientos c WHERE c.id = $1 AND c.paciente_id = $2 AND c.tipo = 'consentimiento'
				        AND c.estado = 'firmado' AND NOT `+hayRevocacion+`)`, *n.RevocaA, n.PacienteID).Scan(&ok); err != nil {
				return err
			}
			if !ok {
				return models.ErrConflicto
			}
		}
		return tx.QueryRow(ctx,
			`INSERT INTO consentimientos (paciente_id, procedimiento_id, profesional_id, plantilla_id, tipo, revoca_a, motivo, titulo, texto,
			        huella_texto, firmante_tipo, firmante_nombre, firmante_documento, creado_por)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14) RETURNING id`,
			n.PacienteID, n.ProcedimientoID, n.ProfesionalID, n.PlantillaID, n.Tipo, n.RevocaA, n.Motivo, n.Titulo, n.Texto,
			models.Huella(n.Texto), n.FirmanteTipo, n.FirmanteNombre, n.FirmanteDocumento, usuarioID).Scan(&id)
	})
	if err != nil {
		return 0, traducirError(err)
	}
	return id, nil
}

// FirmarConsentimiento pasa el consentimiento de pendiente a firmado, una sola vez. El PDF (archivo_id) ya debe existir.
func (pg *PostgresRepository) FirmarConsentimiento(ctx context.Context, pacienteID, id int, f models.DatosFirmaConsentimiento) error {
	tag, err := pg.db.Exec(ctx,
		`UPDATE consentimientos SET estado = 'firmado', modalidad = $3, trazo = $4::jsonb, dispositivo = $5, firmado_en = $7, archivo_id = $6
		  WHERE id = $2 AND paciente_id = $1 AND estado = 'pendiente'`,
		pacienteID, id, f.Modalidad, f.Trazo, f.Dispositivo, f.ArchivoID, f.FirmadoEn)
	if err != nil {
		return traducirError(err)
	}
	if tag.RowsAffected() == 0 {
		if _, err := pg.ObtenerConsentimiento(ctx, pacienteID, id); err != nil {
			return err
		}
		return models.ErrConsentimientoFirmado
	}
	return nil
}

// ProcedimientoDeCatalogo devuelve el nombre y si exige consentimiento; ErrNoEncontrado si no existe o está desactivado.
func (pg *PostgresRepository) ProcedimientoDeCatalogo(ctx context.Context, id int) (nombre string, requiere bool, err error) {
	err = pg.db.QueryRow(ctx, `SELECT nombre, requiere_consentimiento FROM procedimientos WHERE id = $1 AND activo`, id).Scan(&nombre, &requiere)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", false, models.ErrNoEncontrado
	}
	return nombre, requiere, err
}
