import { useEffect, useState } from 'react';
import { ETIQUETA_ROL, type AreaRol } from '../config/permisos';
import { ApiError, apiFetch } from '../services/http';
import AccesoDenegadoPage from './AccesoDenegadoPage';

type Estado =
  | { fase: 'cargando' }
  | { fase: 'permitido' }
  | { fase: 'denegado' }
  | { fase: 'fallo' };

// Marcador de posición del módulo de cada rol: consulta la ruta de prueba del backend, que responde
// 200 solo a ese rol, y muestra lo que decidió el servidor (no solo lo que oculta la interfaz).
const AreaRolPage = ({ area }: { area: AreaRol }) => {
  const [estado, setEstado] = useState<Estado>({ fase: 'cargando' });

  useEffect(() => {
    let vigente = true;
    apiFetch(`/api/prueba/${area.slug}`)
      .then(() => vigente && setEstado({ fase: 'permitido' }))
      .catch((error) => {
        if (!vigente) return;
        setEstado({ fase: error instanceof ApiError && error.status === 403 ? 'denegado' : 'fallo' });
      });
    return () => {
      vigente = false;
    };
  }, [area.slug]);

  if (estado.fase === 'denegado') return <AccesoDenegadoPage />;

  return (
    <section aria-labelledby="titulo-area">
      <h1 id="titulo-area" className="text-pagina font-semibold">{area.label}</h1>
      <div className="vidrio mt-6 max-w-2xl rounded-tarjeta p-6" aria-live="polite">
        {estado.fase === 'cargando' && <p>Verificando el permiso con el servidor…</p>}
        {estado.fase === 'fallo' && (
          <p role="alert" className="text-alerta-valor">No se pudo verificar el permiso con el servidor. Inténtalo de nuevo.</p>
        )}
        {estado.fase === 'permitido' && (
          <div className="space-y-4">
            <p>
              <span className="inline-block rounded-full border border-chip-acento-borde bg-chip-acento-fondo px-3 py-1 text-etiqueta font-medium text-chip-acento-texto">
                Acceso permitido
              </span>
            </p>
            <p className="text-texto-suave">
              El servidor confirmó el acceso del rol {ETIQUETA_ROL[area.rol]}. Los módulos de esta área se agregarán en las
              siguientes etapas.
            </p>
          </div>
        )}
      </div>
    </section>
  );
};

export default AreaRolPage;
