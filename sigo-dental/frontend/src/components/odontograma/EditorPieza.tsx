import { useState, type FormEvent } from 'react';
import { Aviso, Boton, Campo, Chip, Selector, Tarjeta } from '../ui';
import { useAccion } from '../../hooks/useAccion';
import {
  corregirHallazgo, registrarHallazgo, retirarHallazgo, type DatosHallazgo, type Hallazgo,
} from '../../services/odontogramaService';
import { fechaHoraLocal } from '../../utils/fechas';
import DienteSvg from './DienteSvg';
import {
  AYUDA_ESTADO, ETIQUETA_ESTADO, etiquetaSuperficie, nombreDePieza, numeroDePieza, ORDEN_TIPOS, SUPERFICIES, TIPOS,
  type EstadoHallazgo, type Numeracion, type Superficie, type TipoHallazgo,
} from './dientes';

interface Props {
  pacienteId: number;
  pieza: number;
  hallazgos: readonly Hallazgo[];
  numeracion: Numeracion;
  puedeEditar: boolean;
  alCambiar: () => void;
}

// Editor de una pieza: dibujo grande con las cinco superficies (cada una mide más de 44 px, para marcarlas con el
// dedo en una tablet), formulario del hallazgo y lista de lo registrado en el diente.
const EditorPieza = ({ pacienteId, pieza, hallazgos, numeracion, puedeEditar, alCambiar }: Props) => {
  const accion = useAccion();
  const [superficie, setSuperficie] = useState<Superficie | null>(null);
  const [tipo, setTipo] = useState<TipoHallazgo>('caries');
  const [estado, setEstado] = useState<EstadoHallazgo>('requerido');
  const [observacion, setObservacion] = useState('');
  // Si hay un hallazgo en edición, guardar lo corrige (retira el anterior y registra el nuevo).
  const [editando, setEditando] = useState<Hallazgo | null>(null);
  const [aRetirar, setARetirar] = useState<number | null>(null);

  const meta = TIPOS[tipo];
  const etiqueta = (s: Superficie) => `${etiquetaSuperficie(s, pieza)}${superficie === s ? ' (elegida)' : ''}`;

  const cambiarTipo = (nuevo: TipoHallazgo) => {
    setTipo(nuevo);
    if (!TIPOS[nuevo].estados.includes(estado)) setEstado(TIPOS[nuevo].estados[0]);
  };

  const limpiar = () => {
    setEditando(null);
    setSuperficie(null);
    setObservacion('');
  };

  const empezarCorreccion = (h: Hallazgo) => {
    setEditando(h);
    setTipo(h.tipo);
    setEstado(h.estado);
    setSuperficie(h.superficie || null);
    setObservacion(h.observacion);
    setARetirar(null);
  };

  const guardar = async (evento: FormEvent) => {
    evento.preventDefault();
    const datos: DatosHallazgo = { pieza, superficie: meta.porSuperficie ? superficie ?? '' : '', tipo, estado, observacion };
    const bien = await accion.ejecutar(() => (editando ? corregirHallazgo(pacienteId, editando.id, datos) : registrarHallazgo(pacienteId, datos)));
    if (bien) {
      limpiar();
      alCambiar();
    }
  };

  const retirar = async (id: number) => {
    if (await accion.ejecutar(() => retirarHallazgo(pacienteId, id))) {
      setARetirar(null);
      alCambiar();
    }
  };

  const falta = meta.porSuperficie && !superficie;

  return (
    <Tarjeta className="space-y-5">
      <div className="flex flex-wrap items-baseline justify-between gap-2">
        <h2 className="text-seccion font-semibold">
          Pieza <span className="font-mono">{numeroDePieza(pieza, numeracion)}</span>
          <span className="ml-2 text-cuerpo font-normal text-texto-suave">{nombreDePieza(pieza)}{numeracion === 'universal' && <> · FDI <span className="font-mono">{pieza}</span></>}</span>
        </h2>
      </div>

      <div className="grid gap-6 md:grid-cols-[220px_1fr]">
        <div>
          <svg viewBox="-3 -3 46 46" className="mx-auto size-52" role="group" aria-label={`Superficies de la pieza ${pieza}`}>
            <DienteSvg pieza={pieza} hallazgos={hallazgos} alElegirSuperficie={puedeEditar ? setSuperficie : undefined}
              superficieElegida={superficie} etiquetaDeSuperficie={etiqueta} />
          </svg>
          {puedeEditar && (
            <div className="mt-3 flex flex-wrap justify-center gap-2" role="group" aria-label="Elegir superficie">
              {SUPERFICIES.map((s) => (
                <button key={s} type="button" aria-pressed={superficie === s} onClick={() => setSuperficie(s)}
                  className={`min-h-11 rounded-control border px-3 text-etiqueta transition focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-acento ${
                    superficie === s ? 'border-item-activo-borde bg-item-activo font-semibold' : 'border-chip-neutro-borde bg-chip-neutro-fondo text-texto-suave'
                  }`}>
                  {etiquetaSuperficie(s, pieza)}
                </button>
              ))}
            </div>
          )}
        </div>

        <div className="space-y-5">
          {puedeEditar ? (
            <form onSubmit={guardar} className="space-y-4">
              <p className="text-etiqueta text-texto-suave">
                {editando ? `Corrigiendo el hallazgo de ${TIPOS[editando.tipo].etiqueta.toLowerCase()}: el anterior queda en el historial.` : 'Registrar un hallazgo en esta pieza.'}
              </p>
              <div className="grid gap-4 sm:grid-cols-2">
                <Selector id="hallazgo-tipo" etiqueta="Hallazgo" value={tipo} onChange={(e) => cambiarTipo(e.target.value as TipoHallazgo)}>
                  {ORDEN_TIPOS.map((t) => <option key={t} value={t}>{TIPOS[t].etiqueta}</option>)}
                </Selector>
                <Selector id="hallazgo-estado" etiqueta="Estado" ayuda={AYUDA_ESTADO[estado]} value={estado} onChange={(e) => setEstado(e.target.value as EstadoHallazgo)}>
                  {meta.estados.map((e) => <option key={e} value={e}>{ETIQUETA_ESTADO[e]}</option>)}
                </Selector>
              </div>
              <p className="text-etiqueta text-texto-suave">
                {meta.porSuperficie
                  ? superficie ? `Superficie: ${etiquetaSuperficie(superficie, pieza)}.` : 'Elige la superficie en el dibujo o con los botones.'
                  : 'Este hallazgo afecta a la pieza completa.'}
              </p>
              <Campo id="hallazgo-obs" etiqueta="Observación (opcional)" maxLength={200} value={observacion} onChange={(e) => setObservacion(e.target.value)} />
              {accion.error && <Aviso>{accion.error}</Aviso>}
              <div className="flex flex-wrap gap-3">
                <Boton type="submit" disabled={accion.ocupado || falta}>{accion.ocupado ? 'Guardando…' : editando ? 'Guardar corrección' : 'Registrar hallazgo'}</Boton>
                {editando && <Boton variante="secundario" onClick={limpiar}>Cancelar</Boton>}
              </div>
            </form>
          ) : (
            <p className="text-etiqueta text-texto-suave">El odontograma lo registra el odontólogo; tu rol solo lo consulta.</p>
          )}

          <div>
            <h3 className="text-cuerpo font-semibold">Hallazgos de la pieza</h3>
            {hallazgos.length === 0 && <p className="mt-2 text-texto-suave">Sin hallazgos: pieza sana.</p>}
            <ul className="mt-2 space-y-2">
              {hallazgos.map((h) => (
                <li key={h.id} className="rounded-control border border-vidrio-borde p-3">
                  <div className="flex flex-wrap items-center gap-2">
                    <span className="font-medium">{TIPOS[h.tipo].etiqueta}</span>
                    {h.superficie && <Chip>{etiquetaSuperficie(h.superficie, pieza)}</Chip>}
                    <Chip tono={h.estado === 'realizado' ? 'acento' : 'neutro'}>{ETIQUETA_ESTADO[h.estado]}</Chip>
                  </div>
                  {h.observacion && <p className="mt-1 text-etiqueta">{h.observacion}</p>}
                  <p className="mt-1 text-etiqueta text-texto-suave">{h.registrado_por} · {fechaHoraLocal(h.registrado_en)}</p>
                  {puedeEditar && (
                    h.estado === 'realizado' ? (
                      <p className="mt-2 text-etiqueta text-texto-suave">Realizado: no se corrige ni se retira (RN-01). Si hay un error, registra una nota de adenda en Evolución.</p>
                    ) : aRetirar === h.id ? (
                      <div className="mt-2 flex flex-wrap items-center gap-2">
                        <span className="text-etiqueta">¿Retirar este hallazgo? Quedará en el historial.</span>
                        <Boton variante="peligro" disabled={accion.ocupado} onClick={() => retirar(h.id)}>Sí, retirar</Boton>
                        <Boton variante="secundario" onClick={() => setARetirar(null)}>No</Boton>
                      </div>
                    ) : (
                      <div className="mt-2 flex flex-wrap gap-2">
                        <Boton variante="secundario" onClick={() => empezarCorreccion(h)}>Corregir</Boton>
                        <Boton variante="secundario" onClick={() => setARetirar(h.id)}>Retirar</Boton>
                      </div>
                    )
                  )}
                </li>
              ))}
            </ul>
          </div>
        </div>
      </div>
    </Tarjeta>
  );
};

export default EditorPieza;
