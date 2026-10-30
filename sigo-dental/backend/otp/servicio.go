// Paquete otp: códigos de un solo uso para firmar consentimientos y para el ingreso al portal del paciente. El código
// de 6 dígitos se guarda solo como HMAC con la clave del servidor, vence a los 10 minutos y admite 3 intentos.
package otp

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"time"
)

// Propósitos de un código.
const (
	Consentimiento = "consentimiento"
	Portal         = "portal"
)

// Vigencia y número máximo de intentos de cada código.
const (
	Vigencia       = 10 * time.Minute
	MaxIntentos    = 3
	longitudCodigo = 6
)

// Errores de validación. Ninguno revela el código correcto.
var (
	ErrNoExiste = errors.New("no hay un código vigente: solicita uno nuevo")
	ErrExpirado = errors.New("el código venció: solicita uno nuevo")
	ErrUsado    = errors.New("el código ya se usó: solicita uno nuevo")
	ErrAgotado  = errors.New("se agotaron los intentos del código: solicita uno nuevo")
	ErrInvalido = errors.New("el código no es correcto")
)

// Codigo es la fila guardada de un código emitido.
type Codigo struct {
	ID       int
	Hash     string
	ExpiraEn time.Time
	Intentos int
	Usado    bool
}

// Repositorio es el contrato de la capa de datos.
type Repositorio interface {
	// GuardarCodigo registra un código nuevo; los anteriores del mismo propósito y referencia quedan inutilizados.
	GuardarCodigo(ctx context.Context, proposito string, referenciaID int, hash string, expira time.Time) error
	// UltimoCodigo devuelve el código más reciente; ok es false si no hay ninguno.
	UltimoCodigo(ctx context.Context, proposito string, referenciaID int) (c Codigo, ok bool, err error)
	// RegistrarIntentoFallido suma un intento de forma atómica y devuelve los intentos acumulados.
	RegistrarIntentoFallido(ctx context.Context, id int) (int, error)
	// ConsumirCodigo lo marca como usado; false si otra petición ya lo había usado.
	ConsumirCodigo(ctx context.Context, id int) (bool, error)
}

// Servicio emite y valida códigos.
type Servicio struct {
	repo    Repositorio
	secreto []byte
	ahora   func() time.Time
}

// NuevoServicio construye el servicio con la clave del servidor (la misma de los tokens).
func NuevoServicio(repo Repositorio, secreto []byte) *Servicio {
	return &Servicio{repo: repo, secreto: secreto, ahora: time.Now}
}

func (s *Servicio) hash(proposito string, referenciaID int, codigo string) string {
	h := hmac.New(sha256.New, s.secreto)
	fmt.Fprintf(h, "otp|%s|%d|%s", proposito, referenciaID, codigo)
	return hex.EncodeToString(h.Sum(nil))
}

// Emitir genera un código de 6 dígitos nuevo, lo guarda con su vencimiento y lo devuelve para enviarlo al destinatario.
func (s *Servicio) Emitir(ctx context.Context, proposito string, referenciaID int) (string, time.Time, error) {
	limite := new(big.Int).Exp(big.NewInt(10), big.NewInt(longitudCodigo), nil)
	n, err := rand.Int(rand.Reader, limite)
	if err != nil {
		return "", time.Time{}, err
	}
	codigo := fmt.Sprintf("%0*d", longitudCodigo, n)
	expira := s.ahora().Add(Vigencia)
	if err := s.repo.GuardarCodigo(ctx, proposito, referenciaID, s.hash(proposito, referenciaID, codigo), expira); err != nil {
		return "", time.Time{}, err
	}
	return codigo, expira, nil
}

// Validar comprueba el código contra el último emitido. Un código vencido, usado o con los 3 intentos agotados no
// valida aunque sea el correcto; un código correcto se consume (no se puede reutilizar).
func (s *Servicio) Validar(ctx context.Context, proposito string, referenciaID int, codigo string) error {
	c, ok, err := s.repo.UltimoCodigo(ctx, proposito, referenciaID)
	if err != nil {
		return err
	}
	switch {
	case !ok:
		return ErrNoExiste
	case c.Usado:
		return ErrUsado
	case !s.ahora().Before(c.ExpiraEn):
		return ErrExpirado
	case c.Intentos >= MaxIntentos:
		return ErrAgotado
	}
	if !hmac.Equal([]byte(s.hash(proposito, referenciaID, codigo)), []byte(c.Hash)) {
		intentos, err := s.repo.RegistrarIntentoFallido(ctx, c.ID)
		if err != nil {
			return err
		}
		if intentos >= MaxIntentos {
			return ErrAgotado
		}
		return ErrInvalido
	}
	consumido, err := s.repo.ConsumirCodigo(ctx, c.ID)
	if err != nil {
		return err
	}
	if !consumido {
		return ErrUsado
	}
	return nil
}
