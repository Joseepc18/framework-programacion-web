package notificaciones

import (
	"context"
	"fmt"
	"strings"
	"time"

	"sigo-dental-backend/models"
)

// RepositorioCodigos es el contrato de la capa de datos para registrar los códigos enviados al paciente (OTP) en la
// bitácora de notificaciones, que no pertenecen a ninguna cita.
type RepositorioCodigos interface {
	// RegistrarNotificacionCodigo guarda el mensaje como pendiente y devuelve su ID.
	RegistrarNotificacionCodigo(ctx context.Context, n models.NotificacionCodigo) (int, error)
	RegistrarResultado(ctx context.Context, id int, enviada bool, detalle string) error
}

// ConCodigos habilita el envío de códigos de un solo uso (consentimientos y portal del paciente).
func (s *Servicio) ConCodigos(repo RepositorioCodigos) *Servicio {
	s.codigos = repo
	return s
}

// DestinoCodigo es a quién y por dónde se envía un código.
type DestinoCodigo struct {
	PacienteID int
	Canal      string
	Destino    string
	// Referencia describe para qué es el código: "firmar el consentimiento de Exodoncia", "ingresar a tu portal".
	Referencia string
	Tipo       string
}

// EnviarCodigo entrega el código al destinatario y deja el mensaje en la bitácora. Con el adaptador simulado el código
// queda visible en la bitácora para poder probar; con un proveedor real se guarda enmascarado, así quien consulta la
// bitácora no puede usarlo. Devuelve el destino enmascarado para mostrarlo en pantalla.
func (s *Servicio) EnviarCodigo(ctx context.Context, d DestinoCodigo, codigo string, vigencia time.Duration) (string, error) {
	if s.codigos == nil {
		return "", fmt.Errorf("el envío de códigos no está configurado")
	}
	texto := fmt.Sprintf("SIGO-Dental: tu código para %s es %s. Vence en %d minutos. No lo compartas con nadie.", d.Referencia, codigo, int(vigencia.Minutes()))
	guardado := texto
	if _, simulado := s.notificador.(Simulado); !simulado {
		guardado = strings.Replace(texto, codigo, strings.Repeat("•", len(codigo)), 1)
	}
	id, err := s.codigos.RegistrarNotificacionCodigo(ctx, models.NotificacionCodigo{
		PacienteID: d.PacienteID, Tipo: d.Tipo, Canal: d.Canal, Destino: d.Destino, Mensaje: guardado,
	})
	if err != nil {
		return "", err
	}
	if errEnvio := s.notificador.Enviar(ctx, Mensaje{Canal: d.Canal, Destino: d.Destino, Texto: texto}); errEnvio != nil {
		_ = s.codigos.RegistrarResultado(ctx, id, false, errEnvio.Error())
		return "", fmt.Errorf("no se pudo enviar el código: %w", errEnvio)
	}
	return EnmascararDestino(d.Destino), s.codigos.RegistrarResultado(ctx, id, true, "")
}

// EnmascararDestino deja visibles solo los últimos 3 caracteres de un teléfono o los primeros 2 de un correo.
func EnmascararDestino(destino string) string {
	if i := strings.Index(destino, "@"); i > 0 {
		visible := min(2, i)
		return destino[:visible] + strings.Repeat("•", i-visible) + destino[i:]
	}
	if len(destino) <= 3 {
		return destino
	}
	return strings.Repeat("•", len(destino)-3) + destino[len(destino)-3:]
}
