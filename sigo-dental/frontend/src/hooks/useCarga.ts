import { useCallback, useEffect, useState } from 'react';
import { ApiError } from '../services/http';

export const mensajeDeApi = (error: unknown) =>
  error instanceof ApiError ? error.message : 'No se pudo conectar con la API. Verifica tu conexión e inténtalo de nuevo.';

interface Resultado<T> {
  token: string;
  datos: T | null;
  error: string;
}

// Carga datos al montar y cuando cambia `clave`; `recargar` vuelve a pedirlos (por ejemplo tras guardar).
// Mientras recarga conserva los datos anteriores, y `cargando` es verdadero hasta que llega la respuesta
// de la consulta vigente (se deriva del token, sin cambiar estado dentro del efecto).
export function useCarga<T>(cargar: () => Promise<T>, clave: string) {
  const [version, setVersion] = useState(0);
  const [resultado, setResultado] = useState<Resultado<T> | null>(null);
  const token = `${clave}#${version}`;

  useEffect(() => {
    let vigente = true;
    cargar()
      .then((datos) => vigente && setResultado({ token, datos, error: '' }))
      .catch((fallo) => vigente && setResultado({ token, datos: null, error: mensajeDeApi(fallo) }));
    return () => {
      vigente = false;
    };
    // `cargar` se redefine en cada render; la identidad de la consulta la fija `token`.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [token]);

  const recargar = useCallback(() => setVersion((v) => v + 1), []);
  const listo = resultado?.token === token;
  return { datos: resultado?.datos ?? null, error: listo ? resultado.error : '', cargando: !listo, recargar };
}
