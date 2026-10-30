import { useState, type FormEvent } from 'react';
import FormularioFirma from '../../components/FormularioFirma';
import { Aviso, Boton, Campo, Chip, CampoTexto, Selector, Tarjeta } from '../../components/ui';
import { etiquetaSuperficie, type Superficie } from '../../components/odontograma/dientes';
import { useAuth } from '../../context/useAuth';
import { useAccion } from '../../hooks/useAccion';
import { useCarga } from '../../hooks/useCarga';
import { listarProcedimientos, listarProfesionales } from '../../services/catalogoService';
import {
  actualizarNota, cancelarProcedimiento, completarProcedimiento, crearAdenda, crearNota, ETIQUETA_ESTADO_NOTA, ETIQUETA_ESTADO_PROC,
  firmarNota, iniciarProcedimiento, listarNotas, listarProcedimientosRealizados,
  type Adenda, type DatosNota, type Firma, type Nota, type ProcedimientoRealizado,
} from '../../services/evolucionService';
import { fechaHoraLocal } from '../../utils/fechas';

const VACIA: DatosNota = { subjetivo: '', objetivo: '', evaluacion: '', plan: '' };
const SOEP = [
  ['subjetivo', 'S · Subjetivo', 'Lo que refiere el paciente'],
  ['objetivo', 'O · Objetivo', 'Hallazgos del examen'],
  ['evaluacion', 'E · Evaluación', 'Diagnóstico o impresión'],
  ['plan', 'P · Plan', 'Tratamiento y siguiente paso'],
] as const;

const ChipFirma = ({ firma }: { firma: Firma }) => (
  <div className="space-y-1 text-etiqueta">
    <p>
      <Chip tono={firma.valida ? 'acento' : 'alerta'}>{firma.valida ? 'Firma válida' : 'CONTENIDO ALTERADO: la huella no coincide'}</Chip>{' '}
      Firmada por <strong>{firma.nombre_firmante}</strong> (reg. {firma.registro_profesional}) · {fechaHoraLocal(firma.firmado_en)}
    </p>
    <p className="font-mono text-texto-suave">SHA-256 {firma.huella_sha256.slice(0, 16)}…</p>
  </div>
);

const ListaAdendas = ({ adendas }: { adendas: Adenda[] }) => (
  adendas.length === 0 ? null : (
    <ul className="space-y-2 border-l-2 border-chip-acento-borde pl-4">
      {adendas.map((a) => (
        <li key={a.id}>
          <p className="text-etiqueta font-medium">Adenda · {fechaHoraLocal(a.creada_en)}</p>
          <p className="whitespace-pre-wrap">{a.texto}</p>
          {a.firma && <ChipFirma firma={a.firma} />}
        </li>
      ))}
    </ul>
  )
);

// Formulario para escribir o editar una nota SOEP.
const FormularioNota = ({ inicial, profesionales, alGuardar, alCancelar, etiquetaGuardar }: {
  inicial: DatosNota; profesionales: { id: number; nombre: string }[] | null; alGuardar: (d: DatosNota) => Promise<boolean>; alCancelar?: () => void; etiquetaGuardar: string;
}) => {
  const [datos, setDatos] = useState(inicial);
  const [ocupado, setOcupado] = useState(false);
  const enviar = async (e: FormEvent) => {
    e.preventDefault();
    setOcupado(true);
    if (await alGuardar(datos)) setDatos(VACIA);
    setOcupado(false);
  };
  return (
    <form onSubmit={enviar} className="space-y-4">
      <div className="grid gap-4 md:grid-cols-2">
        {SOEP.map(([campo, etiqueta, ayuda]) => (
          <CampoTexto key={campo} id={`soep-${campo}`} etiqueta={etiqueta} ayuda={ayuda} required maxLength={4000} value={datos[campo]}
            onChange={(e) => setDatos({ ...datos, [campo]: e.target.value })} />
        ))}
      </div>
      {profesionales && (
        <Selector id="soep-profesional" etiqueta="Profesional responsable" ayuda="Es quien firmará la nota" required value={datos.profesional_id ?? ''}
          onChange={(e) => setDatos({ ...datos, profesional_id: Number(e.target.value) })}>
          <option value="">Elige…</option>
          {profesionales.map((p) => <option key={p.id} value={p.id}>{p.nombre}</option>)}
        </Selector>
      )}
      <div className="flex flex-wrap gap-3">
        <Boton type="submit" disabled={ocupado}>{ocupado ? 'Guardando…' : etiquetaGuardar}</Boton>
        {alCancelar && <Boton variante="secundario" onClick={alCancelar}>Cancelar</Boton>}
      </div>
    </form>
  );
};

type Panel = { tipo: 'editar' | 'firmar' | 'adenda'; id: number } | null;

const FormularioAdenda = ({ alEnviar, alCancelar }: { alEnviar: (texto: string, password: string) => Promise<unknown>; alCancelar: () => void }) => {
  const [texto, setTexto] = useState('');
  const accion = useAccion();
  const [password, setPassword] = useState('');
  const enviar = async (e: FormEvent) => {
    e.preventDefault();
    await accion.ejecutar(() => alEnviar(texto, password));
  };
  return (
    <form onSubmit={enviar} className="space-y-3 rounded-control border border-vidrio-borde p-4">
      <CampoTexto id="adenda-texto" etiqueta="Texto de la adenda" ayuda="Explica la corrección. La adenda queda enlazada al original y firmada por ti." required maxLength={4000} value={texto} onChange={(e) => setTexto(e.target.value)} />
      <Campo id="adenda-password" etiqueta="Tu contraseña (firma)" type="password" autoComplete="current-password" required value={password} onChange={(e) => setPassword(e.target.value)} />
      {accion.error && <Aviso>{accion.error}</Aviso>}
      <div className="flex flex-wrap gap-3">
        <Boton type="submit" disabled={accion.ocupado}>{accion.ocupado ? 'Firmando…' : 'Firmar adenda'}</Boton>
        <Boton variante="secundario" onClick={alCancelar}>Cancelar</Boton>
      </div>
    </form>
  );
};

const TarjetaNota = ({ nota, pacienteId, alCambiar }: { nota: Nota; pacienteId: number; alCambiar: () => void }) => {
  const { usuario } = useAuth();
  const [panel, setPanel] = useState<Panel>(null);
  const accion = useAccion();
  const cerrar = () => { setPanel(null); alCambiar(); };
  const puedeAdenda = usuario?.rol === 'odontologo' && nota.estado !== 'editable';

  return (
    <Tarjeta className="space-y-4">
      <div className="flex flex-wrap items-center gap-3">
        <h3 className="font-semibold">Nota del {fechaHoraLocal(nota.creada_en)}</h3>
        <Chip tono={nota.estado === 'firmada' ? 'acento' : nota.estado === 'pendiente_firma' ? 'alerta' : 'neutro'}>{ETIQUETA_ESTADO_NOTA[nota.estado]}</Chip>
        <span className="text-etiqueta text-texto-suave">Redactó {nota.autor_nombre} · responsable {nota.profesional_nombre}</span>
      </div>

      {panel?.tipo === 'editar' ? (
        <>
          {accion.error && <Aviso>{accion.error}</Aviso>}
          <FormularioNota inicial={nota} profesionales={null} etiquetaGuardar="Guardar cambios" alCancelar={() => setPanel(null)}
            alGuardar={async (d) => { const bien = await accion.ejecutar(() => actualizarNota(pacienteId, nota.id, d)); if (bien) cerrar(); return false; }} />
        </>
      ) : (
        <dl className="grid gap-4 md:grid-cols-2">
          {SOEP.map(([campo, etiqueta]) => (
            <div key={campo}><dt className="text-etiqueta text-texto-suave">{etiqueta}</dt><dd className="whitespace-pre-wrap">{nota[campo]}</dd></div>
          ))}
        </dl>
      )}

      {nota.estado === 'editable' && <p className="text-etiqueta text-texto-suave">Se puede editar hasta el {fechaHoraLocal(nota.editable_hasta)}; después se bloquea.</p>}
      {nota.estado === 'pendiente_firma' && <p className="text-etiqueta text-texto-suave">Venció el plazo de edición sin firma: ya no se edita. El profesional responsable todavía puede firmarla.</p>}
      {nota.firma && <ChipFirma firma={nota.firma} />}
      <ListaAdendas adendas={nota.adendas} />

      {panel?.tipo === 'firmar' && (
        <FormularioFirma descripcion="esta nota" alCancelar={() => setPanel(null)} alFirmar={async (p) => { await firmarNota(pacienteId, nota.id, p); cerrar(); }} />
      )}
      {panel?.tipo === 'adenda' && (
        <FormularioAdenda alCancelar={() => setPanel(null)} alEnviar={async (texto, password) => { await crearAdenda(pacienteId, { nota_id: nota.id, texto, password }); cerrar(); }} />
      )}
      {!panel && (nota.puede_editar || nota.puede_firmar || puedeAdenda) && (
        <div className="flex flex-wrap gap-3">
          {nota.puede_editar && <Boton variante="secundario" onClick={() => setPanel({ tipo: 'editar', id: nota.id })}>Editar</Boton>}
          {nota.puede_firmar && <Boton onClick={() => setPanel({ tipo: 'firmar', id: nota.id })}>Firmar</Boton>}
          {puedeAdenda && <Boton variante="secundario" onClick={() => setPanel({ tipo: 'adenda', id: nota.id })}>Agregar adenda</Boton>}
        </div>
      )}
    </Tarjeta>
  );
};

// --- Procedimientos realizados ---

const FormularioProcedimiento = ({ pacienteId, alIniciar }: { pacienteId: number; alIniciar: () => void }) => {
  const catalogo = useCarga(() => listarProcedimientos(), 'procedimientos-catalogo');
  const accion = useAccion();
  const [procedimientoId, setProcedimientoId] = useState('');
  const [pieza, setPieza] = useState('');
  const [superficie, setSuperficie] = useState('');
  const [notas, setNotas] = useState('');

  const iniciar = async (e: FormEvent) => {
    e.preventDefault();
    const bien = await accion.ejecutar(() => iniciarProcedimiento(pacienteId, {
      procedimiento_id: Number(procedimientoId), pieza: pieza ? Number(pieza) : undefined, superficie: superficie || undefined, notas,
    }));
    if (bien) { setProcedimientoId(''); setPieza(''); setSuperficie(''); setNotas(''); alIniciar(); }
  };

  return (
    <form onSubmit={iniciar} className="space-y-4">
      <h3 className="font-semibold">Iniciar un procedimiento</h3>
      <div className="grid gap-4 md:grid-cols-4">
        <Selector id="proc-catalogo" etiqueta="Procedimiento" required value={procedimientoId} onChange={(e) => setProcedimientoId(e.target.value)}>
          <option value="">Elige…</option>
          {catalogo.datos?.filter((p) => p.activo).map((p) => <option key={p.id} value={p.id}>{p.nombre}{p.requiere_consentimiento ? ' (requiere consentimiento)' : ''}</option>)}
        </Selector>
        <Campo id="proc-pieza" etiqueta="Pieza FDI (opcional)" inputMode="numeric" maxLength={2} value={pieza} onChange={(e) => setPieza(e.target.value.replace(/\D/g, ''))} />
        <Selector id="proc-superficie" etiqueta="Superficie (opcional)" value={superficie} onChange={(e) => setSuperficie(e.target.value)}>
          <option value="">—</option>
          {(['mesial', 'distal', 'oclusal', 'vestibular', 'lingual'] as Superficie[]).map((s) => <option key={s} value={s}>{etiquetaSuperficie(s, Number(pieza) || 16)}</option>)}
        </Selector>
        <Campo id="proc-notas" etiqueta="Notas (opcional)" maxLength={300} value={notas} onChange={(e) => setNotas(e.target.value)} />
      </div>
      {accion.error && <Aviso>{accion.error}</Aviso>}
      <Boton type="submit" disabled={accion.ocupado || !procedimientoId}>{accion.ocupado ? 'Iniciando…' : 'Iniciar procedimiento'}</Boton>
    </form>
  );
};

const FilaProcedimiento = ({ proc, pacienteId, alCambiar, puedeActuar }: { proc: ProcedimientoRealizado; pacienteId: number; alCambiar: () => void; puedeActuar: boolean }) => {
  const accion = useAccion();
  const [adenda, setAdenda] = useState(false);
  const cambiar = (f: () => Promise<unknown>) => accion.ejecutar(async () => { await f(); alCambiar(); });
  return (
    <li className="space-y-3 rounded-control border border-vidrio-borde p-4">
      <div className="flex flex-wrap items-center gap-3">
        <span className="font-medium">{proc.procedimiento_nombre}</span>
        {proc.pieza && <Chip>Pieza {proc.pieza}{proc.superficie ? ` · ${etiquetaSuperficie(proc.superficie as Superficie, proc.pieza)}` : ''}</Chip>}
        <Chip tono={proc.estado === 'realizado' ? 'acento' : proc.estado === 'cancelado' ? 'alerta' : 'neutro'}>{ETIQUETA_ESTADO_PROC[proc.estado]}</Chip>
        <span className="text-etiqueta text-texto-suave">{proc.profesional_nombre} · {fechaHoraLocal(proc.realizado_en ?? proc.iniciado_en)}</span>
      </div>
      {proc.notas && <p className="text-etiqueta">{proc.notas}</p>}
      {proc.estado === 'realizado' && <p className="text-etiqueta text-texto-suave">Realizado: no se borra ni se cambia (RN-01). Una corrección se hace con una adenda.</p>}
      <ListaAdendas adendas={proc.adendas} />
      {accion.error && <Aviso>{accion.error}</Aviso>}
      {puedeActuar && proc.estado === 'en_curso' && (
        <div className="flex flex-wrap gap-3">
          <Boton disabled={accion.ocupado} onClick={() => cambiar(() => completarProcedimiento(pacienteId, proc.id))}>Marcar como realizado</Boton>
          <Boton variante="secundario" disabled={accion.ocupado} onClick={() => cambiar(() => cancelarProcedimiento(pacienteId, proc.id))}>Cancelar</Boton>
        </div>
      )}
      {puedeActuar && proc.estado === 'realizado' && !adenda && <Boton variante="secundario" onClick={() => setAdenda(true)}>Agregar adenda</Boton>}
      {adenda && (
        <FormularioAdenda alCancelar={() => setAdenda(false)}
          alEnviar={async (texto, password) => { await crearAdenda(pacienteId, { procedimiento_realizado_id: proc.id, texto, password }); setAdenda(false); alCambiar(); }} />
      )}
    </li>
  );
};

// Evolución del paciente (RF-05): bitácora SOEP firmada e inmutable, y procedimientos realizados (RN-01).
const PestanaEvolucion = ({ pacienteId }: { pacienteId: number }) => {
  const { usuario } = useAuth();
  const esOdontologo = usuario?.rol === 'odontologo';
  const notas = useCarga(() => listarNotas(pacienteId), `notas-${pacienteId}`);
  const procs = useCarga(() => listarProcedimientosRealizados(pacienteId), `proc-realizados-${pacienteId}`);
  const profesionales = useCarga(() => (esOdontologo ? Promise.resolve([]) : listarProfesionales()), `profesionales-nota-${esOdontologo}`);
  const accion = useAccion();

  return (
    <div className="space-y-6">
      <Tarjeta className="space-y-4">
        <h2 className="text-seccion font-semibold">Nueva nota de evolución (SOEP)</h2>
        <p className="text-etiqueta text-texto-suave">
          La nota se puede editar durante el plazo configurado (24 horas por defecto) y se bloquea al firmarse o al vencer.
          {!esOdontologo && ' El asistente redacta; la firma la hace el profesional responsable.'}
        </p>
        {accion.error && <Aviso>{accion.error}</Aviso>}
        <FormularioNota inicial={VACIA} etiquetaGuardar="Guardar nota"
          profesionales={esOdontologo ? null : (profesionales.datos ?? []).filter((p) => p.activo)}
          alGuardar={async (d) => { const bien = await accion.ejecutar(() => crearNota(pacienteId, d)); if (bien) notas.recargar(); return bien; }} />
      </Tarjeta>

      <section className="space-y-4" aria-labelledby="bitacora-titulo">
        <h2 id="bitacora-titulo" className="text-seccion font-semibold">Bitácora de evolución</h2>
        {notas.error && <Aviso>{notas.error}</Aviso>}
        {notas.datos?.length === 0 && <p className="text-texto-suave">Todavía no hay notas.</p>}
        {notas.datos?.map((n) => <TarjetaNota key={`${n.id}-${n.estado}-${n.adendas.length}`} nota={n} pacienteId={pacienteId} alCambiar={notas.recargar} />)}
      </section>

      <Tarjeta className="space-y-5">
        <h2 className="text-seccion font-semibold">Procedimientos realizados</h2>
        {esOdontologo && <FormularioProcedimiento pacienteId={pacienteId} alIniciar={procs.recargar} />}
        {procs.error && <Aviso>{procs.error}</Aviso>}
        {procs.datos?.length === 0 && <p className="text-texto-suave">Todavía no hay procedimientos registrados.</p>}
        <ul className="space-y-3">
          {procs.datos?.map((p) => <FilaProcedimiento key={`${p.id}-${p.estado}-${p.adendas.length}`} proc={p} pacienteId={pacienteId} alCambiar={procs.recargar} puedeActuar={esOdontologo} />)}
        </ul>
      </Tarjeta>
    </div>
  );
};

export default PestanaEvolucion;
