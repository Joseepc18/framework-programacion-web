package almacenamiento

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"

	"sigo-dental-backend/models"
)

func TestDiscoGuardaYAbre(t *testing.T) {
	d, err := NuevoDisco(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if err := d.Guardar(ctx, "pacientes/1/a.pdf", bytes.NewReader([]byte("hola")), 4, "application/pdf"); err != nil {
		t.Fatal(err)
	}
	contenido, err := d.Abrir(ctx, "pacientes/1/a.pdf")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = contenido.Close() }()
	if datos, _ := io.ReadAll(contenido); string(datos) != "hola" {
		t.Errorf("contenido = %q", datos)
	}
}

func TestDiscoNoSobrescribeNiSaleDeLaRaiz(t *testing.T) {
	d, _ := NuevoDisco(t.TempDir())
	ctx := context.Background()
	_ = d.Guardar(ctx, "a/b.pdf", bytes.NewReader([]byte("uno")), 3, "")
	if err := d.Guardar(ctx, "a/b.pdf", bytes.NewReader([]byte("dos")), 3, ""); err == nil {
		t.Error("sobrescribió un archivo existente")
	}
	for _, clave := range []string{"../fuera.pdf", "a/../../fuera.pdf", ""} {
		if err := d.Guardar(ctx, clave, bytes.NewReader([]byte("x")), 1, ""); err == nil {
			t.Errorf("aceptó la clave %q", clave)
		}
		if _, err := d.Abrir(ctx, clave); err == nil {
			t.Errorf("abrió la clave %q", clave)
		}
	}
	if _, err := d.Abrir(ctx, "no/existe.pdf"); !errors.Is(err, ErrNoExiste) {
		t.Errorf("error = %v, se esperaba ErrNoExiste", err)
	}
}

type registroFalso struct{ guardado models.NuevoArchivo }

func (r *registroFalso) RegistrarArchivo(_ context.Context, a models.NuevoArchivo) (int, error) {
	r.guardado = a
	return 1, nil
}

func TestServicioGuardaConClaveAleatoriaYHuella(t *testing.T) {
	d, _ := NuevoDisco(t.TempDir())
	registro := &registroFalso{}
	s := NuevoServicio(d, registro)
	contenido := []byte("%PDF-1.4 prueba")
	if _, err := s.Guardar(context.Background(), models.NuevoArchivo{PacienteID: 9, TipoMIME: "application/pdf", NombreOriginal: "../../x.pdf"}, contenido); err != nil {
		t.Fatal(err)
	}
	g := registro.guardado
	if g.TamanoBytes != len(contenido) || len(g.SHA256) != 64 {
		t.Errorf("tamaño o huella mal calculados: %+v", g)
	}
	if g.ClaveAlmacen == "" || bytes.Contains([]byte(g.ClaveAlmacen), []byte("x.pdf")) || !bytes.HasPrefix([]byte(g.ClaveAlmacen), []byte("pacientes/9/")) {
		t.Errorf("clave = %q: no debe contener el nombre original", g.ClaveAlmacen)
	}
}
