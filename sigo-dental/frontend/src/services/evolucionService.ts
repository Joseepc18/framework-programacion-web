// Acceso a la API de la evolución clínica (F2-4): notas SOEP, firma digital, adendas y procedimientos realizados.
import { apiFetch } from './http';

export interface Firma {
  id: number;
  entidad: string;
  entidad_id: number;
  profesional_id: number;
  nombre_firmante: string;
  registro_profesional: string;
  huella_sha256: string;
  firmado_en: string;
  // Falso si el contenido ya no coincide con la huella que se firmó (alteración).
  valida: boolean;
}

export interface Adenda {
  id: number;
  nota_id: number | null;
  procedimiento_realizado_id: number | null;
  profesional_nombre: string;
  texto: string;
  creada_en: string;
  firma: Firma | null;
}

export type EstadoNota = 'editable' | 'firmada' | 'pendiente_firma';

export const ETIQUETA_ESTADO_NOTA: Record<EstadoNota, string> = {
  editable: 'Editable',
  firmada: 'Firmada',
  pendiente_firma: 'Pendiente de firma',
};

export interface Nota {
  id: number;
  paciente_id: number;
  cita_id: number | null;
  profesional_id: number;
  profesional_nombre: string;
  autor_nombre: string;
  subjetivo: string;
  objetivo: string;
  evaluacion: string;
  plan: string;
  creada_en: string;
  estado: EstadoNota;
  editable_hasta: string;
  firma: Firma | null;
  adendas: Adenda[];
  puede_editar: boolean;
  puede_firmar: boolean;
}

export interface DatosNota {
  subjetivo: string;
  objetivo: string;
  evaluacion: string;
  plan: string;
  profesional_id?: number;
}

const base = (pacienteId: number) => `/api/pacientes/${pacienteId}`;

export const listarNotas = (pacienteId: number) => apiFetch<Nota[]>(`${base(pacienteId)}/notas`);
export const crearNota = (pacienteId: number, datos: DatosNota) =>
  apiFetch<Nota>(`${base(pacienteId)}/notas`, { method: 'POST', body: JSON.stringify(datos) });
export const actualizarNota = (pacienteId: number, id: number, datos: DatosNota) =>
  apiFetch<Nota>(`${base(pacienteId)}/notas/${id}`, { method: 'PUT', body: JSON.stringify(datos) });
export const firmarNota = (pacienteId: number, id: number, password: string) =>
  apiFetch<Nota>(`${base(pacienteId)}/notas/${id}/firmar`, { method: 'POST', body: JSON.stringify({ password }) });

export interface DatosAdenda {
  nota_id?: number;
  procedimiento_realizado_id?: number;
  texto: string;
  password: string;
}
export const crearAdenda = (pacienteId: number, datos: DatosAdenda) =>
  apiFetch<Adenda>(`${base(pacienteId)}/adendas`, { method: 'POST', body: JSON.stringify(datos) });

export type EstadoProcedimiento = 'en_curso' | 'realizado' | 'cancelado';

export const ETIQUETA_ESTADO_PROC: Record<EstadoProcedimiento, string> = {
  en_curso: 'En curso',
  realizado: 'Realizado',
  cancelado: 'Cancelado',
};

export interface ProcedimientoRealizado {
  id: number;
  procedimiento_id: number;
  procedimiento_nombre: string;
  pieza: number | null;
  superficie: string;
  profesional_nombre: string;
  cita_id: number | null;
  estado: EstadoProcedimiento;
  notas: string;
  iniciado_en: string;
  realizado_en: string | null;
  adendas: Adenda[];
}

export interface DatosProcedimientoRealizado {
  procedimiento_id: number;
  pieza?: number;
  superficie?: string;
  cita_id?: number;
  presupuesto_item_id?: number;
  notas: string;
}

export const listarProcedimientosRealizados = (pacienteId: number) =>
  apiFetch<ProcedimientoRealizado[]>(`${base(pacienteId)}/procedimientos-realizados`);
export const iniciarProcedimiento = (pacienteId: number, datos: DatosProcedimientoRealizado) =>
  apiFetch<ProcedimientoRealizado>(`${base(pacienteId)}/procedimientos-realizados`, { method: 'POST', body: JSON.stringify(datos) });
export const completarProcedimiento = (pacienteId: number, id: number) =>
  apiFetch<ProcedimientoRealizado>(`${base(pacienteId)}/procedimientos-realizados/${id}/completar`, { method: 'POST' });
export const cancelarProcedimiento = (pacienteId: number, id: number) =>
  apiFetch<ProcedimientoRealizado>(`${base(pacienteId)}/procedimientos-realizados/${id}/cancelar`, { method: 'POST' });
