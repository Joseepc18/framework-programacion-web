import { useCallback, useState } from 'react';
import { mensajeDeApi } from './useCarga';

// Ejecuta una acción que escribe en la API (guardar, anular, cambiar de estado) y guarda
// el estado de espera y el mensaje de error para mostrarlos en el formulario.
export function useAccion() {
  const [error, setError] = useState('');
  const [ocupado, setOcupado] = useState(false);

  // Devuelve true si la acción terminó bien.
  const ejecutar = useCallback(async (accion: () => Promise<unknown>) => {
    setOcupado(true);
    setError('');
    try {
      await accion();
      return true;
    } catch (fallo) {
      setError(mensajeDeApi(fallo));
      return false;
    } finally {
      setOcupado(false);
    }
  }, []);

  const limpiar = useCallback(() => setError(''), []);
  return { error, ocupado, ejecutar, limpiar };
}
