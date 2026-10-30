package almacenamiento

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Postgres guarda el contenido de los archivos en la propia base de datos (tabla archivos_contenido). Sirve donde no
// hay disco persistente ni almacén S3, como el plan gratuito de Render: el contenido sobrevive a los despliegues. Los
// archivos están acotados por el límite de tamaño de la subida, así que se leen enteros a memoria.
type Postgres struct {
	pool *pgxpool.Pool
}

// NuevoPostgres crea el almacén sobre el pool existente; requiere la migración 015.
func NuevoPostgres(pool *pgxpool.Pool) *Postgres {
	return &Postgres{pool: pool}
}

// Guardar inserta el contenido sin sobrescribir: una clave existente es un error.
func (p *Postgres) Guardar(ctx context.Context, clave string, contenido io.Reader, _ int64, tipoMIME string) error {
	datos, err := io.ReadAll(contenido)
	if err != nil {
		return fmt.Errorf("no se pudo leer el contenido del archivo: %w", err)
	}
	_, err = p.pool.Exec(ctx,
		`INSERT INTO archivos_contenido (clave, tipo_mime, contenido) VALUES ($1, $2, $3)`, clave, tipoMIME, datos)
	return err
}

// Abrir devuelve el contenido para leerlo; el llamador lo cierra.
func (p *Postgres) Abrir(ctx context.Context, clave string) (io.ReadCloser, error) {
	var datos []byte
	err := p.pool.QueryRow(ctx, `SELECT contenido FROM archivos_contenido WHERE clave = $1`, clave).Scan(&datos)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNoExiste
	}
	if err != nil {
		return nil, err
	}
	return io.NopCloser(bytes.NewReader(datos)), nil
}
