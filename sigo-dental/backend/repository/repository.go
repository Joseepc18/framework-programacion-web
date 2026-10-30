// Paquete repository: único responsable del acceso a los datos, que viven en PostgreSQL.
package repository

import (
	"context"

	// pgxpool es el pool de conexiones que recibe el repositorio.
	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresRepository implementa el acceso a datos sobre un pool de conexiones de PostgreSQL.
// Se construye una sola vez en main y se inyecta en los controladores por interfaz.
type PostgresRepository struct {
	db *pgxpool.Pool
}

// Nuevo crea el repositorio con el pool de conexiones recibido.
func Nuevo(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{db: pool}
}

// Ping comprueba que PostgreSQL responde (lo usa el health check).
func (pg *PostgresRepository) Ping(ctx context.Context) error {
	return pg.db.Ping(ctx)
}
