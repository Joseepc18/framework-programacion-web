import { useState } from 'react';
import { Aviso, Boton, Campo, Tabla, Td, Th } from '../../components/ui';
import { useAccion } from '../../hooks/useAccion';
import { useCarga } from '../../hooks/useCarga';
import { actualizarParametro, listarParametros } from '../../services/catalogoService';

// Edición en línea de los parámetros del sistema (por ahora, el horario de la agenda).
const PanelParametros = () => {
  const { datos, error, recargar } = useCarga(listarParametros, 'parametros');
  const accion = useAccion();
  const [valores, setValores] = useState<Record<string, string>>({});

  const guardar = async (clave: string) => {
    if (await accion.ejecutar(() => actualizarParametro(clave, valores[clave]))) {
      setValores((v) => {
        const resto = { ...v };
        delete resto[clave];
        return resto;
      });
      recargar();
    }
  };

  return (
    <div className="space-y-6">
      {(error || accion.error) && <Aviso>{error || accion.error}</Aviso>}
      {datos && (
        <Tabla etiqueta="Parámetros del sistema">
          <thead><tr><Th>Parámetro</Th><Th>Valor</Th><Th>Acciones</Th></tr></thead>
          <tbody>
            {datos.map((p) => {
              const valor = valores[p.clave] ?? p.valor;
              return (
                <tr key={p.clave}>
                  <Td>
                    <span className="font-mono">{p.clave}</span>
                    <br /><span className="text-etiqueta text-texto-suave">{p.descripcion}</span>
                  </Td>
                  <Td><Campo id={`par-${p.clave}`} etiqueta={`Valor de ${p.clave}`} value={valor} onChange={(e) => setValores((v) => ({ ...v, [p.clave]: e.target.value }))} /></Td>
                  <Td><Boton disabled={accion.ocupado || valor === p.valor} onClick={() => guardar(p.clave)}>Guardar</Boton></Td>
                </tr>
              );
            })}
          </tbody>
        </Tabla>
      )}
    </div>
  );
};

export default PanelParametros;
