// Este archivo agrupa las liquidaciones a distribuidores (Fase 5).
// Una liquidación es un pago real: cierra comisiones pendientes y las marca
// como pagadas dentro de una ÚNICA transacción, con bloqueo de las filas.
package repository

// Importamos context para el contexto de la petición HTTP.
import (
	"context"
	"math"

	// Importamos pgx para transacciones y errores de fila.
	"github.com/jackc/pgx/v5"

	// Importamos nuestros modelos para Liquidacion, DeudaComisiones y Pagina.
	"multicatalogo-backend/models"
)

// CrearLiquidacion liquida TODO el saldo pendiente de un distribuidor.
// Dentro de la transacción: bloquea las filas con FOR UPDATE, crea la
// liquidación con el total y marca esas filas como pagadas. Así un mismo
// saldo no puede liquidarse dos veces aunque lleguen peticiones a la vez.
func (pg *PostgresRepository) CrearLiquidacion(ctx context.Context, referidoID int) (models.Liquidacion, error) {
	tx, err := pg.db.Begin(ctx)
	if err != nil {
		return models.Liquidacion{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	// Bloqueamos las filas pendientes para serializar liquidaciones simultáneas.
	rows, err := tx.Query(ctx,
		`SELECT id, monto FROM comisiones
		 WHERE referido_id = $1 AND estado = 'pendiente'
		 FOR UPDATE`,
		referidoID,
	)
	if err != nil {
		return models.Liquidacion{}, err
	}

	var total float64
	var cantidad int
	for rows.Next() {
		var id int
		var monto float64
		if err := rows.Scan(&id, &monto); err != nil {
			rows.Close()
			return models.Liquidacion{}, err
		}
		total += monto
		cantidad++
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return models.Liquidacion{}, err
	}

	// Nada que pagar: devolvemos un error de dominio traducible a HTTP 409.
	if cantidad == 0 {
		return models.Liquidacion{}, models.ErrSinComisionesPendientes
	}

	// Redondeamos a dos decimales antes de persistir el total del pago.
	total = aPlata(total)

	var liquidacion models.Liquidacion
	err = tx.QueryRow(ctx,
		`INSERT INTO liquidaciones (referido_id, total, estado, pagada_en)
		 VALUES ($1, $2, 'pagada', NOW())
		 RETURNING id, referido_id, total, estado, creado_en, pagada_en, periodo`,
		referidoID, total,
	).Scan(&liquidacion.ID, &liquidacion.ReferidoID, &liquidacion.Total,
		&liquidacion.Estado, &liquidacion.CreadoEn, &liquidacion.PagadaEn, &liquidacion.Periodo)
	if err != nil {
		return models.Liquidacion{}, err
	}

	// Cerramos exactamente las filas que acabamos de pagar.
	if _, err := tx.Exec(ctx,
		`UPDATE comisiones
		 SET estado = 'pagada', liquidacion_id = $1
		 WHERE referido_id = $2 AND estado = 'pendiente'`,
		liquidacion.ID, referidoID,
	); err != nil {
		return models.Liquidacion{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return models.Liquidacion{}, err
	}

	liquidacion.CantidadComisiones = cantidad
	return liquidacion, nil
}

// aPlata redondea un importe a dos decimales (NUMERIC(10,2)).
func aPlata(valor float64) float64 {
	return math.Round(valor*100) / 100
}

// liquidacionesSelect trae el histórico con el nombre del distribuidor y la
// cantidad de comisiones que cubrió cada pago.
const liquidacionesSelect = `SELECT l.id, l.referido_id, r.nombre, l.total, l.estado,
	l.creado_en, l.pagada_en, l.periodo,
	(SELECT COUNT(*) FROM comisiones c WHERE c.liquidacion_id = l.id)
	FROM liquidaciones l
	JOIN referidos r ON r.id = l.referido_id`

// ObtenerLiquidaciones devuelve el histórico global de pagos (solo admin).
func (pg *PostgresRepository) ObtenerLiquidaciones(ctx context.Context, page, limit int) (models.Pagina, error) {
	pagina, limite, offset := normalizarPagina(itoa(page), itoa(limit))

	var total int
	if err := pg.db.QueryRow(ctx, `SELECT COUNT(*) FROM liquidaciones`).Scan(&total); err != nil {
		return models.Pagina{}, err
	}

	rows, err := pg.db.Query(ctx,
		liquidacionesSelect+` ORDER BY l.creado_en DESC, l.id DESC LIMIT $1 OFFSET $2`,
		limite, offset,
	)
	if err != nil {
		return models.Pagina{}, err
	}
	defer rows.Close()

	items, err := coleccionLiquidaciones(rows)
	if err != nil {
		return models.Pagina{}, err
	}
	return models.Pagina{Data: items, Page: pagina, Limit: limite, Total: total}, nil
}

// ObtenerLiquidacionesMias devuelve solo los pagos recibidos por el distribuidor.
func (pg *PostgresRepository) ObtenerLiquidacionesMias(ctx context.Context, referidoID, page, limit int) (models.Pagina, error) {
	pagina, limite, offset := normalizarPagina(itoa(page), itoa(limit))

	var total int
	if err := pg.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM liquidaciones WHERE referido_id = $1`,
		referidoID,
	).Scan(&total); err != nil {
		return models.Pagina{}, err
	}

	rows, err := pg.db.Query(ctx,
		liquidacionesSelect+` WHERE l.referido_id = $1 ORDER BY l.creado_en DESC, l.id DESC LIMIT $2 OFFSET $3`,
		referidoID, limite, offset,
	)
	if err != nil {
		return models.Pagina{}, err
	}
	defer rows.Close()

	items, err := coleccionLiquidaciones(rows)
	if err != nil {
		return models.Pagina{}, err
	}
	return models.Pagina{Data: items, Page: pagina, Limit: limite, Total: total}, nil
}

// coleccionLiquidaciones materializa las filas de una consulta de pagos.
func coleccionLiquidaciones(rows pgx.Rows) ([]models.Liquidacion, error) {
	items := []models.Liquidacion{}
	for rows.Next() {
		var l models.Liquidacion
		if err := rows.Scan(&l.ID, &l.ReferidoID, &l.Nombre, &l.Total, &l.Estado,
			&l.CreadoEn, &l.PagadaEn, &l.Periodo, &l.CantidadComisiones); err != nil {
			return nil, err
		}
		items = append(items, l)
	}
	return items, rows.Err()
}

// ObtenerDeudas resume cuánto debe cada distribuidor (comisiones pendientes).
// Es la cola de trabajo del back-office: solo aparece quien tiene saldo vivo.
func (pg *PostgresRepository) ObtenerDeudas(ctx context.Context) ([]models.DeudaComisiones, error) {
	rows, err := pg.db.Query(ctx,
		`SELECT r.id, r.nombre, COALESCE(r.email, ''),
		        COALESCE(SUM(c.monto) FILTER (WHERE c.estado = 'pendiente'), 0),
		        COUNT(c.id) FILTER (WHERE c.estado = 'pendiente')
		 FROM referidos r
		 JOIN comisiones c ON c.referido_id = r.id
		 GROUP BY r.id, r.nombre, r.email
		 HAVING COALESCE(SUM(c.monto) FILTER (WHERE c.estado = 'pendiente'), 0) > 0
		 ORDER BY 4 DESC, r.id`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	deudas := []models.DeudaComisiones{}
	for rows.Next() {
		var d models.DeudaComisiones
		if err := rows.Scan(&d.ReferidoID, &d.Nombre, &d.Email, &d.Pendiente, &d.Cantidad); err != nil {
			return nil, err
		}
		deudas = append(deudas, d)
	}
	return deudas, rows.Err()
}
