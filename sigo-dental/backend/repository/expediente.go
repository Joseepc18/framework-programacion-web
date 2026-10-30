// Este archivo agrupa el acceso a datos del expediente: auditoría de accesos, signos vitales y archivos (F2-1).
package repository

import (
	"context"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5"
	"sigo-dental-backend/models"
)

// RegistrarAcceso guarda un evento en la bitácora de solo inserción.
func (pg *PostgresRepository) RegistrarAcceso(ctx context.Context, e models.EventoAuditoria) error {
	var usuario *int
	if e.UsuarioID > 0 {
		usuario = &e.UsuarioID
	}
	_, err := pg.db.Exec(ctx,
		`INSERT INTO auditoria_accesos (usuario_id, usuario_email, rol, accion, entidad, entidad_id, paciente_id, ip, metodo, ruta)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
		usuario, recortar(e.Email, 150), recortar(e.Rol, 30), e.Accion, recortar(e.Entidad, 40), e.EntidadID, e.PacienteID,
		recortar(e.IP, 64), recortar(e.Metodo, 10), recortar(e.Ruta, 200))
	return traducirError(err)
}

// recortar evita que un valor largo (ruta, cabecera) haga fallar el registro de auditoría.
func recortar(texto string, max int) string {
	if len(texto) <= max {
		return texto
	}
	return texto[:max]
}

// ListarAuditoria devuelve una página de la bitácora, del evento más reciente al más antiguo.
func (pg *PostgresRepository) ListarAuditoria(ctx context.Context, f models.FiltroAuditoria) (models.PaginaAuditoria, error) {
	const filtro = ` FROM auditoria_accesos a
		LEFT JOIN pacientes p ON p.id = a.paciente_id
		WHERE ($1 = '' OR a.usuario_email ILIKE $1)
		  AND ($2 = '' OR p.documento ILIKE $2 OR p.numero_historia ILIKE $2
		       OR (p.nombres || ' ' || p.apellidos) ILIKE $2 OR (p.apellidos || ' ' || p.nombres) ILIKE $2)
		  AND ($3 = '' OR a.accion = $3)
		  AND ($4::timestamptz IS NULL OR a.ocurrido_en >= $4)
		  AND ($5::timestamptz IS NULL OR a.ocurrido_en < $5)`
	patron := func(texto string) string {
		texto = strings.TrimSpace(texto)
		if texto == "" {
			return ""
		}
		return "%" + escaparLike.Replace(texto) + "%"
	}
	usuario, paciente := patron(f.Usuario), patron(f.Paciente)

	pagina := models.PaginaAuditoria{Pagina: f.Pagina, Limite: f.Limite, Registros: []models.RegistroAuditoria{}}
	if err := pg.db.QueryRow(ctx, `SELECT COUNT(*)`+filtro, usuario, paciente, f.Accion, f.Desde, f.Hasta).Scan(&pagina.Total); err != nil {
		return pagina, err
	}
	rows, err := pg.db.Query(ctx,
		`SELECT a.id, a.ocurrido_en, a.usuario_email, a.rol, a.accion, a.entidad, a.entidad_id, a.paciente_id,
		        COALESCE(p.nombres || ' ' || p.apellidos, ''), a.ip, a.metodo, a.ruta`+filtro+`
		  ORDER BY a.id DESC LIMIT $6 OFFSET $7`,
		usuario, paciente, f.Accion, f.Desde, f.Hasta, f.Limite, (f.Pagina-1)*f.Limite)
	if err != nil {
		return pagina, err
	}
	defer rows.Close()
	for rows.Next() {
		var r models.RegistroAuditoria
		if err := rows.Scan(&r.ID, &r.OcurridoEn, &r.UsuarioEmail, &r.Rol, &r.Accion, &r.Entidad, &r.EntidadID,
			&r.PacienteID, &r.PacienteNombre, &r.IP, &r.Metodo, &r.Ruta); err != nil {
			return pagina, err
		}
		pagina.Registros = append(pagina.Registros, r)
	}
	return pagina, rows.Err()
}

// ListarSignosVitales devuelve las tomas del paciente, de la más reciente a la más antigua.
func (pg *PostgresRepository) ListarSignosVitales(ctx context.Context, pacienteID int) ([]models.SignosVitales, error) {
	rows, err := pg.db.Query(ctx,
		`SELECT s.id, s.paciente_id, s.cita_id, s.sistolica, s.diastolica, s.frecuencia_cardiaca, s.frecuencia_respiratoria,
		        s.temperatura::float8, s.saturacion, s.peso_kg::float8, s.talla_cm::float8, s.observaciones, u.nombre, s.registrado_en
		   FROM signos_vitales s JOIN usuarios u ON u.id = s.registrado_por
		  WHERE s.paciente_id = $1 ORDER BY s.registrado_en DESC, s.id DESC LIMIT 100`, pacienteID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	lista := []models.SignosVitales{}
	for rows.Next() {
		var s models.SignosVitales
		if err := rows.Scan(&s.ID, &s.PacienteID, &s.CitaID, &s.Sistolica, &s.Diastolica, &s.FrecuenciaCardiaca,
			&s.FrecuenciaRespiratoria, &s.Temperatura, &s.Saturacion, &s.PesoKg, &s.TallaCm, &s.Observaciones,
			&s.RegistradoPor, &s.RegistradoEn); err != nil {
			return nil, err
		}
		lista = append(lista, s)
	}
	return lista, rows.Err()
}

// RegistrarSignosVitales guarda una toma nueva y la devuelve completa. La cita, si se indica, debe ser del mismo paciente.
func (pg *PostgresRepository) RegistrarSignosVitales(ctx context.Context, s models.SignosVitales, usuarioID int) (models.SignosVitales, error) {
	var nombre string
	err := pgx.BeginFunc(ctx, pg.db, func(tx pgx.Tx) error {
		if err := exigirPaciente(ctx, tx, s.PacienteID); err != nil {
			return err
		}
		if s.CitaID != nil {
			var coincide bool
			if err := tx.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM citas WHERE id = $1 AND paciente_id = $2)`,
				*s.CitaID, s.PacienteID).Scan(&coincide); err != nil {
				return err
			}
			if !coincide {
				return models.ErrReferencia
			}
		}
		if err := tx.QueryRow(ctx,
			`INSERT INTO signos_vitales (paciente_id, cita_id, sistolica, diastolica, frecuencia_cardiaca, frecuencia_respiratoria,
			        temperatura, saturacion, peso_kg, talla_cm, observaciones, registrado_por)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12) RETURNING id, registrado_en`,
			s.PacienteID, s.CitaID, s.Sistolica, s.Diastolica, s.FrecuenciaCardiaca, s.FrecuenciaRespiratoria,
			s.Temperatura, s.Saturacion, s.PesoKg, s.TallaCm, s.Observaciones, usuarioID).Scan(&s.ID, &s.RegistradoEn); err != nil {
			return err
		}
		return tx.QueryRow(ctx, `SELECT nombre FROM usuarios WHERE id = $1`, usuarioID).Scan(&nombre)
	})
	if err != nil {
		return models.SignosVitales{}, traducirError(err)
	}
	s.RegistradoPor = nombre
	return s, nil
}

// exigirPaciente devuelve ErrNoEncontrado si el paciente no existe.
func exigirPaciente(ctx context.Context, q interface {
	QueryRow(context.Context, string, ...any) pgx.Row
}, id int) error {
	var existe bool
	if err := q.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM pacientes WHERE id = $1)`, id).Scan(&existe); err != nil {
		return err
	}
	if !existe {
		return models.ErrNoEncontrado
	}
	return nil
}

const columnasArchivo = `f.id, f.paciente_id, f.categoria, f.descripcion, f.nombre_original, f.tipo_mime, f.tamano_bytes,
	f.sha256, f.clave_almacen, u.nombre, f.subido_en`

func escanearArchivo(row pgx.Row, a *models.Archivo) error {
	return row.Scan(&a.ID, &a.PacienteID, &a.Categoria, &a.Descripcion, &a.NombreOriginal, &a.TipoMIME, &a.TamanoBytes,
		&a.SHA256, &a.ClaveAlmacen, &a.SubidoPor, &a.SubidoEn)
}

// RegistrarArchivo guarda los metadatos de un archivo ya almacenado y devuelve su ID.
func (pg *PostgresRepository) RegistrarArchivo(ctx context.Context, a models.NuevoArchivo) (int, error) {
	var id int
	err := pg.db.QueryRow(ctx,
		`INSERT INTO archivos_paciente (paciente_id, categoria, descripcion, nombre_original, tipo_mime, tamano_bytes, sha256,
		        clave_almacen, subido_por)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING id`,
		a.PacienteID, a.Categoria, a.Descripcion, a.NombreOriginal, a.TipoMIME, a.TamanoBytes, a.SHA256, a.ClaveAlmacen,
		a.UsuarioID).Scan(&id)
	if err != nil {
		return 0, traducirError(err)
	}
	return id, nil
}

// ListarArchivos devuelve los archivos del paciente, del más reciente al más antiguo.
func (pg *PostgresRepository) ListarArchivos(ctx context.Context, pacienteID int) ([]models.Archivo, error) {
	rows, err := pg.db.Query(ctx,
		`SELECT `+columnasArchivo+` FROM archivos_paciente f JOIN usuarios u ON u.id = f.subido_por
		  WHERE f.paciente_id = $1 ORDER BY f.subido_en DESC, f.id DESC`, pacienteID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	lista := []models.Archivo{}
	for rows.Next() {
		var a models.Archivo
		if err := escanearArchivo(rows, &a); err != nil {
			return nil, err
		}
		lista = append(lista, a)
	}
	return lista, rows.Err()
}

// ObtenerArchivo devuelve los metadatos de un archivo; ErrNoEncontrado si no existe.
func (pg *PostgresRepository) ObtenerArchivo(ctx context.Context, id int) (models.Archivo, error) {
	var a models.Archivo
	err := escanearArchivo(pg.db.QueryRow(ctx,
		`SELECT `+columnasArchivo+` FROM archivos_paciente f JOIN usuarios u ON u.id = f.subido_por WHERE f.id = $1`, id), &a)
	if errors.Is(err, pgx.ErrNoRows) {
		return models.Archivo{}, models.ErrNoEncontrado
	}
	return a, err
}
