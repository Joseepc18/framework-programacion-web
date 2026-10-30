// Paquete config: configuración externa de la aplicación (base de datos y entorno).
package config

import (
	// context limita el tiempo de espera de la conexión inicial.
	"context"
	// fmt construye la cadena de conexión y envuelve los errores.
	"fmt"
	"net"
	"net/url"
	// os lee las variables de entorno.
	"os"
	"time"

	// pgxpool es el pool de conexiones del driver pgx para PostgreSQL.
	"github.com/jackc/pgx/v5/pgxpool"
	// godotenv carga las variables definidas en el archivo .env.
	"github.com/joho/godotenv"
)

// ConectarDB carga el archivo .env, crea el pool de conexiones a PostgreSQL y verifica que la base responda.
// Quien la llama es responsable de cerrar el pool con pool.Close() al finalizar la aplicación.
func ConectarDB() (*pgxpool.Pool, error) {
	// Si no existe el .env, se usan las variables de entorno del sistema.
	_ = godotenv.Load()

	dsn := cadenaConexionDB()

	// Límite de 5 segundos para no bloquear el arranque indefinidamente.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("no se pudo crear el pool de conexiones: %w", err)
	}

	// El ping confirma que la base de datos está disponible.
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("no se pudo conectar a PostgreSQL: %w", err)
	}

	return pool, nil
}

// DATABASE_URL permite usar la conexión entregada por el proveedor, incluido su modo SSL.
func cadenaConexionDB() string {
	if dsn := os.Getenv("DATABASE_URL"); dsn != "" {
		return dsn
	}
	modoSSL := os.Getenv("DB_SSLMODE")
	if modoSSL == "" {
		modoSSL = "disable"
	}
	conexion := url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD")),
		Host:   net.JoinHostPort(os.Getenv("DB_HOST"), os.Getenv("DB_PORT")),
		Path:   "/" + os.Getenv("DB_NAME"),
	}
	parametros := url.Values{"sslmode": {modoSSL}}
	conexion.RawQuery = parametros.Encode()
	return conexion.String()
}
