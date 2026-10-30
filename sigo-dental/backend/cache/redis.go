// Paquete cache: acceso a Redis. Las demás capas no lo conocen directamente:
// solo usan la interfaz que este paquete implementa.
package cache

import (
	// context limita el tiempo de espera de la conexión inicial y de cada operación.
	"context"
	// fmt envuelve los errores con un mensaje entendible.
	"fmt"
	// os lee las variables de entorno.
	"os"
	"time"

	// godotenv carga las variables definidas en el archivo .env.
	"github.com/joho/godotenv"
	// Cliente oficial de Redis para Go.
	"github.com/redis/go-redis/v9"
)

// Redis es la implementación concreta de la caché sobre el cliente oficial de Redis.
type Redis struct {
	cliente *redis.Client
}

// NuevoRedis carga el .env, crea el cliente desde REDIS_URL o REDIS_ADDR y verifica con un ping
// que el servidor de caché esté disponible. Quien la llama debe cerrar con Cerrar().
func NuevoRedis() (*Redis, error) {
	// Si no existe el .env, se usan las variables de entorno del sistema.
	_ = godotenv.Load()

	opciones, err := opcionesRedis()
	if err != nil {
		return nil, err
	}

	cliente := redis.NewClient(opciones)

	// Límite de 5 segundos para no bloquear el arranque indefinidamente.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := cliente.Ping(ctx).Err(); err != nil {
		_ = cliente.Close()
		return nil, fmt.Errorf("no se pudo conectar a Redis: %w", err)
	}

	return &Redis{cliente: cliente}, nil
}

func opcionesRedis() (*redis.Options, error) {
	if direccion := os.Getenv("REDIS_URL"); direccion != "" {
		opciones, err := redis.ParseURL(direccion)
		if err != nil {
			// No se incluye la URL en el error porque puede contener credenciales.
			return nil, fmt.Errorf("REDIS_URL no válida; use redis:// o rediss://")
		}
		return opciones, nil
	}
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		addr = "localhost:6379"
	}
	return &redis.Options{Addr: addr}, nil
}

// Ping comprueba que Redis responde (lo usa el health check).
func (r *Redis) Ping(ctx context.Context) error {
	return r.cliente.Ping(ctx).Err()
}

// Cerrar libera la conexión con Redis al apagar la aplicación.
func (r *Redis) Cerrar() error {
	return r.cliente.Close()
}
