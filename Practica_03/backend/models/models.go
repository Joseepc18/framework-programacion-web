// Declaramos el paquete models para agrupar las estructuras de datos de la aplicación.
package models

// Importamos time para registrar la fecha de creación de cada pedido.
import "time"

// LoginRequest define la estructura esperada para el cuerpo de la petición (JSON) al iniciar sesión.
type LoginRequest struct {
	// Email representa el correo del usuario; la etiqueta `json:"email"` indica cómo se mapea el JSON a esta variable.
	Email string `json:"email"`
	// Password representa la contraseña del usuario; se extrae del campo "password" del JSON entrante.
	Password string `json:"password"`
}

// Usuario define los datos de una cuenta registrada en el sistema.
type Usuario struct {
	// ID es el identificador único del usuario en la base de datos.
	ID int `json:"id"`
	// Email es el correo con el que el usuario inicia sesión.
	Email string `json:"email"`
	// Password es la contraseña del usuario; el guion en la etiqueta evita que se envíe en las respuestas JSON.
	Password string `json:"-"`
	// Rol indica los permisos del usuario dentro del sistema (admin o cliente).
	Rol string `json:"rol"`
	// Nombre es el nombre visible de la cuenta.
	Nombre string `json:"nombre"`
	// Activo indica si la cuenta puede iniciar sesión (borrado lógico).
	Activo bool `json:"-"`
	// ReferidoID es el nodo de la red multinivel que representa a esta cuenta.
	ReferidoID *int `json:"referido_id,omitempty"`
}

// RegisterRequest define el cuerpo de la petición de alta de una cuenta nueva.
type RegisterRequest struct {
	// Nombre es el nombre visible del nuevo usuario.
	Nombre string `json:"nombre"`
	// Email es el correo con el que iniciará sesión; debe ser único.
	Email string `json:"email"`
	// Password es la contraseña en texto plano; se guarda hasheada con bcrypt.
	Password string `json:"password"`
	// Ref es el código de invitación del patrocinador (opcional).
	Ref string `json:"ref"`
}

// Producto define la estructura de los datos de un artículo en nuestro catálogo.
type Producto struct {
	// ID es el identificador único numérico del producto.
	ID int `json:"id"`
	// Nombre es el nombre comercial del producto.
	Nombre string `json:"nombre"`
	// Descripcion es el texto descriptivo que se muestra en el detalle del producto.
	Descripcion string `json:"descripcion"`
	// Precio es el costo del producto, almacenado como un número con decimales (float64).
	Precio float64 `json:"precio"`
	// Categoria agrupa los productos para los filtros del catálogo (Serum, Crema, ...).
	Categoria string `json:"categoria"`
	// Img es la URL o ruta que apunta a la fotografía o imagen del producto.
	Img string `json:"img"`
	// Galeria contiene las URLs de las imágenes adicionales del producto.
	Galeria []string `json:"galeria"`
	// Stock indica la cantidad disponible en inventario.
	Stock int `json:"stock"`
	// Activo indica si el producto se muestra en el catálogo (borrado lógico).
	Activo bool `json:"activo"`
}

// ItemPedido representa un producto y su cantidad dentro del checkout.
type ItemPedido struct {
	ProductoID int `json:"producto_id"`
	Cantidad   int `json:"cantidad"`
}

// CheckoutRequest define la petición de compra.
type CheckoutRequest struct {
	Nombre      string       `json:"nombre"`
	Email       string       `json:"email"`
	Direccion   string       `json:"direccion"`
	Ciudad      string       `json:"ciudad"`
	Telefono    string       `json:"telefono"`
	MetodoPago  string       `json:"metodoPago"`
	Items       []ItemPedido `json:"items"`
}

// PedidoConfirmado es la respuesta exitosa tras el checkout transaccional.
type PedidoConfirmado struct {
	PedidoID int     `json:"pedido_id"`
	Numero   string  `json:"numero"`
	Total    float64 `json:"total"`
	Message  string  `json:"message"`
}

// Pedido define la cabecera de un pedido registrado en el sistema.
type Pedido struct {
	// ID es el identificador numérico del pedido.
	ID int `json:"id"`
	// Numero es el código legible generado en el checkout (MC-...).
	Numero string `json:"numero"`
	// ClienteNombre es el nombre de quien realizó la compra.
	ClienteNombre string `json:"cliente_nombre"`
	// ClienteEmail es el correo de contacto capturado en el checkout.
	ClienteEmail string `json:"cliente_email"`
	// ClienteDireccion, ClienteCiudad y ClienteTelefono son los datos de envío.
	ClienteDireccion string `json:"cliente_direccion"`
	ClienteCiudad    string `json:"cliente_ciudad"`
	ClienteTelefono  string `json:"cliente_telefono"`
	// MetodoPago indica la forma de pago elegida.
	MetodoPago string `json:"metodo_pago"`
	// Total es la suma de los subtotales de todos los ítems.
	Total float64 `json:"total"`
	// UsuarioID es el dueño real del pedido (claim sub del JWT).
	UsuarioID *int `json:"usuario_id,omitempty"`
	// Estado refleja el ciclo de vida: pendiente, preparando, enviado, entregado o cancelado.
	Estado string `json:"estado"`
	// CreadoEn registra la fecha y hora de alta del pedido.
	CreadoEn time.Time `json:"creado_en"`
}

// ItemPedidoDetalle es una línea de pedido ya calculada con sus precios.
type ItemPedidoDetalle struct {
	// ProductoID referencia al catálogo en el momento de la compra.
	ProductoID int `json:"producto_id"`
	// Nombre es la denominación del producto copiada al historial.
	Nombre string `json:"nombre"`
	// Cantidad es el número de unidades compradas.
	Cantidad int `json:"cantidad"`
	// PrecioUnitario es el precio vigente al comprar (no el actual).
	PrecioUnitario float64 `json:"precio_unitario"`
	// Subtotal es precio_unitario por cantidad.
	Subtotal float64 `json:"subtotal"`
}

// PedidoDetalle agrupa la cabecera del pedido con sus líneas (JOIN con pedido_items).
type PedidoDetalle struct {
	// Pedido contiene los datos de cabecera.
	Pedido
	// Items es el detalle de productos del pedido.
	Items []ItemPedidoDetalle `json:"items"`
}

// Pagina estandariza la respuesta paginada de todos los listados de la API.
type Pagina struct {
	// Data es la colección de la página actual.
	Data interface{} `json:"data"`
	// Page es la página solicitada (comienza en 1).
	Page int `json:"page"`
	// Limit es el tamaño de página aplicado.
	Limit int `json:"limit"`
	// Total es el número total de registros disponibles.
	Total int `json:"total"`
}

// Comision es una fila del ledger de comisiones: quién cobró, de qué pedido,
// en qué nivel, con qué porcentaje y en qué estado está.
type Comision struct {
	// ID es el identificador de la fila en el ledger.
	ID int `json:"id"`
	// PedidoID referencia el pedido que generó la comisión.
	PedidoID int `json:"pedido_id"`
	// PedidoNumero es el código legible del pedido (MC-...).
	PedidoNumero string `json:"pedido_numero"`
	// Nivel es la profundidad en la red del cobrador (1 a 3).
	Nivel int `json:"nivel"`
	// Porcentaje es la tasa aplicada (0.10, 0.05 o 0.02).
	Porcentaje float64 `json:"porcentaje"`
	// Monto es la plata acreditada por esta fila.
	Monto float64 `json:"monto"`
	// Estado es pendiente, aprobada, pagada o revertida.
	Estado string `json:"estado"`
	// CreadoEn registra cuándo se generó la comisión.
	CreadoEn time.Time `json:"creado_en"`
}

// PaginaComisiones agrega los totales por estado a la paginación estándar.
type PaginaComisiones struct {
	// Data es la página de filas del ledger.
	Data []Comision `json:"data"`
	// Page, Limit y Total describen la paginación de las filas.
	Page  int `json:"page"`
	Limit int `json:"limit"`
	Total int `json:"total"`
	// Pendiente es la suma de comisiones aún no pagadas.
	Pendiente float64 `json:"pendiente"`
	// Pagado es la suma de comisiones ya liquidadas.
	Pagado float64 `json:"pagado"`
}

// RedPropia devuelve el subárbol de red del usuario y su código de invitación.
type RedPropia struct {
	// Red es el subárbol que cuelga del nodo del usuario (nil si no tiene nodo).
	Red *Referido `json:"red"`
	// Codigo es el código de invitación del propio nodo (link /registro?ref=).
	Codigo string `json:"codigo"`
	// ReferidoID es el nodo del que cuelga la red (0 si no existe).
	ReferidoID int `json:"referido_id"`
}

// Liquidacion agrupa un pago a un distribuidor: qué nodo cobró, cuánto y cuándo.
type Liquidacion struct {
	// ID es el identificador del pago.
	ID int `json:"id"`
	// ReferidoID es el nodo de la red que recibió el pago.
	ReferidoID int `json:"referido_id"`
	// Nombre es el nombre visible del distribuidor (JOIN con referidos).
	Nombre string `json:"nombre"`
	// Total es la suma de las comisiones liquidadas en este pago.
	Total float64 `json:"total"`
	// Estado es pendiente o pagada.
	Estado string `json:"estado"`
	// CreadoEn registra cuándo se creó la liquidación.
	CreadoEn time.Time `json:"creado_en"`
	// PagadaEn es el momento del pago efectivo (null si sigue pendiente).
	PagadaEn *time.Time `json:"pagada_en,omitempty"`
	// CantidadComisiones indica cuántas filas del ledger cubrió este pago.
	CantidadComisiones int `json:"cantidad_comisiones"`
}

// DeudaComisiones resume lo que un distribuidor tiene pendiente de cobro.
type DeudaComisiones struct {
	// ReferidoID es el nodo deudor.
	ReferidoID int `json:"referido_id"`
	// Nombre es el nombre del distribuidor.
	Nombre string `json:"nombre"`
	// Email es la cuenta real asociada al nodo (vacío si no tiene).
	Email string `json:"email"`
	// Pendiente es la suma de comisiones en estado pendiente.
	Pendiente float64 `json:"pendiente"`
	// Cantidad es el número de filas pendientes.
	Cantidad int `json:"cantidad"`
}

// ResumenVentas agrega los KPIs de negocio de un periodo (pedidos no cancelados).
type ResumenVentas struct {
	// TotalVendido es la suma de total de los pedidos del periodo.
	TotalVendido float64 `json:"total_vendido"`
	// NumPedidos es la cantidad de pedidos del periodo.
	NumPedidos int `json:"num_pedidos"`
	// TicketPromedio es TotalVendido / NumPedidos (0 si no hay pedidos).
	TicketPromedio float64 `json:"ticket_promedio"`
	// ComisionesDevengadas es la suma del ledger generado en el periodo, sin las revertidas.
	ComisionesDevengadas float64 `json:"comisiones_devengadas"`
}

// TopProducto es una fila del ranking de productos por ingresos.
type TopProducto struct {
	ProductoID int     `json:"producto_id"`
	Nombre     string  `json:"nombre"`
	Unidades   int     `json:"unidades"`
	Ingresos   float64 `json:"ingresos"`
}

// TopReferido es una fila del ranking de distribuidores por ventas generadas en su red.
type TopReferido struct {
	ReferidoID int     `json:"referido_id"`
	Nombre     string  `json:"nombre"`
	VentasRed  float64 `json:"ventas_red"`
	Comisiones float64 `json:"comisiones"`
}

// Referido define un nodo del árbol de la red multinivel.
type Referido struct {
	// ID es el identificador único del referido.
	ID int `json:"id"`
	// Nombre es el nombre completo del referido.
	Nombre string `json:"nombre"`
	// Nivel indica la profundidad en la red (0 = usuario raíz, 1 = directo, 2 = indirecto, 3 = tercer nivel).
	Nivel int `json:"nivel"`
	// Ventas son las ventas mensuales del referido en dólares.
	Ventas float64 `json:"ventas"`
	// Hijos son los referidos un nivel más abajo; omitempty evita enviar el campo si no tiene hijos.
	Hijos []Referido `json:"hijos,omitempty"`
}

// APIError estandariza el formato de todas las respuestas de error HTTP de la API.
type APIError struct {
	// Status es el código de estado HTTP del error (400, 401, 404, ...).
	Status int `json:"status"`
	// Message es la descripción legible del error.
	Message string `json:"message"`
	// Details contiene información adicional opcional; omitempty lo oculta cuando está vacío.
	Details interface{} `json:"details,omitempty"`
}
