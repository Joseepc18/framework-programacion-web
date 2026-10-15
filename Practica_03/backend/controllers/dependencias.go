// Seguimos dentro del paquete controllers; este archivo declara las dependencias que los
// controladores reciben inyectadas al arrancar el servidor (Tema 4: inyección de dependencias).
package controllers

import (
	// Importamos context y time porque las operaciones de caché los necesitan en su firma.
	"context"
	"time"

	// Importamos nuestro paquete de modelos para las estructuras que cruzan la interfaz.
	"multicatalogo-backend/models"
)

// DBRepository es el contrato que debe cumplir la capa de datos para que un controlador
// pueda consultar PostgreSQL. Al depender de una interfaz y no de la implementación real,
// el controlador se puede probar con un doble y no sabe cómo se guardan los datos.
type DBRepository interface {
	// BuscarUsuarioPorCredenciales valida un login contra la tabla de usuarios.
	BuscarUsuarioPorCredenciales(email, password string) (models.Usuario, bool, error)
	// ObtenerProductos devuelve el catálogo completo.
	ObtenerProductos() ([]models.Producto, error)
	// ObtenerProductoPorID devuelve un producto concreto.
	ObtenerProductoPorID(id int) (models.Producto, bool, error)
	// ObtenerRed devuelve el árbol de la red multinivel.
	ObtenerRed() (models.Referido, error)
	// RegistrarPedido guarda un pedido completo dentro de una transacción ACID.
	RegistrarPedido(req models.CheckoutRequest) (models.Pedido, error)
	// AcreditarComisiones reparte las comisiones de una compra entre los referidos ascendentes.
	AcreditarComisiones(emailComprador string, total float64) ([]string, error)
}

// CacheRepository es el contrato de la caché (Redis) que usan los controladores.
type CacheRepository interface {
	// Obtener lee una clave y avisa si existía (Cache Hit).
	Obtener(ctx context.Context, clave string) ([]byte, bool)
	// Guardar escribe una clave con su tiempo de vida (Cache Miss).
	Guardar(ctx context.Context, clave string, datos []byte, ttl time.Duration)
	// Eliminar invalida una o varias claves cuando los datos cambian.
	Eliminar(ctx context.Context, claves ...string)
}
