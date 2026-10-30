package middleware

import (
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"sigo-dental-backend/models"
)

var secretoPrueba = []byte("secreto-de-prueba-con-32-bytes-o-mas")

// peticion ejecuta GET /p contra una ruta protegida y devuelve el estado y el cuerpo de error.
// Si roles es nil, solo se aplica RequireAuth.
func peticion(t *testing.T, cabecera string, roles []string) (int, models.APIError) {
	t.Helper()
	app := fiber.New()
	cadena := []fiber.Handler{RequireAuth(secretoPrueba)}
	if roles != nil {
		cadena = append(cadena, RequireRoles(roles...))
	}
	cadena = append(cadena, func(c *fiber.Ctx) error { return c.SendStatus(fiber.StatusOK) })
	app.Get("/p", cadena...)

	req := httptest.NewRequest("GET", "/p", nil)
	if cabecera != "" {
		req.Header.Set("Authorization", cabecera)
	}
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	var cuerpo models.APIError
	_ = json.NewDecoder(resp.Body).Decode(&cuerpo)
	return resp.StatusCode, cuerpo
}

// firmar crea un token con claims arbitrarios, para probar casos que GenerarToken no produce.
func firmar(t *testing.T, metodo jwt.SigningMethod, clave interface{}, claims jwt.MapClaims) string {
	t.Helper()
	token, err := jwt.NewWithClaims(metodo, claims).SignedString(clave)
	if err != nil {
		t.Fatal(err)
	}
	return token
}

func TestRequireAuth(t *testing.T) {
	enUnaHora := time.Now().Add(time.Hour).Unix()
	valido, _ := GenerarToken(secretoPrueba, 1, "a@sigo-dental.test", models.RolAdminTI)
	otraClave, _ := GenerarToken([]byte("otra-clave-distinta-de-32-bytes-o-mas"), 1, "a@sigo-dental.test", models.RolAdminTI)
	expirado := firmar(t, jwt.SigningMethodHS256, secretoPrueba, jwt.MapClaims{
		"sub": 1, "rol": models.RolAdminTI, "exp": time.Now().Add(-time.Minute).Unix(),
	})
	sinExp := firmar(t, jwt.SigningMethodHS256, secretoPrueba, jwt.MapClaims{
		"sub": 1, "rol": models.RolAdminTI,
	})
	sinRol := firmar(t, jwt.SigningMethodHS256, secretoPrueba, jwt.MapClaims{
		"sub": 1, "exp": enUnaHora,
	})
	sinSub := firmar(t, jwt.SigningMethodHS256, secretoPrueba, jwt.MapClaims{
		"rol": models.RolAdminTI, "exp": enUnaHora,
	})
	algNone := firmar(t, jwt.SigningMethodNone, jwt.UnsafeAllowNoneSignatureType, jwt.MapClaims{
		"sub": 1, "rol": models.RolAdminTI, "exp": enUnaHora,
	})
	hs512 := firmar(t, jwt.SigningMethodHS512, secretoPrueba, jwt.MapClaims{
		"sub": 1, "rol": models.RolAdminTI, "exp": enUnaHora,
	})

	casos := []struct {
		nombre, cabecera string
		quiere           int
	}{
		{"sin cabecera", "", 401},
		{"sin esquema Bearer", "Basic " + valido, 401},
		{"solo la palabra Bearer", "Bearer", 401},
		{"token basura", "Bearer abc.def.ghi", 401},
		{"firmado con otra clave", "Bearer " + otraClave, 401},
		{"expirado", "Bearer " + expirado, 401},
		{"sin expiración", "Bearer " + sinExp, 401},
		{"sin claim de rol", "Bearer " + sinRol, 401},
		{"sin claim sub", "Bearer " + sinSub, 401},
		{"algoritmo none", "Bearer " + algNone, 401},
		{"algoritmo HS512", "Bearer " + hs512, 401},
		{"token válido", "Bearer " + valido, 200},
		{"esquema en minúsculas", "bearer " + valido, 200},
	}
	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			if got, _ := peticion(t, caso.cabecera, nil); got != caso.quiere {
				t.Errorf("estado = %d, se esperaba %d", got, caso.quiere)
			}
		})
	}
}

func TestRequireAuthUsaElContratoDeErrores(t *testing.T) {
	estado, cuerpo := peticion(t, "", nil)
	if estado != 401 || cuerpo.Status != 401 || cuerpo.Message == "" {
		t.Fatalf("cuerpo de error inesperado: %d %+v", estado, cuerpo)
	}
}

func TestRequireAuthPublicaLaIdentidad(t *testing.T) {
	token, _ := GenerarToken(secretoPrueba, 7, "o@sigo-dental.test", models.RolOdontologo)
	app := fiber.New()
	app.Get("/p", RequireAuth(secretoPrueba), func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"id": c.Locals("usuario_id"), "email": c.Locals("email"), "rol": c.Locals("rol"),
		})
	})
	req := httptest.NewRequest("GET", "/p", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	var cuerpo struct {
		ID    int    `json:"id"`
		Email string `json:"email"`
		Rol   string `json:"rol"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&cuerpo); err != nil {
		t.Fatal(err)
	}
	if cuerpo.ID != 7 || cuerpo.Email != "o@sigo-dental.test" || cuerpo.Rol != models.RolOdontologo {
		t.Fatalf("identidad inesperada: %+v", cuerpo)
	}
}

func TestRequireRoles(t *testing.T) {
	token := func(rol string) string {
		t, _ := GenerarToken(secretoPrueba, 1, rol+"@sigo-dental.test", rol)
		return t
	}

	casos := []struct {
		nombre, token string
		permitidos    []string
		quiere        int
	}{
		{"rol permitido", token(models.RolOdontologo), []string{models.RolOdontologo}, 200},
		{"uno de varios permitidos", token(models.RolAsistente), []string{models.RolOdontologo, models.RolAsistente}, 200},
		{"rol no permitido", token(models.RolAuditor), []string{models.RolOdontologo}, 403},
		{"rol desconocido", token("superusuario"), []string{models.RolOdontologo}, 403},
		{"lista vacía", token(models.RolAdminTI), []string{}, 403},
	}
	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			estado, cuerpo := peticion(t, "Bearer "+caso.token, caso.permitidos)
			if estado != caso.quiere {
				t.Fatalf("estado = %d, se esperaba %d", estado, caso.quiere)
			}
			if caso.quiere == 403 && cuerpo.Status != 403 {
				t.Errorf("el cuerpo no usa el contrato de errores: %+v", cuerpo)
			}
		})
	}

	t.Run("sin token es 401 y no 403", func(t *testing.T) {
		if estado, _ := peticion(t, "", []string{models.RolAdminTI}); estado != 401 {
			t.Fatalf("estado = %d, se esperaba 401", estado)
		}
	})
}

func TestRequireRolesSinAutenticacionPreviaFallaCerrado(t *testing.T) {
	app := fiber.New()
	app.Get("/p", RequireRoles(models.RolAdminTI), func(c *fiber.Ctx) error { return c.SendStatus(200) })
	resp, err := app.Test(httptest.NewRequest("GET", "/p", nil))
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 403 {
		t.Fatalf("estado = %d, se esperaba 403", resp.StatusCode)
	}
}
