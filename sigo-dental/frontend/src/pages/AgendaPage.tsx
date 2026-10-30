import { useState } from 'react';
import { useSearchParams } from 'react-router-dom';
import { ChevronLeft, ChevronRight, Plus } from 'lucide-react';
import { Aviso, Boton, Campo, Casilla, Selector, Titulo } from '../components/ui';
import { liberaSillon } from '../config/citas';
import { useAuth } from '../context/useAuth';
import { useCarga } from '../hooks/useCarga';
import { listarProfesionales, listarSillones } from '../services/catalogoService';
import { listarCitas, obtenerHorario, type CitaDetalle } from '../services/citaService';
import { aISOFecha, desdeISOFecha, inicioDelDia, inicioDeSemana, nombreDelDia, sumarDias } from '../utils/fechas';
import CuadriculaAgenda, { type Columna } from './agenda/CuadriculaAgenda';
import DetalleCita from './agenda/DetalleCita';
import FormularioCita from './agenda/FormularioCita';

type Vista = 'dia' | 'semana';

type Formulario =
  | { tipo: 'nueva'; inicio: Date; sillonId?: number; tarea?: { pacienteId: number; nombre: string; procedimientoId: number; itemId: number; minutos: number } }
  | { tipo: 'reprogramar'; cita: CitaDetalle };

const formatoRango = new Intl.DateTimeFormat('es', { day: 'numeric', month: 'long', year: 'numeric' });
const formatoDiaCorto = new Intl.DateTimeFormat('es', { day: 'numeric', month: 'short' });

const AgendaPage = () => {
  const { usuario } = useAuth();
  const rol = usuario?.rol ?? 'auditor';
  const esRecepcion = rol === 'recepcionista';

  const [vista, setVista] = useState<Vista>('dia');
  const [ancla, setAncla] = useState(() => inicioDelDia(new Date()));
  const [consultorioId, setConsultorioId] = useState(0);
  const [sillonId, setSillonId] = useState(0);
  const [profesionalId, setProfesionalId] = useState(0);
  const [verInactivas, setVerInactivas] = useState(false);
  const [seleccion, setSeleccion] = useState<number | null>(null);
  // Al venir de las tareas de un presupuesto, la agenda abre el formulario con el paciente, el procedimiento y la duración cargados.
  const [parametros, setParametros] = useSearchParams();
  const [formulario, setFormulario] = useState<Formulario | null>(() => {
    const pacienteId = Number(parametros.get('agendar'));
    const itemId = Number(parametros.get('item'));
    if (!pacienteId || !itemId) return null;
    const manana = new Date();
    manana.setDate(manana.getDate() + 1);
    return {
      tipo: 'nueva', inicio: new Date(manana.getFullYear(), manana.getMonth(), manana.getDate(), 9, 0),
      tarea: { pacienteId, nombre: parametros.get('nombre') ?? '', procedimientoId: Number(parametros.get('procedimiento')) || 0, itemId, minutos: Number(parametros.get('min')) || 30 },
    };
  });
  const [aviso, setAviso] = useState('');

  const sillones = useCarga(() => listarSillones(), 'agenda-sillones');
  const profesionales = useCarga(() => listarProfesionales(), 'agenda-profesionales');
  const horario = useCarga(obtenerHorario, 'agenda-horario');

  const desde = vista === 'dia' ? ancla : inicioDeSemana(ancla);
  const hasta = sumarDias(desde, vista === 'dia' ? 1 : 7);
  const citas = useCarga(
    () => listarCitas(desde, hasta, { consultorioId, sillonId, profesionalId }),
    `agenda|${desde.toISOString()}|${vista}|${consultorioId}|${sillonId}|${profesionalId}`,
  );

  const listaSillones = sillones.datos ?? [];
  const consultorios = [...new Map(listaSillones.map((s) => [s.consultorio_id, s.consultorio_nombre])).entries()];
  const sillonesVisibles = listaSillones.filter(
    (s) => (!consultorioId || s.consultorio_id === consultorioId) && (!sillonId || s.id === sillonId),
  );
  const sillonesDelSelector = listaSillones.filter((s) => !consultorioId || s.consultorio_id === consultorioId);

  const columnas: Columna[] =
    vista === 'dia'
      ? sillonesVisibles.map((s) => ({
          clave: `s${s.id}`,
          titulo: s.nombre,
          subtitulo: s.consultorio_nombre,
          dia: ancla,
          sillonId: s.id,
        }))
      : Array.from({ length: 7 }, (_, i) => {
          const dia = sumarDias(desde, i);
          return { clave: aISOFecha(dia), titulo: nombreDelDia(dia), subtitulo: formatoDiaCorto.format(dia), dia };
        });

  const mover = (pasos: number) => setAncla((a) => sumarDias(a, pasos * (vista === 'dia' ? 1 : 7)));
  const error = sillones.error || profesionales.error || horario.error || citas.error;
  const visibles = (citas.datos ?? []).filter((c) => verInactivas || !liberaSillon(c.estado));
  const titulo =
    vista === 'dia' ? formatoRango.format(ancla) : `${formatoDiaCorto.format(desde)} al ${formatoRango.format(sumarDias(hasta, -1))}`;

  const cerrarFormulario = () => {
    setFormulario(null);
    if (parametros.has('agendar')) setParametros({}, { replace: true });
  };
  const alGuardar = (cita: CitaDetalle) => {
    setFormulario(null);
    if (parametros.has('agendar')) setParametros({}, { replace: true });
    setSeleccion(cita.id);
    setAviso(cita.aviso ?? '');
    citas.recargar();
  };

  const abrirNueva = (inicio?: Date, sillonDelHueco?: number) => {
    const base = inicio ?? new Date(ancla.getFullYear(), ancla.getMonth(), ancla.getDate(), 9, 0);
    setSeleccion(null);
    setAviso('');
    setFormulario({ tipo: 'nueva', inicio: base, sillonId: sillonDelHueco ?? (sillonId || undefined) });
  };

  return (
    <section aria-labelledby="titulo-agenda" className="space-y-6">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <Titulo id="titulo-agenda">Agenda</Titulo>
        {esRecepcion && (
          <Boton onClick={() => abrirNueva()}>
            <Plus className="size-5" aria-hidden="true" /> Nueva cita
          </Boton>
        )}
      </div>

      <div className="flex flex-wrap items-end gap-3">
        <div className="flex gap-2" role="group" aria-label="Periodo">
          <Boton variante="secundario" onClick={() => mover(-1)} aria-label={vista === 'dia' ? 'Día anterior' : 'Semana anterior'}>
            <ChevronLeft className="size-5" aria-hidden="true" />
          </Boton>
          <Boton variante="secundario" onClick={() => setAncla(inicioDelDia(new Date()))}>Hoy</Boton>
          <Boton variante="secundario" onClick={() => mover(1)} aria-label={vista === 'dia' ? 'Día siguiente' : 'Semana siguiente'}>
            <ChevronRight className="size-5" aria-hidden="true" />
          </Boton>
        </div>
        <div className="w-44">
          <Campo id="agenda-fecha" etiqueta="Ir a la fecha" type="date" value={aISOFecha(ancla)} onChange={(e) => e.target.value && setAncla(desdeISOFecha(e.target.value))} />
        </div>
        <div className="flex gap-2" role="group" aria-label="Vista">
          <Boton variante={vista === 'dia' ? 'primario' : 'secundario'} aria-pressed={vista === 'dia'} onClick={() => setVista('dia')}>Día</Boton>
          <Boton variante={vista === 'semana' ? 'primario' : 'secundario'} aria-pressed={vista === 'semana'} onClick={() => setVista('semana')}>Semana</Boton>
        </div>
      </div>

      <div className="grid gap-3 md:grid-cols-3">
        <Selector id="filtro-consultorio" etiqueta="Consultorio" value={consultorioId} onChange={(e) => { setConsultorioId(Number(e.target.value)); setSillonId(0); }}>
          <option value={0}>Todos</option>
          {consultorios.map(([id, nombre]) => <option key={id} value={id}>{nombre}</option>)}
        </Selector>
        <Selector id="filtro-sillon" etiqueta="Sillón" value={sillonId} onChange={(e) => setSillonId(Number(e.target.value))}>
          <option value={0}>Todos</option>
          {sillonesDelSelector.map((s) => <option key={s.id} value={s.id}>{s.consultorio_nombre} · {s.nombre}</option>)}
        </Selector>
        <Selector id="filtro-profesional" etiqueta="Profesional" value={profesionalId} onChange={(e) => setProfesionalId(Number(e.target.value))}>
          <option value={0}>Todos</option>
          {(profesionales.datos ?? []).map((p) => <option key={p.id} value={p.id}>{p.nombre}</option>)}
        </Selector>
      </div>

      <div className="flex flex-wrap items-center justify-between gap-3">
        <h2 className="text-seccion font-semibold first-letter:uppercase" aria-live="polite">{titulo}</h2>
        <Casilla id="ver-canceladas" etiqueta="Mostrar canceladas y no asistidas" checked={verInactivas} onChange={(e) => setVerInactivas(e.target.checked)} />
      </div>

      {error && <Aviso>{error}</Aviso>}
      {aviso && <Aviso tipo="info">{aviso}</Aviso>}

      {formulario && horario.datos && sillones.datos && profesionales.datos && (
        <FormularioCita
          key={formulario.tipo === 'reprogramar' ? `r${formulario.cita.id}` : `n${formulario.inicio.getTime()}`}
          cita={formulario.tipo === 'reprogramar' ? formulario.cita : undefined}
          inicioSugerido={formulario.tipo === 'nueva' ? formulario.inicio : undefined}
          sillonSugerido={formulario.tipo === 'nueva' ? formulario.sillonId : undefined}
          pacienteSugerido={formulario.tipo === 'nueva' && formulario.tarea ? { id: formulario.tarea.pacienteId, nombre: formulario.tarea.nombre } : undefined}
          procedimientoSugerido={formulario.tipo === 'nueva' ? formulario.tarea?.procedimientoId : undefined}
          duracionSugeridaMin={formulario.tipo === 'nueva' ? formulario.tarea?.minutos : undefined}
          presupuestoItemId={formulario.tipo === 'nueva' ? formulario.tarea?.itemId : undefined}
          sillones={listaSillones}
          profesionales={profesionales.datos ?? []}
          alGuardar={alGuardar}
          alCancelar={cerrarFormulario}
        />
      )}

      {seleccion && !formulario && usuario && (
        <DetalleCita
          key={seleccion}
          id={seleccion}
          rol={usuario.rol}
          alReprogramar={(cita) => setFormulario({ tipo: 'reprogramar', cita })}
          alCambiar={citas.recargar}
          alCerrar={() => setSeleccion(null)}
        />
      )}

      {horario.datos && !error ? (
        columnas.length === 0 ? (
          <Aviso tipo="info">No hay sillones activos con ese filtro.</Aviso>
        ) : (
          <CuadriculaAgenda
            columnas={columnas}
            citas={visibles}
            horario={horario.datos}
            puedeAgendar={esRecepcion}
            seleccionada={seleccion ?? undefined}
            alElegirHueco={abrirNueva}
            alElegirCita={(cita) => { setFormulario(null); setAviso(''); setSeleccion(cita.id); }}
          />
        )
      ) : (
        !error && <p>Cargando…</p>
      )}
      <p className="text-etiqueta text-texto-suave">
        Las horas se muestran en la hora local de tu equipo. {esRecepcion && 'Pulsa un hueco del calendario para agendar.'}
      </p>
    </section>
  );
};

export default AgendaPage;
