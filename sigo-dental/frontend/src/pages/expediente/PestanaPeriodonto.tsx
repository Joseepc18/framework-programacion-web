import { useMemo, useState, type FormEvent, type KeyboardEvent } from 'react';
import { Aviso, Boton, Campo, Chip, Selector, Tabla, Tarjeta, Td, Th } from '../../components/ui';
import { ARCADAS } from '../../components/odontograma/dientes';
import { useAccion } from '../../hooks/useAccion';
import { useCarga } from '../../hooks/useCarga';
import {
  crearPeriodontograma, listarPeriodontogramas, media, obtenerPeriodontograma, SITIOS, type DatosExamen, type ExamenPerio,
} from '../../services/periodontoService';
import { fechaLocal, hoyISO } from '../../utils/fechas';

const ORDEN_SUPERIOR: readonly number[] = ARCADAS.permanente.superior;
const ORDEN_INFERIOR: readonly number[] = ARCADAS.permanente.inferior;
const ORDEN = [...ORDEN_SUPERIOR, ...ORDEN_INFERIOR];

// --- Captura (pensada para el sillón: teclado numérico y avance automático) ---

interface SitioForm { pd: string; mg: string; sangra: boolean }
interface DienteForm { mov: string; furca: string; sitios: SitioForm[] }

const dienteVacio = (): DienteForm => ({ mov: '0', furca: '0', sitios: SITIOS.map(() => ({ pd: '', mg: '', sangra: false })) });

// Un número de dos cifras empieza por 1 (10 a 15): con una cifra menor o igual a 9 que no sea 1 ya está completo.
const camposCompletos = (valor: string) => {
  const digitos = valor.replace('-', '');
  if (digitos.length >= 2) return true;
  if (digitos.length === 1 && digitos !== '1') return true;
  return false;
};

// Mueve el foco al siguiente campo de captura en el orden del documento.
function enfocarSiguiente(actual: HTMLElement, atras = false) {
  const campos = [...document.querySelectorAll<HTMLElement>('[data-perio]')];
  const i = campos.indexOf(actual);
  campos[i + (atras ? -1 : 1)]?.focus();
}

const CampoNumero = ({ etiqueta, valor, alCambiar, margen = false, alAlternarSangrado, sangra }: {
  etiqueta: string; valor: string; alCambiar: (v: string) => void; margen?: boolean; alAlternarSangrado?: () => void; sangra?: boolean;
}) => (
  <input
    data-perio
    aria-label={etiqueta}
    inputMode={margen ? 'text' : 'numeric'}
    autoComplete="off"
    maxLength={margen ? 3 : 2}
    placeholder={margen ? '0' : undefined}
    value={valor}
    onFocus={(e) => e.target.select()}
    onChange={(e) => {
      const limpio = e.target.value.replace(margen ? /[^-\d]/g : /\D/g, '');
      alCambiar(limpio);
      if (limpio && camposCompletos(limpio)) enfocarSiguiente(e.target);
    }}
    onKeyDown={(e: KeyboardEvent<HTMLInputElement>) => {
      if (e.key === 'Enter') { e.preventDefault(); enfocarSiguiente(e.currentTarget); }
      // "s" marca o quita el sangrado del punto sin soltar el teclado.
      if ((e.key === 's' || e.key === 'S') && alAlternarSangrado) { e.preventDefault(); alAlternarSangrado(); }
    }}
    className={`h-11 w-11 rounded-control border bg-chip-neutro-fondo text-center font-mono text-cuerpo text-texto focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-acento ${
      sangra ? 'border-chip-alerta-borde' : 'border-chip-neutro-borde'
    }`}
  />
);

const FormularioExamen = ({ pacienteId, ausentes, alGuardar, alCancelar }: {
  pacienteId: number; ausentes: number[]; alGuardar: (id: number) => void; alCancelar: () => void;
}) => {
  const accion = useAccion();
  const [fecha, setFecha] = useState(hoyISO());
  const [notas, setNotas] = useState('');
  const [dientes, setDientes] = useState<Record<number, DienteForm>>({});
  const [errorLocal, setErrorLocal] = useState('');
  const presentes = ORDEN.filter((p) => !ausentes.includes(p));
  const diente = (p: number) => dientes[p] ?? dienteVacio();
  const cambiar = (p: number, cambio: (d: DienteForm) => DienteForm) => setDientes({ ...dientes, [p]: cambio(diente(p)) });
  const cambiarSitio = (p: number, i: number, parte: Partial<SitioForm>) =>
    cambiar(p, (d) => ({ ...d, sitios: d.sitios.map((s, j) => (j === i ? { ...s, ...parte } : s)) }));

  const guardar = async (evento: FormEvent) => {
    evento.preventDefault();
    setErrorLocal('');
    const datos: DatosExamen = { fecha, notas, dientes: [] };
    for (const p of presentes) {
      const d = dientes[p];
      if (!d || d.sitios.every((s) => s.pd === '')) continue;
      if (d.sitios.some((s) => s.pd === '' || Number.isNaN(Number(s.mg)))) {
        setErrorLocal(`Completa los 6 puntos de la pieza ${p} o déjala vacía.`);
        return;
      }
      datos.dientes.push({
        pieza: p, movilidad: Number(d.mov) || 0, furca: Number(d.furca) || 0,
        sitios: d.sitios.map((s, i) => ({ sitio: SITIOS[i], profundidad: Number(s.pd), margen: Number(s.mg) || 0, sangrado: s.sangra })),
      });
    }
    if (datos.dientes.length === 0) {
      setErrorLocal('Captura al menos un diente.');
      return;
    }
    let creado = 0;
    if (await accion.ejecutar(async () => { creado = (await crearPeriodontograma(pacienteId, datos)).id; })) alGuardar(creado);
  };

  return (
    <form onSubmit={guardar} className="space-y-5">
      <Tarjeta className="space-y-4">
        <h2 className="text-seccion font-semibold">Nuevo examen periodontal</h2>
        <p className="text-etiqueta text-texto-suave">
          Escribe la profundidad (mm) y el margen gingival de cada punto: el cursor avanza solo. Con <kbd className="font-mono">s</kbd> marcas sangrado
          en el punto; <kbd className="font-mono">Enter</kbd> salta al siguiente. Margen: positivo si hay recesión y negativo si el margen cubre la
          unión cemento-esmalte. Las piezas ausentes del odontograma no se ofrecen. Una pieza sin profundidades queda fuera del examen.
        </p>
        <div className="grid gap-4 md:grid-cols-2">
          <Campo id="perio-fecha" etiqueta="Fecha del examen" type="date" max={hoyISO()} value={fecha} onChange={(e) => setFecha(e.target.value)} required />
          <Campo id="perio-notas" etiqueta="Notas (opcional)" maxLength={500} value={notas} onChange={(e) => setNotas(e.target.value)} />
        </div>
      </Tarjeta>
      <div className="vidrio-denso overflow-x-auto rounded-tarjeta">
        <table className="w-full min-w-max text-left text-cuerpo">
          <caption className="sr-only">Captura del periodontograma</caption>
          <thead>
            <tr>
              <Th>Pieza</Th><Th>Mov.</Th><Th>Furca</Th>
              {SITIOS.map((s) => <Th key={s}>{s} <span className="font-normal">prof. / margen</span></Th>)}
            </tr>
          </thead>
          <tbody>
            {presentes.map((p) => {
              const d = diente(p);
              return (
                <tr key={p}>
                  <Td className="font-mono font-medium">{p}</Td>
                  <Td><CampoNumero etiqueta={`Pieza ${p}, movilidad 0 a 3`} valor={d.mov} alCambiar={(v) => cambiar(p, (x) => ({ ...x, mov: v.slice(0, 1) }))} /></Td>
                  <Td>{p % 10 >= 4
                    ? <CampoNumero etiqueta={`Pieza ${p}, furca 0 a 3`} valor={d.furca} alCambiar={(v) => cambiar(p, (x) => ({ ...x, furca: v.slice(0, 1) }))} />
                    : <span className="text-texto-suave">—</span>}</Td>
                  {SITIOS.map((s, i) => (
                    <Td key={s}>
                      <div className="flex gap-1">
                        <CampoNumero etiqueta={`Pieza ${p}, ${s}, profundidad al sondaje`} valor={d.sitios[i].pd} sangra={d.sitios[i].sangra}
                          alCambiar={(v) => cambiarSitio(p, i, { pd: v })} alAlternarSangrado={() => cambiarSitio(p, i, { sangra: !d.sitios[i].sangra })} />
                        <CampoNumero etiqueta={`Pieza ${p}, ${s}, margen gingival`} valor={d.sitios[i].mg} margen sangra={d.sitios[i].sangra}
                          alCambiar={(v) => cambiarSitio(p, i, { mg: v })} alAlternarSangrado={() => cambiarSitio(p, i, { sangra: !d.sitios[i].sangra })} />
                      </div>
                      {d.sitios[i].sangra && <span className="text-etiqueta text-chip-alerta-texto">sangra</span>}
                    </Td>
                  ))}
                </tr>
              );
            })}
          </tbody>
        </table>
      </div>
      {(errorLocal || accion.error) && <Aviso>{errorLocal || accion.error}</Aviso>}
      <div className="flex flex-wrap gap-3">
        <Boton type="submit" disabled={accion.ocupado}>{accion.ocupado ? 'Guardando…' : 'Guardar examen'}</Boton>
        <Boton variante="secundario" onClick={alCancelar}>Cancelar</Boton>
      </div>
    </form>
  );
};

// --- Lectura: tabla y gráfica por arcada ---

const Grafica = ({ examen, piezas, titulo }: { examen: ExamenPerio; piezas: readonly number[]; titulo: string }) => {
  const ancho = 640, alto = 180, margenIzq = 28, margenInf = 24, paso = (ancho - margenIzq - 10) / (piezas.length - 1);
  const datos = piezas.map((p) => {
    const d = examen.dientes.find((x) => x.pieza === p);
    return d ? { p, pd: media(d.sitios.map((s) => s.profundidad)), nic: media(d.sitios.map((s) => s.nic)) } : { p, pd: null, nic: null };
  });
  const tope = Math.max(10, Math.ceil(Math.max(...datos.map((d) => d.nic ?? 0))));
  const y = (v: number) => alto - margenInf - (v / tope) * (alto - margenInf - 10);
  const linea = (clave: 'pd' | 'nic') =>
    datos.map((d, i) => (d[clave] === null ? null : `${margenIzq + i * paso},${y(d[clave])}`)).filter(Boolean).join(' ');
  return (
    <figure>
      <figcaption className="mb-1 text-etiqueta text-texto-suave">{titulo}</figcaption>
      <svg viewBox={`0 0 ${ancho} ${alto}`} className="w-full" role="img" aria-label={`${titulo}: profundidad media y nivel de inserción medio por pieza`}>
        {[0, tope / 2, tope].map((v) => (
          <g key={v}>
            <line x1={margenIzq} x2={ancho - 6} y1={y(v)} y2={y(v)} className="stroke-vidrio-borde" />
            <text x={margenIzq - 4} y={y(v) + 4} textAnchor="end" className="fill-texto-suave font-mono text-[10px]">{v}</text>
          </g>
        ))}
        <polyline points={linea('nic')} className="fill-none stroke-odonto-caries" strokeWidth={2} strokeDasharray="5 3" />
        <polyline points={linea('pd')} className="fill-none stroke-acento" strokeWidth={2} />
        {datos.map((d, i) => (
          <text key={d.p} x={margenIzq + i * paso} y={alto - 8} textAnchor="middle" className="fill-texto-suave font-mono text-[10px]">{d.p}</text>
        ))}
      </svg>
    </figure>
  );
};

const VistaExamen = ({ examen, anterior }: { examen: ExamenPerio; anterior: ExamenPerio | null }) => {
  const nicMedio = (e: ExamenPerio | null, pieza: number) => {
    const d = e?.dientes.find((x) => x.pieza === pieza);
    return d ? media(d.sitios.map((s) => s.nic)) : null;
  };
  const ordenados = ORDEN.filter((p) => examen.dientes.some((d) => d.pieza === p));
  return (
    <div className="space-y-5">
      <Tarjeta className="space-y-3">
        <div className="flex flex-wrap items-center gap-3">
          <h2 className="text-seccion font-semibold">Examen del {fechaLocal(`${examen.fecha}T12:00:00`)}</h2>
          <Chip>{examen.dientes.length} dientes</Chip>
          <span className="text-etiqueta text-texto-suave">{examen.creado_por}</span>
        </div>
        {examen.notas && <p className="text-etiqueta">{examen.notas}</p>}
        <div className="grid gap-4 lg:grid-cols-2">
          <Grafica examen={examen} piezas={ORDEN_SUPERIOR} titulo="Arcada superior" />
          <Grafica examen={examen} piezas={ORDEN_INFERIOR} titulo="Arcada inferior" />
        </div>
        <p className="text-etiqueta text-texto-suave">
          <span className="text-acento">━</span> profundidad media · <span className="text-chip-alerta-texto">╍</span> nivel de inserción clínica (NIC) medio. NIC = profundidad + margen gingival.
        </p>
      </Tarjeta>
      <Tabla etiqueta="Detalle del periodontograma">
        <thead>
          <tr>
            <Th>Pieza</Th><Th>Mov.</Th><Th>Furca</Th>
            {SITIOS.map((s) => <Th key={s}>{s} <span className="font-normal">prof. · margen → NIC</span></Th>)}
            {anterior && <Th>Cambio de NIC medio</Th>}
          </tr>
        </thead>
        <tbody>
          {ordenados.map((p) => {
            const d = examen.dientes.find((x) => x.pieza === p)!;
            const antes = nicMedio(anterior, p);
            const ahora = nicMedio(examen, p);
            const delta = antes !== null && ahora !== null ? ahora - antes : null;
            return (
              <tr key={p}>
                <Td className="font-mono font-medium">{p}</Td>
                <Td className="font-mono">{d.movilidad}</Td>
                <Td className="font-mono">{d.furca === 0 ? '—' : ['', 'I', 'II', 'III'][d.furca]}</Td>
                {d.sitios.map((s) => (
                  <Td key={s.sitio} className="font-mono text-etiqueta">
                    {s.profundidad} · {s.margen} → <strong>{s.nic}</strong>{s.sangrado && <span className="ml-1 text-chip-alerta-texto"> sangra</span>}
                  </Td>
                ))}
                {anterior && <Td className="font-mono text-etiqueta">{delta === null ? '—' : `${delta > 0 ? '+' : ''}${delta.toFixed(1)} mm${delta > 0.5 ? ' (empeora)' : delta < -0.5 ? ' (mejora)' : ''}`}</Td>}
              </tr>
            );
          })}
        </tbody>
      </Tabla>
    </div>
  );
};

// Periodontograma del paciente (RF-04): exámenes con fecha para comparar la evolución.
const PestanaPeriodonto = ({ pacienteId }: { pacienteId: number }) => {
  const lista = useCarga(() => listarPeriodontogramas(pacienteId), `perio-${pacienteId}`);
  const [capturando, setCapturando] = useState(false);
  const [elegido, setElegido] = useState<number | null>(null);
  const [compararCon, setCompararCon] = useState<number | null>(null);

  const examenes = useMemo(() => lista.datos?.examenes ?? [], [lista.datos]);
  const idActual = elegido ?? examenes[0]?.id ?? null;
  const examen = useCarga(() => (idActual ? obtenerPeriodontograma(pacienteId, idActual) : Promise.resolve(null)), `perio-${pacienteId}-${idActual}`);
  const otro = useCarga(() => (compararCon ? obtenerPeriodontograma(pacienteId, compararCon) : Promise.resolve(null)), `perio-${pacienteId}-cmp-${compararCon}`);

  if (lista.error) return <Aviso>{lista.error}</Aviso>;
  if (!lista.datos) return <p>Cargando…</p>;

  if (capturando) {
    return (
      <FormularioExamen pacienteId={pacienteId} ausentes={lista.datos.ausentes} alCancelar={() => setCapturando(false)}
        alGuardar={(id) => { setCapturando(false); setElegido(id); lista.recargar(); }} />
    );
  }

  return (
    <div className="space-y-6">
      <Tarjeta className="flex flex-wrap items-end gap-4">
        <Boton onClick={() => setCapturando(true)}>Nuevo examen</Boton>
        {examenes.length > 0 && (
          <>
            <div className="w-64">
              <Selector id="perio-examen" etiqueta="Examen" value={idActual ?? ''} onChange={(e) => setElegido(Number(e.target.value))}>
                {examenes.map((e) => (
                  <option key={e.id} value={e.id}>{fechaLocal(`${e.fecha}T12:00:00`)} · NIC medio {e.nic_medio.toFixed(1)} mm · sangrado {e.sangrado_porcentaje.toFixed(0)} %</option>
                ))}
              </Selector>
            </div>
            <div className="w-64">
              <Selector id="perio-comparar" etiqueta="Comparar con" value={compararCon ?? ''} onChange={(e) => setCompararCon(e.target.value ? Number(e.target.value) : null)}>
                <option value="">Sin comparar</option>
                {examenes.filter((e) => e.id !== idActual).map((e) => <option key={e.id} value={e.id}>{fechaLocal(`${e.fecha}T12:00:00`)}</option>)}
              </Selector>
            </div>
          </>
        )}
      </Tarjeta>
      {examenes.length === 0 && <p className="text-texto-suave">Todavía no hay exámenes periodontales.</p>}
      {examen.error && <Aviso>{examen.error}</Aviso>}
      {examen.datos && <VistaExamen examen={examen.datos} anterior={compararCon ? otro.datos : null} />}
    </div>
  );
};

export default PestanaPeriodonto;
