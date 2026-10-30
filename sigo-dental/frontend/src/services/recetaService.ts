// Acceso a la API de recetas (F2-6).
import { apiBlob, apiFetch } from './http';
import type { Firma } from './evolucionService';

export interface Medicamento {
  nombre: string;
  dosis: string;
  frecuencia: string;
  duracion: string;
  indicaciones: string;
}

export const MEDICAMENTO_VACIO: Medicamento = { nombre: '', dosis: '', frecuencia: '', duracion: '', indicaciones: '' };

export interface AvisoAlergia {
  medicamento: string;
  alergia: string;
}

export type EstadoReceta = 'borrador' | 'firmada' | 'anulada';

export interface Receta {
  id: number;
  numero: string;
  paciente_id: number;
  profesional_nombre: string;
  indicaciones: string;
  creada_en: string;
  medicamentos: Medicamento[];
  estado: EstadoReceta;
  firma: Firma | null;
  anulacion: { motivo: string; usuario_nombre: string; reemplazo_id: number | null; creada_en: string } | null;
  avisos_alergia: AvisoAlergia[];
}

export interface DatosReceta {
  indicaciones: string;
  medicamentos: Medicamento[];
}

const base = (pacienteId: number) => `/api/pacientes/${pacienteId}/recetas`;

export const listarRecetas = (pacienteId: number) => apiFetch<Receta[]>(base(pacienteId));
export const crearReceta = (pacienteId: number, datos: DatosReceta) =>
  apiFetch<Receta>(base(pacienteId), { method: 'POST', body: JSON.stringify(datos) });
export const actualizarReceta = (pacienteId: number, id: number, datos: DatosReceta) =>
  apiFetch<Receta>(`${base(pacienteId)}/${id}`, { method: 'PUT', body: JSON.stringify(datos) });
export const firmarReceta = (pacienteId: number, id: number, password: string) =>
  apiFetch<Receta>(`${base(pacienteId)}/${id}/firmar`, { method: 'POST', body: JSON.stringify({ password }) });
export const anularReceta = (pacienteId: number, id: number, motivo: string, copiar: boolean) =>
  apiFetch<{ receta: Receta; reemplazo_id?: number }>(`${base(pacienteId)}/${id}/anular`, { method: 'POST', body: JSON.stringify({ motivo, copiar }) });
export const verificarAlergias = (pacienteId: number, medicamentos: string[]) =>
  apiFetch<{ avisos: AvisoAlergia[]; alergias: string }>(`${base(pacienteId)}/verificar-alergias`, { method: 'POST', body: JSON.stringify({ medicamentos }) });
export const pdfReceta = (pacienteId: number, id: number) => apiBlob(`${base(pacienteId)}/${id}/pdf`);
