// Declaramos que este es el paquete principal (main), es decir, el archivo ejecutable que inicia el programa.
package main

import (
	// Importamos el framework principal Fiber.
	"github.com/gofiber/fiber/v2"
	// Importamos el middleware CORS para gestionar la seguridad entre distintos puertos/dominios.
	"github.com/gofiber/fiber/v2/middleware/cors"
	// Importamos nuestro propio paquete de rutas para delegarle la configuración de los endpoints.
	"multicatalogo-backend/routes"
)

// func main es el punto de entrada de la aplicación en Go. Todo comienza a ejecutarse aquí.
func main() {
	// Instanciamos una nueva aplicación de Fiber y la guardamos en la variable 'app'.
	app := fiber.New()

	// Implementamos el middleware CORS a nivel global usando app.Use() para interceptar todas las peticiones entrantes.
	app.Use(cors.New(cors.Config{
		// Configuramos el CORS para permitir únicamente peticiones provenientes del frontend local en el puerto 5173.
		AllowOrigins: "http://localhost:5173",
		// Declaramos de forma explícita qué cabeceras (Headers) se permitirán en la comunicación.
		AllowHeaders: "Origin, Content-Type, Accept",
	}))

	// Llamamos a la función SetupRoutes de nuestro paquete 'routes', enviándole la instancia de nuestra 'app'.
	routes.SetupRoutes(app)

	// Ponemos a la aplicación a escuchar peticiones en el puerto 3000 de la máquina local. (Bloquea el hilo de ejecución).
	app.Listen(":3000")
}





// // main.go
// package main

// import (
// 	"github.com/gofiber/fiber/v2"
// 	"github.com/gofiber/fiber/v2/middleware/cors"
// )

// // Estructuras de datos
// type LoginRequest struct {
// 	Email    string `json:"email"`
// 	Password string `json:"password"`
// }

// type Producto struct {
// 	ID     int     `json:"id"`
// 	Nombre string  `json:"nombre"`
// 	Precio float64 `json:"precio"`
// 	Img    string  `json:"img"`
// }

// func main() {
// 	app := fiber.New()

// 	// Middleware CORS para permitir peticiones desde React (Vite corre en el puerto 5173)
// 	app.Use(cors.New(cors.Config{
// 		AllowOrigins: "http://localhost:5173",
// 		AllowHeaders: "Origin, Content-Type, Accept",
// 	}))

// 	// Endpoint para Login
// 	app.Post("/api/login", func(c *fiber.Ctx) error {
// 		var req LoginRequest
// 		if err := c.BodyParser(&req); err != nil {
// 			return c.Status(400).JSON(fiber.Map{"error": "Cuerpo de petición inválido"})
// 		}

// 		if req.Email == "admin@upse.edu.ec" && req.Password == "123456" {
// 			return c.JSON(fiber.Map{"token": "fake-jwt-token-123", "email": req.Email})
// 		}
// 		return c.Status(401).JSON(fiber.Map{"error": "Credenciales incorrectas"})
// 	})

// 	// Endpoint para obtener Productos
// 	app.Get("/api/productos", func(c *fiber.Ctx) error {
// 		productos := []Producto{
// 			{ID: 1, Nombre: "Serum Revitalizante", Precio: 45.00, Img: "https://picsum.photos/seed/serum/150"},
// 			{ID: 2, Nombre: "Crema Hidratante Pro", Precio: 32.50, Img: "https://picsum.photos/seed/crema/150"},
// 			{ID: 3, Nombre: "Tónico Purificante", Precio: 28.00, Img: "https://picsum.photos/seed/tonico/150"},
// 			{ID: 4, Nombre: "Mascarilla Nocturna", Precio: 50.00, Img: "https://picsum.photos/seed/mascarilla/150"},
// 		}
// 		return c.JSON(productos)
// 	})

// 	app.Listen(":3000")
// }