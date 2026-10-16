// Este archivo agrupa el ledger de comisiones y la red propia (Fase 4).
// El reparto se escribe dentro de la MISMA transacción del pedido: si el
// checkout falla, tampoco existen comisiones huérfanas.
package repository

// Importamos context para el contexto de la petición HTTP.
import (
	"context"
	"errors"

	// Importamos pgx para leer filas y transacciones de PostgreSQL.
	"github.com/jackc/pgx/v5"

	// Importamos domain para el reparto puro de comisiones (sin I/O).
	"multicatalogo-backend/domain"
	// Importamos nuestros modelos para Comision, PaginaComisiones y Referido.
	"multicatalogo-backend/models"
)

// repartirComisionesEnTx inserta el ledger del pedido recién creado.
// Recorre el upline desde usuarios.referido_id (la FK real del comprador)
// hasta 3 niveles y escribe una fila de comisiones por cada nivel.
// No tiene efecto si el comprador no pertenece a ninguna red.
func repartirComisionesEnTx(ctx context.Context, tx pgx.Tx, email string, pedidoID int, total float64) error {
	var nodoID *int
	if err := tx.QueryRow(ctx,
		`SELECT referido_id FROM usuarios WHERE email = $1`,
		email,
	).Scan(&nodoID); err != nil || nodoID == nil {
		// Sin nodo de red no hay upline que comisionar.
		return nil
	}

	comisiones := domain.RepartirComisiones(total, domain.TasasUpline)
	if len(comisiones) == 0 {
		return nil
	}

	actual := nodoID
	for _, cm := range comisiones {
		if actual == nil {
			break
		}

		// El patrocinador de cada nivel es el padre del nodo actual.
		var parentID *int
		if err := tx.QueryRow(ctx,
			`SELECT parent_id FROM referidos WHERE id = $1`,
			*actual,
		).Scan(&parentID); err != nil || parentID == nil {
			break
		}

		if _, err := tx.Exec(ctx,
			`INSERT INTO comisiones (pedido_id, referido_id, nivel, porcentaje, monto, estado)
			 VALUES ($1, $2, $3, $4, $5, 'pendiente')`,
			pedidoID, *parentID, cm.Nivel, cm.Porcentaje, cm.Monto,
		); err != nil {
			return err
		}

		// Contador de dinero del nodo: debe coincidir siempre con el ledger.
		if _, err := tx.Exec(ctx,
			`UPDATE referidos SET comisiones = comisiones + $1 WHERE id = $2`,
			cm.Monto, *parentID,
		); err != nil {
			return err
		}

		actual = parentID
	}

	return nil
}

// ComisionesMias pagina el ledger del usuario y agrega los totales por estado.
func ObtenerComisionesMias(ctx context.Context, referidoID, page, limit int) (models.PaginaComisiones, error) {
	pagina, limite, offset := normalizarPagina(itoa(page), itoa(limit))

	var total int
	if err := db.QueryRow(ctx,
		`SELECT COUNT(*) FROM comisiones WHERE referido_id = $1`,
		referidoID,
	).Scan(&total); err != nil {
		return models.PaginaComisiones{}, err
	}

	rows, err := db.Query(ctx,
		`SELECT c.id, c.pedido_id, p.numero, c.nivel, c.porcentaje, c.monto, c.estado, c.creado_en
		 FROM comisiones c
		 JOIN pedidos p ON p.id = c.pedido_id
		 WHERE c.referido_id = $1
		 ORDER BY c.creado_en DESC, c.id DESC
		 LIMIT $2 OFFSET $3`,
		referidoID, limite, offset,
	)
	if err != nil {
		return models.PaginaComisiones{}, err
	}
	defer rows.Close()

	comisiones := []models.Comision{}
	for rows.Next() {
		var cm models.Comision
		if err := rows.Scan(&cm.ID, &cm.PedidoID, &cm.PedidoNumero, &cm.Nivel,
			&cm.Porcentaje, &cm.Monto, &cm.Estado, &cm.CreadoEn); err != nil {
			return models.PaginaComisiones{}, err
		}
		comisiones = append(comisiones, cm)
	}
	if err := rows.Err(); err != nil {
		return models.PaginaComisiones{}, err
	}

	// Totales globales (no solo de la página): lo que se debe y lo ya pagado.
	var pendiente, pagado float64
	if err := db.QueryRow(ctx,
		`SELECT COALESCE(SUM(monto) FILTER (WHERE estado = 'pendiente'), 0),
		        COALESCE(SUM(monto) FILTER (WHERE estado = 'pagada'), 0)
		 FROM comisiones
		 WHERE referido_id = $1`,
		referidoID,
	).Scan(&pendiente, &pagado); err != nil {
		return models.PaginaComisiones{}, err
	}

	return models.PaginaComisiones{
		Data:      comisiones,
		Page:      pagina,
		Limit:     limite,
		Total:     total,
		Pendiente: pendiente,
		Pagado:    pagado,
	}, nil
}

// ObtenerRedPropia devuelve el subárbol de red que cuelga del nodo indicado,
// junto con el código de invitación de ese nodo (para el link /registro?ref=).
// A diferencia de ObtenerRed, NO devuelve el árbol global: cada usuario ve lo suyo.
func ObtenerRedPropia(ctx context.Context, nodoID int) (models.RedPropia, error) {
	filas, err := cargarRedPlana(ctx)
	if err != nil {
		return models.RedPropia{}, err
	}

	// Agrupamos los hijos de cada nodo (la raíz global no interesa aquí).
	hijosDe := map[int][]int{}
	nodos := map[int]models.Referido{}
	for _, f := range filas {
		nodos[f.referido.ID] = f.referido
		if f.parentID != nil {
			hijosDe[*f.parentID] = append(hijosDe[*f.parentID], f.referido.ID)
		}
	}

	// El usuario sin nodo de red recibe una respuesta vacía, no un error.
	if _, existe := nodos[nodoID]; !existe {
		return models.RedPropia{Codigo: ""}, nil
	}

	// El código de invitación del nodo raíz arma el link de referido.
	var codigo string
	if err := db.QueryRow(ctx,
		`SELECT COALESCE(codigo, '') FROM referidos WHERE id = $1`,
		nodoID,
	).Scan(&codigo); err != nil {
		return models.RedPropia{}, err
	}

	arbol := construirArbol(nodoID, nodos, hijosDe)
	return models.RedPropia{Red: &arbol, Codigo: codigo, ReferidoID: nodoID}, nil
}

// cargarRedPlana lee toda la tabla referidos como filas sueltas.
func cargarRedPlana(ctx context.Context) ([]referidoPlano, error) {
	rows, err := db.Query(ctx,
		`SELECT id, nombre, nivel, ventas, parent_id FROM referidos ORDER BY id`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	filas, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (referidoPlano, error) {
		var f referidoPlano
		err := row.Scan(&f.referido.ID, &f.referido.Nombre, &f.referido.Nivel, &f.referido.Ventas, &f.parentID)
		return f, err
	})
	if err != nil {
		return nil, err
	}
	return filas, nil
}

// NodoDeRed devuelve el ID del nodo de la red que representa a un usuario.
// El segundo valor es false si la cuenta no tiene nodo asociado.
func NodoDeRed(ctx context.Context, usuarioID int) (int, bool, error) {
	var nodoID *int
	err := db.QueryRow(ctx,
		`SELECT referido_id FROM usuarios WHERE id = $1`,
		usuarioID,
	).Scan(&nodoID)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, false, nil
	}
	if err != nil {
		return 0, false, err
	}
	if nodoID == nil {
		return 0, false, nil
	}
	return *nodoID, true, nil
}
