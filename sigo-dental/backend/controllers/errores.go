// Este archivo centraliza las respuestas de error compartidas por los controladores.
package controllers

import (
	"errors"
	"log"
	"strings"
	"unicode/utf8"

	"github.com/gofiber/fiber/v2"
	"sigo-dental-backend/models"
)

// errorInterno registra el detalle en el log del servidor y responde 500 con un mensaje genérico,
// para no filtrar a la persona usuaria información interna (consultas, rutas, errores del driver).
func errorInterno(c *fiber.Ctx, err error) error {
	log.Printf("error interno en %s %s: %v", c.Method(), c.Path(), err)
	return c.Status(fiber.StatusInternalServerError).JSON(models.APIError{
		Status:  fiber.StatusInternalServerError,
		Message: "Error interno del servidor",
	})
}

// errorCliente responde un error 4xx con el contrato estándar de la API.
func errorCliente(c *fiber.Ctx, estado int, mensaje string) error {
	return c.Status(estado).JSON(models.APIError{Status: estado, Message: mensaje})
}

// errorDominio traduce los errores de dominio que devuelve la capa de datos a su respuesta HTTP;
// cualquier otro error es un fallo interno.
func errorDominio(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, models.ErrNoEncontrado):
		return errorCliente(c, fiber.StatusNotFound, "Recurso no encontrado")
	case errors.Is(err, models.ErrDuplicado):
		return errorCliente(c, fiber.StatusConflict, "Ya existe un registro con esos datos")
	case errors.Is(err, models.ErrConflicto):
		return errorCliente(c, fiber.StatusConflict, "La operación entra en conflicto con otro registro")
	case errors.Is(err, models.ErrInmutable):
		return errorCliente(c, fiber.StatusConflict, "El registro es inmutable y no admite cambios")
	case errors.Is(err, models.ErrReferencia):
		return errorCliente(c, fiber.StatusBadRequest, "Los datos enviados no son válidos: una referencia no existe o no corresponde")
	}
	return errorInterno(c, err)
}

// idDeRuta lee el parámetro :id de la ruta; ok es false (y ya se respondió 400) si no es un entero positivo.
func idDeRuta(c *fiber.Ctx) (id int, ok bool) {
	id, err := c.ParamsInt("id")
	if err != nil || id <= 0 {
		return 0, false
	}
	return id, true
}

// cuerpoInvalido responde 400 por un JSON que no se puede leer.
func cuerpoInvalido(c *fiber.Ctx) error {
	return errorCliente(c, fiber.StatusBadRequest, "Cuerpo de petición inválido")
}

// textoObligatorio devuelve el texto sin espacios sobrantes y comprueba que no esté vacío ni exceda max caracteres.
func textoObligatorio(valor string, max int) (string, bool) {
	valor = strings.TrimSpace(valor)
	n := utf8.RuneCountInString(valor)
	return valor, n > 0 && n <= max
}

// textoOpcional es como textoObligatorio pero acepta el texto vacío.
func textoOpcional(valor string, max int) (string, bool) {
	valor = strings.TrimSpace(valor)
	return valor, utf8.RuneCountInString(valor) <= max
}

// esAdminTI indica si la cuenta de la petición tiene el rol de administrador TI.
func esAdminTI(c *fiber.Ctx) bool {
	rol, _ := c.Locals("rol").(string)
	return rol == models.RolAdminTI
}
