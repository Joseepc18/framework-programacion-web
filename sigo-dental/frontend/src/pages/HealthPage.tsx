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

const CHIP_NEUTRO = 'border-chip-neutro-borde bg-chip-neutro-fondo text-chip-neutro-texto';
const CHIP_ACENTO = 'border-chip-acento-borde bg-chip-acento-fondo text-chip-acento-texto';
const CHIP_ALERTA = 'border-chip-alerta-borde bg-chip-alerta-fondo text-chip-alerta-texto';

// El color nunca es el único indicador: cada chip lleva también su texto.
const COLOR_ESTADO: Record<HealthStatus['status'], string> = {
  ok: CHIP_ACENTO,
  degraded: CHIP_NEUTRO,
  error: CHIP_ALERTA,
};

const etiquetaServicio = (valor?: 'up' | 'down') =>
  valor === 'up' ? 'Disponible' : valor === 'down' ? 'No disponible' : 'Sin dato';

const colorServicio = (valor?: 'up' | 'down') =>
  valor === 'up' ? CHIP_ACENTO : valor === 'down' ? CHIP_ALERTA : CHIP_NEUTRO;

const Chip = ({ color, children }: { color: string; children: string }) => (
  <span className={`inline-block rounded-full border px-3 py-1 text-etiqueta font-medium ${color}`}>{children}</span>
);

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
        <h1 id="titulo-estado" className="text-pagina font-semibold">Estado del sistema</h1>
        <button
          type="button"
          onClick={consultarDeNuevo}
          disabled={estado.fase === 'cargando'}
          className="inline-flex min-h-11 items-center gap-2 rounded-control border border-chip-neutro-borde bg-chip-neutro-fondo px-4 text-cuerpo font-medium text-texto transition hover:bg-item-activo disabled:opacity-50 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-acento"
        >
          <RefreshCw className="size-4" aria-hidden="true" />
          Actualizar
        </button>
      </div>

      <div className="vidrio mt-6 rounded-tarjeta p-6" aria-live="polite">
        {estado.fase === 'cargando' && <p>Consultando la API…</p>}

        {estado.fase === 'fallo' && (
          <p role="alert" className="text-alerta-valor">{estado.mensaje}</p>
        )}

        {estado.fase === 'listo' && (
          <div className="space-y-4">
            <p>
              <Chip color={COLOR_ESTADO[estado.salud.status]}>{ETIQUETA_ESTADO[estado.salud.status]}</Chip>
            </p>
            <p className="text-texto-suave">{estado.salud.message}</p>
            <dl className="grid gap-3 sm:grid-cols-2">
              <div className="rounded-control border border-vidrio-borde bg-chip-neutro-fondo p-4">
                <dt className="text-etiqueta text-texto-suave">Base de datos (PostgreSQL)</dt>
                <dd className="mt-2">
                  <Chip color={colorServicio(estado.salud.database)}>{etiquetaServicio(estado.salud.database)}</Chip>
                </dd>
              </div>
              <div className="rounded-control border border-vidrio-borde bg-chip-neutro-fondo p-4">
                <dt className="text-etiqueta text-texto-suave">Caché (Redis)</dt>
                <dd className="mt-2">
                  <Chip color={colorServicio(estado.salud.redis)}>{etiquetaServicio(estado.salud.redis)}</Chip>
                </dd>
              </div>
            </dl>
          </div>
        )}
      </div>
    </section>
  );
};

export default HealthPage;
