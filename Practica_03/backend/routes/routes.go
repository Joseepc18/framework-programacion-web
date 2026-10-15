// Declaramos el paquete routes, cuya única responsabilidad es registrar los endpoints de la API.
package routes

import (
	// Importamos Fiber para poder recibir y configurar la instancia principal de la aplicación.
	"github.com/gofiber/fiber/v2"
	// Importamos nuestro paquete de controladores: aquí viven los manejadores de cada ruta.
	"multicatalogo-backend/controllers"
)

// Manejadores agrupa los controladores ya construidos y con sus dependencias inyectadas.
// Se recibe como parámetro para que las rutas no sepan cómo se conectan a la base ni a la caché.
type Manejadores struct {
	// Auth atiende el inicio de sesión.
	Auth *controllers.AuthController
	// Productos atiende el catálogo.
	Productos *controllers.ProductoController
	// Red atiende el árbol de la red multinivel.
	Red *controllers.RedController
	// Checkout atiende la compra.
	Checkout *controllers.CheckoutController
	// Health verifica el estado del servidor.
	Health *controllers.HealthController
}

// SetupRoutes recibe la aplicación Fiber y los controladores, y enlaza cada ruta con su manejador.
func SetupRoutes(app *fiber.App, m Manejadores) {
	// Definimos una ruta HTTP POST en "/api/login" y la enlazamos al manejador Login.
	app.Post("/api/login", m.Auth.Login)

	// Definimos una ruta HTTP GET en "/api/productos" y la enlazamos al manejador GetProductos.
	app.Get("/api/productos", m.Productos.GetProductos)

	// Definimos una ruta HTTP GET en "/api/productos/:id" para obtener un producto por su ID.
	app.Get("/api/productos/:id", m.Productos.GetProductoPorID)

	// Definimos una ruta HTTP GET en "/api/red" para obtener el árbol de la red multinivel.
	app.Get("/api/red", m.Red.GetRed)

	// Definimos una ruta HTTP POST en "/api/checkout" para registrar la compra (transacción ACID).
	app.Post("/api/checkout", m.Checkout.Procesar)

	// Definimos una ruta HTTP GET en "/api/health" para verificar que el servidor está operativo.
	app.Get("/api/health", m.Health.HealthCheck)
}
