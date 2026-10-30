package otp

import (
	"context"
	"errors"
	"testing"
	"time"
)

type repoMemoria struct {
	codigos []*Codigo
}

func (r *repoMemoria) GuardarCodigo(_ context.Context, _ string, _ int, hash string, expira time.Time) error {
	for _, c := range r.codigos {
		c.Usado = true // el nuevo código inutiliza los anteriores
	}
	r.codigos = append(r.codigos, &Codigo{ID: len(r.codigos) + 1, Hash: hash, ExpiraEn: expira})
	return nil
}

func (r *repoMemoria) UltimoCodigo(context.Context, string, int) (Codigo, bool, error) {
	if len(r.codigos) == 0 {
		return Codigo{}, false, nil
	}
	return *r.codigos[len(r.codigos)-1], true, nil
}

func (r *repoMemoria) RegistrarIntentoFallido(_ context.Context, id int) (int, error) {
	r.codigos[id-1].Intentos++
	return r.codigos[id-1].Intentos, nil
}

func (r *repoMemoria) ConsumirCodigo(_ context.Context, id int) (bool, error) {
	if r.codigos[id-1].Usado {
		return false, nil
	}
	r.codigos[id-1].Usado = true
	return true, nil
}

func nuevo(t *testing.T) (*Servicio, *time.Time) {
	t.Helper()
	ahora := time.Date(2026, 10, 30, 10, 0, 0, 0, time.UTC)
	s := NuevoServicio(&repoMemoria{}, []byte("secreto-de-prueba-con-32-bytes-o-mas"))
	s.ahora = func() time.Time { return ahora }
	return s, &ahora
}

func TestElCodigoCorrectoValidaUnaSolaVez(t *testing.T) {
	s, _ := nuevo(t)
	ctx := context.Background()
	codigo, _, err := s.Emitir(ctx, Consentimiento, 7)
	if err != nil || len(codigo) != 6 {
		t.Fatalf("código = %q, error = %v", codigo, err)
	}
	if err := s.Validar(ctx, Consentimiento, 7, codigo); err != nil {
		t.Fatalf("el código correcto debe validar: %v", err)
	}
	if err := s.Validar(ctx, Consentimiento, 7, codigo); !errors.Is(err, ErrUsado) {
		t.Errorf("reutilizar el código: error = %v, se esperaba ErrUsado", err)
	}
}

func TestUnCodigoVencidoNoValida(t *testing.T) {
	s, ahora := nuevo(t)
	ctx := context.Background()
	codigo, _, _ := s.Emitir(ctx, Consentimiento, 7)
	*ahora = ahora.Add(Vigencia - time.Second)
	if err := s.Validar(ctx, Consentimiento, 7, "000000"); !errors.Is(err, ErrInvalido) && !errors.Is(err, ErrAgotado) {
		t.Errorf("a los 9:59 el código sigue vigente: error = %v", err)
	}
	*ahora = ahora.Add(time.Second)
	if err := s.Validar(ctx, Consentimiento, 7, codigo); !errors.Is(err, ErrExpirado) {
		t.Errorf("a los 10 minutos: error = %v, se esperaba ErrExpirado", err)
	}
}

func TestTresIntentosFallidosAgotanElCodigoAunqueDespuesSeaCorrecto(t *testing.T) {
	s, _ := nuevo(t)
	ctx := context.Background()
	codigo, _, _ := s.Emitir(ctx, Consentimiento, 7)
	mal := "999999"
	if codigo == mal {
		mal = "888888"
	}
	for i := 1; i <= 2; i++ {
		if err := s.Validar(ctx, Consentimiento, 7, mal); !errors.Is(err, ErrInvalido) {
			t.Fatalf("intento %d: error = %v, se esperaba ErrInvalido", i, err)
		}
	}
	if err := s.Validar(ctx, Consentimiento, 7, mal); !errors.Is(err, ErrAgotado) {
		t.Fatalf("tercer intento: error = %v, se esperaba ErrAgotado", err)
	}
	if err := s.Validar(ctx, Consentimiento, 7, codigo); !errors.Is(err, ErrAgotado) {
		t.Errorf("el código correcto tras agotar los intentos: error = %v, se esperaba ErrAgotado", err)
	}
}

func TestUnCodigoNuevoInutilizaAlAnteriorYNoSirveParaOtraReferencia(t *testing.T) {
	s, _ := nuevo(t)
	ctx := context.Background()
	primero, _, _ := s.Emitir(ctx, Consentimiento, 7)
	segundo, _, _ := s.Emitir(ctx, Consentimiento, 7)
	if primero != segundo {
		if err := s.Validar(ctx, Consentimiento, 7, primero); err == nil {
			t.Error("el código anterior no debe validar tras emitir otro")
		}
	}
	// El hash incluye el propósito y la referencia: el mismo código no vale para otro consentimiento.
	if s.hash(Consentimiento, 7, segundo) == s.hash(Consentimiento, 8, segundo) || s.hash(Consentimiento, 7, segundo) == s.hash(Portal, 7, segundo) {
		t.Error("el hash debe depender del propósito y de la referencia")
	}
}

func TestSinCodigoEmitidoNoHayNadaQueValidar(t *testing.T) {
	s, _ := nuevo(t)
	if err := s.Validar(context.Background(), Portal, 1, "123456"); !errors.Is(err, ErrNoExiste) {
		t.Errorf("error = %v, se esperaba ErrNoExiste", err)
	}
}
