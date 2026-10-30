import { useEffect, useState, type FormEvent } from 'react';
import { useNavigate, useSearchParams } from 'react-router-dom';
import { Aviso, Boton, BotonEnlace, Campo, Selector, Tarjeta, Titulo } from '../components/ui';
import { useAccion } from '../hooks/useAccion';
import { useCarga } from '../hooks/useCarga';
import { listarProcedimientos, type Procedimiento } from '../services/catalogoService';
import { obtenerCita } from '../services/citaService';
import { crearCobro, type LineaPedida } from '../services/cobroService';
import { listarPacientes } from '../services/pacienteService';
import { formatearDinero } from '../utils/dinero';

interface LineaEditable {
  clave: number;
  procedimientoId: number;
  cantidad: number;
}

// Mismo cálculo entero que el servidor (descuento en puntos básicos, redondeo a la mitad hacia arriba): es solo una
// vista previa; el total definitivo lo calcula el servidor.
const calcular = (subtotal: number, bp: number) => {
  const descuento = Math.floor((subtotal * bp + 5000) / 10000);
  return { descuento, total: subtotal - descuento };
};

// Convierte "12", "12.5" o "12,50" (porcentaje) a puntos básicos; null si no es válido o pasa de 100.
function aPuntosBasicos(texto: string): number | null {
  const limpio = texto.trim().replace(',', '.');
  if (limpio === '') return 0;
  if (!/^\d+(\.\d{1,2})?$/.test(limpio)) return null;
  const [enteros, decimales = ''] = limpio.split('.');
  const bp = Number(enteros) * 100 + Number(decimales.padEnd(2, '0'));
  return bp <= 10000 ? bp : null;
}

const Formulario = ({ procedimientos, citaId, pacienteInicial, procedimientoInicial }: {
  procedimientos: Procedimiento[];
  citaId?: number;
  pacienteInicial?: { id: number; nombre: string };
  procedimientoInicial?: number;
}) => {
  const navegar = useNavigate();
  const accion = useAccion();
  const [texto, setTexto] = useState('');
  const [busqueda, setBusqueda] = useState('');
  const [pacienteId, setPacienteId] = useState(pacienteInicial?.id ?? 0);
  const [lineas, setLineas] = useState<LineaEditable[]>([
    { clave: 1, procedimientoId: procedimientoInicial ?? procedimientos[0]?.id ?? 0, cantidad: 1 },
  ]);
  const [descuento, setDescuento] = useState('');

  useEffect(() => {
    const espera = setTimeout(() => setBusqueda(texto), 300);
    return () => clearTimeout(espera);
  }, [texto]);
  const pacientes = useCarga(() => (citaId ? Promise.resolve([]) : listarPacientes(busqueda)), `pacientes-cobro|${busqueda}`);

  const precioDe = (id: number) => procedimientos.find((p) => p.id === id)?.precio_centavos ?? 0;
  const subtotal = lineas.reduce((suma, l) => suma + precioDe(l.procedimientoId) * l.cantidad, 0);
  const bp = aPuntosBasicos(descuento);
  const { descuento: importeDescuento, total } = calcular(subtotal, bp ?? 0);

  const cambiarLinea = (clave: number, cambio: Partial<LineaEditable>) =>
    setLineas((actuales) => actuales.map((l) => (l.clave === clave ? { ...l, ...cambio } : l)));

  const guardar = async (evento: FormEvent) => {
    evento.preventDefault();
    if (bp === null) return;
    const pedidas: LineaPedida[] = lineas.map((l) => ({ procedimiento_id: l.procedimientoId, cantidad: l.cantidad }));
    let destino = '';
    const bien = await accion.ejecutar(async () => {
      const cobro = await crearCobro({ ...(citaId ? { cita_id: citaId } : { paciente_id: pacienteId }), descuento_bp: bp, lineas: pedidas });
      destino = `/cobros/${cobro.id}`;
    });
    if (bien) navegar(destino);
  };

  return (
    <form onSubmit={guardar} className="space-y-6">
      <Tarjeta>
        <h2 className="text-seccion font-semibold">Paciente</h2>
        {citaId ? (
          <p className="mt-3">{pacienteInicial?.nombre} <span className="text-texto-suave">· cobro de una cita atendida</span></p>
        ) : (
          <div className="mt-5 grid gap-5 md:grid-cols-2">
            <Campo id="cobro-buscar" etiqueta="Buscar paciente" ayuda="Nombre, documento o número de historia" value={texto} onChange={(e) => setTexto(e.target.value)} />
            <Selector id="cobro-paciente" etiqueta="Paciente" required value={pacienteId || ''} onChange={(e) => setPacienteId(Number(e.target.value))}>
              <option value="">{pacientes.cargando ? 'Buscando…' : 'Elige un paciente'}</option>
              {(pacientes.datos ?? []).map((p) => <option key={p.id} value={p.id}>{p.apellidos} {p.nombres} · {p.numero_historia}</option>)}
            </Selector>
          </div>
        )}
      </Tarjeta>

      <Tarjeta>
        <h2 className="text-seccion font-semibold">Procedimientos</h2>
        <div className="mt-5 space-y-4">
          {lineas.map((l, i) => (
            <div key={l.clave} className="grid items-end gap-4 md:grid-cols-[1fr_8rem_9rem_auto]">
              <Selector id={`linea-proc-${l.clave}`} etiqueta={`Procedimiento ${i + 1}`} value={l.procedimientoId} onChange={(e) => cambiarLinea(l.clave, { procedimientoId: Number(e.target.value) })}>
                {procedimientos.map((p) => <option key={p.id} value={p.id}>{p.nombre} · {formatearDinero(p.precio_centavos)}</option>)}
              </Selector>
              <Campo id={`linea-cant-${l.clave}`} etiqueta="Cantidad" type="number" min={1} max={999} required value={l.cantidad} onChange={(e) => cambiarLinea(l.clave, { cantidad: Math.max(1, Math.floor(Number(e.target.value)) || 1) })} />
              <p className="min-h-11 py-2 font-mono md:text-right" aria-label="Importe de la línea">{formatearDinero(precioDe(l.procedimientoId) * l.cantidad)}</p>
              <Boton variante="secundario" disabled={lineas.length === 1} onClick={() => setLineas((actuales) => actuales.filter((x) => x.clave !== l.clave))}>
                Quitar
              </Boton>
            </div>
          ))}
        </div>
        <div className="mt-4">
          <Boton variante="secundario" disabled={lineas.length >= 30} onClick={() => setLineas((actuales) => [...actuales, { clave: Math.max(...actuales.map((x) => x.clave)) + 1, procedimientoId: procedimientos[0]?.id ?? 0, cantidad: 1 }])}>
            Agregar procedimiento
          </Boton>
        </div>
      </Tarjeta>

      <Tarjeta>
        <div className="grid gap-5 md:grid-cols-2">
          <Campo id="cobro-descuento" etiqueta="Descuento (%)" inputMode="decimal" placeholder="0" value={descuento} onChange={(e) => setDescuento(e.target.value)}
            ayuda={bp === null ? undefined : 'Hasta 100 %, con dos decimales'} aria-invalid={bp === null} />
          <dl className="space-y-1 md:text-right">
            <div><dt className="inline text-texto-suave">Subtotal: </dt><dd className="inline font-mono">{formatearDinero(subtotal)}</dd></div>
            <div><dt className="inline text-texto-suave">Descuento: </dt><dd className="inline font-mono">−{formatearDinero(importeDescuento)}</dd></div>
            <div className="text-seccion font-semibold"><dt className="inline">Total: </dt><dd className="inline font-mono">{formatearDinero(total)}</dd></div>
          </dl>
        </div>
        {bp === null && <div className="mt-3"><Aviso>El descuento debe ser un porcentaje entre 0 y 100, con hasta dos decimales.</Aviso></div>}
        <p className="mt-3 text-etiqueta text-texto-suave">Es una vista previa: el precio y el total definitivos los calcula el servidor con el catálogo vigente.</p>
      </Tarjeta>

      {accion.error && <Aviso>{accion.error}</Aviso>}
      <div className="flex flex-wrap gap-3">
        <Boton type="submit" disabled={accion.ocupado || bp === null || (!citaId && !pacienteId) || lineas.some((l) => !l.procedimientoId)}>
          {accion.ocupado ? 'Generando…' : 'Generar cobro'}
        </Boton>
        <BotonEnlace variante="secundario" to="/cobros">Cancelar</BotonEnlace>
      </div>
    </form>
  );
};

// Alta de un cobro: de una cita atendida (?cita=ID, que ya trae al paciente y su procedimiento) o de un paciente.
const CobroFormPage = () => {
  const [parametros] = useSearchParams();
  const citaId = Number(parametros.get('cita')) || undefined;
  const procedimientos = useCarga(() => listarProcedimientos(), 'procedimientos-cobro');
  const cita = useCarga(() => (citaId ? obtenerCita(citaId) : Promise.resolve(null)), `cita-cobro-${citaId ?? 'ninguna'}`);

  const error = procedimientos.error || cita.error;
  const listo = procedimientos.datos && !cita.cargando && (!citaId || cita.datos);

  return (
    <section aria-labelledby="titulo-nuevo-cobro" className="space-y-6">
      <Titulo id="titulo-nuevo-cobro">Nuevo cobro</Titulo>
      {error && <Aviso>{error}</Aviso>}
      {!listo && !error && <p>Cargando…</p>}
      {listo && (
        <Formulario
          procedimientos={procedimientos.datos ?? []}
          citaId={citaId}
          pacienteInicial={cita.datos ? { id: cita.datos.paciente_id, nombre: cita.datos.paciente_nombre } : undefined}
          procedimientoInicial={cita.datos?.procedimiento_id ?? undefined}
        />
      )}
    </section>
  );
};

export default CobroFormPage;
