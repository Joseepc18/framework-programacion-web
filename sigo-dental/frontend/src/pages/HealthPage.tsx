import { useEffect, useState } from 'react';
import { RefreshCw } from 'lucide-react';
import { obtenerEstadoSalud, type HealthStatus } from '../services/healthService';

type Estado =
  | { fase: 'cargando' }
  | { fase: 'listo'; salud: HealthStatus }
  | { fase: 'fallo'; mensaje: string };

const ETIQUETA_ESTADO: Record<HealthStatus['status'], string> = {
  ok: 'Operativo',
  degraded: 'Degradado',
  error: 'Con errores',
};

const COLOR_ESTADO: Record<HealthStatus['status'], string> = {
  ok: 'bg-emerald-100 text-emerald-800',
  degraded: 'bg-amber-100 text-amber-800',
  error: 'bg-red-100 text-red-800',
};

const etiquetaServicio = (valor?: 'up' | 'down') =>
  valor === 'up' ? 'Disponible' : valor === 'down' ? 'No disponible' : 'Sin dato';

const HealthPage = () => {
  const [estado, setEstado] = useState<Estado>({ fase: 'cargando' });
  const [consulta, setConsulta] = useState(0);

  useEffect(() => {
    let vigente = true;
    obtenerEstadoSalud()
      .then((salud) => vigente && setEstado({ fase: 'listo', salud }))
      .catch(() =>
        vigente &&
        setEstado({ fase: 'fallo', mensaje: 'No se pudo conectar con la API. Verifica que esté en ejecución y que VITE_API_URL sea correcta.' }),
      );
    return () => {
      vigente = false;
    };
  }, [consulta]);

  const consultarDeNuevo = () => {
    setEstado({ fase: 'cargando' });
    setConsulta((n) => n + 1);
  };

  return (
    <section aria-labelledby="titulo-estado">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <h1 id="titulo-estado" className="text-2xl font-semibold">Estado del sistema</h1>
        <button
          type="button"
          onClick={consultarDeNuevo}
          disabled={estado.fase === 'cargando'}
          className="inline-flex items-center gap-2 rounded-md border border-slate-300 bg-white px-3 py-2 text-sm font-medium hover:bg-slate-100 disabled:opacity-50 focus-visible:outline-2 focus-visible:outline-clinica-600"
        >
          <RefreshCw className="size-4" aria-hidden="true" />
          Actualizar
        </button>
      </div>

      <div className="mt-6 rounded-lg border border-slate-200 bg-white p-5" aria-live="polite">
        {estado.fase === 'cargando' && <p>Consultando la API…</p>}

        {estado.fase === 'fallo' && (
          <p role="alert" className="text-red-700">{estado.mensaje}</p>
        )}

        {estado.fase === 'listo' && (
          <div className="space-y-4">
            <p>
              <span className={`inline-block rounded-full px-3 py-1 text-sm font-semibold ${COLOR_ESTADO[estado.salud.status]}`}>
                {ETIQUETA_ESTADO[estado.salud.status]}
              </span>
            </p>
            <p>{estado.salud.message}</p>
            <dl className="grid gap-3 sm:grid-cols-2">
              <div className="rounded-md bg-slate-50 p-3">
                <dt className="text-sm text-slate-600">Base de datos (PostgreSQL)</dt>
                <dd className="font-medium">{etiquetaServicio(estado.salud.database)}</dd>
              </div>
              <div className="rounded-md bg-slate-50 p-3">
                <dt className="text-sm text-slate-600">Caché (Redis)</dt>
                <dd className="font-medium">{etiquetaServicio(estado.salud.redis)}</dd>
              </div>
            </dl>
          </div>
        )}
      </div>
    </section>
  );
};

export default HealthPage;
