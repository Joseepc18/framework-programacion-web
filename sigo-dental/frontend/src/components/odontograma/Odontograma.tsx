import type { Hallazgo } from '../../services/odontogramaService';
import DienteSvg from './DienteSvg';
import { ARCADAS, nombreDePieza, numeroDePieza, TIPOS, type Denticion, type Numeracion } from './dientes';

interface Props {
  hallazgos: readonly Hallazgo[];
  denticion: Denticion;
  numeracion: Numeracion;
  seleccion: number | null;
  alElegirPieza: (pieza: number) => void;
}

// Un diente del arco: botón de 44 px como mínimo (tablet junto al sillón) con el dibujo y su número.
const Diente = ({ pieza, hallazgos, numeracion, elegida, alElegir, numeroArriba }: {
  pieza: number; hallazgos: readonly Hallazgo[]; numeracion: Numeracion; elegida: boolean; alElegir: () => void; numeroArriba: boolean;
}) => {
  const resumen = hallazgos.length === 0 ? 'sin hallazgos' : hallazgos.map((h) => `${TIPOS[h.tipo].etiqueta}${h.superficie ? ` (${h.superficie})` : ''}`).join(', ');
  const numero = <span className="font-mono text-etiqueta text-texto-suave">{numeroDePieza(pieza, numeracion)}</span>;
  return (
    <button
      type="button"
      onClick={alElegir}
      aria-pressed={elegida}
      aria-label={`Pieza ${pieza}, ${nombreDePieza(pieza)}: ${resumen}`}
      title={`Pieza ${pieza} · ${nombreDePieza(pieza)}`}
      className={`flex min-h-11 w-11 shrink-0 flex-col items-center gap-1 rounded-control p-0.5 transition focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-acento ${
        elegida ? 'bg-item-activo ring-2 ring-odonto-seleccion' : 'hover:bg-chip-neutro-fondo'
      }`}
    >
      {numeroArriba && numero}
      <svg viewBox="-3 -3 46 46" className="size-10" aria-hidden="true">
        <DienteSvg pieza={pieza} hallazgos={hallazgos} />
      </svg>
      {!numeroArriba && numero}
    </button>
  );
};

const Fila = ({ piezas, etiqueta, hallazgosPorPieza, numeracion, seleccion, alElegirPieza, numeroArriba }: {
  piezas: readonly number[]; etiqueta: string; hallazgosPorPieza: Map<number, Hallazgo[]>; numeracion: Numeracion;
  seleccion: number | null; alElegirPieza: (pieza: number) => void; numeroArriba: boolean;
}) => {
  const mitad = piezas.length / 2;
  const grupo = (lista: readonly number[]) => lista.map((pieza) => (
    <Diente key={pieza} pieza={pieza} hallazgos={hallazgosPorPieza.get(pieza) ?? []} numeracion={numeracion}
      elegida={seleccion === pieza} alElegir={() => alElegirPieza(pieza)} numeroArriba={numeroArriba} />
  ));
  return (
    <div role="group" aria-label={etiqueta} className="flex items-start justify-center gap-4">
      <div className="flex gap-0.5">{grupo(piezas.slice(0, mitad))}</div>
      {/* Línea media: separa el lado derecho del paciente (izquierda de la pantalla) del izquierdo. */}
      <div className="w-px self-stretch bg-vidrio-borde" aria-hidden="true" />
      <div className="flex gap-0.5">{grupo(piezas.slice(mitad))}</div>
    </div>
  );
};

// Odontograma interactivo en SVG: dentición permanente (32 dientes), temporal (20) o mixta (las dos).
const Odontograma = ({ hallazgos, denticion, numeracion, seleccion, alElegirPieza }: Props) => {
  const porPieza = new Map<number, Hallazgo[]>();
  for (const h of hallazgos) porPieza.set(h.pieza, [...(porPieza.get(h.pieza) ?? []), h]);
  const comunes = { hallazgosPorPieza: porPieza, numeracion, seleccion, alElegirPieza };

  return (
    <div className="overflow-x-auto" role="region" aria-label="Odontograma" tabIndex={0}>
      <div className="mx-auto w-max min-w-full space-y-4 py-2">
        {denticion !== 'temporal' && <Fila piezas={ARCADAS.permanente.superior} etiqueta="Arcada superior permanente" numeroArriba={false} {...comunes} />}
        {denticion !== 'permanente' && <Fila piezas={ARCADAS.temporal.superior} etiqueta="Arcada superior temporal" numeroArriba={false} {...comunes} />}
        {denticion === 'mixta' && <div className="h-px bg-vidrio-borde" aria-hidden="true" />}
        {denticion !== 'permanente' && <Fila piezas={ARCADAS.temporal.inferior} etiqueta="Arcada inferior temporal" numeroArriba {...comunes} />}
        {denticion !== 'temporal' && <Fila piezas={ARCADAS.permanente.inferior} etiqueta="Arcada inferior permanente" numeroArriba {...comunes} />}
      </div>
    </div>
  );
};

export default Odontograma;
