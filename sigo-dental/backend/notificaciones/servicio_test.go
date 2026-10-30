// Pruebas del servicio de avisos con un repositorio en memoria que respeta la clave única (cita, tipo) vigente.
package notificaciones

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"sigo-dental-backend/models"
)

type repoMemoria struct {
	cita           models.DatosParaNotificar
	notificaciones []models.Notificacion
}

func (r *repoMemoria) DatosParaNotificar(context.Context, int) (models.DatosParaNotificar, error) {
	return r.cita, nil
}

func (r *repoMemoria) CitasParaRecordar(_ context.Context, desde, hasta time.Time) ([]models.DatosParaNotificar, error) {
	vigente := r.cita.Estado == models.EstadoProgramado || r.cita.Estado == models.EstadoConfirmado
	if vigente && r.cita.Inicio.After(desde) && !r.cita.Inicio.After(hasta) {
		return []models.DatosParaNotificar{r.cita}, nil
	}
	return nil, nil
}

func (r *repoMemoria) ReservarNotificacion(_ context.Context, n models.Notificacion) (int, bool, error) {
	for _, existente := range r.notificaciones {
		if existente.Vigente && existente.CitaID == n.CitaID && existente.Tipo == n.Tipo {
			return 0, false, nil
		}
	}
	n.ID, n.Vigente, n.Estado = len(r.notificaciones)+1, true, models.NotifPendiente
	r.notificaciones = append(r.notificaciones, n)
	return n.ID, true, nil
}

func (r *repoMemoria) RegistrarResultado(_ context.Context, id int, enviada bool, detalle string) error {
	n := &r.notificaciones[id-1]
	n.Estado, n.Detalle = models.NotifFallida, detalle
	if enviada {
		n.Estado = models.NotifEnviada
	}
	return nil
}

func (r *repoMemoria) InvalidarNotificaciones(_ context.Context, citaID int) error {
	for i := range r.notificaciones {
		if r.notificaciones[i].CitaID == citaID {
			r.notificaciones[i].Vigente = false
		}
	}
	return nil
}

func (r *repoMemoria) vigentes(tipo string) int {
	n := 0
	for _, notif := range r.notificaciones {
		if notif.Vigente && notif.Tipo == tipo {
			n++
		}
	}
	return n
}

type notificadorCapturador struct {
	enviados []Mensaje
	err      error
}

func (c *notificadorCapturador) Enviar(_ context.Context, m Mensaje) error {
	c.enviados = append(c.enviados, m)
	return c.err
}

var secretoDePrueba = []byte("secreto-de-prueba-con-32-bytes-o-mas")

// t0 es el momento en que se agenda la cita; la cita es 72 horas después.
var t0 = time.Date(2026, 11, 2, 12, 0, 0, 0, time.UTC)

func nuevoEscenario(inicio time.Time) (*Servicio, *repoMemoria, *notificadorCapturador, *time.Time) {
	repo := &repoMemoria{cita: models.DatosParaNotificar{
		CitaID: 1, Estado: models.EstadoProgramado, Inicio: inicio, Actualizada: t0,
		PacienteNombre: "Ana Prueba", Telefono: "0991112233", Email: "ana@correo.test", Canal: "whatsapp",
		ProfesionalNombre: "Dra. Pérez", Zona: "America/Guayaquil",
	}}
	salida := &notificadorCapturador{}
	servicio := NuevoServicio(repo, salida, secretoDePrueba, "https://sigo.test")
	ahora := t0
	servicio.ahora = func() time.Time { return ahora }
	return servicio, repo, salida, &ahora
}

func TestUnaCitaGeneraUnaConfirmacionYDosRecordatorios(t *testing.T) {
	servicio, repo, salida, ahora := nuevoEscenario(t0.Add(72 * time.Hour))
	ctx := context.Background()

	servicio.CitaAgendada(ctx, 1)
	servicio.CitaAgendada(ctx, 1) // un reintento no duplica la confirmación

	// El planificador corre cada minuto durante cuatro días: debe enviar justo dos recordatorios.
	for *ahora = t0; ahora.Before(t0.Add(73 * time.Hour)); *ahora = ahora.Add(time.Minute) {
		if _, err := servicio.Procesar(ctx); err != nil {
			t.Fatal(err)
		}
	}

	for _, tipo := range []string{models.NotifConfirmacion, models.NotifRecordatorio48, models.NotifRecordatorio24} {
		if n := repo.vigentes(tipo); n != 1 {
			t.Errorf("notificaciones %s = %d, se esperaba 1", tipo, n)
		}
	}
	if len(salida.enviados) != 3 {
		t.Errorf("mensajes entregados = %d, se esperaban 3", len(salida.enviados))
	}
}

func TestLosRecordatoriosSalenEnSuVentana(t *testing.T) {
	servicio, repo, _, ahora := nuevoEscenario(t0.Add(72 * time.Hour))
	ctx := context.Background()

	casos := []struct {
		nombre         string
		cuando         time.Duration
		recordatorio48 int
		recordatorio24 int
	}{
		{"a 71 h de la cita aún no hay recordatorio", 1 * time.Hour, 0, 0},
		{"a 47 h sale el de 48 h", 25 * time.Hour, 1, 0},
		{"a 25 h todavía no sale el de 24 h", 47 * time.Hour, 1, 0},
		{"a 23 h sale el de 24 h", 49 * time.Hour, 1, 1},
	}
	for _, caso := range casos {
		*ahora = t0.Add(caso.cuando)
		if _, err := servicio.Procesar(ctx); err != nil {
			t.Fatal(err)
		}
		if got := repo.vigentes(models.NotifRecordatorio48); got != caso.recordatorio48 {
			t.Errorf("%s: recordatorios de 48 h = %d, se esperaban %d", caso.nombre, got, caso.recordatorio48)
		}
		if got := repo.vigentes(models.NotifRecordatorio24); got != caso.recordatorio24 {
			t.Errorf("%s: recordatorios de 24 h = %d, se esperaban %d", caso.nombre, got, caso.recordatorio24)
		}
	}
}

func TestUnaCitaCanceladaNoGeneraRecordatorios(t *testing.T) {
	servicio, repo, salida, ahora := nuevoEscenario(t0.Add(72 * time.Hour))
	ctx := context.Background()
	servicio.CitaAgendada(ctx, 1)

	repo.cita.Estado = models.EstadoCancelado
	*ahora = t0.Add(49 * time.Hour)
	if n, err := servicio.Procesar(ctx); err != nil || n != 0 {
		t.Fatalf("Procesar = %d, %v; se esperaba 0 sin error", n, err)
	}
	if len(salida.enviados) != 1 {
		t.Errorf("solo debía existir la confirmación previa, hay %d mensajes", len(salida.enviados))
	}
}

func TestUnaCitaReprogramadaRecalculaSusAvisos(t *testing.T) {
	servicio, repo, salida, ahora := nuevoEscenario(t0.Add(72 * time.Hour))
	ctx := context.Background()
	servicio.CitaAgendada(ctx, 1)
	*ahora = t0.Add(25 * time.Hour)
	_, _ = servicio.Procesar(ctx) // sale el recordatorio de 48 h del horario viejo

	// Se reprograma 5 días más tarde: los avisos viejos dejan de ser vigentes y se genera una confirmación nueva.
	*ahora = t0.Add(26 * time.Hour)
	repo.cita.Inicio = t0.Add(26*time.Hour + 120*time.Hour)
	repo.cita.Actualizada = *ahora
	servicio.CitaReprogramada(ctx, 1)
	if n := repo.vigentes(models.NotifConfirmacion); n != 1 {
		t.Errorf("confirmaciones vigentes = %d, se esperaba 1", n)
	}
	if n := repo.vigentes(models.NotifRecordatorio48); n != 0 {
		t.Errorf("el recordatorio del horario viejo sigue vigente (%d)", n)
	}

	// Con el horario nuevo, el recordatorio de 48 h vuelve a salir una sola vez.
	*ahora = repo.cita.Inicio.Add(-47 * time.Hour)
	_, _ = servicio.Procesar(ctx)
	_, _ = servicio.Procesar(ctx)
	if n := repo.vigentes(models.NotifRecordatorio48); n != 1 {
		t.Errorf("recordatorios de 48 h del horario nuevo = %d, se esperaba 1", n)
	}
	if len(salida.enviados) < 4 {
		t.Errorf("mensajes entregados = %d, se esperaban al menos 4", len(salida.enviados))
	}
}

func TestUnaCitaAgendadaDentroDeLaVentanaNoRecibeEseRecordatorio(t *testing.T) {
	// Se agenda con solo 10 h de anticipación: su confirmación ya fue el aviso; no llegan dos mensajes seguidos.
	servicio, repo, _, ahora := nuevoEscenario(t0.Add(10 * time.Hour))
	ctx := context.Background()
	servicio.CitaAgendada(ctx, 1)
	*ahora = t0.Add(time.Minute)
	if _, err := servicio.Procesar(ctx); err != nil {
		t.Fatal(err)
	}
	if repo.vigentes(models.NotifRecordatorio24) != 0 || repo.vigentes(models.NotifRecordatorio48) != 0 {
		t.Error("no debía generar recordatorios para una cita agendada dentro de la ventana")
	}
}

func TestUnFalloDeEntregaQuedaRegistrado(t *testing.T) {
	servicio, repo, salida, _ := nuevoEscenario(t0.Add(72 * time.Hour))
	salida.err = errors.New("proveedor caído")
	servicio.CitaAgendada(context.Background(), 1)

	if len(repo.notificaciones) != 1 || repo.notificaciones[0].Estado != models.NotifFallida || repo.notificaciones[0].Detalle != "proveedor caído" {
		t.Errorf("bitácora = %+v", repo.notificaciones)
	}
}

func TestElMensajeUsaLaHoraLocalYElCanalDelPaciente(t *testing.T) {
	// 15:00 UTC son las 10:00 en Guayaquil.
	servicio, repo, salida, _ := nuevoEscenario(time.Date(2026, 11, 5, 15, 0, 0, 0, time.UTC))
	repo.cita.Canal = "email"
	servicio.CitaAgendada(context.Background(), 1)

	m := salida.enviados[0]
	if m.Canal != "email" || m.Destino != "ana@correo.test" {
		t.Errorf("mensaje = %+v, se esperaba el correo del paciente", m)
	}
	for _, quiere := range []string{"05/11/2026", "10:00", "Ana Prueba", "https://sigo.test/confirmar?token="} {
		if !strings.Contains(m.Texto, quiere) {
			t.Errorf("el texto %q no contiene %q", m.Texto, quiere)
		}
	}
}

func TestElAdaptadorSimuladoNoEnviaSinDestino(t *testing.T) {
	if err := (Simulado{}).Enviar(context.Background(), Mensaje{Canal: "email", Destino: " "}); err == nil {
		t.Error("se esperaba un error por falta de destino")
	}
	if _, err := NuevoNotificador("twilio"); err == nil {
		t.Error("un adaptador real sin credenciales debe rechazarse al arrancar")
	}
	if _, err := NuevoNotificador(""); err != nil {
		t.Errorf("el valor vacío debe usar el simulado: %v", err)
	}
}

func TestTokenDeConfirmacion(t *testing.T) {
	inicio := time.Date(2026, 11, 5, 15, 0, 0, 0, time.UTC)
	token := TokenDeConfirmacion(secretoDePrueba, 7, inicio)

	id, leido, err := LeerToken(secretoDePrueba, token)
	if err != nil || id != 7 || !leido.Equal(inicio) {
		t.Fatalf("LeerToken = %d, %v, %v", id, leido, err)
	}

	otraClave := []byte("otra-clave-distinta-de-32-bytes-o-mas!!")
	adulterado := TokenDeConfirmacion(secretoDePrueba, 8, inicio)
	partes := strings.SplitN(token, ".", 2)
	malos := map[string]string{
		"firmado con otra clave":          TokenDeConfirmacion(otraClave, 7, inicio),
		"carga de otra cita, firma vieja": strings.SplitN(adulterado, ".", 2)[0] + "." + partes[1],
		"sin firma":                       partes[0],
		"vacío":                           "",
		"basura":                          "no.es.un.token",
	}
	for nombre, malo := range malos {
		if _, _, err := LeerToken(secretoDePrueba, malo); !errors.Is(err, ErrTokenInvalido) {
			t.Errorf("%s: err = %v, se esperaba ErrTokenInvalido", nombre, err)
		}
	}
}
