import { useState } from 'react';
import { Link } from 'react-router-dom';
import AlertasMedicas from '../../components/AlertasMedicas';
import ChipEstado from '../../components/ChipEstado';
import { Aviso, Boton, Campo, Tarjeta } from '../../components/ui';
import { ETIQUETA_ACCION, ETIQUETA_ESTADO, estadosSiguientes, sePuedeReprogramar } from '../../config/citas';
import type { Rol } from '../../config/permisos';
import { useAccion } from '../../hooks/useAccion';
import { useCarga } from '../../hooks/useCarga';
import { cambiarEstadoCita, listarNotificaciones, obtenerCita, type CitaDetalle, type EstadoCita, type Notificacion } from '../../services/citaService';
import { horaLocal } from '../../utils/fechas';

interface DetalleCitaProps {
  id: number;
  rol: Rol;
  alReprogramar: (cita: CitaDetalle) => void;
  alCambiar: () => void;
  alCerrar: () => void;
}

const formatoFecha = new Intl.DateTimeFormat('es', { weekday: 'long', day: 'numeric', month: 'long' });
const formatoMarca = new Intl.DateTimeFormat('es', { day: 'numeric', month: 'short', hour: '2-digit', minute: '2-digit' });

const ETIQUETA_TIPO: Record<Notificacion['tipo'], string> = {
  confirmacion: 'Confirmación',
  recordatorio_48h: 'Recordatorio de 48 h',
  recordatorio_24h: 'Recordatorio de 24 h',
};
const ETIQUETA_CANAL_NOTIF: Record<Notificacion['canal'], string> = { whatsapp: 'WhatsApp', sms: 'SMS', email: 'Correo' };
const ETIQUETA_ESTADO_NOTIF: Record<Notificacion['estado'], string> = { pendiente: 'Pendiente', enviada: 'Enviada', fallida: 'Fallida' };

// Bitácora de mensajes al paciente por esta cita; la API solo la entrega a la recepción.
const BitacoraNotificaciones = ({ citaId, version }: { citaId: number; version: string }) => {
  const { datos, error } = useCarga(() => listarNotificaciones(citaId), `notificaciones-${citaId}-${version}`);
  return (
    <>
      <h3 className="mt-6 text-cuerpo font-semibold">Notificaciones al paciente</h3>
      {error && <Aviso>{error}</Aviso>}
      {datos && datos.length === 0 && <p className="mt-2 text-etiqueta text-texto-suave">Aún no se envió ningún mensaje.</p>}
      <ul className="mt-2 space-y-2 text-etiqueta">
        {(datos ?? []).map((n) => (
          <li key={n.id} className={`rounded-control border border-vidrio-borde p-3 ${n.vigente ? '' : 'opacity-60'}`}>
            <p>
              <span className="font-medium">{ETIQUETA_TIPO[n.tipo]}</span> · {ETIQUETA_CANAL_NOTIF[n.canal]} a <span className="font-mono">{n.destino || 'sin destino'}</span> ·{' '}
              <span className={n.estado === 'fallida' ? 'text-alerta-valor' : ''}>{ETIQUETA_ESTADO_NOTIF[n.estado]}</span>
              {!n.vigente && ' · reemplazada por una reprogramación'}
            </p>
            {n.detalle && <p className="text-alerta-valor">{n.detalle}</p>}
            <p className="mt-1 break-words text-texto-suave">{n.mensaje}</p>
          </li>
        ))}
      </ul>
    </>
  );
};

const DetalleCita = ({ id, rol, alReprogramar, alCambiar, alCerrar }: DetalleCitaProps) => {
  const { datos: cita, error, recargar } = useCarga(() => obtenerCita(id), `cita-${id}`);
  const accion = useAccion();
  const [cancelando, setCancelando] = useState(false);
  const [motivo, setMotivo] = useState('');

  if (error) return <Aviso>{error}</Aviso>;
  if (!cita) return <p>Cargando…</p>;

  const cambiar = async (estado: EstadoCita, motivoCambio = '') => {
    if (await accion.ejecutar(() => cambiarEstadoCita(cita.id, estado, motivoCambio))) {
      setCancelando(false);
      setMotivo('');
      recargar();
      alCambiar();
    }
  };

  const inicio = new Date(cita.inicio);
  const fin = new Date(cita.fin);
  const siguientes = estadosSiguientes(cita.estado, rol);

  return (
    <Tarjeta denso>
      <div className="flex flex-wrap items-start justify-between gap-3">
        <div>
          <h2 className="text-seccion font-semibold">{cita.paciente_nombre}</h2>
          <p className="font-mono text-etiqueta text-texto-suave">{cita.numero_historia}</p>
        </div>
        <div className="flex items-center gap-3">
          <ChipEstado estado={cita.estado} />
          <Boton variante="secundario" onClick={alCerrar} aria-label="Cerrar detalle de la cita">Cerrar</Boton>
        </div>
      </div>

      {cita.alertas.length > 0 && <div className="mt-4"><AlertasMedicas alertas={cita.alertas} /></div>}

      <dl className="mt-5 grid gap-4 md:grid-cols-2 xl:grid-cols-4">
        <div>
          <dt className="text-etiqueta text-texto-suave">Horario</dt>
          <dd className="first-letter:uppercase">{formatoFecha.format(inicio)}</dd>
          <dd className="font-mono">{horaLocal(inicio)} a {horaLocal(fin)}</dd>
        </div>
        <div>
          <dt className="text-etiqueta text-texto-suave">Profesional</dt>
          <dd>{cita.profesional_nombre}</dd>
        </div>
        <div>
          <dt className="text-etiqueta text-texto-suave">Sillón</dt>
          <dd>{cita.consultorio_nombre} · {cita.sillon_nombre}</dd>
        </div>
        <div>
          <dt className="text-etiqueta text-texto-suave">Procedimiento o motivo</dt>
          <dd>{cita.procedimiento_nombre || cita.motivo || '—'}</dd>
        </div>
      </dl>
      {cita.notas && <p className="mt-4 text-texto-suave">Notas: {cita.notas}</p>}

      {accion.error && <div className="mt-4"><Aviso>{accion.error}</Aviso></div>}

      {cancelando ? (
        <form
          className="mt-5 flex flex-wrap items-end gap-3"
          onSubmit={(e) => {
            e.preventDefault();
            void cambiar('cancelado', motivo);
          }}
        >
          <div className="min-w-64 flex-1">
            <Campo id="motivo-cancelacion" etiqueta="Motivo de la cancelación" required maxLength={300} value={motivo} onChange={(e) => setMotivo(e.target.value)} />
          </div>
          <Boton type="submit" variante="peligro" disabled={accion.ocupado || !motivo.trim()}>Confirmar cancelación</Boton>
          <Boton variante="secundario" onClick={() => setCancelando(false)}>Volver</Boton>
        </form>
      ) : (
        <div className="mt-5 flex flex-wrap gap-3">
          {siguientes.map((estado) =>
            estado === 'cancelado' ? (
              <Boton key={estado} variante="peligro" onClick={() => setCancelando(true)}>{ETIQUETA_ACCION[estado]}</Boton>
            ) : (
              <Boton key={estado} variante={estado === 'no_asistio' ? 'secundario' : 'primario'} disabled={accion.ocupado} onClick={() => void cambiar(estado)}>
                {ETIQUETA_ACCION[estado]}
              </Boton>
            ),
          )}
          {rol === 'recepcionista' && sePuedeReprogramar(cita.estado) && (
            <Boton variante="secundario" onClick={() => alReprogramar(cita)}>Reprogramar</Boton>
          )}
          <Link className="inline-flex min-h-11 items-center text-chip-acento-texto underline" to={`/pacientes/${cita.paciente_id}`}>
            Abrir expediente
          </Link>
        </div>
      )}

      {rol === 'recepcionista' && <BitacoraNotificaciones citaId={cita.id} version={`${cita.estado}|${cita.inicio}`} />}

      <h3 className="mt-6 text-cuerpo font-semibold">Historial</h3>
      <ol className="mt-2 space-y-1 text-etiqueta text-texto-suave">
        {cita.historial.map((h, i) => (
          <li key={i}>
            <span className="font-mono">{formatoMarca.format(new Date(h.fecha))}</span> · {ETIQUETA_ESTADO[h.estado_nuevo]} · {h.usuario_nombre}
            {h.motivo && ` · ${h.motivo}`}
          </li>
        ))}
      </ol>
    </Tarjeta>
  );
};

export default DetalleCita;
