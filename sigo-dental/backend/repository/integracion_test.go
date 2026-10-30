// Pruebas de integración contra PostgreSQL real: comprueban lo que solo la base garantiza (disparadores de
// inmutabilidad, bloqueos, numeración). Se omiten si no hay base: se activan con DATABASE_URL o con las variables DB_*
// del entorno (por ejemplo las del backend/.env), y exigen que las migraciones estén aplicadas.
//
// Cada prueba trabaja dentro de una transacción que se revierte al terminar, así no deja datos (los registros clínicos
// y de auditoría son inmutables y no se podrían limpiar después). Las pruebas de concurrencia, que necesitan varias
// conexiones, usan el pool y limpian lo que pueden.
package repository

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"sigo-dental-backend/models"
)

// poolDePrueba abre el pool de la base local o salta la prueba si no está configurada.
func poolDePrueba(t *testing.T) *pgxpool.Pool {
	t.Helper()
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" && os.Getenv("DB_HOST") != "" {
		dsn = (&url.URL{
			Scheme:   "postgres",
			User:     url.UserPassword(os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD")),
			Host:     net.JoinHostPort(os.Getenv("DB_HOST"), os.Getenv("DB_PORT")),
			Path:     "/" + os.Getenv("DB_NAME"),
			RawQuery: "sslmode=disable",
		}).String()
	}
	if dsn == "" {
		t.Skip("sin base de datos: define DATABASE_URL o DB_HOST para las pruebas de integración")
	}
	ctx, cancelar := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelar()
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Skipf("no se pudo conectar: %v", err)
	}
	if err := pool.Ping(ctx); err != nil {
		t.Skipf("la base no responde: %v", err)
	}
	t.Cleanup(pool.Close)
	return pool
}

// repoTransaccional devuelve un repositorio que opera dentro de una transacción revertida al terminar la prueba.
func repoTransaccional(t *testing.T) (*PostgresRepository, pgx.Tx) {
	t.Helper()
	pool := poolDePrueba(t)
	tx, err := pool.Begin(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = tx.Rollback(context.Background()) })
	return &PostgresRepository{db: tx, pool: pool}, tx
}

// debeFallarComoInmutable ejecuta la sentencia en un punto de guardado y comprueba que la base la rechace como
// registro inmutable; el punto de guardado se revierte para que la transacción siga viva.
func debeFallarComoInmutable(t *testing.T, tx pgx.Tx, descripcion, sentencia string, args ...any) {
	t.Helper()
	punto, err := tx.Begin(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = punto.Rollback(context.Background()) }()
	if _, err := punto.Exec(context.Background(), sentencia, args...); !errors.Is(traducirError(err), models.ErrInmutable) {
		t.Errorf("%s: error = %v, se esperaba un registro inmutable", descripcion, err)
	}
}

// intentar ejecuta la operación en un punto de guardado que se revierte, para comprobar un fallo sin abortar la transacción
// de la prueba.
func intentar(t *testing.T, tx pgx.Tx, operacion func(pg *PostgresRepository) error) error {
	t.Helper()
	punto, err := tx.Begin(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = punto.Rollback(context.Background()) }()
	return operacion(&PostgresRepository{db: punto})
}

var contadorDocumentos int

// pacienteDePrueba crea un paciente dentro de la transacción de la prueba y devuelve su ID y el del odontólogo de prueba.
func pacienteDePrueba(t *testing.T, tx pgx.Tx) (pacienteID, usuarioID int) {
	t.Helper()
	ctx := context.Background()
	if err := tx.QueryRow(ctx, `SELECT id FROM usuarios WHERE email = 'odontologo@sigo-dental.test'`).Scan(&usuarioID); err != nil {
		t.Skipf("faltan los datos de prueba (900_datos_prueba.sql): %v", err)
	}
	contadorDocumentos++
	err := tx.QueryRow(ctx,
		`INSERT INTO pacientes (documento, nombres, apellidos, fecha_nacimiento, sexo, telefono, emergencia_nombre, emergencia_telefono)
		 VALUES ($1, 'Paciente', 'Integración', DATE '1990-01-01', 'otro', '0990000000', 'Contacto', '0991111111') RETURNING id`,
		fmt.Sprintf("T%d-%d", time.Now().UnixNano(), contadorDocumentos)).Scan(&pacienteID)
	if err != nil {
		t.Fatal(err)
	}
	return pacienteID, usuarioID
}

// RNF-06: ni la API ni el dueño de la tabla pueden modificar o borrar una fila de auditoría.
func TestAuditoriaEsDeSoloInsercion(t *testing.T) {
	pg, tx := repoTransaccional(t)
	ctx := context.Background()
	correo := fmt.Sprintf("integracion-%d@sigo-dental.test", time.Now().UnixNano())
	if err := pg.RegistrarAcceso(ctx, models.EventoAuditoria{
		Email: correo, Rol: models.RolAuditor, Accion: models.AccionConsulta, Entidad: "prueba",
		IP: "203.0.113.5", Metodo: "GET", Ruta: "/api/prueba",
	}); err != nil {
		t.Fatal(err)
	}
	pagina, err := pg.ListarAuditoria(ctx, models.FiltroAuditoria{Usuario: correo, Limite: 5, Pagina: 1})
	if err != nil || len(pagina.Registros) != 1 {
		t.Fatalf("no se lee lo registrado: %v %v", pagina, err)
	}
	if pagina.Registros[0].IP != "203.0.113.5" {
		t.Errorf("IP = %q", pagina.Registros[0].IP)
	}
	debeFallarComoInmutable(t, tx, "UPDATE", `UPDATE auditoria_accesos SET rol = 'x'`)
	debeFallarComoInmutable(t, tx, "DELETE", `DELETE FROM auditoria_accesos`)
	debeFallarComoInmutable(t, tx, "TRUNCATE", `TRUNCATE auditoria_accesos`)
}

// RN-01 en el odontograma: corregir conserva el historial y un hallazgo realizado no se toca ni desde la base.
func TestOdontogramaConservaHistorialYProtegeLoRealizado(t *testing.T) {
	pg, tx := repoTransaccional(t)
	ctx := context.Background()
	pacienteID, usuarioID := pacienteDePrueba(t, tx)

	caries, err := pg.RegistrarHallazgo(ctx, models.NuevoHallazgo{PacienteID: pacienteID, Pieza: 16, Superficie: "oclusal", Tipo: "caries", Estado: "requerido"}, usuarioID)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := pg.RegistrarHallazgo(ctx, models.NuevoHallazgo{PacienteID: pacienteID, Pieza: 16, Superficie: "oclusal", Tipo: "caries", Estado: "requerido"}, usuarioID); !errors.Is(err, models.ErrDuplicado) {
		t.Errorf("repetido: error = %v, se esperaba ErrDuplicado", err)
	}

	// Corregir: la caries pasa a mesial. La anterior queda en el historial, retirada.
	corregida, err := pg.CorregirHallazgo(ctx, caries.ID, models.NuevoHallazgo{PacienteID: pacienteID, Pieza: 16, Superficie: "mesial", Tipo: "caries", Estado: "requerido"}, usuarioID)
	if err != nil {
		t.Fatal(err)
	}
	if corregida.ReemplazaA == nil || *corregida.ReemplazaA != caries.ID {
		t.Errorf("la corrección debe apuntar al hallazgo que retiró: %+v", corregida)
	}
	vigentes, _ := pg.ListarHallazgos(ctx, pacienteID, nil)
	historial, _ := pg.HistorialHallazgos(ctx, pacienteID)
	if len(vigentes) != 1 || vigentes[0].Superficie != "mesial" || len(historial) != 2 {
		t.Errorf("vigentes = %d, historial = %d; se esperaban 1 y 2", len(vigentes), len(historial))
	}
	if err := pg.RetirarHallazgo(ctx, pacienteID, caries.ID, usuarioID); !errors.Is(err, models.ErrHallazgoRetirado) {
		t.Errorf("retirar uno ya retirado: error = %v", err)
	}

	// Un tratamiento realizado: no se corrige, no se retira y la base lo impide aunque se salte la API.
	endo, err := pg.RegistrarHallazgo(ctx, models.NuevoHallazgo{PacienteID: pacienteID, Pieza: 36, Tipo: "endodoncia", Estado: "realizado"}, usuarioID)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := pg.CorregirHallazgo(ctx, endo.ID, models.NuevoHallazgo{PacienteID: pacienteID, Pieza: 36, Tipo: "endodoncia", Estado: "requerido"}, usuarioID); !errors.Is(err, models.ErrInmutable) {
		t.Errorf("corregir un realizado: error = %v, se esperaba ErrInmutable", err)
	}
	if err := pg.RetirarHallazgo(ctx, pacienteID, endo.ID, usuarioID); !errors.Is(err, models.ErrInmutable) {
		t.Errorf("retirar un realizado: error = %v, se esperaba ErrInmutable", err)
	}
	debeFallarComoInmutable(t, tx, "UPDATE de un realizado", `UPDATE odontograma_hallazgos SET retirado_en = NOW(), retirado_por = $2 WHERE id = $1`, endo.ID, usuarioID)
	debeFallarComoInmutable(t, tx, "cambiar el tipo", `UPDATE odontograma_hallazgos SET tipo = 'sellante' WHERE id = $1`, corregida.ID)
	debeFallarComoInmutable(t, tx, "DELETE", `DELETE FROM odontograma_hallazgos WHERE id = $1`, corregida.ID)
}

func TestOdontogramaPermiteVerUnaVisitaAnterior(t *testing.T) {
	pg, tx := repoTransaccional(t)
	ctx := context.Background()
	pacienteID, usuarioID := pacienteDePrueba(t, tx)

	h, err := pg.RegistrarHallazgo(ctx, models.NuevoHallazgo{PacienteID: pacienteID, Pieza: 11, Superficie: "mesial", Tipo: "caries", Estado: "requerido"}, usuarioID)
	if err != nil {
		t.Fatal(err)
	}
	// Dentro de una transacción NOW() es constante: se fijan instantes distintos a mano para simular dos visitas
	// (con los disparadores desactivados solo para este ajuste de la prueba).
	if _, err := tx.Exec(ctx, `SET LOCAL session_replication_role = replica`); err != nil {
		t.Skipf("se necesita un rol que pueda desactivar disparadores: %v", err)
	}
	if _, err := tx.Exec(ctx, `UPDATE odontograma_hallazgos SET registrado_en = NOW() - interval '10 days' WHERE id = $1`, h.ID); err != nil {
		t.Fatal(err)
	}
	if err := pg.RetirarHallazgo(ctx, pacienteID, h.ID, usuarioID); err != nil {
		t.Fatal(err)
	}
	if _, err := tx.Exec(ctx, `UPDATE odontograma_hallazgos SET retirado_en = NOW() - interval '2 days' WHERE id = $1`, h.ID); err != nil {
		t.Fatal(err)
	}

	if _, err := tx.Exec(ctx, `SET LOCAL session_replication_role = DEFAULT`); err != nil {
		t.Fatal(err)
	}
	hace5 := time.Now().Add(-5 * 24 * time.Hour)
	hoy := time.Now()
	antes, _ := pg.ListarHallazgos(ctx, pacienteID, &hace5)
	ahora, _ := pg.ListarHallazgos(ctx, pacienteID, &hoy)
	vigente, _ := pg.ListarHallazgos(ctx, pacienteID, nil)
	if len(antes) != 1 || len(ahora) != 0 || len(vigente) != 0 {
		t.Errorf("hace 5 días = %d, hoy = %d, vigentes = %d; se esperaban 1, 0 y 0", len(antes), len(ahora), len(vigente))
	}
}

// RN-01: un procedimiento realizado no se borra ni se cambia, y una nota firmada no se modifica, ni desde la base.
func TestEvolucionProtegeProcedimientosRealizadosYNotasFirmadas(t *testing.T) {
	pg, tx := repoTransaccional(t)
	ctx := context.Background()
	pacienteID, usuarioID := pacienteDePrueba(t, tx)
	var profesionalID, catalogoID int
	if err := tx.QueryRow(ctx, `SELECT id FROM profesionales WHERE usuario_id = $1`, usuarioID).Scan(&profesionalID); err != nil {
		t.Skipf("falta el profesional de prueba: %v", err)
	}
	if err := tx.QueryRow(ctx, `SELECT id FROM procedimientos WHERE activo ORDER BY id LIMIT 1`).Scan(&catalogoID); err != nil {
		t.Skipf("falta el catálogo de prueba: %v", err)
	}

	pieza := 16
	id, err := pg.IniciarProcedimiento(ctx, models.NuevoProcedimientoRealizado{PacienteID: pacienteID, ProcedimientoID: catalogoID, Pieza: &pieza, Superficie: "oclusal"}, profesionalID, usuarioID)
	if err != nil {
		t.Fatal(err)
	}
	if err := pg.CambiarEstadoProcedimiento(ctx, pacienteID, id, models.ProcRealizado); err != nil {
		t.Fatal(err)
	}
	if err := pg.CambiarEstadoProcedimiento(ctx, pacienteID, id, models.ProcCancelado); !errors.Is(err, models.ErrProcedimientoNoEnCurso) {
		t.Errorf("cancelar uno realizado: error = %v", err)
	}
	debeFallarComoInmutable(t, tx, "DELETE de un procedimiento realizado", `DELETE FROM procedimientos_realizados WHERE id = $1`, id)
	debeFallarComoInmutable(t, tx, "UPDATE de un procedimiento realizado", `UPDATE procedimientos_realizados SET notas = 'otra cosa' WHERE id = $1`, id)

	notaID, err := pg.CrearNota(ctx, models.Nota{PacienteID: pacienteID, ProfesionalID: profesionalID, AutorID: usuarioID,
		Subjetivo: "s", Objetivo: "o", Evaluacion: "e", Plan: "p"})
	if err != nil {
		t.Fatal(err)
	}
	if err := pg.ActualizarNota(ctx, pacienteID, notaID, models.Nota{Subjetivo: "s2", Objetivo: "o2", Evaluacion: "e2", Plan: "p2"}); err != nil {
		t.Fatalf("una nota sin firmar se edita: %v", err)
	}
	nota, _ := pg.ObtenerNota(ctx, pacienteID, notaID)
	firmaNota := models.Firma{Entidad: models.FirmaNota, EntidadID: notaID, UsuarioID: usuarioID, ProfesionalID: profesionalID,
		NombreFirmante: "X", RegistroProfesional: "R", HuellaSHA256: models.Huella(models.ContenidoCanonicoNota(nota.Nota)), ContenidoCanonico: models.ContenidoCanonicoNota(nota.Nota)}
	if _, err := pg.GuardarFirma(ctx, firmaNota); err != nil {
		t.Fatal(err)
	}
	err = intentar(t, tx, func(p *PostgresRepository) error { _, err := p.GuardarFirma(ctx, firmaNota); return err })
	if !errors.Is(err, models.ErrDuplicado) {
		t.Errorf("firmar dos veces: error = %v", err)
	}
	err = intentar(t, tx, func(p *PostgresRepository) error {
		return p.ActualizarNota(ctx, pacienteID, notaID, models.Nota{Subjetivo: "s3", Objetivo: "o3", Evaluacion: "e3", Plan: "p3"})
	})
	if !errors.Is(err, models.ErrInmutable) {
		t.Errorf("editar una nota firmada: error = %v, se esperaba ErrInmutable", err)
	}
	debeFallarComoInmutable(t, tx, "DELETE de una nota", `DELETE FROM notas_clinicas WHERE id = $1`, notaID)
	debeFallarComoInmutable(t, tx, "UPDATE de la firma", `UPDATE firmas SET huella_sha256 = repeat('0', 64) WHERE entidad = 'nota'`)
}
