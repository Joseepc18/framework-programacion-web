// Paquete almacenamiento: guarda y entrega los archivos del paciente (radiografías, fotos y PDF) detrás de una
// interfaz. El disco local sirve para desarrollo; en producción se usa un almacén compatible con S3 porque el disco
// de Render gratuito es efímero. Cambiar de uno a otro no toca el resto del código.
package almacenamiento

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// ErrNoExiste indica que la clave pedida no está en el almacén.
var ErrNoExiste = errors.New("el archivo no existe en el almacén")

// AlmacenArchivos es el puerto de salida del almacenamiento. Las claves las genera siempre el servidor.
// No hay operación de borrado: en el expediente nada se elimina.
type AlmacenArchivos interface {
	Guardar(ctx context.Context, clave string, contenido io.Reader, tamano int64, tipoMIME string) error
	Abrir(ctx context.Context, clave string) (io.ReadCloser, error)
}

// Disco guarda los archivos bajo un directorio local.
type Disco struct {
	raiz string
}

// NuevoDisco crea el almacén en disco; crea el directorio si no existe.
func NuevoDisco(raiz string) (*Disco, error) {
	absoluta, err := filepath.Abs(raiz)
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(absoluta, 0o750); err != nil {
		return nil, fmt.Errorf("no se pudo crear el directorio de archivos: %w", err)
	}
	return &Disco{raiz: absoluta}, nil
}

// ruta resuelve la clave dentro de la raíz y rechaza cualquier intento de salir de ella.
func (d *Disco) ruta(clave string) (string, error) {
	limpia := filepath.Join(d.raiz, filepath.FromSlash(clave))
	if clave == "" || !strings.HasPrefix(limpia, d.raiz+string(os.PathSeparator)) {
		return "", fmt.Errorf("clave de archivo inválida: %q", clave)
	}
	return limpia, nil
}

// Guardar escribe el contenido sin sobrescribir: una clave existente es un error.
func (d *Disco) Guardar(_ context.Context, clave string, contenido io.Reader, _ int64, _ string) error {
	destino, err := d.ruta(clave)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(destino), 0o750); err != nil {
		return err
	}
	archivo, err := os.OpenFile(destino, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o640)
	if err != nil {
		return err
	}
	if _, err := io.Copy(archivo, contenido); err != nil {
		_ = archivo.Close()
		_ = os.Remove(destino)
		return err
	}
	return archivo.Close()
}

// Abrir devuelve el contenido para leerlo; el llamador lo cierra.
func (d *Disco) Abrir(_ context.Context, clave string) (io.ReadCloser, error) {
	origen, err := d.ruta(clave)
	if err != nil {
		return nil, err
	}
	archivo, err := os.Open(origen)
	if errors.Is(err, os.ErrNotExist) {
		return nil, ErrNoExiste
	}
	return archivo, err
}
