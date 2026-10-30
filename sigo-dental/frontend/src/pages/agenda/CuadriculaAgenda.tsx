import type { MouseEvent } from 'react';
import { ESTILO_ESTADO, ETIQUETA_ESTADO } from '../../config/citas';
import type { Cita, HorarioAgenda } from '../../services/citaService';
import { horaLocal, minutosDe } from '../../utils/fechas';

// Alto de una hora en píxeles y paso de los huecos que se pueden pulsar para agendar.
const PX_POR_HORA = 84;
const PASO_MINUTOS = 30;
const ANCHO_MINIMO_COLUMNA = 150;

export interface Columna {
  clave: string;
  titulo: string;
  subtitulo?: string;
  // Medianoche local del día que representa la columna.
  dia: Date;
  sillonId?: number;
}

interface CuadriculaProps {
  columnas: Columna[];
  citas: Cita[];
  horario: HorarioAgenda;
  puedeAgendar: boolean;
  alElegirHueco: (inicio: Date, sillonId?: number) => void;
  alElegirCita: (cita: Cita) => void;
  seleccionada?: number;
}

interface Colocada {
  cita: Cita;
  carril: number;
  carriles: number;
}

const minutosDesdeMedianoche = (fecha: Date) => fecha.getHours() * 60 + fecha.getMinutes();

// Reparte en carriles las citas de una columna que se cruzan en el tiempo, para que se vean lado a lado.
function colocar(citas: Cita[]): Colocada[] {
  const ordenadas = [...citas].sort((a, b) => a.inicio.localeCompare(b.inicio) || a.id - b.id);
  const resultado: Colocada[] = [];
  let grupo: Colocada[] = [];
  let finesPorCarril: number[] = [];
  let finGrupo = -Infinity;

  const cerrarGrupo = () => {
    grupo.forEach((c) => (c.carriles = finesPorCarril.length));
    resultado.push(...grupo);
    grupo = [];
    finesPorCarril = [];
  };

  for (const cita of ordenadas) {
    const inicio = new Date(cita.inicio).getTime();
    const fin = new Date(cita.fin).getTime();
    if (inicio >= finGrupo) cerrarGrupo();
    let carril = finesPorCarril.findIndex((fc) => fc <= inicio);
    if (carril === -1) carril = finesPorCarril.length;
    finesPorCarril[carril] = fin;
    finGrupo = Math.max(finGrupo === -Infinity ? fin : finGrupo, fin);
    grupo.push({ cita, carril, carriles: 1 });
  }
  cerrarGrupo();
  return resultado;
}

const CuadriculaAgenda = ({ columnas, citas, horario, puedeAgendar, alElegirHueco, alElegirCita, seleccionada }: CuadriculaProps) => {
  const abre = minutosDe(horario.inicio);
  const cierra = minutosDe(horario.fin);
  const horas = Array.from({ length: Math.ceil((cierra - abre) / 60) }, (_, i) => abre + i * 60);
  const alto = ((cierra - abre) / 60) * PX_POR_HORA;
  const plantilla = `56px repeat(${columnas.length}, minmax(${ANCHO_MINIMO_COLUMNA}px, 1fr))`;

  const pulsarHueco = (evento: MouseEvent<HTMLDivElement>, columna: Columna) => {
    const desplazamiento = evento.clientY - evento.currentTarget.getBoundingClientRect().top;
    const minutos = abre + Math.floor(((desplazamiento / PX_POR_HORA) * 60) / PASO_MINUTOS) * PASO_MINUTOS;
    const inicio = new Date(columna.dia);
    inicio.setHours(0, minutos, 0, 0);
    alElegirHueco(inicio, columna.sillonId);
  };

  return (
    <div className="vidrio-denso overflow-x-auto rounded-tarjeta">
      <div className="grid min-w-max" style={{ gridTemplateColumns: plantilla }}>
        <div className="border-b border-vidrio-borde" />
        {columnas.map((columna) => (
          <div key={columna.clave} className="border-b border-l border-vidrio-borde px-3 py-3">
            <p className="font-medium capitalize">{columna.titulo}</p>
            {columna.subtitulo && <p className="text-etiqueta text-texto-suave">{columna.subtitulo}</p>}
          </div>
        ))}

        <div className="relative" style={{ height: alto }} aria-hidden="true">
          {horas.map((minutos) => (
            <span
              key={minutos}
              className="absolute right-2 -translate-y-1/2 font-mono text-etiqueta text-texto-suave"
              style={{ top: ((minutos - abre) / 60) * PX_POR_HORA, display: minutos === abre ? 'none' : undefined }}
            >
              {String(Math.floor(minutos / 60)).padStart(2, '0')}:00
            </span>
          ))}
        </div>

        {columnas.map((columna) => {
          const delDia = citas.filter((c) => {
            if (columna.sillonId !== undefined && c.sillon_id !== columna.sillonId) return false;
            const inicio = new Date(c.inicio);
            return inicio.toDateString() === columna.dia.toDateString();
          });
          return (
            <div
              key={columna.clave}
              className={`relative border-l border-vidrio-borde ${puedeAgendar ? 'cursor-pointer hover:bg-item-activo/30' : ''}`}
              style={{
                height: alto,
                backgroundImage: 'linear-gradient(to bottom, var(--color-vidrio-borde) 1px, transparent 1px)',
                backgroundSize: `100% ${PX_POR_HORA}px`,
              }}
              onClick={puedeAgendar ? (e) => pulsarHueco(e, columna) : undefined}
            >
              {colocar(delDia).map(({ cita, carril, carriles }) => {
                const inicio = new Date(cita.inicio);
                const fin = new Date(cita.fin);
                const desde = Math.max(minutosDesdeMedianoche(inicio), abre);
                const hasta = Math.min(minutosDesdeMedianoche(fin), cierra);
                const ancho = 100 / carriles;
                return (
                  <button
                    key={cita.id}
                    type="button"
                    onClick={(e) => {
                      e.stopPropagation();
                      alElegirCita(cita);
                    }}
                    aria-label={`${horaLocal(inicio)} a ${horaLocal(fin)}, ${cita.paciente_nombre}, ${ETIQUETA_ESTADO[cita.estado]}`}
                    aria-pressed={seleccionada === cita.id}
                    className={`absolute overflow-hidden rounded-control border px-2 py-1 text-left text-etiqueta leading-tight transition hover:brightness-125 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-acento ${ESTILO_ESTADO[cita.estado]} ${seleccionada === cita.id ? 'ring-2 ring-acento' : ''}`}
                    style={{
                      top: ((desde - abre) / 60) * PX_POR_HORA + 1,
                      height: Math.max(((hasta - desde) / 60) * PX_POR_HORA - 2, 26),
                      left: `calc(${carril * ancho}% + 2px)`,
                      width: `calc(${ancho}% - 4px)`,
                    }}
                  >
                    <span className="font-mono">{horaLocal(inicio)}</span>{' '}
                    <span className="font-medium">{cita.paciente_nombre}</span>
                    {cita.alertas.length > 0 && <span className="ml-1 font-semibold text-alerta-valor">· Alerta médica</span>}
                    <span className="block">{ETIQUETA_ESTADO[cita.estado]}</span>
                  </button>
                );
              })}
            </div>
          );
        })}
      </div>
    </div>
  );
};

export default CuadriculaAgenda;
