// Declaramos el paquete repository, cuya única responsabilidad es proveer el acceso a los datos.
// Los datos viven en PostgreSQL; este paquete traduce cada consulta SQL a las estructuras de models.
package repository

import (
	// Importamos context para propagar el contexto de cada consulta.
	"context"
	// Importamos errors para distinguir el caso "sin resultados" de un error real.
	"errors"

	// Importamos pgx para reconocer el error pgx.ErrNoRows.
	"github.com/jackc/pgx/v5"
	// Importamos pgxpool para recibir el pool de conexiones.
	"github.com/jackc/pgx/v5/pgxpool"
	// Importamos nuestro paquete de modelos para usar las estructuras Usuario, Producto y Referido.
	"multicatalogo-backend/models"
)

// db es el pool de conexiones inyectado desde main al arrancar la aplicación.
var db *pgxpool.Pool

// Inicializar recibe el pool de conexiones que usarán todas las funciones del repositorio.
func Inicializar(pool *pgxpool.Pool) {
	db = pool
}

// BuscarUsuarioPorCredenciales devuelve el usuario cuyo email y contraseña coinciden.
// El booleano indica si se encontró; el error indica un fallo de la base de datos.
func BuscarUsuarioPorCredenciales(email, password string) (models.Usuario, bool, error) {
	var u models.Usuario

	// Consultamos con parámetros ($1, $2) para evitar inyección SQL.
	err := db.QueryRow(context.Background(),
		`SELECT id, email, rol FROM usuarios WHERE email = $1 AND password = $2`,
		email, password,
	).Scan(&u.ID, &u.Email, &u.Rol)

	// Si no hay filas, las credenciales no coinciden: no es un error de la base de datos.
	if errors.Is(err, pgx.ErrNoRows) {
		return models.Usuario{}, false, nil
	}
	if err != nil {
		return models.Usuario{}, false, err
	}
	return u, true, nil
}

// ObtenerProductos devuelve el catálogo completo ordenado por ID.
func ObtenerProductos() ([]models.Producto, error) {
	rows, err := db.Query(context.Background(),
		`SELECT id, nombre, descripcion, precio, categoria, img, galeria FROM productos ORDER BY id`,
	)
	if err != nil {
		return nil, err
	}

	// CollectRows recorre las filas, las mapea a Producto y cierra el cursor automáticamente.
	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (models.Producto, error) {
		var p models.Producto
		err := row.Scan(&p.ID, &p.Nombre, &p.Descripcion, &p.Precio, &p.Categoria, &p.Img, &p.Galeria)
		return p, err
	})
}

// ObtenerProductoPorID devuelve el producto con el ID indicado.
// El booleano indica si existe; el error indica un fallo de la base de datos.
func ObtenerProductoPorID(id int) (models.Producto, bool, error) {
	var p models.Producto

	err := db.QueryRow(context.Background(),
		`SELECT id, nombre, descripcion, precio, categoria, img, galeria FROM productos WHERE id = $1`,
		id,
	).Scan(&p.ID, &p.Nombre, &p.Descripcion, &p.Precio, &p.Categoria, &p.Img, &p.Galeria)

	// Si no hay filas, el producto no existe.
	if errors.Is(err, pgx.ErrNoRows) {
		return models.Producto{}, false, nil
	}
	if err != nil {
		return models.Producto{}, false, err
	}
	return p, true, nil
}

// referidoPlano representa una fila de la tabla referidos antes de armar el árbol.
type referidoPlano struct {
	referido models.Referido
	parentID *int
}

// ObtenerRed devuelve el árbol completo de la red multinivel.
// Se consulta la tabla como una lista plana y el árbol se reconstruye en memoria usando parent_id.
func ObtenerRed() (models.Referido, error) {
	rows, err := db.Query(context.Background(),
		`SELECT id, nombre, nivel, ventas, parent_id FROM referidos ORDER BY id`,
	)
	if err != nil {
		return models.Referido{}, err
	}

	filas, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (referidoPlano, error) {
		var f referidoPlano
		err := row.Scan(&f.referido.ID, &f.referido.Nombre, &f.referido.Nivel, &f.referido.Ventas, &f.parentID)
		return f, err
	})
	if err != nil {
		return models.Referido{}, err
	}

	// Agrupamos los IDs de los hijos de cada nodo y ubicamos la raíz (la fila sin parent_id).
	hijosDe := map[int][]int{}
	nodos := map[int]models.Referido{}
	raizID := -1
	for _, f := range filas {
		nodos[f.referido.ID] = f.referido
		if f.parentID == nil {
			raizID = f.referido.ID
		} else {
			hijosDe[*f.parentID] = append(hijosDe[*f.parentID], f.referido.ID)
		}
	}

	// Si la tabla está vacía, no hay red que devolver.
	if raizID == -1 {
		return models.Referido{}, errors.New("la red no tiene un nodo raíz")
	}

	return construirArbol(raizID, nodos, hijosDe), nil
}

// construirArbol arma recursivamente el nodo indicado junto con todos sus descendientes.
func construirArbol(id int, nodos map[int]models.Referido, hijosDe map[int][]int) models.Referido {
	nodo := nodos[id]
	for _, hijoID := range hijosDe[id] {
		nodo.Hijos = append(nodo.Hijos, construirArbol(hijoID, nodos, hijosDe))
	}
	return nodo
}
