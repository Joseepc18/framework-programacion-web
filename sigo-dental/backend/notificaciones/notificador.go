// Paquete notificaciones: avisos automáticos de citas (F1-4: RF-11). El servicio redacta y registra los
// mensajes y se apoya en un Notificador para entregarlos; cambiar de proveedor es cambiar de adaptador.
package notificaciones

import (
	"context"
	"fmt"
	"log"
	"strings"
)

// Mensaje es lo que se entrega por un canal: whatsapp, sms o email.
type Mensaje struct {
	Canal   string
	Destino string
	Texto   string
}

// Notificador es el puerto de salida: entrega un mensaje por su canal o devuelve el motivo del fallo.
type Notificador interface {
	Enviar(ctx context.Context, m Mensaje) error
}

// Simulado es el adaptador de desarrollo: no contacta a ningún proveedor, solo deja el mensaje en el log del
// servidor (la bitácora de la base de datos lo registra siempre). Los adaptadores reales (Meta Cloud API,
// Twilio, SMTP) implementan la misma interfaz y se eligen con NOTIFICADOR cuando haya credenciales.
type Simulado struct{}

// Enviar registra el mensaje en el log; falla si no hay a dónde enviarlo, como lo haría un proveedor real.
func (Simulado) Enviar(_ context.Context, m Mensaje) error {
	if strings.TrimSpace(m.Destino) == "" {
		return fmt.Errorf("el paciente no tiene destino registrado para el canal %s", m.Canal)
	}
	log.Printf("[notificación simulada] %s a %s: %s", m.Canal, m.Destino, m.Texto)
	return nil
}

// NuevoNotificador devuelve el adaptador con ese nombre (variable NOTIFICADOR; vacío = simulado).
func NuevoNotificador(nombre string) (Notificador, error) {
	switch strings.ToLower(strings.TrimSpace(nombre)) {
	case "", "simulado":
		return Simulado{}, nil
	}
	return nil, fmt.Errorf("NOTIFICADOR %q no está disponible: los adaptadores reales requieren credenciales del proveedor", nombre)
}
