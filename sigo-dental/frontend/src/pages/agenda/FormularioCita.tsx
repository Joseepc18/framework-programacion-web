import { useEffect, useState, type FormEvent } from 'react';
import { Aviso, Boton, Campo, CampoTexto, Selector, Tarjeta } from '../../components/ui';
import { useAccion } from '../../hooks/useAccion';
import { useCarga } from '../../hooks/useCarga';
import { listarProcedimientos, type Procedimiento, type Profesional, type Sillon } from '../../services/catalogoService';
import { crearCita, reprogramarCita, type CitaDetalle } from '../../services/citaService';
import { listarPacientes } from '../../services/pacienteService';
import { aISOFecha, combinarFechaHora, horaLocal } from '../../utils/fechas';

interface FormularioCitaProps {
  // Si hay una cita se reprograma; si no, se agenda una nueva con el hueco elegido.
  cita?: CitaDetalle;
  inicioSugerido?: Date;
  sillonSugerido?: number;
  pacienteSugerido?: { id: number; nombre: string };
  sillones: Sillon[];
  profesionales: Profesional[];
  alGuardar: (cita: CitaDetalle) => void;
  alCancelar: () => void;
}

const FormularioCita = ({
  cita,
  inicioSugerido,
  sillonSugerido,
  pacienteSugerido,
  sillones,
  profesionales,
  alGuardar,
  alCancelar,
}: FormularioCitaProps) => {
  const accion = useAccion();
  const procedimientos = useCarga(() => listarProcedimientos(), 'procedimientos-activos');

  const inicio = cita ? new Date(cita.inicio) : (inicioSugerido ?? new Date());
  const fin = cita ? new Date(cita.fin) : new Date(inicio.getTime() + 30 * 60_000);

  const [texto, setTexto] = useState('');
  const [busqueda, setBusqueda] = useState('');
  const [pacienteId, setPacienteId] = useState(cita?.paciente_id ?? pacienteSugerido?.id ?? 0);
  const [fecha, setFecha] = useState(aISOFecha(inicio));
  const [horaInicio, setHoraInicio] = useState(horaLocal(inicio));
  const [horaFin, setHoraFin] = useState(horaLocal(fin));
  const [profesionalId, setProfesionalId] = useState(cita?.profesional_id ?? profesionales[0]?.id ?? 0);
  const [sillonId, setSillonId] = useState(cita?.sillon_id ?? sillonSugerido ?? sillones[0]?.id ?? 0);
  const [procedimientoId, setProcedimientoId] = useState(cita?.procedimiento_id ?? 0);
  const [motivo, setMotivo] = useState(cita?.motivo ?? '');
  const [notas, setNotas] = useState(cita?.notas ?? '');

  // La búsqueda de pacientes se envía 300 ms después de dejar de escribir; no hace falta al reprogramar.
  useEffect(() => {
    const espera = setTimeout(() => setBusqueda(texto), 300);
    return () => clearTimeout(espera);
  }, [texto]);
  const pacientes = useCarga(() => (cita ? Promise.resolve([]) : listarPacientes(busqueda)), `pacientes-cita|${busqueda}`);

  // Al elegir un procedimiento, el fin se propone según su duración estimada.
  const elegirProcedimiento = (id: number, lista: Procedimiento[]) => {
    setProcedimientoId(id);
    const elegido = lista.find((p) => p.id === id);
    if (elegido) {
      const nuevoFin = new Date(combinarFechaHora(fecha, horaInicio).getTime() + elegido.duracion_min * 60_000);
      setHoraFin(horaLocal(nuevoFin));
      if (!motivo) setMotivo(elegido.nombre);
    }
  };

  const guardar = async (evento: FormEvent) => {
    evento.preventDefault();
    const datos = {
      paciente_id: pacienteId,
      profesional_id: profesionalId,
      sillon_id: sillonId,
      procedimiento_id: procedimientoId || null,
      motivo,
      notas,
      inicio: combinarFechaHora(fecha, horaInicio).toISOString(),
      fin: combinarFechaHora(fecha, horaFin).toISOString(),
    };
    let guardada: CitaDetalle | undefined;
    const bien = await accion.ejecutar(async () => {
      guardada = cita ? await reprogramarCita(cita.id, datos) : await crearCita(datos);
    });
    if (bien && guardada) alGuardar(guardada);
  };

  const listaProcedimientos = procedimientos.datos ?? [];
  const resultados = pacientes.datos ?? [];

  return (
    <Tarjeta denso>
      <form onSubmit={guardar} className="space-y-5">
        <h2 className="text-seccion font-semibold">{cita ? 'Reprogramar cita' : 'Nueva cita'}</h2>

        {cita ? (
          <p>
            Paciente: <span className="font-medium">{cita.paciente_nombre}</span>{' '}
            <span className="font-mono text-texto-suave">{cita.numero_historia}</span>
          </p>
        ) : (
          <div className="grid gap-5 md:grid-cols-2">
            <Campo id="cita-buscar" etiqueta="Buscar paciente" ayuda="Nombre, documento o número de historia" value={texto} onChange={(e) => setTexto(e.target.value)} />
            <Selector id="cita-paciente" etiqueta="Paciente" required value={pacienteId || ''} onChange={(e) => setPacienteId(Number(e.target.value))}>
              <option value="">{pacientes.cargando ? 'Buscando…' : 'Elige un paciente'}</option>
              {pacienteSugerido && !resultados.some((p) => p.id === pacienteSugerido.id) && (
                <option value={pacienteSugerido.id}>{pacienteSugerido.nombre}</option>
              )}
              {resultados.map((p) => (
                <option key={p.id} value={p.id}>
                  {p.apellidos} {p.nombres} · {p.numero_historia}
                  {p.alertas.length > 0 ? ' · ALERTA MÉDICA' : ''}
                </option>
              ))}
            </Selector>
          </div>
        )}

        <div className="grid gap-5 md:grid-cols-3">
          <Campo id="cita-fecha" etiqueta="Fecha" type="date" required value={fecha} onChange={(e) => setFecha(e.target.value)} />
          <Campo id="cita-inicio" etiqueta="Hora de inicio" type="time" step={300} required value={horaInicio} onChange={(e) => setHoraInicio(e.target.value)} />
          <Campo id="cita-fin" etiqueta="Hora de fin" type="time" step={300} required value={horaFin} onChange={(e) => setHoraFin(e.target.value)} />
        </div>

        <div className="grid gap-5 md:grid-cols-3">
          <Selector id="cita-profesional" etiqueta="Profesional" required value={profesionalId} onChange={(e) => setProfesionalId(Number(e.target.value))}>
            {profesionales.map((p) => <option key={p.id} value={p.id}>{p.nombre}</option>)}
          </Selector>
          <Selector id="cita-sillon" etiqueta="Sillón" required value={sillonId} onChange={(e) => setSillonId(Number(e.target.value))}>
            {sillones.map((s) => <option key={s.id} value={s.id}>{s.consultorio_nombre} · {s.nombre}</option>)}
          </Selector>
          <Selector id="cita-procedimiento" etiqueta="Procedimiento (opcional)" value={procedimientoId} onChange={(e) => elegirProcedimiento(Number(e.target.value), listaProcedimientos)}>
            <option value={0}>Sin procedimiento</option>
            {listaProcedimientos.map((p) => <option key={p.id} value={p.id}>{p.nombre} · {p.duracion_min} min</option>)}
          </Selector>
        </div>

        <div className="grid gap-5 md:grid-cols-2">
          <Campo id="cita-motivo" etiqueta="Motivo (opcional)" maxLength={200} value={motivo} onChange={(e) => setMotivo(e.target.value)} />
          <CampoTexto id="cita-notas" etiqueta="Notas (opcional)" maxLength={1000} value={notas} onChange={(e) => setNotas(e.target.value)} />
        </div>

        {accion.error && <Aviso>{accion.error}</Aviso>}
        <div className="flex flex-wrap gap-3">
          <Boton type="submit" disabled={accion.ocupado || !pacienteId}>
            {accion.ocupado ? 'Guardando…' : cita ? 'Guardar cambios' : 'Agendar cita'}
          </Boton>
          <Boton variante="secundario" onClick={alCancelar}>Cancelar</Boton>
        </div>
      </form>
    </Tarjeta>
  );
};

export default FormularioCita;
