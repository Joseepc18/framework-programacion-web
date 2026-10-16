// Declaramos el paquete mailer: envío de correos transaccionales.
// Hoy los "envía" escribiéndolos en el log; cambiar a SMTP real solo exige otra implementación
// de la misma interfaz, sin tocar los controladores.
package mailer

import (
	"context"
	"log"
)

// Log simula el envío de correos escribiéndolos en el log del servidor.
type Log struct{}

// NuevoLog construye el mailer de simulación.
func NuevoLog() *Log {
	return &Log{}
}

// EnviarBienvenida simula el correo de bienvenida tras el registro.
func (Log) EnviarBienvenida(_ context.Context, nombre, email string) {
	log.Printf("[mailer] bienvenida a %s <%s>", nombre, email)
}

// EnviarConfirmacionPedido simula el correo de confirmación de un pedido.
func (Log) EnviarConfirmacionPedido(_ context.Context, email, numero string, total float64) {
	log.Printf("[mailer] confirmación del pedido %s por $%.2f a <%s>", numero, total, email)
}
