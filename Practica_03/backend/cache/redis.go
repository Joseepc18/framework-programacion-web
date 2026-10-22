// Declaramos el paquete cache, responsable de guardar y recuperar datos en el
// almacenamiento en caché. Las demás capas no conocen a Redis directamente:
// solo usan la interfaz que este paquete implementa.
package cache

import (
	// Importamos context para limitar el tiempo de espera de la conexión inicial y de cada operación.
	"context"
	// Importamos fmt para envolver los errores con un mensaje entendible.
	"fmt"
	// Importamos log para reportar fallos de escritura o borrado sin interrumpir la petición.
	"log"
	// Importamos os para leer las variables de entorno.
	"os"
	// Importamos time para definir los tiempos de espera y el TTL de las claves.
	"time"

	// Importamos godotenv para cargar las variables definidas en el archivo .env.
	"github.com/joho/godotenv"
	// Importamos el cliente oficial de Redis para Go.
	"github.com/redis/go-redis/v9"
)

// Redis es la implementación concreta de la caché sobre el cliente oficial de Redis.
type Redis struct {
	// cliente es la conexión con el servidor de Redis.
	cliente *redis.Client
}

// NuevoRedis carga el .env, crea el cliente desde REDIS_URL o REDIS_ADDR y verifica con un ping
// que el servidor de caché esté disponible. Quien la llama debe cerrar con Cerrar().
func NuevoRedis() (*Redis, error) {
	// Cargamos el .env; si no existe, se usan las variables de entorno del sistema.
	_ = godotenv.Load()

	opciones, err := opcionesRedis()
	if err != nil {
		return nil, err
	}

	// Creamos el cliente de conexión a Redis.
	cliente := redis.NewClient(opciones)

	// Limitamos la espera del ping inicial para no bloquear el arranque indefinidamente.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := cliente.Ping(ctx).Err(); err != nil {
		// Si Redis no responde, liberamos el cliente y reportamos el fallo.
		_ = cliente.Close()
		return nil, fmt.Errorf("no se pudo conectar a Redis: %w", err)
	}

	return &Redis{cliente: cliente}, nil
}

func opcionesRedis() (*redis.Options, error) {
	if direccion := os.Getenv("REDIS_URL"); direccion != "" {
		opciones, err := redis.ParseURL(direccion)
		if err != nil {
			// No incluimos la URL en el error porque puede contener credenciales.
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

// Obtener devuelve el valor guardado en la clave; el booleano indica si existía.
func (r *Redis) Obtener(ctx context.Context, clave string) ([]byte, bool) {
	datos, err := r.cliente.Get(ctx, clave).Bytes()
	if err != nil {
		// Una clave ausente o un Redis caído se tratan igual: no hay caché y se consulta la base.
		return nil, false
	}
	return datos, true
}

// Guardar escribe la clave con un tiempo de vida (TTL).
func (r *Redis) Guardar(ctx context.Context, clave string, datos []byte, ttl time.Duration) {
	// Si la escritura falla, la petición continúa con normalidad: la caché es opcional.
	if err := r.cliente.Set(ctx, clave, datos, ttl).Err(); err != nil {
		log.Printf("no se pudo guardar la clave %s en la caché: %v", clave, err)
	}
}

// Eliminar borra las claves indicadas; se usa para invalidar la caché tras cambiar los datos.
func (r *Redis) Eliminar(ctx context.Context, claves ...string) {
	if len(claves) == 0 {
		return
	}
	if err := r.cliente.Del(ctx, claves...).Err(); err != nil {
		log.Printf("no se pudieron invalidar las claves %v: %v", claves, err)
	}
}

// Cerrar libera la conexión con Redis al apagar la aplicación.
func (r *Redis) Cerrar() error {
	return r.cliente.Close()
}

// Ping comprueba que Redis responde (lo usa el health check).
func (r *Redis) Ping(ctx context.Context) error {
	return r.cliente.Ping(ctx).Err()
}
