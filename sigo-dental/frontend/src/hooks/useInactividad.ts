import { useEffect, useRef } from 'react';

// RNF-05: cierre de sesión por inactividad a los 15 minutos en las estaciones de trabajo.
export const INACTIVIDAD_MS = 15 * 60 * 1000;
// Se compara contra la hora de la última actividad (no con un temporizador largo) para que también
// funcione si el equipo se suspendió o la pestaña estuvo en segundo plano.
const REVISION_MS = 15 * 1000;

const EVENTOS_ACTIVIDAD = ['pointerdown', 'pointermove', 'keydown', 'wheel', 'touchstart', 'scroll'] as const;

export function useInactividad(activo: boolean, alInactivo: () => void) {
  const accion = useRef(alInactivo);
  useEffect(() => {
    accion.current = alInactivo;
  }, [alInactivo]);

  useEffect(() => {
    if (!activo) return;

    let ultimaActividad = Date.now();
    const registrar = () => {
      ultimaActividad = Date.now();
    };
    const revisar = () => {
      if (Date.now() - ultimaActividad >= INACTIVIDAD_MS) accion.current();
    };

    // `capture` permite oír el scroll de <main>, que no sube hasta window.
    EVENTOS_ACTIVIDAD.forEach((evento) => window.addEventListener(evento, registrar, { capture: true, passive: true }));
    document.addEventListener('visibilitychange', revisar);
    const intervalo = window.setInterval(revisar, REVISION_MS);

    return () => {
      EVENTOS_ACTIVIDAD.forEach((evento) => window.removeEventListener(evento, registrar, { capture: true }));
      document.removeEventListener('visibilitychange', revisar);
      window.clearInterval(intervalo);
    };
  }, [activo]);
}
