package cache

import "testing"

func TestRedisURLAutenticacionYTLS(t *testing.T) {
	for _, esquema := range []string{"redis", "rediss"} {
		t.Run(esquema, func(t *testing.T) {
			t.Setenv("REDIS_URL", esquema+"://usuario:p%40ss@redis.example.com:6380/2")
			t.Setenv("REDIS_ADDR", "localhost:6379")
			opciones, err := opcionesRedis()
			if err != nil {
				t.Fatal(err)
			}
			if opciones.Addr != "redis.example.com:6380" || opciones.Username != "usuario" || opciones.Password != "p@ss" || opciones.DB != 2 {
				t.Fatal("REDIS_URL debe conservar dirección, credenciales y base")
			}
			if (opciones.TLSConfig != nil) != (esquema == "rediss") {
				t.Fatal("rediss debe habilitar TLS")
			}
		})
	}
}

func TestRedisDireccionLocal(t *testing.T) {
	t.Setenv("REDIS_URL", "")
	for _, direccion := range []string{"", "redis:6379"} {
		t.Setenv("REDIS_ADDR", direccion)
		opciones, err := opcionesRedis()
		if err != nil {
			t.Fatal(err)
		}
		esperada := direccion
		if esperada == "" {
			esperada = "localhost:6379"
		}
		if opciones.Addr != esperada {
			t.Fatal("no se conservó la dirección local de Redis")
		}
	}
}

func TestRedisURLInvalida(t *testing.T) {
	t.Setenv("REDIS_URL", "https://usuario:secreto@redis.example.com")
	if _, err := opcionesRedis(); err == nil || err.Error() != "REDIS_URL no válida; use redis:// o rediss://" {
		t.Fatal("una URL inválida debe fallar sin exponer credenciales")
	}
}
