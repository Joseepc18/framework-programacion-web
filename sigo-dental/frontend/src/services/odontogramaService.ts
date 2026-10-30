// Acceso a la API del odontograma (F2-2). Las piezas viajan siempre en numeración FDI.
import type { EstadoHallazgo, Superficie, TipoHallazgo } from '../components/odontograma/dientes';
import { apiFetch } from './http';

export interface Hallazgo {
  id: number;
  paciente_id: number;
  pieza: number;
  superficie: Superficie | '';
  tipo: TipoHallazgo;
  estado: EstadoHallazgo;
  observacion: string;
  cita_id: number | null;
  registrado_por: string;
  registrado_en: string;
  retirado_por: string;
  retirado_en: string | null;
  reemplaza_a: number | null;
}

export interface DatosHallazgo {
  pieza: number;
  superficie: Superficie | '';
  tipo: TipoHallazgo;
  estado: EstadoHallazgo;
  observacion: string;
}

const base = (pacienteId: number) => `/api/pacientes/${pacienteId}/odontograma`;

// Con `al` (instante ISO) devuelve el odontograma como estaba entonces.
export const obtenerOdontograma = (pacienteId: number, al?: string) =>
  apiFetch<Hallazgo[]>(`${base(pacienteId)}${al ? `?al=${encodeURIComponent(al)}` : ''}`);

export const obtenerHistorial = (pacienteId: number) => apiFetch<Hallazgo[]>(`${base(pacienteId)}/historial`);

export const registrarHallazgo = (pacienteId: number, datos: DatosHallazgo) =>
  apiFetch<Hallazgo>(`${base(pacienteId)}/hallazgos`, { method: 'POST', body: JSON.stringify(datos) });

export const corregirHallazgo = (pacienteId: number, id: number, datos: DatosHallazgo) =>
  apiFetch<Hallazgo>(`${base(pacienteId)}/hallazgos/${id}`, { method: 'PUT', body: JSON.stringify(datos) });

export const retirarHallazgo = (pacienteId: number, id: number) =>
  apiFetch<void>(`${base(pacienteId)}/hallazgos/${id}/retirar`, { method: 'POST' });
