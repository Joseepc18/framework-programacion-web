// Pruebas de integración contra PostgreSQL real: comprueban lo que solo la base garantiza (disparadores de
// inmutabilidad, bloqueos, numeración). Se omiten si no hay base: se activan con DATABASE_URL o con las variables DB_*
// del entorno (por ejemplo las del backend/.env), y exigen que las migraciones estén aplicadas.
package repository

import (
	"context"
	"errors"
	"net"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"sigo-dental-backend/models"
)

// repoDePrueba abre el repositorio sobre la base local o salta la prueba si no está configurada.
func repoDePrueba(t *testing.T) *PostgresRepository {
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
	return Nuevo(pool)
}

// RNF-06: ni la API ni el dueño de la tabla pueden modificar o borrar una fila de auditoría.
func TestAuditoriaEsDeSoloInsercion(t *testing.T) {
	pg := repoDePrueba(t)
	ctx := context.Background()
	if err := pg.RegistrarAcceso(ctx, models.EventoAuditoria{
		Email: "integracion@sigo-dental.test", Rol: models.RolAuditor, Accion: models.AccionConsulta, Entidad: "prueba",
		IP: "203.0.113.5", Metodo: "GET", Ruta: "/api/prueba",
	}); err != nil {
		t.Fatal(err)
	}
	pagina, err := pg.ListarAuditoria(ctx, models.FiltroAuditoria{Usuario: "integracion@sigo-dental.test", Limite: 5, Pagina: 1})
	if err != nil || len(pagina.Registros) == 0 {
		t.Fatalf("no se lee lo registrado: %v %v", pagina, err)
	}
	if pagina.Registros[0].IP != "203.0.113.5" {
		t.Errorf("IP = %q", pagina.Registros[0].IP)
	}

	for nombre, sentencia := range map[string]string{
		"UPDATE":   `UPDATE auditoria_accesos SET rol = 'x'`,
		"DELETE":   `DELETE FROM auditoria_accesos`,
		"TRUNCATE": `TRUNCATE auditoria_accesos`,
	} {
		_, err := pg.db.Exec(ctx, sentencia)
		if !errors.Is(traducirError(err), models.ErrInmutable) {
			t.Errorf("%s: error = %v, se esperaba ErrInmutable", nombre, err)
		}
	}
}
