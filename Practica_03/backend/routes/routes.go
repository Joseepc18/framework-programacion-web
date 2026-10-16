// Declaramos el paquete routes, cuya única responsabilidad es registrar los endpoints de la API.
package routes

import (
	// Importamos Fiber para poder recibir y configurar la instancia principal de la aplicación.
	"github.com/gofiber/fiber/v2"
	// Importamos nuestro paquete de controladores, donde residen las funciones que se ejecutarán en cada ruta.
	"multicatalogo-backend/controllers"
	// Importamos el middleware de autenticación para proteger rutas sensibles.
	"multicatalogo-backend/middleware"
)

// SetupRoutes es una función que recibe un puntero a la aplicación Fiber (*fiber.App) para inyectarle las rutas.
func SetupRoutes(app *fiber.App) {
	// Definimos una ruta HTTP POST en "/api/login" y la enlazamos a la función Login del paquete controllers.
	app.Post("/api/login", controllers.Login)

	// Registro público de cuentas de cliente (opcionalmente con código de referido).
	app.Post("/api/register", controllers.Register)

	// Definimos una ruta HTTP GET en "/api/productos" y la enlazamos a la función GetProductos del paquete controllers.
	app.Get("/api/productos", controllers.GetProductos)

	// Definimos una ruta HTTP GET en "/api/productos/:id" para obtener un producto por su ID.
	app.Get("/api/productos/:id", controllers.GetProductoPorID)

	// Catálogo administrativo: incluye productos inactivos y acepta ?q= de búsqueda.
	app.Get("/api/admin/productos", middleware.RequireAuth(), middleware.RequireAdmin(), controllers.GetProductosAdmin)

	// Escrituras del catálogo: solo administradores y con invalidación de caché.
	app.Post("/api/productos", middleware.RequireAuth(), middleware.RequireAdmin(), controllers.CrearProducto)
	app.Put("/api/productos/:id", middleware.RequireAuth(), middleware.RequireAdmin(), controllers.ActualizarProducto)
	// Borrado lógico: ?activar=true vuelve a publicar el producto.
	app.Delete("/api/productos/:id", middleware.RequireAuth(), middleware.RequireAdmin(), controllers.DesactivarProducto)

	// Definimos una ruta HTTP GET en "/api/health" para verificar que el servidor está operativo.
	app.Get("/api/health", controllers.HealthCheck)

	// Rutas protegidas mediante JWT (RequireAuth)
	// Árbol de referidos protegido: solo usuarios autenticados pueden consultarlo.
	app.Get("/api/red", middleware.RequireAuth(), controllers.GetRed)

	// Checkout transaccional protegido: valida inventario y asigna compra a la identidad del token.
	app.Post("/api/checkout", middleware.RequireAuth(), controllers.Checkout)
}
