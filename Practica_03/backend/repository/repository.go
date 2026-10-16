// Declaramos el paquete repository, cuya única responsabilidad es proveer el acceso a los datos.
// Los datos viven en PostgreSQL; este paquete traduce cada consulta SQL a las estructuras de models.
package repository

import (
	// Importamos context para propagar el contexto de cada consulta.
	"context"
	// Importamos errors para distinguir el caso "sin resultados" de un error real.
	"errors"
	// Importamos strings para normalizar el código de invitación.
	"strings"
	// Importamos time para marcas de tiempo y códigos de pedidos.
	"time"

	// Importamos pgx para reconocer el error pgx.ErrNoRows.
	"github.com/jackc/pgx/v5"
	// Importamos pgconn para reconocer violaciones de unicidad (código 23505).
	"github.com/jackc/pgx/v5/pgconn"
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

// GetDB devuelve el pool de conexiones actual (usado para verificaciones de health).
func GetDB() *pgxpool.Pool {
	return db
}

// BuscarUsuarioPorEmail busca un usuario por su dirección de correo para verificar su contraseña hasheada.
func BuscarUsuarioPorEmail(email string) (models.Usuario, bool, error) {
	var u models.Usuario

	err := db.QueryRow(context.Background(),
		`SELECT id, email, password, rol, nombre, activo, referido_id FROM usuarios WHERE email = $1`,
		email,
	).Scan(&u.ID, &u.Email, &u.Password, &u.Rol, &u.Nombre, &u.Activo, &u.ReferidoID)

	if errors.Is(err, pgx.ErrNoRows) {
		return models.Usuario{}, false, nil
	}
	if err != nil {
		return models.Usuario{}, false, err
	}
	return u, true, nil
}

// Errores de dominio del alta de cuentas.
var (
	ErrEmailYaExiste      = errors.New("el correo electrónico ya está registrado")
	ErrCodigoInvalido     = errors.New("el código de referido no existe")
	ErrRegistroInvalido   = errors.New("nombre, correo y contraseña son obligatorios")
	ErrCuentaDesactivada  = errors.New("cuenta desactivada")
)

// esViolacionDeUnicidad reconoce el error 23505 de PostgreSQL (clave duplicada).
func esViolacionDeUnicidad(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

// generarCodigoPropuesto arma un código de invitación legible a partir del correo
// (parte local en mayúsculas) y le agrega un sufijo numérico hasta que sea único.
func generarCodigoPropuesto(ctx context.Context, tx pgx.Tx, email string) (string, error) {
	base := strings.ToUpper(strings.SplitN(email, "@", 2)[0])
	base = strings.Map(func(r rune) rune {
		if (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			return r
		}
		return -1
	}, base)
	if base == "" {
		base = "INVITADO"
	}
	if len(base) > 20 {
		base = base[:20]
	}

	for sufijo := 0; sufijo < 100; sufijo++ {
		codigo := base
		if sufijo > 0 {
			codigo = base + string(rune('A'+sufijo-1))
		}
		var existe bool
		err := tx.QueryRow(ctx,
			`SELECT EXISTS (SELECT 1 FROM referidos WHERE codigo = $1)`,
			codigo,
		).Scan(&existe)
		if err != nil {
			return "", err
		}
		if !existe {
			return codigo, nil
		}
	}
	return base, nil
}

// CrearUsuario da de alta una cuenta de cliente junto con su nodo en la red
// multinivel dentro de una única transacción:
//   - con código de referido válido, el nodo cuelga del patrocinador;
//   - sin código, cuelga de la raíz de la semilla para no romper el árbol.
//
// Devuelve ErrEmailYaExiste (409) o ErrCodigoInvalido (400) según el caso.
func CrearUsuario(ctx context.Context, req models.RegisterRequest, hashPassword string) (models.Usuario, error) {
	nombre := strings.TrimSpace(req.Nombre)
	email := strings.TrimSpace(req.Email)
	if nombre == "" || email == "" || hashPassword == "" {
		return models.Usuario{}, ErrRegistroInvalido
	}

	tx, err := db.Begin(ctx)
	if err != nil {
		return models.Usuario{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	// Resolvemos el patrocinador a partir del código de invitación.
	var parentID *int
	nivel := 0
	codigoRef := strings.ToUpper(strings.TrimSpace(req.Ref))
	if codigoRef != "" {
		var padreID int
		err := tx.QueryRow(ctx,
			`SELECT id FROM referidos WHERE codigo = $1`,
			codigoRef,
		).Scan(&padreID)
		if errors.Is(err, pgx.ErrNoRows) {
			return models.Usuario{}, ErrCodigoInvalido
		}
		if err != nil {
			return models.Usuario{}, err
		}
		parentID = &padreID
		err = tx.QueryRow(ctx,
			`SELECT nivel FROM referidos WHERE id = $1`,
			padreID,
		).Scan(&nivel)
		if err != nil {
			return models.Usuario{}, err
		}
		nivel++
	} else {
		// Sin código: buscamos la raíz sembrada (parent_id NULL) para no crear una segunda raíz.
		var raizID int
		err := tx.QueryRow(ctx,
			`SELECT id FROM referidos WHERE parent_id IS NULL ORDER BY id LIMIT 1`,
		).Scan(&raizID)
		if err == nil {
			parentID = &raizID
			nivel = 1
		} else if !errors.Is(err, pgx.ErrNoRows) {
			return models.Usuario{}, err
		}
	}

	// Creamos la cuenta con rol cliente y contraseña ya hasheada con bcrypt.
	var usuarioID int
	err = tx.QueryRow(ctx,
		`INSERT INTO usuarios (email, password, rol, nombre) VALUES ($1, $2, 'cliente', $3) RETURNING id`,
		email, hashPassword, nombre,
	).Scan(&usuarioID)
	if esViolacionDeUnicidad(err) {
		return models.Usuario{}, ErrEmailYaExiste
	}
	if err != nil {
		return models.Usuario{}, err
	}

	// Creamos el nodo de red que representará a la cuenta.
	codigo, err := generarCodigoPropuesto(ctx, tx, email)
	if err != nil {
		return models.Usuario{}, err
	}

	var referidoID int
	err = tx.QueryRow(ctx,
		`INSERT INTO referidos (nombre, nivel, ventas, parent_id, email, codigo)
		 VALUES ($1, $2, 0, $3, $4, $5) RETURNING id`,
		nombre, nivel, parentID, email, codigo,
	).Scan(&referidoID)
	if err != nil {
		return models.Usuario{}, err
	}

	// Enlazamos la cuenta con su nodo: es la FK que después usan las comisiones.
	_, err = tx.Exec(ctx,
		`UPDATE usuarios SET referido_id = $1 WHERE id = $2`,
		referidoID, usuarioID,
	)
	if err != nil {
		return models.Usuario{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return models.Usuario{}, err
	}

	return models.Usuario{
		ID:         usuarioID,
		Email:      email,
		Rol:        "cliente",
		Nombre:     nombre,
		Activo:     true,
		ReferidoID: &referidoID,
	}, nil
}

// ObtenerProductos devuelve el catálogo público: solo productos activos, ordenados por ID.
// Los inactivos (borrado lógico) quedan ocultos para el cliente pero siguen en pedido_items.
func ObtenerProductos() ([]models.Producto, error) {
	rows, err := db.Query(context.Background(),
		`SELECT id, nombre, descripcion, precio, categoria, img, galeria, stock
		 FROM productos WHERE activo = TRUE ORDER BY id`,
	)
	if err != nil {
		return nil, err
	}

	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (models.Producto, error) {
		var p models.Producto
		err := row.Scan(&p.ID, &p.Nombre, &p.Descripcion, &p.Precio, &p.Categoria, &p.Img, &p.Galeria, &p.Stock)
		return p, err
	})
}

// ObtenerProductoPorID devuelve el producto con el ID indicado incluyendo su stock.
func ObtenerProductoPorID(id int) (models.Producto, bool, error) {
	var p models.Producto

	err := db.QueryRow(context.Background(),
		`SELECT id, nombre, descripcion, precio, categoria, img, galeria, stock FROM productos WHERE id = $1`,
		id,
	).Scan(&p.ID, &p.Nombre, &p.Descripcion, &p.Precio, &p.Categoria, &p.Img, &p.Galeria, &p.Stock)

	if errors.Is(err, pgx.ErrNoRows) {
		return models.Producto{}, false, nil
	}
	if err != nil {
		return models.Producto{}, false, err
	}
	return p, true, nil
}

// ObtenerProductosAdmin devuelve el catálogo completo (activos e inactivos)
// para la vista administrativa, opcionalmente filtrado por un texto de búsqueda.
func ObtenerProductosAdmin(busqueda string) ([]models.Producto, error) {
	q := strings.TrimSpace(busqueda)
	filas := []models.Producto{}
	if q == "" {
		rows, err := db.Query(context.Background(),
			`SELECT id, nombre, descripcion, precio, categoria, img, galeria, stock, activo
			 FROM productos ORDER BY id`)
		if err != nil {
			return nil, err
		}
		filas, err = pgx.CollectRows(rows, escanearProductoAdmin)
		return filas, err
	}

	// ILIKE con comodín escapado: el usuario puede escribir % o _ sin romper el filtro.
	patron := "%" + strings.ReplaceAll(strings.ReplaceAll(q, "%", "\\%"), "_", "\\_") + "%"
	rows, err := db.Query(context.Background(),
		`SELECT id, nombre, descripcion, precio, categoria, img, galeria, stock, activo
		 FROM productos
		 WHERE nombre ILIKE $1 OR categoria ILIKE $1 OR descripcion ILIKE $1
		 ORDER BY id`, patron)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, escanearProductoAdmin)
}

// escanearProductoAdmin mapea una fila del catálogo administrativo (incluye activo).
func escanearProductoAdmin(row pgx.CollectableRow) (models.Producto, error) {
	var p models.Producto
	err := row.Scan(&p.ID, &p.Nombre, &p.Descripcion, &p.Precio, &p.Categoria, &p.Img, &p.Galeria, &p.Stock, &p.Activo)
	return p, err
}

// CrearProducto inserta un artículo del catálogo con borrado lógico activo por defecto.
func CrearProducto(ctx context.Context, p models.Producto) (models.Producto, error) {
	err := db.QueryRow(ctx,
		`INSERT INTO productos (nombre, descripcion, precio, categoria, img, galeria, stock, activo)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, TRUE) RETURNING id`,
		p.Nombre, p.Descripcion, p.Precio, p.Categoria, p.Img, p.Galeria, p.Stock,
	).Scan(&p.ID)
	if err != nil {
		return models.Producto{}, err
	}
	p.Activo = true
	return p, nil
}

// ActualizarProducto edita los campos del producto indicado.
// El booleano indica si existía un producto con ese ID.
func ActualizarProducto(ctx context.Context, id int, p models.Producto) (bool, error) {
	res, err := db.Exec(ctx,
		`UPDATE productos
		 SET nombre = $1, descripcion = $2, precio = $3, categoria = $4,
		     img = $5, galeria = $6, stock = $7
		 WHERE id = $8`,
		p.Nombre, p.Descripcion, p.Precio, p.Categoria, p.Img, p.Galeria, p.Stock, id,
	)
	if err != nil {
		return false, err
	}
	return res.RowsAffected() > 0, nil
}

// CambiarEstadoProducto aplica el borrado lógico (activo = false) o lo revierte.
// No borra la fila para conservar el historial de pedido_items.
func CambiarEstadoProducto(ctx context.Context, id int, activo bool) (bool, error) {
	res, err := db.Exec(ctx,
		`UPDATE productos SET activo = $1 WHERE id = $2`,
		activo, id,
	)
	if err != nil {
		return false, err
	}
	return res.RowsAffected() > 0, nil
}

// Errores de dominio para el proceso de checkout
var (
	ErrStockInsuficiente = errors.New("stock insuficiente")
	ErrCantidadInvalida  = errors.New("cantidad debe ser mayor que cero")
	ErrCarritoVacio      = errors.New("el carrito no contiene artículos")
)

// CrearPedido ejecuta una transacción ACID con FOR UPDATE para validar stock, descontarlo y registrar el pedido.
// CrearPedido ejecuta la compra transaccional: valida stock con FOR UPDATE,
// descuenta inventario y asocia el pedido a la identidad del token (usuarioID).
func CrearPedido(ctx context.Context, usuarioID *int, email string, req models.CheckoutRequest) (models.PedidoConfirmado, error) {
	if len(req.Items) == 0 {
		return models.PedidoConfirmado{}, ErrCarritoVacio
	}

	tx, err := db.Begin(ctx)
	if err != nil {
		return models.PedidoConfirmado{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	type itemCalculado struct {
		productoID     int
		nombre         string
		cantidad       int
		precioUnitario float64
		subtotal       float64
	}

	var itemsCalculados []itemCalculado
	var total float64

	// Validamos cada ítem bloqueando la fila correspondiente en productos con FOR UPDATE
	for _, it := range req.Items {
		if it.Cantidad <= 0 {
			return models.PedidoConfirmado{}, ErrCantidadInvalida
		}

		var p models.Producto
		err := tx.QueryRow(ctx,
			`SELECT id, nombre, precio, stock, activo FROM productos WHERE id = $1 FOR UPDATE`,
			it.ProductoID,
		).Scan(&p.ID, &p.Nombre, &p.Precio, &p.Stock, &p.Activo)

		if errors.Is(err, pgx.ErrNoRows) {
			return models.PedidoConfirmado{}, errors.New("producto no encontrado en el catálogo")
		}
		if err != nil {
			return models.PedidoConfirmado{}, err
		}

		if !p.Activo {
			return models.PedidoConfirmado{}, errors.New("Producto desactivado: " + p.Nombre)
		}

		if p.Stock < it.Cantidad {
			return models.PedidoConfirmado{}, errors.New("Stock insuficiente para: " + p.Nombre)
		}

		subtotal := p.Precio * float64(it.Cantidad)
		total += subtotal

		// Descontamos el stock
		_, err = tx.Exec(ctx,
			`UPDATE productos SET stock = stock - $1 WHERE id = $2`,
			it.Cantidad, p.ID,
		)
		if err != nil {
			return models.PedidoConfirmado{}, err
		}

		itemsCalculados = append(itemsCalculados, itemCalculado{
			productoID:     p.ID,
			nombre:         p.Nombre,
			cantidad:       it.Cantidad,
			precioUnitario: p.Precio,
			subtotal:       subtotal,
		})
	}

	numeroPedido := "MC-" + time.Now().Format("060102150405")

	// Insertamos la cabecera del pedido asegurando que la identidad cliente_email
	// provenga del token verificado y que usuario_id apunte al dueño real.
	var pedidoID int
	err = tx.QueryRow(ctx,
		`INSERT INTO pedidos (numero, cliente_nombre, cliente_email, cliente_direccion, cliente_ciudad, cliente_telefono, metodo_pago, total, usuario_id, estado)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, 'pendiente')
		 RETURNING id`,
		numeroPedido, req.Nombre, email, req.Direccion, req.Ciudad, req.Telefono, req.MetodoPago, total, usuarioID,
	).Scan(&pedidoID)
	if err != nil {
		return models.PedidoConfirmado{}, err
	}

	// Insertamos los ítems en pedido_items
	for _, it := range itemsCalculados {
		_, err = tx.Exec(ctx,
			`INSERT INTO pedido_items (pedido_id, producto_id, nombre, cantidad, precio_unitario, subtotal)
			 VALUES ($1, $2, $3, $4, $5, $6)`,
			pedidoID, it.productoID, it.nombre, it.cantidad, it.precioUnitario, it.subtotal,
		)
		if err != nil {
			return models.PedidoConfirmado{}, err
		}
	}

	// El ledger de comisiones nace junto con el pedido, dentro de la MISMA
	// transacción: o se registran ambas cosas o no se registra ninguna.
	if err := repartirComisionesEnTx(ctx, tx, email, pedidoID, total); err != nil {
		return models.PedidoConfirmado{}, err
	}

	// Confirmamos la transacción
	if err := tx.Commit(ctx); err != nil {
		return models.PedidoConfirmado{}, err
	}

	return models.PedidoConfirmado{
		PedidoID: pedidoID,
		Numero:   numeroPedido,
		Total:    total,
		Message:  "Pedido registrado y stock reservado con éxito",
	}, nil
}

// ActualizarComisionesUpline suma las ventas al upline multinivel (hasta 3 niveles) en la tabla referidos.
// El punto de partida es la FK usuarios.referido_id, no una correspondencia de IDs por email.
func ActualizarComisionesUpline(ctx context.Context, email string, monto float64) error {
	var nodoID *int
	err := db.QueryRow(ctx,
		`SELECT referido_id FROM usuarios WHERE email = $1`,
		email,
	).Scan(&nodoID)
	if err != nil || nodoID == nil {
		return nil
	}

	currentID := nodoID
	for nivel := 1; nivel <= 3 && currentID != nil; nivel++ {
		var parentID *int
		err := db.QueryRow(ctx,
			`SELECT parent_id FROM referidos WHERE id = $1`,
			*currentID,
		).Scan(&parentID)
		if err != nil || parentID == nil {
			break
		}

		// Acreditamos venta al patrocinador
		_, _ = db.Exec(ctx,
			`UPDATE referidos SET ventas = ventas + $1 WHERE id = $2`,
			monto, *parentID,
		)

		currentID = parentID
	}

	return nil
}

// referidoPlano representa una fila de la tabla referidos antes de armar el árbol.
type referidoPlano struct {
	referido models.Referido
	parentID *int
}

// ObtenerRed devuelve el árbol completo de la red multinivel.
// Se consulta la tabla como una lista plana y el árbol se reconstruye en memoria usando parent_id.
func ObtenerRed() (models.Referido, error) {
	filas, err := cargarRedPlana(context.Background())
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
