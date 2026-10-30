// Paquete repository: único responsable del acceso a los datos, que viven en PostgreSQL.
package repository

import (
	"context"
	"errors"

	// pgx aporta ErrNoRows; pgxpool es el pool de conexiones que recibe el repositorio.
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"sigo-dental-backend/models"
)

// baseDatos es lo que el repositorio usa de la conexión: el pool en producción y una transacción en las pruebas de
// integración (que se revierte al terminar para no dejar datos).
type baseDatos interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Begin(ctx context.Context) (pgx.Tx, error)
}

// PostgresRepository implementa el acceso a datos sobre un pool de conexiones de PostgreSQL.
// Se construye una sola vez en main y se inyecta en los controladores por interfaz.
type PostgresRepository struct {
	db   baseDatos
	pool *pgxpool.Pool
}

// Nuevo crea el repositorio con el pool de conexiones recibido.
func Nuevo(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{db: pool, pool: pool}
}

// Ping comprueba que PostgreSQL responde (lo usa el health check).
func (pg *PostgresRepository) Ping(ctx context.Context) error {
	return pg.pool.Ping(ctx)
}

// BuscarUsuarioPorEmail busca la cuenta por correo junto con el código de su rol.
// El booleano es false cuando no existe ninguna cuenta con ese correo.
func (pg *PostgresRepository) BuscarUsuarioPorEmail(ctx context.Context, email string) (models.Usuario, bool, error) {
	var u models.Usuario
	err := pg.db.QueryRow(ctx,
		`SELECT u.id, u.email, u.password_hash, r.codigo, u.activo
		   FROM usuarios u
		   JOIN roles r ON r.id = u.rol_id
		  WHERE u.email = $1`, email,
	).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Rol, &u.Activo)
	if errors.Is(err, pgx.ErrNoRows) {
		return models.Usuario{}, false, nil
	}
	if err != nil {
		return models.Usuario{}, false, err
	}
	return u, true, nil
}
