// Este archivo agrupa el acceso a datos del portal del paciente (F2-8).
package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"sigo-dental-backend/models"
)

// BuscarPacientePorDocumento devuelve el ID del paciente activo con ese documento de identidad.
func (pg *PostgresRepository) BuscarPacientePorDocumento(ctx context.Context, documento string) (int, bool, error) {
	var id int
	err := pg.db.QueryRow(ctx, `SELECT id FROM pacientes WHERE documento = $1 AND activo`, documento).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, false, nil
	}
	return id, err == nil, err
}

// UsuarioDePaciente devuelve la cuenta del portal del paciente y la crea si no existe. hashInicial es el hash de una
// contraseña aleatoria que nadie conoce: esa cuenta solo entra con código OTP.
func (pg *PostgresRepository) UsuarioDePaciente(ctx context.Context, pacienteID int, hashInicial string) (int, string, error) {
	var usuarioID int
	var email string
	err := pgx.BeginFunc(ctx, pg.db, func(tx pgx.Tx) error {
		err := tx.QueryRow(ctx, `SELECT id, email FROM usuarios WHERE paciente_id = $1 AND activo`, pacienteID).Scan(&usuarioID, &email)
		if err == nil {
			return nil
		}
		if !errors.Is(err, pgx.ErrNoRows) {
			return err
		}
		var historia, nombre string
		if err := tx.QueryRow(ctx, `SELECT numero_historia, nombres || ' ' || apellidos FROM pacientes WHERE id = $1`, pacienteID).Scan(&historia, &nombre); err != nil {
			return err
		}
		email = fmt.Sprintf("paciente.%s@portal.invalid", historia)
		return tx.QueryRow(ctx,
			`INSERT INTO usuarios (email, password_hash, nombre, rol_id, paciente_id)
			 VALUES ($1, $2, $3, (SELECT id FROM roles WHERE codigo = 'paciente'), $4)
			 ON CONFLICT (paciente_id) DO UPDATE SET paciente_id = EXCLUDED.paciente_id RETURNING id`,
			email, hashInicial, nombre, pacienteID).Scan(&usuarioID)
	})
	return usuarioID, email, traducirError(err)
}

// PacienteDePortal devuelve el paciente al que pertenece la cuenta del portal; ok es false si no es una cuenta de paciente.
func (pg *PostgresRepository) PacienteDePortal(ctx context.Context, usuarioID int) (int, bool, error) {
	var id int
	err := pg.db.QueryRow(ctx, `SELECT paciente_id FROM usuarios WHERE id = $1 AND paciente_id IS NOT NULL AND activo`, usuarioID).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, false, nil
	}
	return id, err == nil, err
}

// CitasDePaciente devuelve las citas recientes y próximas del paciente (sin alertas médicas ni notas internas).
func (pg *PostgresRepository) CitasDePaciente(ctx context.Context, pacienteID int) ([]models.CitaPortal, error) {
	rows, err := pg.db.Query(ctx,
		`SELECT c.id, c.estado, c.inicio, c.fin, u.nombre, co.nombre, COALESCE(pr.nombre, ''), c.motivo
		   FROM citas c
		   JOIN profesionales pf ON pf.id = c.profesional_id JOIN usuarios u ON u.id = pf.usuario_id
		   JOIN sillones s ON s.id = c.sillon_id JOIN consultorios co ON co.id = s.consultorio_id
		   LEFT JOIN procedimientos pr ON pr.id = c.procedimiento_id
		  WHERE c.paciente_id = $1 AND c.inicio > NOW() - INTERVAL '90 days'
		  ORDER BY c.inicio DESC LIMIT 50`, pacienteID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	lista := []models.CitaPortal{}
	for rows.Next() {
		var c models.CitaPortal
		if err := rows.Scan(&c.ID, &c.Estado, &c.Inicio, &c.Fin, &c.ProfesionalNombre, &c.ConsultorioNombre, &c.ProcedimientoNombre, &c.Motivo); err != nil {
			return nil, err
		}
		lista = append(lista, c)
	}
	return lista, rows.Err()
}
