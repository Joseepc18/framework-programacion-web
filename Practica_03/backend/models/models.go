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
	// Email es el correo con el que el usuario inicia sesión.
	Email string `json:"email"`
	// Password es la contraseña del usuario; el guion en la etiqueta evita que se envíe en las respuestas JSON.
	Password string `json:"-"`
	// Rol indica los permisos del usuario dentro del sistema (admin o cliente).
	Rol string `json:"rol"`
	// Token es el token ficticio que se entrega al iniciar sesión correctamente.
	Token string `json:"token"`
}

// Producto define la estructura de los datos de un artículo en nuestro catálogo.
type Producto struct {
	// ID es el identificador único numérico del producto.
	ID int `json:"id"`
	// Nombre es la descripción en texto del producto.
	Nombre string `json:"nombre"`
	// Precio es el costo del producto, almacenado como un número con decimales (float64).
	Precio float64 `json:"precio"`
	// Img es la URL o ruta que apunta a la fotografía o imagen del producto.
	Img string `json:"img"`
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
