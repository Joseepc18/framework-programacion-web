// Paquete principal: punto de entrada del ejecutable. Aquí se arma la inyección de dependencias.
package main

import (
	"context"
	"log"
	// os y os/signal detectan cuando se detiene la aplicación (Ctrl+C).
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
	// Incrusta la base de zonas horarias: la agenda la necesita aunque el sistema no la tenga (Windows, contenedores mínimos).
	_ "time/tzdata"

	"github.com/gofiber/fiber/v2"
	// CORS gestiona la seguridad entre distintos puertos y dominios.
	"github.com/gofiber/fiber/v2/middleware/cors"
	"sigo-dental-backend/almacenamiento"
	"sigo-dental-backend/cache"
	"sigo-dental-backend/config"
	"sigo-dental-backend/controllers"
	"sigo-dental-backend/middleware"
	"sigo-dental-backend/notificaciones"
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

	// Avisos de citas (F1-4): el adaptador se elige con NOTIFICADOR; sin credenciales de proveedor se usa el simulado.
	notificador, err := notificaciones.NuevoNotificador(os.Getenv("NOTIFICADOR"))
	if err != nil {
		log.Fatal(err)
	}
	urlFrontend := strings.TrimRight(strings.TrimSpace(os.Getenv("FRONTEND_URL")), "/")
	if urlFrontend == "" {
		urlFrontend = "http://localhost:5173"
	}
	avisos := notificaciones.NuevoServicio(repositorio, notificador, secretoJWT, urlFrontend)

	// Archivos del paciente (F2-1): disco en desarrollo, S3 en producción (ALMACEN_ARCHIVOS).
	almacenArchivos, err := almacenamiento.NuevoDesdeEntorno()
	if err != nil {
		log.Fatal(err)
	}
	archivos := almacenamiento.NuevoServicio(almacenArchivos, repositorio)

	// La auditoría anota la IP del cliente; detrás de un proxy de confianza se toma de PROXY_HEADER (RNF-06).
	saltosProxy, _ := strconv.Atoi(os.Getenv("PROXY_SALTOS"))
	middleware.ConfigurarIPDeProxy(os.Getenv("PROXY_HEADER"), saltosProxy)

	manejadores := routes.Manejadores{
		Health:     controllers.NuevoHealthController(repositorio, almacen),
		Auth:       auth,
		Catalogo:   controllers.NuevoCatalogoController(repositorio),
		Pacientes:  controllers.NuevoPacientesController(repositorio),
		Citas:      controllers.NuevoCitasController(repositorio, avisos, secretoJWT),
		Cobros:     controllers.NuevoCobrosController(repositorio),
		Expediente: controllers.NuevoExpedienteController(repositorio, archivos),
		Auditoria:  controllers.NuevoAuditoriaController(repositorio),
		Auditor:    repositorio,
		JWTSecret:  secretoJWT,
	}

	// El límite del cuerpo (16 MB) deja subir archivos de hasta 15 MB; Fiber trae 4 MB por defecto.
	app := fiber.New(fiber.Config{BodyLimit: 16 * 1024 * 1024})
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

	// El planificador envía los recordatorios de 48 h y 24 h; se detiene junto con la API.
	ctxPlanificador, detenerPlanificador := context.WithCancel(context.Background())
	defer detenerPlanificador()
	go planificarRecordatorios(ctxPlanificador, avisos, intervaloDeRecordatorios())

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

// intervaloDeRecordatorios lee cada cuánto revisa el planificador (segundos en NOTIFICACIONES_INTERVALO; 60 por defecto).
func intervaloDeRecordatorios() time.Duration {
	segundos, err := strconv.Atoi(os.Getenv("NOTIFICACIONES_INTERVALO"))
	if err != nil || segundos < 5 {
		segundos = 60
	}
	return time.Duration(segundos) * time.Second
}

// planificarRecordatorios revisa las citas próximas al arrancar y luego cada intervalo, hasta que se cancele el contexto.
// Procesar es idempotente, así que una revisión repetida o una reiniciada no duplica mensajes.
func planificarRecordatorios(ctx context.Context, servicio *notificaciones.Servicio, intervalo time.Duration) {
	revisar := func() {
		if enviados, err := servicio.Procesar(ctx); err != nil {
			log.Printf("recordatorios: %v", err)
		} else if enviados > 0 {
			log.Printf("recordatorios enviados: %d", enviados)
		}
	}
	revisar()
	reloj := time.NewTicker(intervalo)
	defer reloj.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-reloj.C:
			revisar()
		}
	}
}
