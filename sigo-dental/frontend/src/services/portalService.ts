// Acceso a la API del portal del paciente (F2-8). El paciente se deduce siempre del token: nunca se envía su ID.
import { esRol, type Rol } from '../config/permisos';
import type { Sesion } from './authService';
import { apiBlob, apiFetch } from './http';
import type { EstadoCita } from './citaService';
import type { Consentimiento } from './consentimientoService';
import type { Receta } from './recetaService';

export const solicitarCodigo = (documento: string) =>
  apiFetch<{ mensaje: string }>('/api/portal/codigo', { method: 'POST', body: JSON.stringify({ documento }) });

export async function ingresarAlPortal(documento: string, codigo: string): Promise<Sesion> {
  const datos = await apiFetch<{ token?: string; email?: string; rol?: Rol }>('/api/portal/ingresar', {
    method: 'POST', body: JSON.stringify({ documento, codigo }),
  });
  if (!datos.token || !datos.email || !esRol(datos.rol)) throw new Error('Respuesta de ingreso inválida');
  return { token: datos.token, email: datos.email, rol: datos.rol };
}

export interface PerfilPortal {
  nombres: string;
  apellidos: string;
  numero_historia: string;
}

export interface CitaPortal {
  id: number;
  estado: EstadoCita;
  inicio: string;
  fin: string;
  profesional_nombre: string;
  consultorio_nombre: string;
  procedimiento_nombre: string;
  motivo: string;
}

export const obtenerPerfil = () => apiFetch<PerfilPortal>('/api/portal/yo');
export const listarMisCitas = () => apiFetch<CitaPortal[]>('/api/portal/citas');
export const confirmarMiCita = (id: number) => apiFetch<CitaPortal>(`/api/portal/citas/${id}/confirmar`, { method: 'POST' });
export const cancelarMiCita = (id: number, motivo: string) =>
  apiFetch<CitaPortal>(`/api/portal/citas/${id}/cancelar`, { method: 'POST', body: JSON.stringify({ motivo }) });
export const listarMisRecetas = () => apiFetch<Receta[]>('/api/portal/recetas');
export const pdfMiReceta = (id: number) => apiBlob(`/api/portal/recetas/${id}/pdf`);
export const listarMisConsentimientos = () => apiFetch<Consentimiento[]>('/api/portal/consentimientos');
export const pdfMiConsentimiento = (id: number) => apiBlob(`/api/portal/consentimientos/${id}/pdf`);
