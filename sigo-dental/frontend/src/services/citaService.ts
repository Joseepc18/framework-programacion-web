// Acceso a la API de la agenda. Las horas viajan como ISO-8601 con zona y se guardan en UTC.
import { apiFetch } from './http';

export type EstadoCita = 'programado' | 'confirmado' | 'en_sala' | 'en_atencion' | 'atendido' | 'cancelado' | 'no_asistio';

export interface Cita {
  id: number;
  estado: EstadoCita;
  paciente_id: number;
  paciente_nombre: string;
  numero_historia: string;
  alertas: string[];
  profesional_id: number;
  profesional_nombre: string;
  sillon_id: number;
  sillon_nombre: string;
  consultorio_id: number;
  consultorio_nombre: string;
  procedimiento_id: number | null;
  procedimiento_nombre: string;
  motivo: string;
  notas: string;
  inicio: string;
  fin: string;
}

export interface CambioEstado {
  estado_anterior: string;
  estado_nuevo: EstadoCita;
  motivo: string;
  usuario_nombre: string;
  fecha: string;
}

export interface CitaDetalle extends Cita {
  historial: CambioEstado[];
  // Solo al agendar o reprogramar: avisa sin bloquear (el profesional ya tiene otra cita a esa hora).
  aviso?: string;
}

export interface DatosCita {
  paciente_id: number;
  profesional_id: number;
  sillon_id: number;
  procedimiento_id: number | null;
  motivo: string;
  notas: string;
  inicio: string;
  fin: string;
}

export interface HorarioAgenda {
  inicio: string;
  fin: string;
  zona: string;
}

export interface FiltroCitas {
  consultorioId: number;
  sillonId: number;
  profesionalId: number;
}

export function listarCitas(desde: Date, hasta: Date, filtro: FiltroCitas) {
  const params = new URLSearchParams({ desde: desde.toISOString(), hasta: hasta.toISOString() });
  if (filtro.consultorioId) params.set('consultorio_id', String(filtro.consultorioId));
  if (filtro.sillonId) params.set('sillon_id', String(filtro.sillonId));
  if (filtro.profesionalId) params.set('profesional_id', String(filtro.profesionalId));
  return apiFetch<Cita[]>(`/api/citas?${params}`);
}

export const obtenerCita = (id: number) => apiFetch<CitaDetalle>(`/api/citas/${id}`);
export const obtenerHorario = () => apiFetch<HorarioAgenda>('/api/agenda/horario');

export const crearCita = (datos: DatosCita) =>
  apiFetch<CitaDetalle>('/api/citas', { method: 'POST', body: JSON.stringify(datos) });

export const reprogramarCita = (id: number, datos: DatosCita) =>
  apiFetch<CitaDetalle>(`/api/citas/${id}`, { method: 'PUT', body: JSON.stringify(datos) });

export const cambiarEstadoCita = (id: number, estado: EstadoCita, motivo = '') =>
  apiFetch<CitaDetalle>(`/api/citas/${id}/estado`, { method: 'POST', body: JSON.stringify({ estado, motivo }) });
