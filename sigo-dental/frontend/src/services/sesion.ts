// Almacén de la sesión activa. El token vive solo en memoria: no se escribe en localStorage ni en
// sessionStorage, así que no queda una copia que un script malicioso o la siguiente persona en un
// equipo compartido pueda leer. Recargar la página cierra la sesión (ver docs/architecture.md).

let token: string | null = null;
let alExpirar: (() => void) | null = null;

export const obtenerToken = () => token;

export const guardarToken = (nuevo: string) => {
  token = nuevo;
};

export const borrarToken = () => {
  token = null;
};

// El contexto de autenticación se registra aquí para enterarse de los 401 que detecta http.ts.
export const alSesionExpirada = (accion: () => void) => {
  alExpirar = accion;
  return () => {
    if (alExpirar === accion) alExpirar = null;
  };
};

// Un 401 con token enviado significa que la sesión ya no vale (expiró o se invalidó).
// Se ignora si entretanto hubo otro inicio de sesión: la respuesta es de un token anterior.
export const notificarSesionExpirada = (tokenEnviado: string) => {
  if (tokenEnviado !== token) return;
  token = null;
  alExpirar?.();
};
