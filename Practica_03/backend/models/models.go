// Declaramos el paquete models para agrupar las estructuras de datos de la aplicación.
package models

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
	// Comisiones es el saldo acumulado que el referido ha ganado por las compras de su red.
	Comisiones float64 `json:"comisiones"`
	// Hijos son los referidos un nivel más abajo; omitempty evita enviar el campo si no tiene hijos.
	Hijos []Referido `json:"hijos,omitempty"`
}

// ItemCarrito representa un producto del carrito tal como lo envía Checkout.tsx.
type ItemCarrito struct {
	// ID es el identificador del producto que se está comprando.
	ID int `json:"id"`
	// Nombre es el nombre del producto en el momento de la compra.
	Nombre string `json:"nombre"`
	// Cantidad es el número de unidades pedidas.
	Cantidad int `json:"cantidad"`
	// Precio es el precio unitario en dólares.
	Precio float64 `json:"precio"`
}

// DatosCliente agrupa los datos del formulario de envío del checkout.
type DatosCliente struct {
	// Nombre es el nombre de quien recibe el pedido.
	Nombre string `json:"nombre"`
	// Email es el correo de contacto del pedido.
	Email string `json:"email"`
	// Direccion es la dirección de entrega.
	Direccion string `json:"direccion"`
	// Ciudad es la ciudad de entrega.
	Ciudad string `json:"ciudad"`
	// Telefono es el teléfono de contacto.
	Telefono string `json:"telefono"`
}

// CheckoutRequest es el cuerpo JSON que recibe POST /api/checkout.
type CheckoutRequest struct {
	// Cliente son los datos del formulario de envío.
	Cliente DatosCliente `json:"cliente"`
	// UsuarioEmail es el correo de la cuenta autenticada: identifica su nodo en la red.
	UsuarioEmail string `json:"usuarioEmail"`
	// MetodoPago es la forma de pago elegida (tarjeta, transferencia, efectivo).
	MetodoPago string `json:"metodoPago"`
	// Items es el arreglo de productos del carrito.
	Items []ItemCarrito `json:"items"`
}

// Pedido es el registro que se guarda en PostgreSQL y se devuelve al frontend tras un checkout exitoso.
type Pedido struct {
	// ID es el identificador del pedido generado por la base de datos.
	ID int `json:"id"`
	// Numero es el código legible del pedido (MC-000001).
	Numero string `json:"numero"`
	// Estado indica la situación del pedido; tras el COMMIT es "confirmado".
	Estado string `json:"estado"`
	// Total es la suma de los subtotales del pedido.
	Total float64 `json:"total"`
	// Fecha es la fecha de registro ya formateada para mostrarla en la confirmación.
	Fecha string `json:"fecha"`
	// MetodoPago es la forma de pago con la que se registró el pedido.
	MetodoPago string `json:"metodoPago"`
	// Cliente son los datos del formulario de envío.
	Cliente DatosCliente `json:"cliente"`
	// Items es el detalle del pedido tal como quedó registrado.
	Items []ItemCarrito `json:"items"`
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
