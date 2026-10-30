package firma

import (
	"context"
	"errors"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"
	"sigo-dental-backend/models"
)

type repoFalso struct {
	hash        string
	activo      bool
	profesional bool
}

func (r repoFalso) CredencialesDeUsuario(context.Context, int) (string, bool, bool, error) {
	return r.hash, r.activo, true, nil
}

func (r repoFalso) ProfesionalDeUsuario(context.Context, int) (models.ProfesionalDeUsuario, bool, error) {
	return models.ProfesionalDeUsuario{ID: 3, Nombre: "Dra. Prueba", RegistroProfesional: "REG-1"}, r.profesional, nil
}

func servicio(t *testing.T, r repoFalso) *Servicio {
	t.Helper()
	hash, err := bcrypt.GenerateFromPassword([]byte("clave-correcta"), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}
	r.hash = string(hash)
	return NuevoServicio(r)
}

func TestAutenticarExigeContrasenaYProfesionalActivo(t *testing.T) {
	ctx := context.Background()
	f, err := servicio(t, repoFalso{activo: true, profesional: true}).Autenticar(ctx, 1, "clave-correcta")
	if err != nil || f.ProfesionalID != 3 || f.RegistroProfesional != "REG-1" {
		t.Fatalf("firmante = %+v, error = %v", f, err)
	}
	if _, err := servicio(t, repoFalso{activo: true, profesional: true}).Autenticar(ctx, 1, "otra"); !errors.Is(err, models.ErrPasswordIncorrecta) {
		t.Errorf("contraseña errónea: error = %v", err)
	}
	if _, err := servicio(t, repoFalso{activo: false, profesional: true}).Autenticar(ctx, 1, "clave-correcta"); !errors.Is(err, models.ErrPasswordIncorrecta) {
		t.Errorf("cuenta desactivada: error = %v", err)
	}
	if _, err := servicio(t, repoFalso{activo: true, profesional: false}).Autenticar(ctx, 1, "clave-correcta"); !errors.Is(err, models.ErrNoEsProfesional) {
		t.Errorf("sin ficha de profesional: error = %v", err)
	}
}

func TestAutenticarBloqueaTrasVariosFallos(t *testing.T) {
	ctx := context.Background()
	s := servicio(t, repoFalso{activo: true, profesional: true})
	ahora := time.Date(2026, 10, 30, 9, 0, 0, 0, time.UTC)
	s.ahora = func() time.Time { return ahora }
	for i := 0; i < 5; i++ {
		_, _ = s.Autenticar(ctx, 1, "mal")
	}
	if _, err := s.Autenticar(ctx, 1, "clave-correcta"); !errors.Is(err, models.ErrDemasiadosIntentos) {
		t.Errorf("error = %v, se esperaba el bloqueo aun con la clave correcta", err)
	}
	// Otra cuenta no se ve afectada, y pasada la ventana la cuenta vuelve a poder firmar.
	if _, err := s.Autenticar(ctx, 2, "clave-correcta"); err != nil {
		t.Errorf("otra cuenta: %v", err)
	}
	ahora = ahora.Add(6 * time.Minute)
	if _, err := s.Autenticar(ctx, 1, "clave-correcta"); err != nil {
		t.Errorf("tras la ventana: %v", err)
	}
}

// Alterar el contenido de lo firmado hace que la huella deje de coincidir y se detecta.
func TestAlterarElContenidoRompeLaHuella(t *testing.T) {
	nota := models.Nota{ID: 7, PacienteID: 2, ProfesionalID: 3, AutorID: 1, CreadaEn: time.Date(2026, 10, 30, 9, 0, 0, 0, time.UTC),
		Subjetivo: "Dolor", Objetivo: "Caries", Evaluacion: "Pulpitis", Plan: "Endodoncia"}
	f := Construir(Firmante{UsuarioID: 1, ProfesionalID: 3, Nombre: "Dra. Prueba", RegistroProfesional: "REG-1"}, models.FirmaNota, nota.ID, models.ContenidoCanonicoNota(nota))
	if !Verificar(f, models.ContenidoCanonicoNota(nota)) {
		t.Fatal("la nota sin cambios debe verificar")
	}
	alterada := nota
	alterada.Plan = "Exodoncia"
	if Verificar(f, models.ContenidoCanonicoNota(alterada)) {
		t.Error("una nota alterada no debe verificar")
	}
	if len(f.HuellaSHA256) != 64 {
		t.Errorf("huella = %q", f.HuellaSHA256)
	}
}

func TestEstadoDeNotaConElRelojControlado(t *testing.T) {
	creada := time.Date(2026, 10, 30, 9, 0, 0, 0, time.UTC)
	plazo := 24 * time.Hour
	casos := []struct {
		nombre  string
		ahora   time.Time
		firmada bool
		estado  string
	}{
		{"recién creada", creada, false, models.NotaEditable},
		{"un minuto antes del plazo", creada.Add(24*time.Hour - time.Minute), false, models.NotaEditable},
		{"en el instante del plazo", creada.Add(24 * time.Hour), false, models.NotaPendienteFirma},
		{"vencida sin firma", creada.Add(48 * time.Hour), false, models.NotaPendienteFirma},
		{"firmada antes del plazo", creada.Add(time.Hour), true, models.NotaFirmada},
		{"firmada después del plazo", creada.Add(72 * time.Hour), true, models.NotaFirmada},
	}
	for _, c := range casos {
		if got := models.EstadoDeNota(creada, c.firmada, c.ahora, plazo); got != c.estado {
			t.Errorf("%s: estado = %s, se esperaba %s", c.nombre, got, c.estado)
		}
	}
}
