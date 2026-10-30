// Pruebas del login con un repositorio falso: no hace falta PostgreSQL.
package controllers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"sigo-dental-backend/models"
)

var secretoPrueba = []byte("secreto-de-prueba-con-32-bytes-o-mas")

// repoUsuarios es un doble de UsuarioRepository que sirve una lista fija de cuentas.
type repoUsuarios struct {
	cuentas map[string]models.Usuario
	err     error
}

func (r repoUsuarios) BuscarUsuarioPorEmail(_ context.Context, email string) (models.Usuario, bool, error) {
	if r.err != nil {
		return models.Usuario{}, false, r.err
	}
	u, ok := r.cuentas[email]
	return u, ok, nil
}

// loginPrueba envía el cuerpo a POST /api/login y devuelve el estado y el cuerpo JSON.
func loginPrueba(t *testing.T, repo UsuarioRepository, cuerpo string) (int, map[string]any) {
	t.Helper()
	ctrl, err := NuevoAuthController(repo, secretoPrueba)
	if err != nil {
		t.Fatal(err)
	}
	app := fiber.New()
	app.Post("/api/login", ctrl.Login)

	req := httptest.NewRequest("POST", "/api/login", strings.NewReader(cuerpo))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	var datos map[string]any
	_ = json.NewDecoder(resp.Body).Decode(&datos)
	return resp.StatusCode, datos
}

func TestLogin(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("clave-de-prueba"), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}
	repo := repoUsuarios{cuentas: map[string]models.Usuario{
		"odonto@sigo-dental.test": {ID: 3, Email: "odonto@sigo-dental.test", PasswordHash: string(hash), Rol: models.RolOdontologo, Activo: true},
		"baja@sigo-dental.test":   {ID: 4, Email: "baja@sigo-dental.test", PasswordHash: string(hash), Rol: models.RolAsistente, Activo: false},
	}}

	casos := []struct {
		nombre, cuerpo string
		repo           UsuarioRepository
		estado         int
		mensaje        string
	}{
		{"credenciales válidas", `{"email":"odonto@sigo-dental.test","password":"clave-de-prueba"}`, repo, 200, ""},
		{"correo con espacios alrededor", `{"email":" odonto@sigo-dental.test ","password":"clave-de-prueba"}`, repo, 200, ""},
		{"contraseña incorrecta", `{"email":"odonto@sigo-dental.test","password":"otra"}`, repo, 401, "Credenciales incorrectas"},
		{"correo inexistente", `{"email":"nadie@sigo-dental.test","password":"clave-de-prueba"}`, repo, 401, "Credenciales incorrectas"},
		{"cuenta desactivada", `{"email":"baja@sigo-dental.test","password":"clave-de-prueba"}`, repo, 401, "Credenciales incorrectas"},
		{"cuerpo que no es JSON", `no-es-json`, repo, 400, ""},
		{"sin contraseña", `{"email":"odonto@sigo-dental.test"}`, repo, 400, ""},
		{"sin correo", `{"password":"clave-de-prueba"}`, repo, 400, ""},
		{"falla la base de datos", `{"email":"odonto@sigo-dental.test","password":"clave-de-prueba"}`,
			repoUsuarios{err: errors.New("conexión rechazada")}, 500, "Error interno del servidor"},
	}
	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			estado, datos := loginPrueba(t, caso.repo, caso.cuerpo)
			if estado != caso.estado {
				t.Fatalf("estado = %d, se esperaba %d (%v)", estado, caso.estado, datos)
			}
			if caso.mensaje != "" && datos["message"] != caso.mensaje {
				t.Errorf("message = %v, se esperaba %q", datos["message"], caso.mensaje)
			}
			if caso.estado != 200 {
				if _, hayToken := datos["token"]; hayToken {
					t.Error("una respuesta fallida no debe incluir token")
				}
			}
		})
	}

	t.Run("el token lleva el rol y se puede verificar", func(t *testing.T) {
		_, datos := loginPrueba(t, repo, `{"email":"odonto@sigo-dental.test","password":"clave-de-prueba"}`)
		token, _ := datos["token"].(string)
		if datos["rol"] != models.RolOdontologo || datos["email"] != "odonto@sigo-dental.test" {
			t.Fatalf("respuesta inesperada: %v", datos)
		}
		analizado, err := jwt.Parse(token, func(*jwt.Token) (any, error) { return secretoPrueba, nil },
			jwt.WithValidMethods([]string{"HS256"}))
		if err != nil || !analizado.Valid {
			t.Fatalf("token inválido: %v", err)
		}
		claims := analizado.Claims.(jwt.MapClaims)
		if claims["rol"] != models.RolOdontologo || claims["sub"] != float64(3) {
			t.Errorf("claims inesperados: %v", claims)
		}
	})

	t.Run("no distingue entre usuario inexistente y contraseña errónea", func(t *testing.T) {
		_, noExiste := loginPrueba(t, repo, `{"email":"nadie@sigo-dental.test","password":"x"}`)
		_, malaClave := loginPrueba(t, repo, `{"email":"odonto@sigo-dental.test","password":"x"}`)
		if noExiste["message"] != malaClave["message"] || noExiste["status"] != malaClave["status"] {
			t.Errorf("las respuestas difieren: %v vs %v", noExiste, malaClave)
		}
	})
}
