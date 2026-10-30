// Acceso a la API de pacientes y anamnesis. Las fechas viajan como AAAA-MM-DD.
import { apiFetch } from './http';

export type Sexo = 'femenino' | 'masculino' | 'otro';
export type Canal = 'whatsapp' | 'sms' | 'email';

export const ETIQUETA_CANAL: Record<Canal, string> = { whatsapp: 'WhatsApp', sms: 'SMS', email: 'Correo electrónico' };

export interface DatosPaciente {
  documento: string;
  nombres: string;
  apellidos: string;
  fecha_nacimiento: string;
  sexo: Sexo;
  telefono: string;
  email: string;
  direccion: string;
  emergencia_nombre: string;
  emergencia_telefono: string;
  emergencia_parentesco: string;
  tutor_nombre: string;
  tutor_documento: string;
  tutor_telefono: string;
  convenio_id: number | null;
  canal_preferido: Canal;
  activo: boolean;
}

export interface Paciente extends DatosPaciente {
  id: number;
  numero_historia: string;
  edad: number;
  convenio_nombre: string;
}

export interface Anamnesis {
  enfermedades_sistemicas: string;
  alergias_farmacos: string;
  antecedentes_quirurgicos: string;
  coagulopatias: string;
  condiciones_criticas: number[];
}

export const ANAMNESIS_VACIA: Anamnesis = {
  enfermedades_sistemicas: '',
  alergias_farmacos: '',
  antecedentes_quirurgicos: '',
  coagulopatias: '',
  condiciones_criticas: [],
};

export interface PacienteDetalle extends Paciente {
  anamnesis: Anamnesis;
  alertas: string[];
}

export interface PacienteResumen {
  id: number;
  numero_historia: string;
  documento: string;
  nombres: string;
  apellidos: string;
  edad: number;
  telefono: string;
  convenio_nombre: string;
  activo: boolean;
  alertas: string[];
}

export const nombreCompleto = (p: { nombres: string; apellidos: string }) => `${p.nombres} ${p.apellidos}`;

export function listarPacientes(busqueda: string, todos = false) {
  const params = new URLSearchParams();
  if (busqueda.trim()) params.set('q', busqueda.trim());
  if (todos) params.set('todos', 'true');
  const consulta = params.toString();
  return apiFetch<PacienteResumen[]>(`/api/pacientes${consulta ? `?${consulta}` : ''}`);
}

export const obtenerPaciente = (id: number) => apiFetch<PacienteDetalle>(`/api/pacientes/${id}`);

export const crearPaciente = (datos: DatosPaciente, anamnesis: Anamnesis) =>
  apiFetch<PacienteDetalle>('/api/pacientes', { method: 'POST', body: JSON.stringify({ ...datos, anamnesis }) });

export const actualizarPaciente = (id: number, datos: DatosPaciente) =>
  apiFetch<PacienteDetalle>(`/api/pacientes/${id}`, { method: 'PUT', body: JSON.stringify(datos) });

export const actualizarAnamnesis = (id: number, anamnesis: Anamnesis) =>
  apiFetch<PacienteDetalle>(`/api/pacientes/${id}/anamnesis`, { method: 'PUT', body: JSON.stringify(anamnesis) });
