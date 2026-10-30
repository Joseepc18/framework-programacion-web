// Este archivo agrupa el acceso a datos de cobros, pagos y comprobantes (F1-5).
package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"sigo-dental-backend/models"
)

// limiteCobros acota las filas que devuelve el listado de cobros.
const limiteCobros = 200

// pagadoDe es la subconsulta con lo pagado de un cobro; el estado de pago nunca se guarda, se deriva de los pagos.
const pagadoDe = `COALESCE((SELECT SUM(pg.monto_centavos) FROM pagos pg WHERE pg.cobro_id = c.id), 0)::int`

const columnasResumenCobro = `c.id, c.numero, c.paciente_id, p.nombres || ' ' || p.apellidos, c.total_centavos, ` + pagadoDe + `,
	EXISTS (SELECT 1 FROM cobro_anulaciones a WHERE a.cobro_id = c.id), c.creado_en`

func escanearResumenCobro(row pgx.Row, r *models.CobroResumen) error {
	if err := row.Scan(&r.ID, &r.Numero, &r.PacienteID, &r.PacienteNombre, &r.TotalCentavos, &r.PagadoCentavos, &r.Anulado, &r.CreadoEn); err != nil {
		return err
	}
	r.SaldoCentavos = r.TotalCentavos - r.PagadoCentavos
	r.Estado = models.EstadoDePago(r.TotalCentavos, r.PagadoCentavos)
	return nil
}

// ListarCobros busca por número de comprobante, nombre o documento del paciente (vacío = todos), del más reciente al más antiguo.
func (pg *PostgresRepository) ListarCobros(ctx context.Context, busqueda string, pacienteID int) ([]models.CobroResumen, error) {
	patron := "%" + escaparLike.Replace(strings.TrimSpace(busqueda)) + "%"
	rows, err := pg.db.Query(ctx,
		`SELECT `+columnasResumenCobro+`
		   FROM cobros c JOIN pacientes p ON p.id = c.paciente_id
		  WHERE ($2 = 0 OR c.paciente_id = $2)
		    AND (c.numero ILIKE $1 OR p.documento ILIKE $1 OR (p.nombres || ' ' || p.apellidos) ILIKE $1
		         OR (p.apellidos || ' ' || p.nombres) ILIKE $1)
		  ORDER BY c.id DESC
		  LIMIT $3`, patron, pacienteID, limiteCobros)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	lista := []models.CobroResumen{}
	for rows.Next() {
		var r models.CobroResumen
		if err := escanearResumenCobro(rows, &r); err != nil {
			return nil, err
		}
		lista = append(lista, r)
	}
	return lista, rows.Err()
}

// ObtenerCobro devuelve el cobro con sus líneas, pagos y anulación; ErrNoEncontrado si no existe.
func (pg *PostgresRepository) ObtenerCobro(ctx context.Context, id int) (models.Cobro, error) {
	var cobro models.Cobro
	err := pg.db.QueryRow(ctx,
		`SELECT `+columnasResumenCobro+`, c.cita_id, p.documento, c.subtotal_centavos, c.descuento_bp, c.descuento_centavos,
		        COALESCE((SELECT valor FROM parametros WHERE clave = 'moneda'), 'USD'), u.nombre
		   FROM cobros c
		   JOIN pacientes p ON p.id = c.paciente_id
		   JOIN usuarios u ON u.id = c.creado_por
		  WHERE c.id = $1`, id,
	).Scan(&cobro.ID, &cobro.Numero, &cobro.PacienteID, &cobro.PacienteNombre, &cobro.TotalCentavos, &cobro.PagadoCentavos,
		&cobro.Anulado, &cobro.CreadoEn, &cobro.CitaID, &cobro.PacienteDocumento, &cobro.SubtotalCentavos,
		&cobro.DescuentoBP, &cobro.DescuentoCentavos, &cobro.Moneda, &cobro.CreadoPor)
	if errors.Is(err, pgx.ErrNoRows) {
		return models.Cobro{}, models.ErrNoEncontrado
	}
	if err != nil {
		return models.Cobro{}, err
	}
	cobro.SaldoCentavos = cobro.TotalCentavos - cobro.PagadoCentavos
	cobro.Estado = models.EstadoDePago(cobro.TotalCentavos, cobro.PagadoCentavos)

	lineas, err := pg.db.Query(ctx,
		`SELECT id, procedimiento_id, descripcion, cantidad, precio_unitario_centavos, total_centavos
		   FROM cobro_lineas WHERE cobro_id = $1 ORDER BY id`, id)
	if err != nil {
		return models.Cobro{}, err
	}
	defer lineas.Close()
	cobro.Lineas = []models.CobroLinea{}
	for lineas.Next() {
		var l models.CobroLinea
		if err := lineas.Scan(&l.ID, &l.ProcedimientoID, &l.Descripcion, &l.Cantidad, &l.PrecioUnitarioCentavos, &l.TotalCentavos); err != nil {
			return models.Cobro{}, err
		}
		cobro.Lineas = append(cobro.Lineas, l)
	}
	if err := lineas.Err(); err != nil {
		return models.Cobro{}, err
	}

	pagos, err := pg.db.Query(ctx,
		`SELECT pg.id, pg.metodo, pg.monto_centavos, pg.referencia, u.nombre, pg.creado_en
		   FROM pagos pg JOIN usuarios u ON u.id = pg.creado_por WHERE pg.cobro_id = $1 ORDER BY pg.id`, id)
	if err != nil {
		return models.Cobro{}, err
	}
	defer pagos.Close()
	cobro.Pagos = []models.Pago{}
	for pagos.Next() {
		var p models.Pago
		if err := pagos.Scan(&p.ID, &p.Metodo, &p.MontoCentavos, &p.Referencia, &p.UsuarioNombre, &p.CreadoEn); err != nil {
			return models.Cobro{}, err
		}
		cobro.Pagos = append(cobro.Pagos, p)
	}
	if err := pagos.Err(); err != nil {
		return models.Cobro{}, err
	}

	if cobro.Anulado {
		var a models.Anulacion
		err := pg.db.QueryRow(ctx,
			`SELECT a.motivo, u.nombre, a.creado_en FROM cobro_anulaciones a JOIN usuarios u ON u.id = a.usuario_id
			  WHERE a.cobro_id = $1`, id).Scan(&a.Motivo, &a.UsuarioNombre, &a.CreadoEn)
		if err != nil {
			return models.Cobro{}, err
		}
		cobro.Anulacion = &a
	}
	return cobro, nil
}

// CrearCobro calcula en el servidor las líneas con el precio vigente del catálogo, el descuento y el total, asigna
// el número de comprobante y lo guarda todo en una sola transacción. El número sale de un contador que se incrementa
// dentro de esa misma transacción: dos solicitudes simultáneas se serializan por el bloqueo de la fila del contador,
// así que no se repiten, y si algo falla el número no se consume (no hay saltos).
func (pg *PostgresRepository) CrearCobro(ctx context.Context, n models.NuevoCobro, usuarioID int) (int, error) {
	var id int
	err := pgx.BeginFunc(ctx, pg.db, func(tx pgx.Tx) error {
		pacienteID := n.PacienteID
		if n.CitaID != nil {
			var estado string
			err := tx.QueryRow(ctx, `SELECT paciente_id, estado FROM citas WHERE id = $1 FOR UPDATE`, *n.CitaID).Scan(&pacienteID, &estado)
			if errors.Is(err, pgx.ErrNoRows) {
				return models.ErrReferencia
			}
			if err != nil {
				return err
			}
			if estado != models.EstadoAtendido {
				return models.ErrCitaNoCobrable
			}
			var yaCobrada bool
			if err := tx.QueryRow(ctx,
				`SELECT EXISTS (SELECT 1 FROM cobros c WHERE c.cita_id = $1
				                 AND NOT EXISTS (SELECT 1 FROM cobro_anulaciones a WHERE a.cobro_id = c.id))`, *n.CitaID).Scan(&yaCobrada); err != nil {
				return err
			}
			if yaCobrada {
				return models.ErrCitaYaCobrada
			}
		}

		var existe bool
		if err := tx.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM pacientes WHERE id = $1)`, pacienteID).Scan(&existe); err != nil {
			return err
		}
		if !existe {
			return models.ErrReferencia
		}

		lineas := make([]models.CobroLinea, 0, len(n.Lineas))
		subtotal := 0
		for _, pedida := range n.Lineas {
			var l models.CobroLinea
			err := tx.QueryRow(ctx,
				`SELECT id, nombre, precio_centavos FROM procedimientos WHERE id = $1 AND activo`, pedida.ProcedimientoID,
			).Scan(&l.ProcedimientoID, &l.Descripcion, &l.PrecioUnitarioCentavos)
			if errors.Is(err, pgx.ErrNoRows) {
				return models.ErrReferencia
			}
			if err != nil {
				return err
			}
			l.Cantidad = pedida.Cantidad
			l.TotalCentavos = l.Cantidad * l.PrecioUnitarioCentavos
			subtotal += l.TotalCentavos
			lineas = append(lineas, l)
		}
		descuento, total := models.CalcularTotales(subtotal, n.DescuentoBP)

		var correlativo int64
		if err := tx.QueryRow(ctx, `UPDATE contadores SET valor = valor + 1 WHERE clave = 'comprobante' RETURNING valor`).Scan(&correlativo); err != nil {
			return err
		}

		if err := tx.QueryRow(ctx,
			`INSERT INTO cobros (numero, paciente_id, cita_id, subtotal_centavos, descuento_bp, descuento_centavos, total_centavos, creado_por)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id`,
			fmt.Sprintf("C-%06d", correlativo), pacienteID, n.CitaID, subtotal, n.DescuentoBP, descuento, total, usuarioID,
		).Scan(&id); err != nil {
			return err
		}
		for _, l := range lineas {
			if _, err := tx.Exec(ctx,
				`INSERT INTO cobro_lineas (cobro_id, procedimiento_id, descripcion, cantidad, precio_unitario_centavos, total_centavos)
				 VALUES ($1, $2, $3, $4, $5, $6)`,
				id, l.ProcedimientoID, l.Descripcion, l.Cantidad, l.PrecioUnitarioCentavos, l.TotalCentavos); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return 0, traducirError(err)
	}
	return id, nil
}

// RegistrarPago agrega un abono. Bloquea la fila del cobro para que dos pagos simultáneos no superen el total.
func (pg *PostgresRepository) RegistrarPago(ctx context.Context, cobroID int, metodo string, monto int, referencia string, usuarioID int) error {
	err := pgx.BeginFunc(ctx, pg.db, func(tx pgx.Tx) error {
		// Primero se bloquea la fila del cobro y solo después se suma lo pagado, en otra consulta: dentro de la
		// misma, PostgreSQL calcularía la suma con la instantánea anterior a la espera del bloqueo y no vería los
		// pagos de la solicitud que acaba de terminar, permitiendo pagar de más.
		var total int
		err := tx.QueryRow(ctx, `SELECT total_centavos FROM cobros WHERE id = $1 FOR UPDATE`, cobroID).Scan(&total)
		if errors.Is(err, pgx.ErrNoRows) {
			return models.ErrNoEncontrado
		}
		if err != nil {
			return err
		}
		var pagado int
		var anulado bool
		if err := tx.QueryRow(ctx,
			`SELECT COALESCE(SUM(monto_centavos), 0)::int,
			        EXISTS (SELECT 1 FROM cobro_anulaciones WHERE cobro_id = $1)
			   FROM pagos WHERE cobro_id = $1`, cobroID).Scan(&pagado, &anulado); err != nil {
			return err
		}
		if anulado {
			return models.ErrCobroAnulado
		}
		if pagado+monto > total {
			return models.ErrExcedeSaldo
		}
		_, err = tx.Exec(ctx,
			`INSERT INTO pagos (cobro_id, metodo, monto_centavos, referencia, creado_por) VALUES ($1, $2, $3, $4, $5)`,
			cobroID, metodo, monto, referencia, usuarioID)
		return err
	})
	return traducirError(err)
}

// AnularCobro deja sin efecto el cobro con un registro nuevo (motivo, quién y cuándo); nada se borra.
func (pg *PostgresRepository) AnularCobro(ctx context.Context, cobroID int, motivo string, usuarioID int) error {
	err := pgx.BeginFunc(ctx, pg.db, func(tx pgx.Tx) error {
		var existe bool
		if err := tx.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM cobros WHERE id = $1)`, cobroID).Scan(&existe); err != nil {
			return err
		}
		if !existe {
			return models.ErrNoEncontrado
		}
		_, err := tx.Exec(ctx,
			`INSERT INTO cobro_anulaciones (cobro_id, motivo, usuario_id) VALUES ($1, $2, $3)`, cobroID, motivo, usuarioID)
		return err
	})
	// La unicidad de cobro_id en las anulaciones convierte un segundo intento en ErrDuplicado.
	if errors.Is(traducirError(err), models.ErrDuplicado) {
		return models.ErrCobroAnulado
	}
	return traducirError(err)
}
