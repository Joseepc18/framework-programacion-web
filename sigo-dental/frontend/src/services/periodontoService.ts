// Acceso a la API del periodontograma (F2-3). El NIC lo calcula el servidor: aquí solo se muestra.
import { apiFetch } from './http';

export const SITIOS = ['DV', 'V', 'MV', 'DL', 'L', 'ML'] as const;
export type CodigoSitio = (typeof SITIOS)[number];

export interface Sitio {
  sitio: CodigoSitio;
  profundidad: number;
  margen: number;
  sangrado: boolean;
  nic: number;
}

export interface DientePerio {
  pieza: number;
  movilidad: number;
  furca: number;
  sitios: Sitio[];
}

export interface ExamenPerio {
  id: number;
  paciente_id: number;
  fecha: string;
  notas: string;
  creado_por: string;
  creado_en: string;
  dientes: DientePerio[];
}

export interface ResumenExamen {
  id: number;
  fecha: string;
  creado_por: string;
  dientes: number;
  profundidad_media: number;
  nic_medio: number;
  sangrado_porcentaje: number;
}

export interface ListaPerio {
  examenes: ResumenExamen[];
  ausentes: number[];
}

export interface DatosExamen {
  fecha: string;
  notas: string;
  dientes: { pieza: number; movilidad: number; furca: number; sitios: Omit<Sitio, 'nic'>[] }[];
}

const base = (pacienteId: number) => `/api/pacientes/${pacienteId}/periodontogramas`;

export const listarPeriodontogramas = (pacienteId: number) => apiFetch<ListaPerio>(base(pacienteId));
export const obtenerPeriodontograma = (pacienteId: number, id: number) => apiFetch<ExamenPerio>(`${base(pacienteId)}/${id}`);
export const crearPeriodontograma = (pacienteId: number, datos: DatosExamen) =>
  apiFetch<ExamenPerio>(base(pacienteId), { method: 'POST', body: JSON.stringify(datos) });

export const media = (valores: number[]) => (valores.length === 0 ? 0 : valores.reduce((a, b) => a + b, 0) / valores.length);
