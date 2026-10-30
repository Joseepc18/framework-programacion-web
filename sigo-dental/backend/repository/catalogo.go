// Este archivo agrupa el acceso a datos del catálogo médico y de los parámetros del sistema.
package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"sigo-dental-backend/models"
)

// Códigos de error de PostgreSQL que se traducen a errores de dominio.
const (
	pgUnicidad   = "23505"
	pgForanea    = "23503"
	pgCheck      = "23514"
	pgExclusion  = "23P01"
	pgNoNulo     = "23502"
	pgTextoLargo = "22001"
	// pgInmutable lo lanzan los disparadores que protegen los registros clínicos y de auditoría (migración 007).
	pgInmutable = "SG001"
)

// traducirError convierte las violaciones de restricciones de PostgreSQL en errores de dominio,
// para que los controladores respondan 409 o 400 sin conocer el driver. Otros errores pasan igual.
func traducirError(err error) error {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return err
	}
	switch pgErr.Code {
	case pgUnicidad:
		return models.ErrDuplicado
	case pgInmutable:
		return models.ErrInmutable
	case pgExclusion:
		return models.ErrConflicto
	case pgForanea, pgCheck, pgNoNulo, pgTextoLargo:
		return models.ErrReferencia
	}
	return err
}

// tablasSimples es la lista cerrada de catálogos de solo nombre: el recurso de la URL nunca se
// concatena a la consulta, solo se usa para elegir una tabla de aquí.
var tablasSimples = map[string]string{
	"consultorios":         "consultorios",
	"convenios":            "convenios",
	"condiciones-criticas": "condiciones_criticas",
}

// ListarSimples devuelve los elementos del catálogo simple, ordenados por nombre.
func (pg *PostgresRepository) ListarSimples(ctx context.Context, recurso string, soloActivos bool) ([]models.ItemCatalogo, error) {
	tabla, ok := tablasSimples[recurso]
	if !ok {
		return nil, models.ErrNoEncontrado
	}
	rows, err := pg.db.Query(ctx,
		`SELECT id, nombre, activo FROM `+tabla+` WHERE ($1 OR activo) ORDER BY nombre`, !soloActivos)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []models.ItemCatalogo{}
	for rows.Next() {
		var it models.ItemCatalogo
		if err := rows.Scan(&it.ID, &it.Nombre, &it.Activo); err != nil {
			return nil, err
		}
		items = append(items, it)
	}
	return items, rows.Err()
}

// CrearSimple agrega un elemento activo al catálogo simple.
func (pg *PostgresRepository) CrearSimple(ctx context.Context, recurso, nombre string) (models.ItemCatalogo, error) {
	tabla, ok := tablasSimples[recurso]
	if !ok {
		return models.ItemCatalogo{}, models.ErrNoEncontrado
	}
	var it models.ItemCatalogo
	err := pg.db.QueryRow(ctx,
		`INSERT INTO `+tabla+` (nombre) VALUES ($1) RETURNING id, nombre, activo`, nombre,
	).Scan(&it.ID, &it.Nombre, &it.Activo)
	if err != nil {
		return models.ItemCatalogo{}, traducirError(err)
	}
	return it, nil
}

// ActualizarSimple cambia el nombre y el estado de un elemento; ErrNoEncontrado si no existe.
func (pg *PostgresRepository) ActualizarSimple(ctx context.Context, recurso string, id int, nombre string, activo bool) (models.ItemCatalogo, error) {
	tabla, ok := tablasSimples[recurso]
	if !ok {
		return models.ItemCatalogo{}, models.ErrNoEncontrado
	}
	var it models.ItemCatalogo
	err := pg.db.QueryRow(ctx,
		`UPDATE `+tabla+` SET nombre = $2, activo = $3 WHERE id = $1 RETURNING id, nombre, activo`,
		id, nombre, activo,
	).Scan(&it.ID, &it.Nombre, &it.Activo)
	if errors.Is(err, pgx.ErrNoRows) {
		return models.ItemCatalogo{}, models.ErrNoEncontrado
	}
	if err != nil {
		return models.ItemCatalogo{}, traducirError(err)
	}
	return it, nil
}

const columnasSillon = `s.id, s.consultorio_id, c.nombre, s.nombre, s.activo`

// ListarSillones devuelve los sillones, opcionalmente de un solo consultorio (consultorioID 0 = todos).
func (pg *PostgresRepository) ListarSillones(ctx context.Context, consultorioID int, soloActivos bool) ([]models.Sillon, error) {
	rows, err := pg.db.Query(ctx,
		`SELECT `+columnasSillon+`
		   FROM sillones s JOIN consultorios c ON c.id = s.consultorio_id
		  WHERE ($1 = 0 OR s.consultorio_id = $1) AND ($2 OR s.activo)
		  ORDER BY c.nombre, s.nombre`, consultorioID, !soloActivos)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	sillones := []models.Sillon{}
	for rows.Next() {
		var s models.Sillon
		if err := rows.Scan(&s.ID, &s.ConsultorioID, &s.ConsultorioNombre, &s.Nombre, &s.Activo); err != nil {
			return nil, err
		}
		sillones = append(sillones, s)
	}
	return sillones, rows.Err()
}

// CrearSillon agrega un sillón activo a un consultorio existente.
func (pg *PostgresRepository) CrearSillon(ctx context.Context, consultorioID int, nombre string) (models.Sillon, error) {
	var id int
	err := pg.db.QueryRow(ctx,
		`INSERT INTO sillones (consultorio_id, nombre) VALUES ($1, $2) RETURNING id`, consultorioID, nombre,
	).Scan(&id)
	if err != nil {
		return models.Sillon{}, traducirError(err)
	}
	return pg.obtenerSillon(ctx, id)
}

// ActualizarSillon cambia el nombre y el estado de un sillón; el consultorio no se cambia.
func (pg *PostgresRepository) ActualizarSillon(ctx context.Context, id int, nombre string, activo bool) (models.Sillon, error) {
	tag, err := pg.db.Exec(ctx, `UPDATE sillones SET nombre = $2, activo = $3 WHERE id = $1`, id, nombre, activo)
	if err != nil {
		return models.Sillon{}, traducirError(err)
	}
	if tag.RowsAffected() == 0 {
		return models.Sillon{}, models.ErrNoEncontrado
	}
	return pg.obtenerSillon(ctx, id)
}

func (pg *PostgresRepository) obtenerSillon(ctx context.Context, id int) (models.Sillon, error) {
	var s models.Sillon
	err := pg.db.QueryRow(ctx,
		`SELECT `+columnasSillon+` FROM sillones s JOIN consultorios c ON c.id = s.consultorio_id WHERE s.id = $1`, id,
	).Scan(&s.ID, &s.ConsultorioID, &s.ConsultorioNombre, &s.Nombre, &s.Activo)
	if errors.Is(err, pgx.ErrNoRows) {
		return models.Sillon{}, models.ErrNoEncontrado
	}
	return s, err
}

const columnasProfesional = `p.id, p.usuario_id, u.nombre, u.email, p.especialidad, p.registro_profesional, p.activo`

// ListarProfesionales devuelve los profesionales con el nombre y correo de su cuenta.
func (pg *PostgresRepository) ListarProfesionales(ctx context.Context, soloActivos bool) ([]models.Profesional, error) {
	rows, err := pg.db.Query(ctx,
		`SELECT `+columnasProfesional+`
		   FROM profesionales p JOIN usuarios u ON u.id = p.usuario_id
		  WHERE ($1 OR p.activo) ORDER BY u.nombre`, !soloActivos)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	lista := []models.Profesional{}
	for rows.Next() {
		var p models.Profesional
		if err := rows.Scan(&p.ID, &p.UsuarioID, &p.Nombre, &p.Email, &p.Especialidad, &p.RegistroProfesional, &p.Activo); err != nil {
			return nil, err
		}
		lista = append(lista, p)
	}
	return lista, rows.Err()
}

// CrearProfesional vincula una cuenta con rol odontólogo como profesional.
// Si la cuenta no existe o no es odontólogo, devuelve ErrReferencia.
func (pg *PostgresRepository) CrearProfesional(ctx context.Context, usuarioID int, especialidad, registro string) (models.Profesional, error) {
	var id int
	err := pg.db.QueryRow(ctx,
		`INSERT INTO profesionales (usuario_id, especialidad, registro_profesional)
		 SELECT u.id, $2, $3
		   FROM usuarios u JOIN roles r ON r.id = u.rol_id
		  WHERE u.id = $1 AND r.codigo = $4
		 RETURNING id`, usuarioID, especialidad, registro, models.RolOdontologo,
	).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return models.Profesional{}, models.ErrReferencia
	}
	if err != nil {
		return models.Profesional{}, traducirError(err)
	}
	return pg.obtenerProfesional(ctx, id)
}

// ActualizarProfesional cambia especialidad, registro y estado de un profesional.
func (pg *PostgresRepository) ActualizarProfesional(ctx context.Context, id int, especialidad, registro string, activo bool) (models.Profesional, error) {
	tag, err := pg.db.Exec(ctx,
		`UPDATE profesionales SET especialidad = $2, registro_profesional = $3, activo = $4 WHERE id = $1`,
		id, especialidad, registro, activo)
	if err != nil {
		return models.Profesional{}, traducirError(err)
	}
	if tag.RowsAffected() == 0 {
		return models.Profesional{}, models.ErrNoEncontrado
	}
	return pg.obtenerProfesional(ctx, id)
}

func (pg *PostgresRepository) obtenerProfesional(ctx context.Context, id int) (models.Profesional, error) {
	var p models.Profesional
	err := pg.db.QueryRow(ctx,
		`SELECT `+columnasProfesional+` FROM profesionales p JOIN usuarios u ON u.id = p.usuario_id WHERE p.id = $1`, id,
	).Scan(&p.ID, &p.UsuarioID, &p.Nombre, &p.Email, &p.Especialidad, &p.RegistroProfesional, &p.Activo)
	if errors.Is(err, pgx.ErrNoRows) {
		return models.Profesional{}, models.ErrNoEncontrado
	}
	return p, err
}

// ListarUsuariosOdontologos devuelve las cuentas activas con rol odontólogo que aún no son profesionales.
func (pg *PostgresRepository) ListarUsuariosOdontologos(ctx context.Context) ([]models.UsuarioResumen, error) {
	rows, err := pg.db.Query(ctx,
		`SELECT u.id, u.nombre, u.email
		   FROM usuarios u JOIN roles r ON r.id = u.rol_id
		  WHERE r.codigo = $1 AND u.activo
		    AND NOT EXISTS (SELECT 1 FROM profesionales p WHERE p.usuario_id = u.id)
		  ORDER BY u.nombre`, models.RolOdontologo)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	lista := []models.UsuarioResumen{}
	for rows.Next() {
		var u models.UsuarioResumen
		if err := rows.Scan(&u.ID, &u.Nombre, &u.Email); err != nil {
			return nil, err
		}
		lista = append(lista, u)
	}
	return lista, rows.Err()
}

const columnasProcedimiento = `id, codigo, nombre, especialidad, precio_centavos, duracion_min, requiere_consentimiento, activo`

func escanearProcedimiento(row pgx.Row) (models.Procedimiento, error) {
	var p models.Procedimiento
	err := row.Scan(&p.ID, &p.Codigo, &p.Nombre, &p.Especialidad, &p.PrecioCentavos, &p.DuracionMin, &p.RequiereConsentimiento, &p.Activo)
	return p, err
}

// ListarProcedimientos devuelve el catálogo de procedimientos ordenado por código.
func (pg *PostgresRepository) ListarProcedimientos(ctx context.Context, soloActivos bool) ([]models.Procedimiento, error) {
	rows, err := pg.db.Query(ctx,
		`SELECT `+columnasProcedimiento+` FROM procedimientos WHERE ($1 OR activo) ORDER BY codigo`, !soloActivos)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	lista := []models.Procedimiento{}
	for rows.Next() {
		p, err := escanearProcedimiento(rows)
		if err != nil {
			return nil, err
		}
		lista = append(lista, p)
	}
	return lista, rows.Err()
}

// CrearProcedimiento agrega un procedimiento activo; el código no puede repetirse.
func (pg *PostgresRepository) CrearProcedimiento(ctx context.Context, p models.Procedimiento) (models.Procedimiento, error) {
	creado, err := escanearProcedimiento(pg.db.QueryRow(ctx,
		`INSERT INTO procedimientos (codigo, nombre, especialidad, precio_centavos, duracion_min, requiere_consentimiento)
		 VALUES ($1, $2, $3, $4, $5, $6) RETURNING `+columnasProcedimiento,
		p.Codigo, p.Nombre, p.Especialidad, p.PrecioCentavos, p.DuracionMin, p.RequiereConsentimiento))
	if err != nil {
		return models.Procedimiento{}, traducirError(err)
	}
	return creado, nil
}

// ActualizarProcedimiento reemplaza los datos editables de un procedimiento.
func (pg *PostgresRepository) ActualizarProcedimiento(ctx context.Context, id int, p models.Procedimiento) (models.Procedimiento, error) {
	actualizado, err := escanearProcedimiento(pg.db.QueryRow(ctx,
		`UPDATE procedimientos
		    SET codigo = $2, nombre = $3, especialidad = $4, precio_centavos = $5,
		        duracion_min = $6, requiere_consentimiento = $7, activo = $8
		  WHERE id = $1 RETURNING `+columnasProcedimiento,
		id, p.Codigo, p.Nombre, p.Especialidad, p.PrecioCentavos, p.DuracionMin, p.RequiereConsentimiento, p.Activo))
	if errors.Is(err, pgx.ErrNoRows) {
		return models.Procedimiento{}, models.ErrNoEncontrado
	}
	if err != nil {
		return models.Procedimiento{}, traducirError(err)
	}
	return actualizado, nil
}

// ListarParametros devuelve todos los parámetros del sistema.
func (pg *PostgresRepository) ListarParametros(ctx context.Context) ([]models.Parametro, error) {
	rows, err := pg.db.Query(ctx, `SELECT clave, valor, descripcion FROM parametros ORDER BY clave`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	lista := []models.Parametro{}
	for rows.Next() {
		var p models.Parametro
		if err := rows.Scan(&p.Clave, &p.Valor, &p.Descripcion); err != nil {
			return nil, err
		}
		lista = append(lista, p)
	}
	return lista, rows.Err()
}

// ObtenerParametro devuelve el valor de un parámetro; ErrNoEncontrado si no existe.
func (pg *PostgresRepository) ObtenerParametro(ctx context.Context, clave string) (string, error) {
	var valor string
	err := pg.db.QueryRow(ctx, `SELECT valor FROM parametros WHERE clave = $1`, clave).Scan(&valor)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", models.ErrNoEncontrado
	}
	return valor, err
}

// ActualizarParametro cambia el valor de un parámetro existente.
func (pg *PostgresRepository) ActualizarParametro(ctx context.Context, clave, valor string) (models.Parametro, error) {
	var p models.Parametro
	err := pg.db.QueryRow(ctx,
		`UPDATE parametros SET valor = $2 WHERE clave = $1 RETURNING clave, valor, descripcion`, clave, valor,
	).Scan(&p.Clave, &p.Valor, &p.Descripcion)
	if errors.Is(err, pgx.ErrNoRows) {
		return models.Parametro{}, models.ErrNoEncontrado
	}
	return p, err
}
