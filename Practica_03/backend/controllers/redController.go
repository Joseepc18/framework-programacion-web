// Seguimos dentro del paquete controllers, ahora para la sección de la red multinivel.
package controllers

import (
	// Importamos context para propagar el contexto de las operaciones de caché.
	"context"
	// Importamos encoding/json para serializar y deserializar el árbol de referidos.
	"encoding/json"
	// Importamos time para definir el tiempo de vida (TTL) de la caché.
	"time"

	// Importamos Fiber para manejar la respuesta HTTP.
	"github.com/gofiber/fiber/v2"
)

// RedController agrupa los manejadores de la red multinivel, con el repositorio y la caché
// inyectados al arrancar el servidor.
type RedController struct {
	// repo es el acceso a PostgreSQL.
	repo DBRepository
	// cache es el almacenamiento en caché (Redis).
	cache CacheRepository
}

// NuevoRedController construye el controlador con las dependencias que recibirá.
func NuevoRedController(repo DBRepository, cache CacheRepository) *RedController {
	return &RedController{repo: repo, cache: cache}
}

// GetRed devuelve el árbol completo de la red multinivel del usuario.
// Reconstruir el árbol es costoso, así que aplica Cache-Aside con un TTL de 10 minutos.
func (ctrl *RedController) GetRed(c *fiber.Ctx) error {
	// Tiempo de vida del árbol guardado en Redis.
	const ttl = 10 * time.Minute

	// El árbol pertenece a un usuario: su email identifica la clave en Redis.
	// Si la petición no envía el email, se usa una clave general.
	usuario := c.Query("email")
	if usuario == "" {
		usuario = "global"
	}
	clave := "red:tree:" + usuario

	ctx := context.Background()

	// 1) Cache Hit: devolvemos el árbol guardado sin reconstruirlo ni tocar PostgreSQL.
	if datos, ok := ctrl.cache.Obtener(ctx, clave); ok {
		c.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
		return c.Send(datos)
	}

	// 2) Cache Miss: consultamos y reconstruimos el árbol desde la base de datos.
	red, err := ctrl.repo.ObtenerRed()
	if err != nil {
		// Si la consulta falla, respondemos HTTP 500.
		return errorInterno(c, err)
	}

	// 3) Guardamos el árbol serializado en Redis durante 10 minutos.
	if datos, err := json.Marshal(red); err == nil {
		ctrl.cache.Guardar(ctx, clave, datos, ttl)
	}

	// Fiber serializa la estructura recursiva de referidos a JSON.
	return c.JSON(red)
}
