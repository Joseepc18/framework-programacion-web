import { useMemo, useState } from 'react';
import { Aviso, Boton, Chip, Selector, Tarjeta } from '../../components/ui';
import EditorPieza from '../../components/odontograma/EditorPieza';
import Odontograma from '../../components/odontograma/Odontograma';
import {
  AYUDA_ESTADO, denticionPorEdad, ETIQUETA_ESTADO, ORDEN_TIPOS, TIPOS, type Denticion, type EstadoHallazgo, type Numeracion,
} from '../../components/odontograma/dientes';
import { useCarga } from '../../hooks/useCarga';
import { obtenerHistorial, obtenerOdontograma, type Hallazgo } from '../../services/odontogramaService';
import { fechaHoraLocal, fechaLocal } from '../../utils/fechas';

const Leyenda = () => (
  <Tarjeta denso className="space-y-4">
    <h2 className="text-seccion font-semibold">Leyenda</h2>
    <ul className="grid gap-x-6 gap-y-2 sm:grid-cols-2 xl:grid-cols-3">
      {ORDEN_TIPOS.map((tipo) => {
        const meta = TIPOS[tipo];
        return (
          <li key={tipo} className="flex items-center gap-3">
            <svg viewBox="0 0 20 20" className="size-6 shrink-0" aria-hidden="true">
              <rect x="3" y="3" width="14" height="14" rx="3" strokeWidth={tipo === 'ausente' ? 1.6 : 1.4}
                strokeDasharray={tipo === 'ausente' ? '3 2' : undefined} className={`${meta.relleno} ${meta.trazo}`} />
            </svg>
            <span><span className="font-medium">{meta.etiqueta}</span> <span className="text-etiqueta text-texto-suave">· {meta.forma}</span></span>
          </li>
        );
      })}
    </ul>
    <ul className="grid gap-1 text-etiqueta text-texto-suave sm:grid-cols-3">
      {(Object.keys(ETIQUETA_ESTADO) as EstadoHallazgo[]).map((e) => (
        <li key={e}><span className="font-medium text-texto">{ETIQUETA_ESTADO[e]}</span>: {AYUDA_ESTADO[e]}
          {e === 'requerido' && ' (trazo punteado)'}{e === 'realizado' && ' (trazo grueso)'}</li>
      ))}
    </ul>
  </Tarjeta>
);

// Agrupa los eventos del historial por día: cada visita muestra lo que se registró y lo que se retiró.
function agruparPorVisita(historial: readonly Hallazgo[]) {
  const eventos: { cuando: string; texto: string; quien: string; cita: number | null }[] = [];
  for (const h of historial) {
    const que = `${TIPOS[h.tipo].etiqueta.toLowerCase()} en la pieza ${h.pieza}${h.superficie ? ` (${h.superficie})` : ''} · ${ETIQUETA_ESTADO[h.estado].toLowerCase()}`;
    eventos.push({ cuando: h.registrado_en, texto: `Registró ${que}`, quien: h.registrado_por, cita: h.cita_id });
    if (h.retirado_en) eventos.push({ cuando: h.retirado_en, texto: `${h.reemplaza_a === null ? 'Retiró' : 'Retiró (corregido o retirado)'} ${que}`, quien: h.retirado_por, cita: null });
  }
  eventos.sort((a, b) => b.cuando.localeCompare(a.cuando));
  const dias = new Map<string, typeof eventos>();
  for (const e of eventos) dias.set(e.cuando.slice(0, 10), [...(dias.get(e.cuando.slice(0, 10)) ?? []), e]);
  return [...dias.entries()].map(([dia, lista]) => ({ dia, ultimo: lista[0].cuando, lista }));
}

interface Props {
  pacienteId: number;
  edad: number;
  puedeEditar: boolean;
}

// Odontograma del paciente (RF-03): vista interactiva, editor de pieza, leyenda e historial por visita.
const PestanaOdontograma = ({ pacienteId, edad, puedeEditar }: Props) => {
  const [denticion, setDenticion] = useState<Denticion>(denticionPorEdad(edad));
  const [numeracion, setNumeracion] = useState<Numeracion>('fdi');
  const [pieza, setPieza] = useState<number | null>(null);
  // Instante que se está revisando (visita anterior); null = odontograma actual.
  const [al, setAl] = useState<string | null>(null);

  const odontograma = useCarga(() => obtenerOdontograma(pacienteId, al ?? undefined), `odonto-${pacienteId}-${al ?? 'actual'}`);
  const historial = useCarga(() => obtenerHistorial(pacienteId), `odonto-historial-${pacienteId}`);
  const visitas = useMemo(() => agruparPorVisita(historial.datos ?? []), [historial.datos]);

  const recargar = () => {
    odontograma.recargar();
    historial.recargar();
  };

  const hallazgos = odontograma.datos ?? [];
  const deLaPieza = pieza === null ? [] : hallazgos.filter((h) => h.pieza === pieza);

  return (
    <div className="space-y-6">
      <Tarjeta className="space-y-4">
        <div className="flex flex-wrap items-end gap-4">
          <div className="w-52">
            <Selector id="odonto-denticion" etiqueta="Dentición" value={denticion} onChange={(e) => setDenticion(e.target.value as Denticion)}>
              <option value="permanente">Permanente (32)</option>
              <option value="temporal">Temporal (20)</option>
              <option value="mixta">Mixta (recambio)</option>
            </Selector>
          </div>
          <div className="w-52">
            <Selector id="odonto-numeracion" etiqueta="Numeración" ayuda="Solo cambia lo que se ve; se guarda FDI." value={numeracion} onChange={(e) => setNumeracion(e.target.value as Numeracion)}>
              <option value="fdi">FDI</option>
              <option value="universal">Universal</option>
            </Selector>
          </div>
          {al && (
            <div className="flex flex-wrap items-center gap-3">
              <Chip tono="alerta">Solo lectura: odontograma al {fechaHoraLocal(al)}</Chip>
              <Boton variante="secundario" onClick={() => { setAl(null); setPieza(null); }}>Volver al actual</Boton>
            </div>
          )}
        </div>
        {odontograma.error && <Aviso>{odontograma.error}</Aviso>}
        {odontograma.cargando && !odontograma.datos ? <p>Cargando odontograma…</p> : (
          <Odontograma hallazgos={hallazgos} denticion={denticion} numeracion={numeracion} seleccion={pieza} alElegirPieza={setPieza} />
        )}
        <p className="text-etiqueta text-texto-suave">Toca una pieza para ver y registrar sus hallazgos.</p>
      </Tarjeta>

      {pieza !== null && (
        <EditorPieza key={`${pieza}-${al ?? 'actual'}`} pacienteId={pacienteId} pieza={pieza} hallazgos={deLaPieza} numeracion={numeracion}
          puedeEditar={puedeEditar && al === null} alCambiar={recargar} />
      )}

      <Leyenda />

      <Tarjeta denso className="space-y-4">
        <h2 className="text-seccion font-semibold">Historial por visita</h2>
        <p className="text-etiqueta text-texto-suave">Nada se sobrescribe: cada registro y cada retiro queda aquí con quién y cuándo.</p>
        {historial.error && <Aviso>{historial.error}</Aviso>}
        {visitas.length === 0 && !historial.cargando && <p className="text-texto-suave">Todavía no hay hallazgos registrados.</p>}
        <ul className="space-y-4">
          {visitas.map((v) => (
            <li key={v.dia}>
              <div className="flex flex-wrap items-center gap-3">
                <h3 className="font-semibold">{fechaLocal(v.ultimo)}</h3>
                <Boton variante="secundario" onClick={() => { setAl(v.ultimo); setPieza(null); }}>Ver el odontograma de ese día</Boton>
              </div>
              <ul className="mt-2 space-y-1 text-etiqueta">
                {v.lista.map((e, i) => (
                  <li key={i}>{e.texto} <span className="text-texto-suave">· {e.quien}{e.cita ? ` · cita #${e.cita}` : ''}</span></li>
                ))}
              </ul>
            </li>
          ))}
        </ul>
      </Tarjeta>
    </div>
  );
};

export default PestanaOdontograma;
