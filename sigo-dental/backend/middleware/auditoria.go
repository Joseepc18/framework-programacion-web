// Este archivo registra en la bitácora inmutable cada acceso a datos clínicos (F2-1: RNF-06).
package middleware

import (
	"context"
	"log"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"sigo-dental-backend/models"
)

// Auditor es el contrato de la capa de datos que guarda los eventos de auditoría.
type Auditor interface {
	RegistrarAcceso(ctx context.Context, e models.EventoAuditoria) error
}

const (
	claveAuditoriaPaciente = "auditoria_paciente_id"
	claveAuditoriaEntidad  = "auditoria_entidad_id"
)

// MarcarAuditoria lo llama un controlador cuando el paciente o la entidad afectados solo se conocen al procesar la
// petición (por ejemplo, el ID del registro recién creado). Un valor menor o igual a cero se ignora.
func MarcarAuditoria(c *fiber.Ctx, pacienteID, entidadID int) {
	if pacienteID > 0 {
		c.Locals(claveAuditoriaPaciente, pacienteID)
	}
	if entidadID > 0 {
		c.Locals(claveAuditoriaEntidad, entidadID)
	}
}

type configAuditoria struct {
	paramPaciente string
	paramEntidad  string
}

// OpcionAuditoria ajusta de dónde sale el paciente o la entidad del evento.
type OpcionAuditoria func(*configAuditoria)

// PacienteDeRuta toma el paciente del parámetro de la ruta (por ejemplo "id" en /api/pacientes/:id/...).
func PacienteDeRuta(param string) OpcionAuditoria {
	return func(cfg *configAuditoria) { cfg.paramPaciente = param }
}

// EntidadDeRuta toma el ID de la entidad del parámetro de la ruta.
func EntidadDeRuta(param string) OpcionAuditoria {
	return func(cfg *configAuditoria) { cfg.paramEntidad = param }
}

// Auditar registra el acceso cuando la petición terminó bien (estado menor a 400). Debe ir después de RequireAuth y
// de RequireRoles. Falla cerrado: si la bitácora no acepta el evento, una consulta o exportación no devuelve los
// datos (responde 500), porque entregar información clínica sin dejar rastro incumpliría el RNF-06. Una escritura que
// ya se aplicó no se puede deshacer aquí, así que solo se deja constancia en el log del servidor.
func Auditar(auditor Auditor, accion, entidad string, opciones ...OpcionAuditoria) fiber.Handler {
	var cfg configAuditoria
	for _, opcion := range opciones {
		opcion(&cfg)
	}
	return func(c *fiber.Ctx) error {
		if err := c.Next(); err != nil {
			return err
		}
		if c.Response().StatusCode() >= fiber.StatusBadRequest {
			return nil
		}

		evento := models.EventoAuditoria{
			Accion:  accion,
			Entidad: entidad,
			IP:      resolverIP(c),
			Metodo:  c.Method(),
			Ruta:    c.Path(),
		}
		evento.UsuarioID, _ = c.Locals("usuario_id").(int)
		evento.Email, _ = c.Locals("email").(string)
		evento.Rol, _ = c.Locals("rol").(string)
		evento.PacienteID = identificador(c, cfg.paramPaciente, claveAuditoriaPaciente)
		evento.EntidadID = identificador(c, cfg.paramEntidad, claveAuditoriaEntidad)

		// El contexto de la petición puede cancelarse al cerrar la conexión; la auditoría no debe perderse por eso.
		ctx, cancelar := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancelar()
		if err := auditor.RegistrarAcceso(ctx, evento); err != nil {
			log.Printf("auditoría: no se pudo registrar %s %s: %v", accion, c.Path(), err)
			if accion == models.AccionConsulta || accion == models.AccionExportacion {
				c.Response().ResetBody()
				return errorAPI(c, fiber.StatusInternalServerError, "No se pudo registrar la auditoría del acceso")
			}
		}
		return nil
	}
}

// identificador lee primero el parámetro de la ruta y, si no hay, el valor que dejó el controlador.
func identificador(c *fiber.Ctx, param, clave string) *int {
	if param != "" {
		if id, err := strconv.Atoi(c.Params(param)); err == nil && id > 0 {
			return &id
		}
	}
	if id, ok := c.Locals(clave).(int); ok && id > 0 {
		return &id
	}
	return nil
}

// resolverIP decide qué dirección se anota como origen. Por defecto es la de la conexión.
var resolverIP = func(c *fiber.Ctx) string { return c.Context().RemoteIP().String() }

// ConfigurarIPDeProxy hace que la auditoría tome la IP del cliente de una cabecera que añade el proxy de confianza
// (en Render, X-Forwarded-For). Se cuentan las entradas desde la derecha, que son las que añaden los proxies propios:
// con saltos = 1 se toma la última (la que anotó el primer proxy), y las anteriores, que el cliente puede falsear, se
// ignoran. Con la cabecera vacía se usa la dirección de la conexión.
func ConfigurarIPDeProxy(cabecera string, saltos int) {
	cabecera = strings.TrimSpace(cabecera)
	if cabecera == "" {
		resolverIP = func(c *fiber.Ctx) string { return c.Context().RemoteIP().String() }
		return
	}
	if saltos < 1 {
		saltos = 1
	}
	resolverIP = func(c *fiber.Ctx) string {
		partes := strings.Split(c.Get(cabecera), ",")
		indice := len(partes) - saltos
		if indice < 0 {
			indice = 0
		}
		if ip := strings.TrimSpace(partes[indice]); ip != "" && net.ParseIP(ip) != nil {
			return ip
		}
		return c.Context().RemoteIP().String()
	}
}

// IPDelCliente devuelve la IP de origen de la petición con la misma regla que la auditoría (proxy de confianza).
func IPDelCliente(c *fiber.Ctx) string { return resolverIP(c) }
