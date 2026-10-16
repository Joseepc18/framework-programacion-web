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

	// ===== Fase 3: consulta y gestión de pedidos =====
	// Historial propio: cada token solo recibe los pedidos de su propio usuario_id.
	app.Get("/api/pedidos/mios", middleware.RequireAuth(), controllers.GetPedidosMios)

	// Ciclo de vida de los pedidos: exclusivo del administrador.
	app.Get("/api/pedidos", middleware.RequireAuth(), middleware.RequireAdmin(), controllers.GetPedidosAdmin)
	app.Patch("/api/pedidos/:id/estado", middleware.RequireAuth(), middleware.RequireAdmin(), controllers.CambiarEstadoPedido)

	// Detalle con líneas: lo consulta el dueño (token) o el administrador.
	app.Get("/api/pedidos/:id", middleware.RequireAuth(), controllers.GetPedidoDetalle)

	// ===== Fase 4: ledger de comisiones y red propia =====
	// Cada usuario ve SU subárbol de referidos (no el árbol global de la semilla).
	app.Get("/api/red/mia", middleware.RequireAuth(), controllers.GetRedMia)

	// Ledger de comisiones del usuario: totales pendiente y pagado incluidos.
	app.Get("/api/comisiones/mias", middleware.RequireAuth(), controllers.GetComisionesMias)

	// ===== Fase 5: liquidaciones a distribuidores =====
	// Deuda por distribuidor y histórico de pagos: exclusivos del admin.
	app.Get("/api/liquidaciones/deudas", middleware.RequireAuth(), middleware.RequireAdmin(), controllers.GetDeudas)
	app.Post("/api/liquidaciones", middleware.RequireAuth(), middleware.RequireAdmin(), controllers.CrearLiquidacion)
	app.Get("/api/liquidaciones", middleware.RequireAuth(), middleware.RequireAdmin(), controllers.GetLiquidaciones)

	// Cada distribuidor consulta sus propios cobros con su token.
	app.Get("/api/liquidaciones/mias", middleware.RequireAuth(), controllers.GetLiquidacionesMias)

	// ===== Fase 7: administración de usuarios (solo admin) =====
	app.Get("/api/usuarios", middleware.RequireAuth(), middleware.RequireAdmin(), controllers.GetUsuarios)
	app.Patch("/api/usuarios/:id", middleware.RequireAuth(), middleware.RequireAdmin(), controllers.ActualizarUsuario)

	// ===== Fase 6: reportes de negocio (solo admin) =====
	app.Get("/api/reportes/resumen", middleware.RequireAuth(), middleware.RequireAdmin(), controllers.GetReporteResumen)
	app.Get("/api/reportes/top-productos", middleware.RequireAuth(), middleware.RequireAdmin(), controllers.GetReporteTopProductos)
	app.Get("/api/reportes/top-referidos", middleware.RequireAuth(), middleware.RequireAdmin(), controllers.GetReporteTopReferidos)
	app.Get("/api/reportes/comisiones-pendientes", middleware.RequireAuth(), middleware.RequireAdmin(), controllers.GetReporteComisionesPendientes)
}
