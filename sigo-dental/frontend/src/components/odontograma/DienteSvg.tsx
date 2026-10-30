import type { ReactNode } from 'react';
import type { Hallazgo } from '../../services/odontogramaService';
import { POLIGONOS, SUPERFICIES, TIPOS, zonaDeSuperficie, type Superficie, type Zona } from './dientes';

interface Props {
  pieza: number;
  hallazgos: readonly Hallazgo[];
  // Si se indica, las cinco superficies son botones (editor de pieza); si no, el dibujo es solo de lectura.
  alElegirSuperficie?: (superficie: Superficie) => void;
  superficieElegida?: Superficie | null;
  // Etiqueta accesible de cada superficie (en el editor).
  etiquetaDeSuperficie?: (superficie: Superficie) => string;
}

// Los hallazgos por superficie pintan su zona; los de pieza completa dibujan una forma sobre el diente. El estado
// "requerido" lleva el trazo punteado y "realizado" un trazo más grueso, así que el color no es el único indicador.
const trazoDeEstado = (hallazgo: Hallazgo | undefined) => {
  if (!hallazgo) return { strokeWidth: 1.2, strokeDasharray: undefined };
  if (hallazgo.estado === 'requerido') return { strokeWidth: 1.8, strokeDasharray: '3 2' };
  if (hallazgo.estado === 'realizado') return { strokeWidth: 2.6, strokeDasharray: undefined };
  return { strokeWidth: 1.8, strokeDasharray: undefined };
};

const DienteSvg = ({ pieza, hallazgos, alElegirSuperficie, superficieElegida, etiquetaDeSuperficie }: Props) => {
  const porSuperficie = new Map<Zona, Hallazgo>();
  const completos: Hallazgo[] = [];
  for (const h of hallazgos) {
    if (h.superficie) porSuperficie.set(zonaDeSuperficie(h.superficie, pieza), h);
    else completos.push(h);
  }
  const ausente = completos.some((h) => h.tipo === 'ausente');
  const de = (tipo: string) => completos.find((h) => h.tipo === tipo);

  const zonas: Zona[] = ['arriba', 'abajo', 'izquierda', 'derecha', 'centro'];
  const superficieDe = (zona: Zona) => SUPERFICIES.find((s) => zonaDeSuperficie(s, pieza) === zona) as Superficie;

  const capas: ReactNode[] = zonas.map((zona) => {
    const h = porSuperficie.get(zona);
    const relleno = ausente ? 'fill-transparent' : h ? TIPOS[h.tipo].relleno : 'fill-odonto-sano';
    const trazo = ausente ? 'stroke-odonto-ausente-borde' : h ? TIPOS[h.tipo].trazo : 'stroke-odonto-sano-borde';
    const { strokeWidth, strokeDasharray } = ausente ? { strokeWidth: 1.4, strokeDasharray: '3 2' } : trazoDeEstado(h);
    const superficie = superficieDe(zona);
    const elegida = superficieElegida === superficie;
    return (
      <polygon
        key={zona}
        points={POLIGONOS[zona]}
        className={`${relleno} ${trazo} ${alElegirSuperficie ? 'cursor-pointer' : ''}`}
        strokeWidth={strokeWidth}
        strokeDasharray={strokeDasharray}
        strokeLinejoin="round"
        onClick={alElegirSuperficie ? () => alElegirSuperficie(superficie) : undefined}
        role={alElegirSuperficie ? 'button' : undefined}
        tabIndex={alElegirSuperficie ? 0 : undefined}
        aria-label={alElegirSuperficie && etiquetaDeSuperficie ? etiquetaDeSuperficie(superficie) : undefined}
        aria-pressed={alElegirSuperficie ? elegida : undefined}
        onKeyDown={alElegirSuperficie ? (e) => { if (e.key === 'Enter' || e.key === ' ') { e.preventDefault(); alElegirSuperficie(superficie); } } : undefined}
        style={elegida ? { filter: 'drop-shadow(0 0 2px var(--color-odonto-seleccion))' } : undefined}
      >
        {elegida && <title>seleccionada</title>}
      </polygon>
    );
  });

  const endodoncia = de('endodoncia');
  const implante = de('implante');
  const fija = de('protesis_fija');
  const removible = de('protesis_removible');
  const exodoncia = de('exodoncia');

  return (
    <>
      {capas}
      {/* El contorno de la superficie elegida se repite encima para que se vea sobre las demás. */}
      {superficieElegida && (
        <polygon points={POLIGONOS[zonaDeSuperficie(superficieElegida, pieza)]} className="pointer-events-none fill-none stroke-odonto-seleccion" strokeWidth={3} strokeLinejoin="round" />
      )}
      {fija && <rect x={-1} y={-1} width={42} height={42} rx={3} className="pointer-events-none fill-none stroke-odonto-protesis-fija" strokeWidth={trazoDeEstado(fija).strokeWidth + 1} strokeDasharray={trazoDeEstado(fija).strokeDasharray} />}
      {removible && <rect x={-1} y={-1} width={42} height={42} rx={9} className="pointer-events-none fill-none stroke-odonto-protesis-removible" strokeWidth={trazoDeEstado(removible).strokeWidth + 1} strokeDasharray="5 3" />}
      {endodoncia && (
        <polygon points="20,13 28,27 12,27" className={`pointer-events-none ${TIPOS.endodoncia.relleno} ${TIPOS.endodoncia.trazo}`} strokeWidth={trazoDeEstado(endodoncia).strokeWidth} strokeDasharray={trazoDeEstado(endodoncia).strokeDasharray} strokeLinejoin="round" />
      )}
      {implante && (
        <g className="pointer-events-none" strokeLinecap="round">
          <rect x={16} y={9} width={8} height={22} rx={2} className={`${TIPOS.implante.relleno} ${TIPOS.implante.trazo}`} strokeWidth={trazoDeEstado(implante).strokeWidth} strokeDasharray={trazoDeEstado(implante).strokeDasharray} />
          <path d="M14 15h12M14 20h12M14 25h12" className="fill-none stroke-fondo" strokeWidth={1.4} />
        </g>
      )}
      {exodoncia && <path d="M3 3L37 37M37 3L3 37" className="pointer-events-none fill-none stroke-odonto-exodoncia" strokeWidth={trazoDeEstado(exodoncia).strokeWidth + 1.2} strokeLinecap="round" strokeDasharray={trazoDeEstado(exodoncia).strokeDasharray} />}
    </>
  );
};

export default DienteSvg;
