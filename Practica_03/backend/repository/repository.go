// Declaramos el paquete repository, cuya única responsabilidad es proveer el acceso a los datos.
// Por ahora los datos son estáticos (en memoria); en la siguiente fase se reemplazarán por una base de datos relacional.
package repository

import (
	// Importamos nuestro paquete de modelos para usar las estructuras Usuario, Producto y Referido.
	"multicatalogo-backend/models"
)

// usuarios contiene las cuentas registradas con su rol y token ficticio.
var usuarios = []models.Usuario{
	// Cuenta del administrador.
	{Email: "admin@upse.edu.ec", Password: "123456", Rol: "admin", Token: "fake-jwt-token-123"},
	// Cuenta de un cliente registrado.
	{Email: "cliente@upse.edu.ec", Password: "123456", Rol: "cliente", Token: "fake-jwt-token-456"},
}

// productos contiene el catálogo de artículos disponibles.
var productos = []models.Producto{
	// Agregamos el primer producto con sus respectivos valores para ID, Nombre, Precio e Img.
	{ID: 1, Nombre: "Serum Revitalizante", Precio: 45.00, Img: "https://picsum.photos/seed/serum/600/400"},
	// Agregamos el segundo producto a la lista.
	{ID: 2, Nombre: "Crema Hidratante Pro", Precio: 32.50, Img: "https://picsum.photos/seed/crema/600/400"},
	// Agregamos el tercer producto a la lista.
	{ID: 3, Nombre: "Tónico Purificante", Precio: 28.00, Img: "https://picsum.photos/seed/tonico/600/400"},
	// Agregamos el cuarto producto a la lista.
	{ID: 4, Nombre: "Mascarilla Nocturna", Precio: 50.00, Img: "https://picsum.photos/seed/mascarilla/600/400"},
}

// red contiene el árbol de la red multinivel; la raíz es el usuario autenticado (nivel 0).
var red = models.Referido{
	ID: 0, Nombre: "Tú", Nivel: 0, Ventas: 2400,
	Hijos: []models.Referido{
		{
			ID: 1, Nombre: "Ana García", Nivel: 1, Ventas: 1200,
			Hijos: []models.Referido{
				{
					ID: 4, Nombre: "Carlos Ruiz", Nivel: 2, Ventas: 500,
					Hijos: []models.Referido{
						{ID: 7, Nombre: "Diana Paz", Nivel: 3, Ventas: 300},
					},
				},
				{ID: 5, Nombre: "Sofía León", Nivel: 2, Ventas: 430},
			},
		},
		{
			ID: 2, Nombre: "Luis Poveda", Nivel: 1, Ventas: 850,
			Hijos: []models.Referido{
				{ID: 6, Nombre: "Marco Díaz", Nivel: 2, Ventas: 380},
			},
		},
		{ID: 3, Nombre: "Marta Sánchez", Nivel: 1, Ventas: 430},
	},
}

// BuscarUsuario devuelve el usuario cuyas credenciales coinciden; el booleano indica si se encontró.
func BuscarUsuario(email, password string) (models.Usuario, bool) {
	// Recorremos las cuentas registradas comparando correo y contraseña.
	for _, u := range usuarios {
		if u.Email == email && u.Password == password {
			return u, true
		}
	}
	// Si ninguna cuenta coincide, devolvemos un usuario vacío y false.
	return models.Usuario{}, false
}

// ObtenerProductos devuelve el catálogo completo.
func ObtenerProductos() []models.Producto {
	return productos
}

// ObtenerProductoPorID devuelve el producto con el ID indicado; el booleano indica si existe.
func ObtenerProductoPorID(id int) (models.Producto, bool) {
	// Recorremos el catálogo buscando el ID solicitado.
	for _, p := range productos {
		if p.ID == id {
			return p, true
		}
	}
	// Si no existe, devolvemos un producto vacío y false.
	return models.Producto{}, false
}

// ObtenerRed devuelve el árbol completo de la red multinivel.
func ObtenerRed() models.Referido {
	return red
}
