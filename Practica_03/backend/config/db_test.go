package config

import (
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

func TestConexionDBPriorizaURLProveedor(t *testing.T) {
	url := "postgres://usuario:clave@db.example.com/catalogo?sslmode=require"
	t.Setenv("DATABASE_URL", url)
	t.Setenv("DB_SSLMODE", "disable")
	if obtenido := cadenaConexionDB(); obtenido != url {
		t.Fatal("DATABASE_URL debe conservarse sin cambios")
	}
}

func TestConexionDBVariablesLocalesYSSL(t *testing.T) {
	for _, modo := range []string{"", "require"} {
		t.Run("ssl_"+modo, func(t *testing.T) {
			t.Setenv("DATABASE_URL", "")
			t.Setenv("DB_HOST", "localhost")
			t.Setenv("DB_PORT", "5432")
			t.Setenv("DB_USER", "usuario@ejemplo")
			t.Setenv("DB_PASSWORD", "p@ss:/?#%& palabra")
			t.Setenv("DB_NAME", "catalogo")
			t.Setenv("DB_SSLMODE", modo)
			config, err := pgxpool.ParseConfig(cadenaConexionDB())
			if err != nil {
				t.Fatal(err)
			}
			if config.ConnConfig.User != "usuario@ejemplo" || config.ConnConfig.Password != "p@ss:/?#%& palabra" {
				t.Fatal("las credenciales deben conservar sus caracteres especiales")
			}
			if config.ConnConfig.Database != "catalogo" || config.ConnConfig.Host != "localhost" || config.ConnConfig.Port != 5432 {
				t.Fatal("la configuración local de PostgreSQL no se conservó")
			}
			if (config.ConnConfig.TLSConfig != nil) != (modo == "require") {
				t.Fatal("el modo SSL no se aplicó")
			}
		})
	}
}
