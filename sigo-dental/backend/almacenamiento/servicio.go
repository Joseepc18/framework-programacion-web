package almacenamiento

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"path"
	"strings"

	"sigo-dental-backend/models"
)

// Registro es el contrato de la capa de datos para los metadatos de los archivos.
type Registro interface {
	RegistrarArchivo(ctx context.Context, a models.NuevoArchivo) (int, error)
}

// Servicio guarda el contenido en el almacén y sus metadatos en la base. Lo usan la subida de archivos y los
// documentos que genera el sistema (consentimientos, recetas y presupuestos en PDF).
type Servicio struct {
	almacen  AlmacenArchivos
	registro Registro
}

// NuevoServicio construye el servicio con el almacén y el registro inyectados.
func NuevoServicio(almacen AlmacenArchivos, registro Registro) *Servicio {
	return &Servicio{almacen: almacen, registro: registro}
}

// Almacen expone el almacén para leer el contenido.
func (s *Servicio) Almacen() AlmacenArchivos { return s.almacen }

// Guardar almacena el contenido bajo una clave nueva y aleatoria, y devuelve el ID del registro. El nombre original
// solo se conserva como metadato: nunca forma parte de la clave, así no se puede usar para recorrer directorios.
func (s *Servicio) Guardar(ctx context.Context, n models.NuevoArchivo, contenido []byte) (int, error) {
	huella := sha256.Sum256(contenido)
	aleatorio := make([]byte, 16)
	if _, err := rand.Read(aleatorio); err != nil {
		return 0, err
	}
	n.SHA256 = hex.EncodeToString(huella[:])
	n.TamanoBytes = len(contenido)
	n.ClaveAlmacen = path.Join("pacientes", fmt.Sprint(n.PacienteID), hex.EncodeToString(aleatorio)+extensionDe(n.TipoMIME))
	if err := s.almacen.Guardar(ctx, n.ClaveAlmacen, bytes.NewReader(contenido), int64(len(contenido)), n.TipoMIME); err != nil {
		return 0, fmt.Errorf("no se pudo guardar el archivo: %w", err)
	}
	return s.registro.RegistrarArchivo(ctx, n)
}

func extensionDe(tipoMIME string) string {
	switch strings.ToLower(tipoMIME) {
	case "image/jpeg":
		return ".jpg"
	case "image/png":
		return ".png"
	case "image/webp":
		return ".webp"
	case "application/pdf":
		return ".pdf"
	}
	return ""
}
