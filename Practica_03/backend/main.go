// Declaramos que este es el paquete principal (main), es decir, el archivo ejecutable que inicia el programa.
package main

import (
	"context"
	// Importamos log para reportar errores fatales de arranque.
	"log"
	// Importamos os y os/signal para detectar cuando se detiene la aplicación (Ctrl+C).
	"os"
	"os/signal"
	"strings"
	// Importamos syscall para escuchar también la señal de terminación del sistema.
	"syscall"

	// Importamos el framework principal Fiber.
	"github.com/gofiber/fiber/v2"
	// Importamos el middleware CORS para gestionar la seguridad entre distintos puertos/dominios.
	"github.com/gofiber/fiber/v2/middleware/cors"
	// Importamos el paquete cache con el cliente de Redis.
	"multicatalogo-backend/cache"
	// Importamos el paquete config para conectarnos a PostgreSQL.
	"multicatalogo-backend/config"
	// Importamos los controladores que se van a instanciar al arrancar.
	"multicatalogo-backend/controllers"
	"multicatalogo-backend/mailer"
	// Importamos el repositorio para inyectarle la conexión a la base de datos.
	"multicatalogo-backend/repository"
	// Importamos nuestro propio paquete de rutas para delegarle la configuración de los endpoints.
	"multicatalogo-backend/routes"
)

// func main es el punto de entrada de la aplicación en Go. Todo comienza a ejecutarse aquí.
func main() {
	// Abrimos el pool de conexiones a PostgreSQL; si la base no responde, no tiene sentido levantar la API.
	pool, err := config.ConectarDB()
	if err != nil {
		log.Fatal(err)
	}
	// Cerramos el pool cuando main termine, liberando todas las conexiones abiertas.
	defer pool.Close()

	// Abrimos la conexión con Redis, el almacenamiento en caché de las consultas de lectura.
	almacen, err := cache.NuevoRedis()
	if err != nil {
		log.Fatal(err)
	}
	// Cerramos el cliente de Redis cuando main termine.
	defer func() { _ = almacen.Cerrar() }()

	// La clave JWT es obligatoria (salvo APP_ENV=dev): sin ella no se firman tokens seguros.
	secreto, err := config.CargarJWTSecret()
	if err != nil {
		log.Fatal(err)
	}

	// --- Inyección de dependencias (Tema 4) ---
	// Construimos una sola vez las implementaciones concretas...
	repositorio := repository.Nuevo(pool)
	correo := mailer.NuevoLog()

	// Una cuenta desactivada pierde acceso aunque su token no haya expirado; ante un error de base se deniega.
	cuentaActiva := func(ctx context.Context, usuarioID int) bool {
		activa, err := repositorio.CuentaActiva(ctx, usuarioID)
		return err == nil && activa
	}

	// ...y se las pasamos a cada controlador, que solo conoce las interfaces que necesita.
	manejadores := routes.Manejadores{
		Auth:          controllers.NuevoAuthController(repositorio, secreto, correo),
		Productos:     controllers.NuevoProductoController(repositorio, almacen),
		Red:           controllers.NuevoRedController(repositorio, almacen),
		Checkout:      controllers.NuevoCheckoutController(repositorio, almacen, correo),
		Health:        controllers.NuevoHealthController(repositorio, almacen),
		Pedidos:       controllers.NuevoPedidoController(repositorio, almacen),
		Liquidaciones: controllers.NuevoLiquidacionController(repositorio),
		Reportes:      controllers.NuevoReporteController(repositorio),
		Usuarios:      controllers.NuevoUsuarioController(repositorio),
		JWTSecret:     secreto,
		CuentaActiva:  cuentaActiva,
	}

	// Instanciamos una nueva aplicación de Fiber y la guardamos en la variable 'app'.
	app := fiber.New()
	origenes := strings.TrimSpace(os.Getenv("CORS_ALLOWED_ORIGINS"))
	if origenes == "" {
		origenes = "http://localhost:5173,http://127.0.0.1:5173"
	}
	if strings.Contains(origenes, "*") {
		log.Fatal("CORS_ALLOWED_ORIGINS requiere orígenes explícitos, sin comodines")
	}

	// Implementamos el middleware CORS a nivel global usando app.Use() para interceptar todas las peticiones entrantes.
	app.Use(cors.New(cors.Config{
		// En producción se indican los orígenes HTTPS del frontend separados por comas.
		AllowOrigins: origenes,
		// Declaramos de forma explícita qué cabeceras se permitirán, incluyendo Authorization para el token JWT.
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
	}))

	// Llamamos a la función SetupRoutes de nuestro paquete 'routes', enviándole la instancia de 'app'
	// y los controladores ya construidos.
	routes.SetupRoutes(app, manejadores)

	// Al recibir Ctrl+C (o la señal de terminación), apagamos Fiber de forma ordenada.
	// Así app.Listen retorna, main termina y se ejecuta el defer que cierra el pool.
	go func() {
		senal := make(chan os.Signal, 1)
		signal.Notify(senal, os.Interrupt, syscall.SIGTERM)
		<-senal
		_ = app.Shutdown()
	}()

	// Escuchamos en el puerto 3001 (3000 lo ocupa otro proyecto); se puede cambiar con la variable PORT.
	puerto := os.Getenv("PORT")
	if puerto == "" {
		puerto = "3001"
	}
	if err := app.Listen(":" + puerto); err != nil {
		log.Println(err)
	}
}
