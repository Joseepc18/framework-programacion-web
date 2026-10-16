// Este archivo agrupa la consulta y el ciclo de vida de los pedidos (Fase 3).
// Todo acceso a pedidos pasa por PostgreSQL y respeta la paginación estándar.
package repository

// Importamos context para recibir el contexto de la petición HTTP.
import (
	"context"
	"errors"
	"strconv"
	"strings"

	// Importamos pgx para leer filas con sus tipos nativos de PostgreSQL.
	"github.com/jackc/pgx/v5"

	// Importamos nuestros modelos para Pedido, PedidoDetalle y Pagina.
	"multicatalogo-backend/models"
)

// Errores de dominio reconocidos por los controladores para traducirlos a HTTP.
var (
	// ErrPedidoNoIndica dueño cuando el pedido no existe o no pertenece al token.
	ErrPedidoNoEncontrado = errors.New("pedido no encontrado")
	// ErrTransicionInvalida indica un salto de estado fuera del ciclo de vida.
	ErrTransicionInvalida = errors.New("transición de estado inválida")
)

// transicionesValidas describe el ciclo de vida permitido del pedido:
// pendiente → preparando → enviado → entregado, y cancelado desde etapas abiertas.
var transicionesValidas = map[string][]string{
	"pendiente":  {"preparando", "cancelado"},
	"preparando": {"enviado", "cancelado"},
	"enviado":    {"entregado", "cancelado"},
	"entregado":  {},
	"cancelado":  {},
}

// normalizarPagina convierte los parámetros ?page y ?limit a valores seguros.
// Devuelve la página (>=1), el límite (entre 1 y 100) y el desplazamiento SQL.
func normalizarPagina(pageStr, limitStr string) (int, int, int) {
	pagina, err := strconv.Atoi(strings.TrimSpace(pageStr))
	if err != nil || pagina < 1 {
		pagina = 1
	}

	limite, err := strconv.Atoi(strings.TrimSpace(limitStr))
	if err != nil || limite < 1 {
		limite = 10
	}
	if limite > 100 {
		limite = 100
	}

	return pagina, limite, (pagina - 1) * limite
}

// itoa convierte un entero a texto para reutilizar normalizarPagina.
func itoa(n int) string {
	return strconv.Itoa(n)
}

// seleccionPedidos define las columnas comunes a cabecera de pedido.
const seleccionPedidos = `id, numero, cliente_nombre, cliente_email, cliente_direccion,
	cliente_ciudad, cliente_telefono, metodo_pago, total, usuario_id, estado, creado_en`

// filaPedido es el mínimo común entre una fila suelta y un registro de un rows.
type filaPedido interface {
	Scan(dest ...any) error
}

// escanearPedido lee una cabecera de pedido desde cualquier fila consultable.
func escanearPedido(row filaPedido) (models.Pedido, error) {
	var p models.Pedido
	err := row.Scan(
		&p.ID, &p.Numero, &p.ClienteNombre, &p.ClienteEmail, &p.ClienteDireccion,
		&p.ClienteCiudad, &p.ClienteTelefono, &p.MetodoPago, &p.Total,
		&p.UsuarioID, &p.Estado, &p.CreadoEn,
	)
	return p, err
}

// ObtenerPedidosMios devuelve el historial de pedidos del usuario del token,
// paginado y ordenado del más reciente al más antiguo.
func ObtenerPedidosMios(ctx context.Context, usuarioID, page, limit int) (models.Pagina, error) {
	pagina, limite, offset := normalizarPagina(strconv.Itoa(page), strconv.Itoa(limit))

	var total int
	err := db.QueryRow(ctx,
		`SELECT COUNT(*) FROM pedidos WHERE usuario_id = $1`,
		usuarioID,
	).Scan(&total)
	if err != nil {
		return models.Pagina{}, err
	}

	rows, err := db.Query(ctx,
		`SELECT `+seleccionPedidos+`
		 FROM pedidos
		 WHERE usuario_id = $1
		 ORDER BY creado_en DESC, id DESC
		 LIMIT $2 OFFSET $3`,
		usuarioID, limite, offset,
	)
	if err != nil {
		return models.Pagina{}, err
	}
	defer rows.Close()

	pedidos := []models.Pedido{}
	for rows.Next() {
		p, err := escanearPedido(rows)
		if err != nil {
			return models.Pagina{}, err
		}
		pedidos = append(pedidos, p)
	}
	if err := rows.Err(); err != nil {
		return models.Pagina{}, err
	}

	return models.Pagina{Data: pedidos, Page: pagina, Limit: limite, Total: total}, nil
}

// ObtenerPedidosAdmin lista todos los pedidos del back-office con filtro
// opcional por estado (?estado=) y paginación estándar.
func ObtenerPedidosAdmin(ctx context.Context, estado string, page, limit int) (models.Pagina, error) {
	pagina, limite, offset := normalizarPagina(strconv.Itoa(page), strconv.Itoa(limit))
	filtro := strings.TrimSpace(estado)

	var total int
	err := db.QueryRow(ctx,
		`SELECT COUNT(*) FROM pedidos WHERE ($1 = '' OR estado = $1)`,
		filtro,
	).Scan(&total)
	if err != nil {
		return models.Pagina{}, err
	}

	rows, err := db.Query(ctx,
		`SELECT `+seleccionPedidos+`
		 FROM pedidos
		 WHERE ($1 = '' OR estado = $1)
		 ORDER BY creado_en DESC, id DESC
		 LIMIT $2 OFFSET $3`,
		filtro, limite, offset,
	)
	if err != nil {
		return models.Pagina{}, err
	}
	defer rows.Close()

	pedidos := []models.Pedido{}
	for rows.Next() {
		p, err := escanearPedido(rows)
		if err != nil {
			return models.Pagina{}, err
		}
		pedidos = append(pedidos, p)
	}
	if err := rows.Err(); err != nil {
		return models.Pagina{}, err
	}

	return models.Pagina{Data: pedidos, Page: pagina, Limit: limite, Total: total}, nil
}

// ObtenerPedidoDetalle devuelve la cabecera y las líneas de un pedido.
// El parámetro esAdmin permite a un administrador consultar cualquier pedido;
// un cliente solo recibe los suyos (dueño por usuario_id o por su email).
func ObtenerPedidoDetalle(ctx context.Context, id int, usuarioID int, email string, esAdmin bool) (models.PedidoDetalle, error) {
	p, err := escanearPedido(db.QueryRow(ctx,
		`SELECT `+seleccionPedidos+` FROM pedidos WHERE id = $1`,
		id,
	))
	if errors.Is(err, pgx.ErrNoRows) {
		return models.PedidoDetalle{}, ErrPedidoNoEncontrado
	}
	if err != nil {
		return models.PedidoDetalle{}, err
	}

	if !esAdmin {
		esDueno := p.UsuarioID != nil && *p.UsuarioID == usuarioID
		if !esDueno && !strings.EqualFold(p.ClienteEmail, email) {
			// Devolvemos el mismo error que para un ID inexistente para no
			// filtrar la existencia de pedidos ajenos.
			return models.PedidoDetalle{}, ErrPedidoNoEncontrado
		}
	}

	items, err := obtenerItemsPedido(ctx, id)
	if err != nil {
		return models.PedidoDetalle{}, err
	}

	return models.PedidoDetalle{Pedido: p, Items: items}, nil
}

// obtenerItemsPedido carga las líneas del pedido con el precio vigente al comprar.
func obtenerItemsPedido(ctx context.Context, pedidoID int) ([]models.ItemPedidoDetalle, error) {
	rows, err := db.Query(ctx,
		`SELECT producto_id, nombre, cantidad, precio_unitario, subtotal
		 FROM pedido_items
		 WHERE pedido_id = $1
		 ORDER BY id`,
		pedidoID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []models.ItemPedidoDetalle{}
	for rows.Next() {
		var it models.ItemPedidoDetalle
		if err := rows.Scan(&it.ProductoID, &it.Nombre, &it.Cantidad, &it.PrecioUnitario, &it.Subtotal); err != nil {
			return nil, err
		}
		items = append(items, it)
	}
	return items, rows.Err()
}

// CambiarEstadoPedido aplica una transición del ciclo de vida dentro de una
// transacción ACID. La cancelación es el único caso especial: devuelve el
// stock de cada ítem y revierte las comisiones que el pedido había generado.
func CambiarEstadoPedido(ctx context.Context, id int, nuevoEstado string) (models.Pedido, error) {
	nuevoEstado = strings.TrimSpace(strings.ToLower(nuevoEstado))

	tx, err := db.Begin(ctx)
	if err != nil {
		return models.Pedido{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	// Bloqueamos la fila para que dos administradores no transicionen a la vez.
	p, err := escanearPedido(tx.QueryRow(ctx,
		`SELECT `+seleccionPedidos+` FROM pedidos WHERE id = $1 FOR UPDATE`,
		id,
	))
	if errors.Is(err, pgx.ErrNoRows) {
		return models.Pedido{}, ErrPedidoNoEncontrado
	}
	if err != nil {
		return models.Pedido{}, err
	}

	// Mismo estado = sin cambios; lo tratamos como éxito idempotente.
	if nuevoEstado == p.Estado {
		return p, nil
	}

	permitidas, esValida := transicionesValidas[p.Estado]
	transicionPermitida := false
	for _, e := range permitidas {
		if e == nuevoEstado {
			transicionPermitida = true
			break
		}
	}
	if !esValida || !transicionPermitida {
		return models.Pedido{}, ErrTransicionInvalida
	}

	if _, err := tx.Exec(ctx,
		`UPDATE pedidos SET estado = $1 WHERE id = $2`,
		nuevoEstado, id,
	); err != nil {
		return models.Pedido{}, err
	}

	// Espejo ACID de la cancelación: stock de vuelta y comisiones revertidas.
	if nuevoEstado == "cancelado" {
		if _, err := tx.Exec(ctx,
			`UPDATE productos p
			 SET stock = stock + i.cantidad
			 FROM pedido_items i
			 WHERE i.pedido_id = $1 AND p.id = i.producto_id`,
			id,
		); err != nil {
			return models.Pedido{}, err
		}

		if _, err := tx.Exec(ctx,
			`UPDATE comisiones SET estado = 'revertida'
			 WHERE pedido_id = $1 AND estado IN ('pendiente', 'aprobada')`,
			id,
		); err != nil {
			return models.Pedido{}, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return models.Pedido{}, err
	}

	p.Estado = nuevoEstado
	return p, nil
}
