// Paquete firma: firma digital de los documentos clínicos (notas, adendas, recetas y consentimientos). Es una firma
// electrónica simple, no un certificado de firma electrónica avanzada: el profesional se reautentica con su
// contraseña y queda la huella SHA-256 del contenido canónico, su identidad y la fecha. Si el contenido cambiara
// por fuera de la API, la huella dejaría de coincidir y se detecta con Verificar.
package firma

import (
	"context"
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"
	"sigo-dental-backend/models"
)

// Repositorio es el contrato de la capa de datos que necesita el servicio.
type Repositorio interface {
	// CredencialesDeUsuario devuelve el hash de la contraseña y si la cuenta está activa; ok es false si no existe.
	CredencialesDeUsuario(ctx context.Context, usuarioID int) (hash string, activo bool, ok bool, err error)
	// ProfesionalDeUsuario devuelve el profesional activo de la cuenta; ok es false si no es un profesional activo.
	ProfesionalDeUsuario(ctx context.Context, usuarioID int) (p models.ProfesionalDeUsuario, ok bool, err error)
}

// Firmante es la identidad ya verificada de quien firma.
type Firmante struct {
	UsuarioID           int
	ProfesionalID       int
	Nombre              string
	RegistroProfesional string
}

// Servicio verifica la identidad de quien firma y construye la firma.
type Servicio struct {
	repo  Repositorio
	ahora func() time.Time

	mu       sync.Mutex
	fallos   map[int][]time.Time
	maxFallo int
	ventana  time.Duration
}

// NuevoServicio construye el servicio. Tras 5 contraseñas erróneas en 5 minutos, la cuenta no puede firmar hasta que
// pase la ventana: limita los intentos de adivinar la contraseña a través de la firma.
func NuevoServicio(repo Repositorio) *Servicio {
	return &Servicio{repo: repo, ahora: time.Now, fallos: map[int][]time.Time{}, maxFallo: 5, ventana: 5 * time.Minute}
}

// Autenticar reautentica a la cuenta con su contraseña y comprueba que sea un profesional activo.
func (s *Servicio) Autenticar(ctx context.Context, usuarioID int, password string) (Firmante, error) {
	if s.bloqueada(usuarioID) {
		return Firmante{}, models.ErrDemasiadosIntentos
	}
	hash, activo, ok, err := s.repo.CredencialesDeUsuario(ctx, usuarioID)
	if err != nil {
		return Firmante{}, err
	}
	if !ok || !activo {
		return Firmante{}, models.ErrPasswordIncorrecta
	}
	if bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) != nil {
		s.registrarFallo(usuarioID)
		return Firmante{}, models.ErrPasswordIncorrecta
	}
	profesional, ok, err := s.repo.ProfesionalDeUsuario(ctx, usuarioID)
	if err != nil {
		return Firmante{}, err
	}
	if !ok {
		return Firmante{}, models.ErrNoEsProfesional
	}
	s.limpiar(usuarioID)
	return Firmante{UsuarioID: usuarioID, ProfesionalID: profesional.ID, Nombre: profesional.Nombre, RegistroProfesional: profesional.RegistroProfesional}, nil
}

// Construir arma la firma del contenido canónico para la entidad dada; la fecha la pone la base al guardarla.
func Construir(f Firmante, entidad string, entidadID int, canonico string) models.Firma {
	return models.Firma{
		Entidad: entidad, EntidadID: entidadID, UsuarioID: f.UsuarioID, ProfesionalID: f.ProfesionalID,
		NombreFirmante: f.Nombre, RegistroProfesional: f.RegistroProfesional,
		HuellaSHA256: models.Huella(canonico), ContenidoCanonico: canonico,
	}
}

// Verificar indica si el contenido actual del documento sigue teniendo la huella que se firmó.
func Verificar(f models.Firma, canonicoActual string) bool {
	return f.HuellaSHA256 != "" && f.HuellaSHA256 == models.Huella(canonicoActual)
}

func (s *Servicio) recientes(usuarioID int) []time.Time {
	corte := s.ahora().Add(-s.ventana)
	var vigentes []time.Time
	for _, t := range s.fallos[usuarioID] {
		if t.After(corte) {
			vigentes = append(vigentes, t)
		}
	}
	return vigentes
}

func (s *Servicio) bloqueada(usuarioID int) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.recientes(usuarioID)) >= s.maxFallo
}

func (s *Servicio) registrarFallo(usuarioID int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.fallos[usuarioID] = append(s.recientes(usuarioID), s.ahora())
}

func (s *Servicio) limpiar(usuarioID int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.fallos, usuarioID)
}
