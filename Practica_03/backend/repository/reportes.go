// Este archivo agrupa los reportes de negocio (Fase 6): agregaciones SQL
// (SUM/COUNT/GROUP BY) sobre pedidos y el ledger de comisiones.
// Los pedidos cancelados y las comisiones revertidas nunca cuentan como venta.
package repository

import (
	"context"

	"multicatalogo-backend/models"
)

// ObtenerResumenVentas calcula total vendido, nº de pedidos, ticket promedio y comisiones devengadas.
func (pg *PostgresRepository) ObtenerResumenVentas(ctx context.Context, r models.RangoFechas) (models.ResumenVentas, error) {
	var res models.ResumenVentas
	err := pg.db.QueryRow(ctx,
		`SELECT COALESCE(SUM(total), 0), COUNT(*)
		 FROM pedidos
		 WHERE estado <> 'cancelado'
		   AND ($1::timestamptz IS NULL OR creado_en >= $1)
		   AND ($2::timestamptz IS NULL OR creado_en <  $2)`,
		r.Desde, r.Hasta,
	).Scan(&res.TotalVendido, &res.NumPedidos)
	if err != nil {
		return res, err
	}

	err = pg.db.QueryRow(ctx,
		`SELECT COALESCE(SUM(c.monto), 0)
		 FROM comisiones c
		 JOIN pedidos p ON p.id = c.pedido_id
		 WHERE c.estado <> 'revertida'
		   AND ($1::timestamptz IS NULL OR p.creado_en >= $1)
		   AND ($2::timestamptz IS NULL OR p.creado_en <  $2)`,
		r.Desde, r.Hasta,
	).Scan(&res.ComisionesDevengadas)
	if err != nil {
		return res, err
	}

	res.TotalVendido = aPlata(res.TotalVendido)
	res.ComisionesDevengadas = aPlata(res.ComisionesDevengadas)
	if res.NumPedidos > 0 {
		res.TicketPromedio = aPlata(res.TotalVendido / float64(res.NumPedidos))
	}
	return res, nil
}

// ObtenerTopProductos ordena los productos por ingresos (SUM cantidad × precio).
func (pg *PostgresRepository) ObtenerTopProductos(ctx context.Context, r models.RangoFechas, limite int) ([]models.TopProducto, error) {
	rows, err := pg.db.Query(ctx,
		`SELECT i.producto_id, MAX(i.nombre), SUM(i.cantidad), SUM(i.cantidad * i.precio_unitario)
		 FROM pedido_items i
		 JOIN pedidos p ON p.id = i.pedido_id
		 WHERE p.estado <> 'cancelado'
		   AND ($1::timestamptz IS NULL OR p.creado_en >= $1)
		   AND ($2::timestamptz IS NULL OR p.creado_en <  $2)
		 GROUP BY i.producto_id
		 ORDER BY 4 DESC, 1
		 LIMIT $3`,
		r.Desde, r.Hasta, limite,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	top := []models.TopProducto{}
	for rows.Next() {
		var t models.TopProducto
		if err := rows.Scan(&t.ProductoID, &t.Nombre, &t.Unidades, &t.Ingresos); err != nil {
			return nil, err
		}
		top = append(top, t)
	}
	return top, rows.Err()
}

// ObtenerTopReferidos ordena a los distribuidores por las ventas que generó su red.
// Cada fila del ledger une a un cobrador con un pedido de su red (una por nivel),
// así que sumar p.total por cobrador da las ventas de su red sin duplicados.
func (pg *PostgresRepository) ObtenerTopReferidos(ctx context.Context, r models.RangoFechas, limite int) ([]models.TopReferido, error) {
	rows, err := pg.db.Query(ctx,
		`SELECT rf.id, rf.nombre, SUM(p.total), SUM(c.monto)
		 FROM comisiones c
		 JOIN pedidos p    ON p.id = c.pedido_id
		 JOIN referidos rf ON rf.id = c.referido_id
		 WHERE c.estado <> 'revertida'
		   AND ($1::timestamptz IS NULL OR p.creado_en >= $1)
		   AND ($2::timestamptz IS NULL OR p.creado_en <  $2)
		 GROUP BY rf.id, rf.nombre
		 ORDER BY 3 DESC, 1
		 LIMIT $3`,
		r.Desde, r.Hasta, limite,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	top := []models.TopReferido{}
	for rows.Next() {
		var t models.TopReferido
		if err := rows.Scan(&t.ReferidoID, &t.Nombre, &t.VentasRed, &t.Comisiones); err != nil {
			return nil, err
		}
		top = append(top, t)
	}
	return top, rows.Err()
}
