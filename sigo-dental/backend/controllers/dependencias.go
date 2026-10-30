// Paquete controllers: este archivo declara las dependencias que los controladores reciben
// inyectadas al arrancar el servidor. Al depender de interfaces y no de implementaciones,
// cada controlador se prueba con un doble y no sabe cómo se guardan los datos.
package controllers

import "context"

// DBRepository es el contrato que debe cumplir la capa de datos (PostgreSQL).
// Crecerá con un contrato pequeño por cada área del negocio.
type DBRepository interface {
	// Ping comprueba que la base de datos responde (health check).
	Ping(ctx context.Context) error
}

// CacheRepository es el contrato de la caché (Redis).
type CacheRepository interface {
	// Ping comprueba que Redis responde (health check).
	Ping(ctx context.Context) error
}
