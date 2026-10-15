// Seguimos dentro del paquete config; aquí centralizamos la conexión al sistema de caché (Redis).
package config

import (
	// Importamos context para limitar el tiempo de espera de la conexión inicial.
	"context"
	// Importamos fmt para envolver los errores con un mensaje entendible.
	"fmt"
	// Importamos os para leer las variables de entorno.
	"os"
	// Importamos time para definir el tiempo máximo de espera.
	"time"

	// Importamos el cliente oficial de Redis para Go.
	"github.com/redis/go-redis/v9"
	// Importamos godotenv para cargar las variables definidas en el archivo .env.
	"github.com/joho/godotenv"
)

// Cache es el cliente de Redis compartido por las capas que necesitan caché.
// Se inicializa una sola vez al arrancar la aplicación con ConectarRedis.
var Cache *redis.Client

// ConectarRedis crea el cliente de Redis usando la variable de entorno REDIS_ADDR
// y verifica con un ping que el servidor de caché esté disponible.
func ConectarRedis() error {
	// Cargamos el .env; si no existe, se usan las variables de entorno del sistema.
	_ = godotenv.Load()

	// Leemos el host:puerto de Redis; por defecto el contenedor local expuesto en 6379.
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		addr = "localhost:6379"
	}

	// Creamos el cliente de conexión a Redis.
	Cache = redis.NewClient(&redis.Options{Addr: addr})

	// Limitamos la espera del ping inicial para no bloquear el arranque indefinidamente.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := Cache.Ping(ctx).Err(); err != nil {
		// Si Redis no responde, liberamos el cliente y reportamos el fallo.
		_ = Cache.Close()
		Cache = nil
		return fmt.Errorf("no se pudo conectar a Redis: %w", err)
	}

	return nil
}
