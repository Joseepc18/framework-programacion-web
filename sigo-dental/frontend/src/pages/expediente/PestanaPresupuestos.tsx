import { useState, type FormEvent } from 'react';
import { useNavigate } from 'react-router-dom';
import { Aviso, Boton, Campo, CampoTexto, Casilla, Chip, Selector, Tarjeta } from '../../components/ui';
import { useAuth } from '../../context/useAuth';
import { useAccion } from '../../hooks/useAccion';
import { mensajeDeApi, useCarga } from '../../hooks/useCarga';
import { listarProcedimientos, type Procedimiento } from '../../services/catalogoService';
import { iniciarProcedimiento } from '../../services/evolucionService';
import {
  actualizarPresupuesto, crearPresupuesto, ETIQUETA_ESTADO_PRESUPUESTO, listarPresupuestos, listarTareas, nuevaVersionPresupuesto, pdfPresupuesto,
  presentarPresupuesto, resolverPresupuesto, type ItemPedido, type Presupuesto, type Tarea,
} from '../../services/presupuestoService';
import { abrirBlob } from '../../utils/archivos';
import { aCentavos, deCentavos, formatearDinero } from '../../utils/dinero';
import { fechaHoraLocal } from '../../utils/fechas';

// --- Tareas agendables (RF-08) ---

const ListaTareas = ({ pacienteId, pacienteNombre, tareas, alCambiar }: { pacienteId: number; pacienteNombre: string; tareas: Tarea[]; alCambiar: () => void }) => {
  const { usuario } = useAuth();
  const navegar = useNavigate();
  const accion = useAccion();
  const puedeAgendar = usuario?.rol === 'recepcionista';
  const puedeIniciar = usuario?.rol === 'odontologo';

  const agendar = (t: Tarea) => {
    const params = new URLSearchParams({ agendar: String(pacienteId), nombre: pacienteNombre, procedimiento: String(t.procedimiento_id), item: String(t.item_id), min: String(t.duracion_min) });
    navegar(`/agenda?${params}`);
  };

  const iniciar = (t: Tarea) => accion.ejecutar(async () => {
    await iniciarProcedimiento(pacienteId, { procedimiento_id: t.procedimiento_id, pieza: t.pieza ?? undefined, superficie: t.superficie || undefined, presupuesto_item_id: t.item_id, notas: '' });
    alCambiar();
  });

  return (
    <Tarjeta className="space-y-4">
      <h2 className="text-seccion font-semibold">Tratamientos aprobados por agendar</h2>
      {tareas.length === 0 && <p className="text-texto-suave">No hay tratamientos aprobados. Cuando se apruebe un presupuesto, sus ítems aparecen aquí.</p>}
      {accion.error && <Aviso>{accion.error}</Aviso>}
      <ul className="space-y-2">
        {tareas.map((t) => (
          <li key={t.item_id} className="flex flex-wrap items-center gap-3 rounded-control border border-vidrio-borde p-3">
            <span className="font-medium">{t.descripcion}</span>
            {t.pieza && <Chip>Pieza {t.pieza}</Chip>}
            <span className="text-etiqueta text-texto-suave">Fase {t.fase} · sesión {t.sesion} · {t.presupuesto_numero} v{t.version} · {t.duracion_min} min</span>
            <Chip tono={t.estado === 'cumplido' ? 'acento' : t.estado === 'agendado' ? 'neutro' : 'alerta'}>
              {t.estado === 'cumplido' ? 'Cumplido' : t.estado === 'agendado' ? `Agendado${t.cita_inicio ? ` · ${fechaHoraLocal(t.cita_inicio)}` : ''}` : 'Por agendar'}
            </Chip>
            {t.estado === 'aprobado' && puedeAgendar && <Boton onClick={() => agendar(t)}>Agendar cita</Boton>}
            {t.estado !== 'cumplido' && puedeIniciar && <Boton variante="secundario" disabled={accion.ocupado} onClick={() => iniciar(t)}>Iniciar procedimiento</Boton>}
          </li>
        ))}
      </ul>
    </Tarjeta>
  );
};

// --- Constructor de presupuestos ---

interface FilaItem {
  id?: number;
  fase: string;
  sesion: string;
  procedimientoId: string;
  pieza: string;
  superficie: string;
  honorario: string;
  laboratorio: string;
  descuento: string;
}

const FILA_VACIA = (fase = '1', sesion = '1'): FilaItem => ({ fase, sesion, procedimientoId: '', pieza: '', superficie: '', honorario: '', laboratorio: '0.00', descuento: '0.00' });

const filaDe = (it: Presupuesto['items'][number]): FilaItem => ({
  id: it.id, fase: String(it.fase), sesion: String(it.sesion), procedimientoId: String(it.procedimiento_id), pieza: it.pieza ? String(it.pieza) : '',
  superficie: it.superficie, honorario: deCentavos(it.honorario_centavos), laboratorio: deCentavos(it.laboratorio_centavos), descuento: deCentavos(it.descuento_centavos),
});

const Constructor = ({ pacienteId, inicial, procedimientos, alTerminar, alCancelar }: {
  pacienteId: number; inicial?: Presupuesto; procedimientos: Procedimiento[]; alTerminar: () => void; alCancelar: () => void;
}) => {
  const accion = useAccion();
  const [filas, setFilas] = useState<FilaItem[]>(inicial?.items.map(filaDe) ?? [FILA_VACIA()]);
  const [notas, setNotas] = useState(inicial?.notas ?? '');
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const moneda = inicial?.moneda ?? 'USD';

  const cambiar = (i: number, campo: keyof FilaItem, valor: string) => setFilas(filas.map((f, j) => (j === i ? { ...f, [campo]: valor } : f)));
  const elegirProcedimiento = (i: number, id: string) => {
    const proc = procedimientos.find((p) => p.id === Number(id));
    setFilas(filas.map((f, j) => (j === i ? { ...f, procedimientoId: id, honorario: proc ? deCentavos(proc.precio_centavos) : '' } : f)));
  };

  // Vista previa de los totales; el servidor los vuelve a calcular y es quien manda.
  const centavos = (texto: string) => aCentavos(texto || '0') ?? 0;
  const totalFila = (f: FilaItem) => centavos(f.honorario) + centavos(f.laboratorio) - centavos(f.descuento);
  const total = filas.reduce((suma, f) => suma + totalFila(f), 0);
  const hayDescuento = filas.some((f) => centavos(f.descuento) > 0);

  const guardar = async (evento: FormEvent) => {
    evento.preventDefault();
    const items: ItemPedido[] = filas.map((f) => ({
      id: f.id, fase: Number(f.fase), sesion: Number(f.sesion), procedimiento_id: Number(f.procedimientoId),
      pieza: f.pieza ? Number(f.pieza) : undefined, superficie: f.superficie || undefined,
      honorario_centavos: aCentavos(f.honorario || '0') ?? 0, laboratorio_centavos: centavos(f.laboratorio), descuento_centavos: centavos(f.descuento),
    }));
    const datos = { notas, items, autorizacion: hayDescuento && email ? { email, password } : undefined };
    if (await accion.ejecutar(() => (inicial ? actualizarPresupuesto(pacienteId, inicial.id, datos) : crearPresupuesto(pacienteId, datos)))) alTerminar();
  };

  return (
    <form onSubmit={guardar} className="space-y-4">
      {filas.map((f, i) => (
        <fieldset key={i} className="space-y-3 rounded-control border border-vidrio-borde p-4">
          <legend className="px-2 text-etiqueta text-texto-suave">Ítem {i + 1} · total {formatearDinero(totalFila(f), moneda)}</legend>
          <div className="grid gap-3 md:grid-cols-6">
            <Campo id={`it-fase-${i}`} etiqueta="Fase" inputMode="numeric" required value={f.fase} onChange={(e) => cambiar(i, 'fase', e.target.value.replace(/\D/g, ''))} />
            <Campo id={`it-sesion-${i}`} etiqueta="Sesión" inputMode="numeric" required value={f.sesion} onChange={(e) => cambiar(i, 'sesion', e.target.value.replace(/\D/g, ''))} />
            <div className="md:col-span-3">
              <Selector id={`it-proc-${i}`} etiqueta="Procedimiento" required value={f.procedimientoId} onChange={(e) => elegirProcedimiento(i, e.target.value)}>
                <option value="">Elige…</option>
                {procedimientos.filter((p) => p.activo).map((p) => <option key={p.id} value={p.id}>{p.nombre} · {formatearDinero(p.precio_centavos, moneda)}</option>)}
              </Selector>
            </div>
            <Campo id={`it-pieza-${i}`} etiqueta="Pieza FDI" inputMode="numeric" maxLength={2} value={f.pieza} onChange={(e) => cambiar(i, 'pieza', e.target.value.replace(/\D/g, ''))} />
          </div>
          <div className="grid gap-3 md:grid-cols-4">
            <Campo id={`it-hon-${i}`} etiqueta="Honorarios" ayuda="Editable sobre el precio base" inputMode="decimal" required value={f.honorario} onChange={(e) => cambiar(i, 'honorario', e.target.value)} />
            <Campo id={`it-lab-${i}`} etiqueta="Costo de laboratorio" inputMode="decimal" value={f.laboratorio} onChange={(e) => cambiar(i, 'laboratorio', e.target.value)} />
            <Campo id={`it-desc-${i}`} etiqueta="Descuento" ayuda="Requiere autorización" inputMode="decimal" value={f.descuento} onChange={(e) => cambiar(i, 'descuento', e.target.value)} />
            <Selector id={`it-sup-${i}`} etiqueta="Superficie" value={f.superficie} onChange={(e) => cambiar(i, 'superficie', e.target.value)}>
              <option value="">—</option>
              {['mesial', 'distal', 'oclusal', 'vestibular', 'lingual'].map((s) => <option key={s} value={s}>{s}</option>)}
            </Selector>
          </div>
          {filas.length > 1 && <Boton variante="secundario" onClick={() => setFilas(filas.filter((_, j) => j !== i))}>Quitar ítem</Boton>}
        </fieldset>
      ))}
      <div className="flex flex-wrap gap-3">
        <Boton variante="secundario" onClick={() => { const ult = filas[filas.length - 1]; setFilas([...filas, FILA_VACIA(ult?.fase ?? '1', String((Number(ult?.sesion) || 1) + 1))]); }}>Agregar ítem</Boton>
        <Boton variante="secundario" onClick={() => { const ult = filas[filas.length - 1]; setFilas([...filas, FILA_VACIA(String((Number(ult?.fase) || 1) + 1), '1')]); }}>Agregar fase nueva</Boton>
      </div>
      <CampoTexto id="pres-notas" etiqueta="Notas (opcional)" maxLength={500} value={notas} onChange={(e) => setNotas(e.target.value)} />
      {hayDescuento && (
        <div className="space-y-3 rounded-control border border-chip-acento-borde p-4">
          <p className="text-etiqueta">Hay descuentos: necesitan la autorización de un odontólogo o administrador. Escribe su correo y contraseña (si ya estaban autorizados y no cambiaron, no hace falta).</p>
          <div className="grid gap-3 md:grid-cols-2">
            <Campo id="aut-email" etiqueta="Correo de quien autoriza" type="email" value={email} onChange={(e) => setEmail(e.target.value)} />
            <Campo id="aut-pass" etiqueta="Contraseña de quien autoriza" type="password" autoComplete="off" value={password} onChange={(e) => setPassword(e.target.value)} />
          </div>
        </div>
      )}
      <p className="text-seccion font-semibold">Total (vista previa): {formatearDinero(total, moneda)}</p>
      {accion.error && <Aviso>{accion.error}</Aviso>}
      <div className="flex flex-wrap gap-3">
        <Boton type="submit" disabled={accion.ocupado}>{accion.ocupado ? 'Guardando…' : inicial ? 'Guardar cambios' : 'Guardar borrador'}</Boton>
        <Boton variante="secundario" onClick={alCancelar}>Cancelar</Boton>
      </div>
    </form>
  );
};

// --- Tarjeta de un presupuesto ---

const TarjetaPresupuesto = ({ p, pacienteId, procedimientos, alCambiar }: { p: Presupuesto; pacienteId: number; procedimientos: Procedimiento[]; alCambiar: () => void }) => {
  const { usuario } = useAuth();
  const accion = useAccion();
  const [editando, setEditando] = useState(false);
  const [resolviendo, setResolviendo] = useState(false);
  const [aprobados, setAprobados] = useState<number[]>(p.items.map((i) => i.id));
  const [errorPdf, setErrorPdf] = useState('');
  const esOdontologo = usuario?.rol === 'odontologo';
  const cambiar = (f: () => Promise<unknown>) => accion.ejecutar(async () => { await f(); setEditando(false); setResolviendo(false); alCambiar(); });

  const abrirPdf = async () => {
    setErrorPdf('');
    try {
      abrirBlob(await pdfPresupuesto(pacienteId, p.id), `presupuesto-${p.numero}.pdf`);
    } catch (fallo) {
      setErrorPdf(mensajeDeApi(fallo));
    }
  };

  const fases = [...new Set(p.items.map((i) => i.fase))];

  return (
    <Tarjeta className="space-y-4">
      <div className="flex flex-wrap items-center gap-3">
        <h3 className="font-mono text-seccion font-semibold">{p.numero} · v{p.version}</h3>
        <Chip tono={p.estado === 'borrador' || p.estado === 'presentado' ? 'neutro' : p.estado === 'rechazado' ? 'alerta' : 'acento'}>{ETIQUETA_ESTADO_PRESUPUESTO[p.estado]}</Chip>
        {p.reemplazado && <Chip>Reemplazado por una versión nueva</Chip>}
        <span className="text-etiqueta text-texto-suave">{p.profesional_nombre} · {fechaHoraLocal(p.creado_en)}</span>
      </div>

      {editando ? (
        <Constructor pacienteId={pacienteId} inicial={p} procedimientos={procedimientos} alTerminar={() => { setEditando(false); alCambiar(); }} alCancelar={() => setEditando(false)} />
      ) : (
        <div className="space-y-3">
          {fases.map((fase) => (
            <div key={fase}>
              <h4 className="text-etiqueta font-semibold text-texto-suave">Fase {fase}</h4>
              <ul className="mt-1 space-y-1">
                {p.items.filter((i) => i.fase === fase).map((it) => (
                  <li key={it.id} className="flex flex-wrap items-center gap-2">
                    {resolviendo && <Casilla id={`ap-${it.id}`} etiqueta="" aria-label={`Aprobar ${it.descripcion}`} checked={aprobados.includes(it.id)} onChange={(e) => setAprobados(e.target.checked ? [...aprobados, it.id] : aprobados.filter((x) => x !== it.id))} />}
                    <span>Sesión {it.sesion} · <strong>{it.descripcion}</strong>{it.pieza ? ` · pieza ${it.pieza}${it.superficie ? ` (${it.superficie})` : ''}` : ''}</span>
                    <span className="font-mono">{formatearDinero(it.total_centavos, p.moneda)}</span>
                    {it.descuento_centavos > 0 && <span className="text-etiqueta text-texto-suave">(honorarios {formatearDinero(it.honorario_centavos, p.moneda)} + lab. {formatearDinero(it.laboratorio_centavos, p.moneda)} − descuento {formatearDinero(it.descuento_centavos, p.moneda)}, autorizó {it.descuento_autorizado_por})</span>}
                    {it.aprobado === true && <Chip tono="acento">{it.estado === 'cumplido' ? 'Cumplido' : it.estado === 'agendado' ? 'Agendado' : 'Aprobado'}</Chip>}
                    {it.aprobado === false && <Chip tono="alerta">Rechazado</Chip>}
                  </li>
                ))}
              </ul>
            </div>
          ))}
          <p className="text-etiqueta text-texto-suave">
            Honorarios {formatearDinero(p.honorarios_centavos, p.moneda)} · laboratorio {formatearDinero(p.laboratorio_centavos, p.moneda)} · descuentos −{formatearDinero(p.descuentos_centavos, p.moneda)}
          </p>
          <p className="text-seccion font-semibold">Total {formatearDinero(p.total_centavos, p.moneda)}{p.estado === 'aprobado_parcial' && ` · aprobado ${formatearDinero(p.aprobado_centavos, p.moneda)}`}</p>
          {p.notas && <p className="text-etiqueta">{p.notas}</p>}
          {p.presentado_en && <p className="text-etiqueta text-texto-suave">Presentado por {p.presentado_por} · {fechaHoraLocal(p.presentado_en)}</p>}
          {p.resuelto_en && <p className="text-etiqueta text-texto-suave">Resuelto por {p.resuelto_por} · {fechaHoraLocal(p.resuelto_en)}</p>}
        </div>
      )}

      {errorPdf && <Aviso>{errorPdf}</Aviso>}
      {accion.error && <Aviso>{accion.error}</Aviso>}
      {!editando && (
        <div className="flex flex-wrap gap-3">
          <Boton variante="secundario" onClick={abrirPdf}>Ver PDF</Boton>
          {p.estado === 'borrador' && esOdontologo && <Boton variante="secundario" onClick={() => setEditando(true)}>Editar</Boton>}
          {p.estado === 'borrador' && <Boton disabled={accion.ocupado} onClick={() => cambiar(() => presentarPresupuesto(pacienteId, p.id))}>Presentar al paciente</Boton>}
          {p.estado === 'presentado' && !resolviendo && <Boton onClick={() => setResolviendo(true)}>Registrar la decisión del paciente</Boton>}
          {resolviendo && (
            <>
              <Boton disabled={accion.ocupado} onClick={() => cambiar(() => resolverPresupuesto(pacienteId, p.id, aprobados))}>
                {aprobados.length === p.items.length ? 'Aprobar todo' : aprobados.length === 0 ? 'Rechazar el presupuesto' : `Aprobar ${aprobados.length} de ${p.items.length}`}
              </Boton>
              <Boton variante="secundario" onClick={() => setResolviendo(false)}>Cancelar</Boton>
            </>
          )}
          {p.estado !== 'borrador' && !p.reemplazado && esOdontologo && <Boton variante="secundario" disabled={accion.ocupado} onClick={() => cambiar(() => nuevaVersionPresupuesto(pacienteId, p.id))}>Crear versión nueva</Boton>}
        </div>
      )}
      {resolviendo && <p className="text-etiqueta text-texto-suave">Marca los ítems que el paciente aprueba. Los aprobados pasan a ser tratamientos por agendar; los demás quedan rechazados.</p>}
    </Tarjeta>
  );
};

// Presupuestos del paciente (RF-07 y RF-08): plan por fases y sesiones, aprobación total o parcial y tareas agendables.
const PestanaPresupuestos = ({ pacienteId, pacienteNombre }: { pacienteId: number; pacienteNombre: string }) => {
  const { usuario } = useAuth();
  const presupuestos = useCarga(() => listarPresupuestos(pacienteId), `presupuestos-${pacienteId}`);
  const tareas = useCarga(() => listarTareas(pacienteId), `tareas-${pacienteId}`);
  const catalogo = useCarga(() => listarProcedimientos(true), 'procedimientos-todos');
  const [nuevo, setNuevo] = useState(false);
  const recargar = () => { presupuestos.recargar(); tareas.recargar(); };

  return (
    <div className="space-y-6">
      {tareas.error && <Aviso>{tareas.error}</Aviso>}
      {tareas.datos && <ListaTareas pacienteId={pacienteId} pacienteNombre={pacienteNombre} tareas={tareas.datos} alCambiar={recargar} />}

      {usuario?.rol === 'odontologo' && (
        nuevo && catalogo.datos ? (
          <Tarjeta className="space-y-4">
            <h2 className="text-seccion font-semibold">Nuevo presupuesto</h2>
            <Constructor pacienteId={pacienteId} procedimientos={catalogo.datos} alTerminar={() => { setNuevo(false); recargar(); }} alCancelar={() => setNuevo(false)} />
          </Tarjeta>
        ) : <Boton onClick={() => setNuevo(true)}>Nuevo presupuesto</Boton>
      )}

      {presupuestos.error && <Aviso>{presupuestos.error}</Aviso>}
      {presupuestos.datos?.length === 0 && <p className="text-texto-suave">El paciente aún no tiene presupuestos.</p>}
      {catalogo.datos && presupuestos.datos?.map((p) => <TarjetaPresupuesto key={`${p.id}-${p.estado}`} p={p} pacienteId={pacienteId} procedimientos={catalogo.datos ?? []} alCambiar={recargar} />)}
    </div>
  );
};

export default PestanaPresupuestos;
