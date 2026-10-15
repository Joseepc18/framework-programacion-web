// Declaramos el paquete repository, cuya única responsabilidad es proveer el acceso a los datos.
// Los datos viven en PostgreSQL; este paquete traduce cada consulta SQL a las estructuras de models.
package repository

import (
	// Importamos context para propagar el contexto de cada consulta.
	"context"
	// Importamos errors para distinguir el caso "sin resultados" de un error real.
	"errors"
	// Importamos fmt para construir el número del pedido y describir errores de negocio.
	"fmt"
	// Importamos time para leer la fecha de creación que genera PostgreSQL.
	"time"

	// Importamos pgx para reconocer el error pgx.ErrNoRows.
	"github.com/jackc/pgx/v5"
	// Importamos pgxpool para recibir el pool de conexiones.
	"github.com/jackc/pgx/v5/pgxpool"
	// Importamos nuestro paquete de modelos para usar las estructuras Usuario, Producto y Referido.
	"multicatalogo-backend/models"
)

// Repositorio es la puerta de entrada a los datos: guarda el pool de conexiones que recibe
// inyectado desde main.go. Al ser una estructura (y no variables globales) se puede sustituir
// por una implementación falsa en las pruebas y se conoce explícitamente de dónde vienen los datos.
type Repositorio struct {
	// db es el pool de conexiones a PostgreSQL.
	db *pgxpool.Pool
}

// Nuevo construye el repositorio con el pool de conexiones que usará en cada consulta.
func Nuevo(pool *pgxpool.Pool) *Repositorio {
	return &Repositorio{db: pool}
}

// tasasComision define, por nivel ascendente, el porcentaje que recibe cada patrocinador
// cuando alguien de su red compra: nivel 1 = 10 %, nivel 2 = 5 %, nivel 3 = 2 %.
var tasasComision = []float64{0.10, 0.05, 0.02}

// BuscarUsuarioPorCredenciales devuelve el usuario cuyo email y contraseña coinciden.
// El booleano indica si se encontró; el error indica un fallo de la base de datos.
func (r *Repositorio) BuscarUsuarioPorCredenciales(email, password string) (models.Usuario, bool, error) {
	var u models.Usuario

	// Consultamos con parámetros ($1, $2) para evitar inyección SQL.
	err := r.db.QueryRow(context.Background(),
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
func (r *Repositorio) ObtenerProductos() ([]models.Producto, error) {
	rows, err := r.db.Query(context.Background(),
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
func (r *Repositorio) ObtenerProductoPorID(id int) (models.Producto, bool, error) {
	var p models.Producto

	err := r.db.QueryRow(context.Background(),
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
func (r *Repositorio) ObtenerRed() (models.Referido, error) {
	rows, err := r.db.Query(context.Background(),
		`SELECT id, nombre, nivel, ventas, comisiones, parent_id FROM referidos ORDER BY id`,
	)
	if err != nil {
		return models.Referido{}, err
	}

	filas, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (referidoPlano, error) {
		var f referidoPlano
		err := row.Scan(&f.referido.ID, &f.referido.Nombre, &f.referido.Nivel, &f.referido.Ventas, &f.referido.Comisiones, &f.parentID)
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

// totalDelPedido suma los subtotales de los items recibidos desde el frontend.
func totalDelPedido(items []models.ItemCarrito) float64 {
	total := 0.0
	for _, item := range items {
		total += item.Precio * float64(item.Cantidad)
	}
	return total
}

// RegistrarPedido guarda un pedido completo aplicando una transacción ACID:
// o se guardan la cabecera, el detalle y el stock en una sola operación, o no se guarda nada.
func (r *Repositorio) RegistrarPedido(req models.CheckoutRequest) (models.Pedido, error) {
	ctx := context.Background()

	// 1) Abrimos la transacción: desde aquí nada es visible para los demás hasta el COMMIT.
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return models.Pedido{}, err
	}
	// 2) El defer garantiza el ROLLBACK si alguna sentencia falla o si se sale antes del COMMIT.
	// Después de un COMMIT exitoso esta llamada no tiene efecto (pgx devuelve ErrTxClosed).
	defer func() { _ = tx.Rollback(ctx) }()

	// 3) Calculamos el total a partir de los items del carrito.
	total := totalDelPedido(req.Items)

	// 4) Insertamos la cabecera del pedido y recuperamos el id y la fecha que genera PostgreSQL.
	// La fecha se convierte a la zona horaria local para mostrarla en la confirmación.
	var pedido models.Pedido
	var creadoEn time.Time
	err = tx.QueryRow(ctx,
		`INSERT INTO pedidos (cliente_nombre, cliente_email, cliente_direccion, cliente_ciudad, cliente_telefono, metodo_pago, total, estado)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, 'confirmado')
		 RETURNING id, creado_en AT TIME ZONE 'America/Guayaquil'`,
		req.Cliente.Nombre, req.Cliente.Email, req.Cliente.Direccion,
		req.Cliente.Ciudad, req.Cliente.Telefono, req.MetodoPago, total,
	).Scan(&pedido.ID, &creadoEn)
	if err != nil {
		return models.Pedido{}, err
	}

	// 5) El número legible del pedido se deriva del id: MC-000001, MC-000002, ...
	pedido.Numero = fmt.Sprintf("MC-%06d", pedido.ID)
	if _, err := tx.Exec(ctx, `UPDATE pedidos SET numero = $1 WHERE id = $2`, pedido.Numero, pedido.ID); err != nil {
		return models.Pedido{}, err
	}

	// 6) Insertamos el detalle y descontamos el stock de cada producto.
	for _, item := range req.Items {
		subtotal := item.Precio * float64(item.Cantidad)

		if _, err := tx.Exec(ctx,
			`INSERT INTO pedido_items (pedido_id, producto_id, nombre, cantidad, precio_unitario, subtotal)
			 VALUES ($1, $2, $3, $4, $5, $6)`,
			pedido.ID, item.ID, item.Nombre, item.Cantidad, item.Precio, subtotal,
		); err != nil {
			// Un producto inexistente viola la clave foránea y aborta todo el pedido.
			return models.Pedido{}, err
		}

		// El stock no puede quedar negativo: la restricción CHECK de la tabla lo impide.
		if _, err := tx.Exec(ctx,
			`UPDATE productos SET stock = stock - $1 WHERE id = $2`,
			item.Cantidad, item.ID,
		); err != nil {
			return models.Pedido{}, err
		}
	}

	// 7) COMMIT: recién aquí el pedido, el detalle y el stock quedan guardados de forma permanente.
	if err := tx.Commit(ctx); err != nil {
		return models.Pedido{}, err
	}

	// 8) Completamos la respuesta que verá el cliente en la pantalla de confirmación.
	pedido.Estado = "confirmado"
	pedido.Total = total
	pedido.MetodoPago = req.MetodoPago
	pedido.Cliente = req.Cliente
	pedido.Items = req.Items
	pedido.Fecha = creadoEn.Format("02/01/2006 15:04:05")

	return pedido, nil
}

// ascendiente es un nodo de la red ubicado por encima del comprador.
type ascendiente struct {
	id    int
	email *string
	salto int
}

// AcreditarComisiones reparte la comisión de una compra entre los referidos ascendentes
// (niveles 1, 2 y 3) y devuelve los correos de los usuarios cuyos KPIs cambiaron.
// Se ejecuta en segundo plano (Goroutine) después del COMMIT del pedido.
func (r *Repositorio) AcreditarComisiones(emailComprador string, total float64) ([]string, error) {
	ctx := context.Background()

	// La acreditación también es transaccional: o se pagan todas las comisiones o ninguna.
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	// Una sola consulta recursiva sube por parent_id desde el nodo del comprador:
	// salto 0 es él mismo, salto 1 su patrocinador directo, y así hasta el tercer nivel.
	filas, err := tx.Query(ctx,
		`WITH RECURSIVE ascendencia AS (
		     SELECT id, email, parent_id, 0 AS salto
		     FROM referidos
		     WHERE email = $1
		     UNION ALL
		     SELECT r.id, r.email, r.parent_id, a.salto + 1
		     FROM referidos r
		     JOIN ascendencia a ON r.id = a.parent_id
		     WHERE a.salto < $2
		 )
		 SELECT id, email, salto FROM ascendencia ORDER BY salto`,
		emailComprador, len(tasasComision),
	)
	if err != nil {
		return nil, err
	}

	ascendientes, err := pgx.CollectRows(filas, func(row pgx.CollectableRow) (ascendiente, error) {
		var a ascendiente
		err := row.Scan(&a.id, &a.email, &a.salto)
		return a, err
	})
	if err != nil {
		return nil, err
	}

	// Si el comprador no pertenece a la red, no hay comisiones que repartir.
	if len(ascendientes) == 0 {
		return nil, fmt.Errorf("el usuario %s no pertenece a la red de referidos", emailComprador)
	}

	// Correos de los usuarios afectados: se usan después para invalidar su caché.
	afectados := []string{}

	for _, a := range ascendientes {
		if a.salto == 0 {
			// El nodo del comprador acumula la venta: alimenta el KPI "Ventas de la Red".
			if _, err := tx.Exec(ctx, `UPDATE referidos SET ventas = ventas + $1 WHERE id = $2`, total, a.id); err != nil {
				return nil, err
			}
		} else {
			// Cada nivel ascendente recibe su porcentaje de las comisiones del mes.
			comision := total * tasasComision[a.salto-1]
			if _, err := tx.Exec(ctx, `UPDATE referidos SET comisiones = comisiones + $1 WHERE id = $2`, comision, a.id); err != nil {
				return nil, err
			}
		}

		// Solo los nodos asociados a una cuenta real tienen una clave de caché que invalidar.
		if a.email != nil && *a.email != "" {
			afectados = append(afectados, *a.email)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return afectados, nil
}
