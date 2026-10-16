// Este archivo agrupa la administración de cuentas (Fase 7): listado paginado
// y cambio de rol / estado activo, siempre bajo RequireAdmin.
package repository

import (
	"context"
	"strings"

	"multicatalogo-backend/models"
)

// ObtenerUsuarios devuelve una página de cuentas con búsqueda server-side por nombre o correo.
func (pg *PostgresRepository) ObtenerUsuarios(ctx context.Context, busqueda string, page, limit int) (models.Pagina, error) {
	pagina, limite, offset := normalizarPagina(itoa(page), itoa(limit))

	q := strings.TrimSpace(busqueda)
	patron := "%" + strings.ReplaceAll(strings.ReplaceAll(q, "%", "\\%"), "_", "\\_") + "%"
	const filtro = `WHERE (nombre ILIKE $1 OR email ILIKE $1)`

	var total int
	if err := pg.db.QueryRow(ctx, `SELECT COUNT(*) FROM usuarios `+filtro, patron).Scan(&total); err != nil {
		return models.Pagina{}, err
	}

	rows, err := pg.db.Query(ctx,
		`SELECT id, email, nombre, rol, activo, referido_id, creado_en
		 FROM usuarios `+filtro+` ORDER BY id LIMIT $2 OFFSET $3`,
		patron, limite, offset,
	)
	if err != nil {
		return models.Pagina{}, err
	}
	defer rows.Close()

	usuarios := []models.UsuarioAdmin{}
	for rows.Next() {
		var u models.UsuarioAdmin
		if err := rows.Scan(&u.ID, &u.Email, &u.Nombre, &u.Rol, &u.Activo, &u.ReferidoID, &u.CreadoEn); err != nil {
			return models.Pagina{}, err
		}
		usuarios = append(usuarios, u)
	}
	if err := rows.Err(); err != nil {
		return models.Pagina{}, err
	}
	return models.Pagina{Data: usuarios, Page: pagina, Limit: limite, Total: total}, nil
}

// ActualizarUsuario cambia el rol y/o el estado activo de una cuenta (los nil no se tocan).
// Devuelve la cuenta resultante y false si el ID no existe.
func (pg *PostgresRepository) ActualizarUsuario(ctx context.Context, id int, rol *string, activo *bool) (models.UsuarioAdmin, bool, error) {
	var u models.UsuarioAdmin
	rows, err := pg.db.Query(ctx,
		`UPDATE usuarios
		 SET rol = COALESCE($1, rol), activo = COALESCE($2, activo)
		 WHERE id = $3
		 RETURNING id, email, nombre, rol, activo, referido_id, creado_en`,
		rol, activo, id,
	)
	if err != nil {
		return u, false, err
	}
	defer rows.Close()

	if !rows.Next() {
		return u, false, rows.Err()
	}
	if err := rows.Scan(&u.ID, &u.Email, &u.Nombre, &u.Rol, &u.Activo, &u.ReferidoID, &u.CreadoEn); err != nil {
		return u, false, err
	}
	return u, true, nil
}

// CuentaActiva indica si la cuenta existe y está habilitada. Una cuenta inexistente cuenta como inactiva.
func (pg *PostgresRepository) CuentaActiva(ctx context.Context, usuarioID int) (bool, error) {
	var activo bool
	rows, err := pg.db.Query(ctx, `SELECT activo FROM usuarios WHERE id = $1`, usuarioID)
	if err != nil {
		return false, err
	}
	defer rows.Close()
	if !rows.Next() {
		return false, rows.Err()
	}
	if err := rows.Scan(&activo); err != nil {
		return false, err
	}
	return activo, nil
}
