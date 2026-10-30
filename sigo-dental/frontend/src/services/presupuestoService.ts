// Acceso a la API de presupuestos y tareas agendables (F2-7). El dinero viaja en centavos y los totales los calcula el servidor.
import { apiBlob, apiFetch } from './http';

export type EstadoPresupuesto = 'borrador' | 'presentado' | 'aprobado_total' | 'aprobado_parcial' | 'rechazado';

export const ETIQUETA_ESTADO_PRESUPUESTO: Record<EstadoPresupuesto, string> = {
  borrador: 'Borrador',
  presentado: 'Presentado',
  aprobado_total: 'Aprobado total',
  aprobado_parcial: 'Aprobado parcial',
  rechazado: 'Rechazado',
};

export type EstadoItem = 'pendiente' | 'aprobado' | 'rechazado' | 'agendado' | 'cumplido';

export interface ItemPresupuesto {
  id: number;
  fase: number;
  sesion: number;
  procedimiento_id: number;
  descripcion: string;
  pieza: number | null;
  superficie: string;
  duracion_min: number;
  precio_base_centavos: number;
  honorario_centavos: number;
  laboratorio_centavos: number;
  descuento_centavos: number;
  descuento_autorizado_por: string;
  total_centavos: number;
  aprobado: boolean | null;
  estado: EstadoItem;
  cita_id: number | null;
}

export interface Presupuesto {
  id: number;
  numero: string;
  version: number;
  origen_id: number | null;
  paciente_id: number;
  profesional_nombre: string;
  estado: EstadoPresupuesto;
  notas: string;
  creado_en: string;
  presentado_en: string | null;
  presentado_por: string;
  resuelto_en: string | null;
  resuelto_por: string;
  reemplazado: boolean;
  moneda: string;
  honorarios_centavos: number;
  laboratorio_centavos: number;
  descuentos_centavos: number;
  total_centavos: number;
  aprobado_centavos: number;
  items: ItemPresupuesto[];
}

export interface ItemPedido {
  id?: number;
  fase: number;
  sesion: number;
  procedimiento_id: number;
  pieza?: number;
  superficie?: string;
  honorario_centavos?: number;
  laboratorio_centavos: number;
  descuento_centavos: number;
}

export interface DatosPresupuesto {
  notas: string;
  items: ItemPedido[];
  autorizacion?: { email: string; password: string };
}

export interface Tarea {
  item_id: number;
  presupuesto_id: number;
  presupuesto_numero: string;
  version: number;
  fase: number;
  sesion: number;
  procedimiento_id: number;
  descripcion: string;
  pieza: number | null;
  superficie: string;
  duracion_min: number;
  total_centavos: number;
  estado: EstadoItem;
  cita_id: number | null;
  cita_inicio: string | null;
}

const base = (pacienteId: number) => `/api/pacientes/${pacienteId}/presupuestos`;

export const listarPresupuestos = (pacienteId: number) => apiFetch<Presupuesto[]>(base(pacienteId));
export const listarTareas = (pacienteId: number) => apiFetch<Tarea[]>(`/api/pacientes/${pacienteId}/tareas`);
export const crearPresupuesto = (pacienteId: number, datos: DatosPresupuesto) =>
  apiFetch<Presupuesto>(base(pacienteId), { method: 'POST', body: JSON.stringify(datos) });
export const actualizarPresupuesto = (pacienteId: number, id: number, datos: DatosPresupuesto) =>
  apiFetch<Presupuesto>(`${base(pacienteId)}/${id}`, { method: 'PUT', body: JSON.stringify(datos) });
export const presentarPresupuesto = (pacienteId: number, id: number) =>
  apiFetch<Presupuesto>(`${base(pacienteId)}/${id}/presentar`, { method: 'POST' });
export const resolverPresupuesto = (pacienteId: number, id: number, aprobados: number[]) =>
  apiFetch<Presupuesto>(`${base(pacienteId)}/${id}/resolver`, { method: 'POST', body: JSON.stringify({ aprobados }) });
export const nuevaVersionPresupuesto = (pacienteId: number, id: number) =>
  apiFetch<Presupuesto>(`${base(pacienteId)}/${id}/nueva-version`, { method: 'POST' });
export const pdfPresupuesto = (pacienteId: number, id: number) => apiBlob(`${base(pacienteId)}/${id}/pdf`);
