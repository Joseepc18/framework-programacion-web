package almacenamiento

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// S3 guarda los archivos en un bucket compatible con S3 (AWS S3, Cloudflare R2, MinIO, Backblaze B2...).
type S3 struct {
	cliente *minio.Client
	bucket  string
}

// NuevoS3 conecta con el almacén. endpoint va sin esquema (por ejemplo "s3.amazonaws.com" o
// "cuenta.r2.cloudflarestorage.com"); usarSSL debe ser verdadero en producción.
func NuevoS3(endpoint, bucket, acceso, secreto, region string, usarSSL bool) (*S3, error) {
	cliente, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(acceso, secreto, ""),
		Secure: usarSSL,
		Region: region,
	})
	if err != nil {
		return nil, fmt.Errorf("no se pudo configurar el almacén S3: %w", err)
	}
	return &S3{cliente: cliente, bucket: bucket}, nil
}

// Guardar sube el objeto; tamano es el largo exacto del contenido.
func (s *S3) Guardar(ctx context.Context, clave string, contenido io.Reader, tamano int64, tipoMIME string) error {
	_, err := s.cliente.PutObject(ctx, s.bucket, clave, contenido, tamano, minio.PutObjectOptions{ContentType: tipoMIME})
	return err
}

// Abrir descarga el objeto. Se lee entero a memoria: los archivos están acotados por el límite de tamaño y así un
// objeto inexistente se detecta aquí, no a mitad de la respuesta HTTP.
func (s *S3) Abrir(ctx context.Context, clave string) (io.ReadCloser, error) {
	objeto, err := s.cliente.GetObject(ctx, s.bucket, clave, minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}
	defer func() { _ = objeto.Close() }()
	datos, err := io.ReadAll(objeto)
	if err != nil {
		if minio.ToErrorResponse(err).Code == "NoSuchKey" {
			return nil, ErrNoExiste
		}
		return nil, err
	}
	return io.NopCloser(bytes.NewReader(datos)), nil
}

// NuevoDesdeEntorno elige el adaptador según ALMACEN_ARCHIVOS: "disco" (por defecto, en ARCHIVOS_DIR) o "s3"
// (con S3_ENDPOINT, S3_BUCKET, S3_ACCESS_KEY, S3_SECRET_KEY, S3_REGION y S3_USE_SSL).
func NuevoDesdeEntorno() (AlmacenArchivos, error) {
	switch strings.ToLower(strings.TrimSpace(os.Getenv("ALMACEN_ARCHIVOS"))) {
	case "", "disco":
		raiz := strings.TrimSpace(os.Getenv("ARCHIVOS_DIR"))
		if raiz == "" {
			raiz = "datos/archivos"
		}
		return NuevoDisco(raiz)
	case "s3":
		endpoint, bucket := os.Getenv("S3_ENDPOINT"), os.Getenv("S3_BUCKET")
		acceso, secreto := os.Getenv("S3_ACCESS_KEY"), os.Getenv("S3_SECRET_KEY")
		if endpoint == "" || bucket == "" || acceso == "" || secreto == "" {
			return nil, fmt.Errorf("ALMACEN_ARCHIVOS=s3 requiere S3_ENDPOINT, S3_BUCKET, S3_ACCESS_KEY y S3_SECRET_KEY")
		}
		return NuevoS3(endpoint, bucket, acceso, secreto, os.Getenv("S3_REGION"), !strings.EqualFold(os.Getenv("S3_USE_SSL"), "false"))
	}
	return nil, fmt.Errorf("ALMACEN_ARCHIVOS %q no es válido: usa disco o s3", os.Getenv("ALMACEN_ARCHIVOS"))
}
