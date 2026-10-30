// Paquete principal: punto de entrada del ejecutable. Aquí se arma la inyección de dependencias.
package main

import (
	"log"
	// os y os/signal detectan cuando se detiene la aplicación (Ctrl+C).
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/gofiber/fiber/v2"
	// CORS gestiona la seguridad entre distintos puertos y dominios.
	"github.com/gofiber/fiber/v2/middleware/cors"
	"sigo-dental-backend/cache"
	"sigo-dental-backend/config"
	"sigo-dental-backend/controllers"
	"sigo-dental-backend/repository"
	"sigo-dental-backend/routes"
)

func main() {
	// Si la base no responde, no tiene sentido levantar la API.
	pool, err := config.ConectarDB()
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	// Sin una JWT_SECRET válida (presente y de al menos 32 caracteres) la API no arranca.
	// ConectarDB ya cargó el .env, así que el secreto también se puede definir ahí.
	secretoJWT, err := config.CargarJWTSecret()
	if err != nil {
		log.Fatal(err)
	}

	almacen, err := cache.NuevoRedis()
	if err != nil {
		log.Fatal(err)
	}
	defer func() { _ = almacen.Cerrar() }()

	// --- Inyección de dependencias ---
	// Se construyen una sola vez las implementaciones concretas...
	repositorio := repository.Nuevo(pool)

	// ...y se pasan a cada controlador, que solo conoce las interfaces que necesita.
	auth, err := controllers.NuevoAuthController(repositorio, secretoJWT)
	if err != nil {
		log.Fatal(err)
	}
	manejadores := routes.Manejadores{
		Health:    controllers.NuevoHealthController(repositorio, almacen),
		Auth:      auth,
		JWTSecret: secretoJWT,
	}

	app := fiber.New()
	origenes := strings.TrimSpace(os.Getenv("CORS_ALLOWED_ORIGINS"))
	if origenes == "" {
		origenes = "http://localhost:5173,http://127.0.0.1:5173"
	}
	if strings.Contains(origenes, "*") {
		log.Fatal("CORS_ALLOWED_ORIGINS requiere orígenes explícitos, sin comodines")
	}

	app.Use(cors.New(cors.Config{
		// En producción se indican los orígenes HTTPS del frontend separados por comas.
		AllowOrigins: origenes,
		// Authorization queda permitida para el token JWT.
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
	}))

	routes.SetupRoutes(app, manejadores)

	// Al recibir Ctrl+C (o la señal de terminación) Fiber se apaga de forma ordenada:
	// app.Listen retorna, main termina y se ejecutan los defer que cierran las conexiones.
	go func() {
		senal := make(chan os.Signal, 1)
		signal.Notify(senal, os.Interrupt, syscall.SIGTERM)
		<-senal
		_ = app.Shutdown()
	}()

	// Puerto 3001 por defecto (3000 lo ocupa otro proyecto); se cambia con la variable PORT.
	puerto := os.Getenv("PORT")
	if puerto == "" {
		puerto = "3001"
	}
	if err := app.Listen(":" + puerto); err != nil {
		log.Println(err)
	}
}
