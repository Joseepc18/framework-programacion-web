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
// Se compone de contratos más pequeños, uno por área del negocio.
type DBRepository interface {
	UsuarioRepository
	CatalogoRepository
	PedidoRepository
	RedRepository
	LiquidacionRepository
	ReporteRepository
	// Ping comprueba que la base de datos responde (health check).
	Ping(ctx context.Context) error
}

// UsuarioRepository cubre cuentas: login, registro y administración.
type UsuarioRepository interface {
	// CuentaActiva indica si la cuenta existe y está habilitada (lo consulta RequireAuth en cada petición).
	CuentaActiva(ctx context.Context, usuarioID int) (bool, error)
	// BuscarUsuarioPorEmail devuelve la cuenta con su hash de contraseña para verificar el login.
	BuscarUsuarioPorEmail(email string) (models.Usuario, bool, error)
	// CrearUsuario da de alta una cuenta de cliente y su nodo de red en una transacción.
	CrearUsuario(ctx context.Context, req models.RegisterRequest, hashPassword string) (models.Usuario, error)
	// ObtenerUsuarios devuelve una página de cuentas con búsqueda por nombre o correo.
	ObtenerUsuarios(ctx context.Context, busqueda string, page, limit int) (models.Pagina, error)
	// ActualizarUsuario cambia rol y/o estado activo de una cuenta.
	ActualizarUsuario(ctx context.Context, id int, rol *string, activo *bool) (models.UsuarioAdmin, bool, error)
}

// CatalogoRepository cubre la lectura pública y la administración del catálogo.
type CatalogoRepository interface {
	// ObtenerProductos devuelve el catálogo público completo (solo activos).
	ObtenerProductos() ([]models.Producto, error)
	// ObtenerProductoPorID devuelve un producto concreto.
	ObtenerProductoPorID(id int) (models.Producto, bool, error)
	// ObtenerProductosPagina devuelve una página del catálogo con búsqueda server-side.
	ObtenerProductosPagina(ctx context.Context, busqueda string, soloActivos bool, page, limit int) (models.Pagina, error)
	// CrearProducto da de alta un producto.
	CrearProducto(ctx context.Context, p models.Producto) (models.Producto, error)
	// ActualizarProducto edita un producto; false si el ID no existe.
	ActualizarProducto(ctx context.Context, id int, p models.Producto) (bool, error)
	// CambiarEstadoProducto aplica o revierte el borrado lógico.
	CambiarEstadoProducto(ctx context.Context, id int, activo bool) (bool, error)
}

// PedidoRepository cubre el checkout transaccional y la gestión de pedidos.
type PedidoRepository interface {
	// CrearPedido ejecuta la compra ACID: stock, pedido, ítems y ledger de comisiones.
	CrearPedido(ctx context.Context, usuarioID *int, email string, req models.CheckoutRequest) (models.PedidoConfirmado, error)
	// ActualizarComisionesUpline suma las ventas al upline en el contador referidos.ventas.
	ActualizarComisionesUpline(ctx context.Context, email string, monto float64) error
	// ObtenerPedidosMios devuelve el historial del usuario dueño.
	ObtenerPedidosMios(ctx context.Context, usuarioID, page, limit int) (models.Pagina, error)
	// ObtenerPedidosAdmin devuelve todos los pedidos con filtro opcional por estado.
	ObtenerPedidosAdmin(ctx context.Context, estado string, page, limit int) (models.Pagina, error)
	// ObtenerPedidoDetalle devuelve cabecera e ítems si el pedido es del solicitante o este es admin.
	ObtenerPedidoDetalle(ctx context.Context, id int, usuarioID int, email string, esAdmin bool) (models.PedidoDetalle, error)
	// CambiarEstadoPedido valida la transición y, si cancela, devuelve stock y revierte comisiones.
	CambiarEstadoPedido(ctx context.Context, id int, nuevoEstado string) (models.Pedido, error)
}

// RedRepository cubre el árbol multinivel y el ledger de comisiones.
type RedRepository interface {
	// ObtenerRed devuelve el árbol global de la red.
	ObtenerRed() (models.Referido, error)
	// NodoDeRed devuelve el nodo de red de una cuenta, si lo tiene.
	NodoDeRed(ctx context.Context, usuarioID int) (int, bool, error)
	// ObtenerRedPropia devuelve el subárbol de un nodo y su código de invitación.
	ObtenerRedPropia(ctx context.Context, nodoID int) (models.RedPropia, error)
	// ObtenerComisionesMias devuelve el ledger de un nodo con totales por estado.
	ObtenerComisionesMias(ctx context.Context, referidoID, page, limit int) (models.PaginaComisiones, error)
}

// LiquidacionRepository cubre el pago de comisiones a distribuidores.
type LiquidacionRepository interface {
	// CrearLiquidacion paga todo el saldo pendiente de un distribuidor.
	CrearLiquidacion(ctx context.Context, referidoID int) (models.Liquidacion, error)
	// ObtenerLiquidaciones devuelve el histórico global de pagos.
	ObtenerLiquidaciones(ctx context.Context, page, limit int) (models.Pagina, error)
	// ObtenerLiquidacionesMias devuelve los pagos recibidos por un distribuidor.
	ObtenerLiquidacionesMias(ctx context.Context, referidoID, page, limit int) (models.Pagina, error)
	// ObtenerDeudas devuelve el saldo pendiente por distribuidor.
	ObtenerDeudas(ctx context.Context) ([]models.DeudaComisiones, error)
}

// ReporteRepository cubre las agregaciones del dashboard.
type ReporteRepository interface {
	// ObtenerResumenVentas devuelve los KPIs del periodo.
	ObtenerResumenVentas(ctx context.Context, r models.RangoFechas) (models.ResumenVentas, error)
	// ObtenerTopProductos devuelve el ranking de productos por ingresos.
	ObtenerTopProductos(ctx context.Context, r models.RangoFechas, limite int) ([]models.TopProducto, error)
	// ObtenerTopReferidos devuelve el ranking de distribuidores por ventas de su red.
	ObtenerTopReferidos(ctx context.Context, r models.RangoFechas, limite int) ([]models.TopReferido, error)
}

// CacheRepository es el contrato de la caché (Redis) que usan los controladores.
type CacheRepository interface {
	// Obtener lee una clave y avisa si existía (Cache Hit).
	Obtener(ctx context.Context, clave string) ([]byte, bool)
	// Guardar escribe una clave con su tiempo de vida (Cache Miss).
	Guardar(ctx context.Context, clave string, datos []byte, ttl time.Duration)
	// Eliminar invalida una o varias claves cuando los datos cambian.
	Eliminar(ctx context.Context, claves ...string)
	// Ping comprueba que la caché responde (health check).
	Ping(ctx context.Context) error
}

// Mailer es el contrato de los correos transaccionales. Los controladores lo reciben
// inyectado; en desarrollo basta una implementación que escriba en el log.
type Mailer interface {
	// EnviarBienvenida avisa a un usuario recién registrado.
	EnviarBienvenida(ctx context.Context, nombre, email string)
	// EnviarConfirmacionPedido confirma una compra al comprador.
	EnviarConfirmacionPedido(ctx context.Context, email, numero string, total float64)
}
