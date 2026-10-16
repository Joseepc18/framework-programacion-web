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
