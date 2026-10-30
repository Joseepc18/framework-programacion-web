package notificaciones

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// ErrTokenInvalido: el enlace de confirmación está alterado, vencido o ya no corresponde a la cita.
var ErrTokenInvalido = errors.New("enlace de confirmación inválido o vencido")

// TokenDeConfirmacion firma "cita:inicio" con HMAC-SHA256 y la clave del servidor. Incluir el inicio hace que
// el enlace deje de servir si la cita se reprograma, y la firma impide fabricar enlaces de otras citas.
func TokenDeConfirmacion(secreto []byte, citaID int, inicio time.Time) string {
	carga := fmt.Sprintf("%d:%d", citaID, inicio.Unix())
	return base64.RawURLEncoding.EncodeToString([]byte(carga)) + "." + firmar(secreto, carga)
}

func firmar(secreto []byte, carga string) string {
	h := hmac.New(sha256.New, secreto)
	h.Write([]byte(carga))
	return base64.RawURLEncoding.EncodeToString(h.Sum(nil))
}

// LeerToken comprueba la firma y devuelve la cita y el inicio con los que se emitió el enlace.
func LeerToken(secreto []byte, token string) (citaID int, inicio time.Time, err error) {
	parteCarga, parteFirma, ok := strings.Cut(token, ".")
	if !ok {
		return 0, time.Time{}, ErrTokenInvalido
	}
	bruto, errCarga := base64.RawURLEncoding.DecodeString(parteCarga)
	if errCarga != nil {
		return 0, time.Time{}, ErrTokenInvalido
	}
	carga := string(bruto)
	if !hmac.Equal([]byte(firmar(secreto, carga)), []byte(parteFirma)) {
		return 0, time.Time{}, ErrTokenInvalido
	}
	idTexto, inicioTexto, ok := strings.Cut(carga, ":")
	id, errID := strconv.Atoi(idTexto)
	unix, errInicio := strconv.ParseInt(inicioTexto, 10, 64)
	if !ok || errID != nil || errInicio != nil || id <= 0 {
		return 0, time.Time{}, ErrTokenInvalido
	}
	return id, time.Unix(unix, 0).UTC(), nil
}
