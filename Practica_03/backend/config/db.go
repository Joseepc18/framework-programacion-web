// Declaramos el paquete config, responsable de la configuración externa de la aplicación (base de datos, entorno).
package config

import (
	// Importamos context para limitar el tiempo de espera de la conexión inicial.
	"context"
	// Importamos fmt para construir la cadena de conexión y envolver errores.
	"fmt"
	// Importamos os para leer las variables de entorno.
	"os"
	// Importamos time para definir el tiempo máximo de espera.
	"time"

	// Importamos pgxpool, el pool de conexiones del driver pgx para PostgreSQL.
	"github.com/jackc/pgx/v5/pgxpool"
	// Importamos godotenv para cargar las variables definidas en el archivo .env.
	"github.com/joho/godotenv"
)

// ConectarDB carga el archivo .env, crea el pool de conexiones a PostgreSQL y verifica que la base responda.
// Quien la llama es responsable de cerrar el pool con pool.Close() al finalizar la aplicación.
func ConectarDB() (*pgxpool.Pool, error) {
	// Cargamos el .env; si no existe, se usan las variables de entorno del sistema.
	_ = godotenv.Load()

	// Construimos la cadena de conexión con los datos del entorno.
	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"),
	)

	// Creamos un contexto con un límite de 5 segundos para no bloquear el arranque indefinidamente.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Creamos el pool: reutiliza conexiones abiertas en lugar de abrir una nueva por cada petición.
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("no se pudo crear el pool de conexiones: %w", err)
	}

	// Hacemos un ping para confirmar que la base de datos está disponible.
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("no se pudo conectar a PostgreSQL: %w", err)
	}

	return pool, nil
}
