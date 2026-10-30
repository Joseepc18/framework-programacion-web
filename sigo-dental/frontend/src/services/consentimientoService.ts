// Acceso a la API de consentimientos informados (F2-5) y a la bitácora de códigos enviados.
import { apiFetch } from './http';

export interface Consentimiento {
  id: number;
  paciente_id: number;
  procedimiento_id: number;
  procedimiento_nombre: string;
  profesional_nombre: string;
  tipo: 'consentimiento' | 'revocacion';
  revoca_a: number | null;
  motivo: string;
  titulo: string;
  texto: string;
  huella_texto: string;
  estado: 'pendiente' | 'firmado';
  firmante_tipo: 'paciente' | 'tutor';
  firmante_nombre: string;
  firmante_documento: string;
  modalidad: '' | 'manuscrita' | 'otp';
  dispositivo: string;
  firmado_en: string | null;
  archivo_id: number | null;
  creado_por: string;
  creado_en: string;
  vigente: boolean;
  revocado: boolean;
}

export interface Plantilla {
  id: number;
  procedimiento_id: number | null;
  procedimiento_nombre: string;
  titulo: string;
  cuerpo: string;
  activo: boolean;
}

export interface Punto { x: number; y: number; t: number }

export interface FirmaManuscrita {
  ancho: number;
  alto: number;
  trazos: Punto[][];
}

export interface NotificacionCodigo {
  id: number;
  tipo: string;
  canal: string;
  destino: string;
  mensaje: string;
  estado: string;
  detalle: string;
  creado_en: string;
}

const base = (pacienteId: number) => `/api/pacientes/${pacienteId}/consentimientos`;

export const listarConsentimientos = (pacienteId: number) => apiFetch<Consentimiento[]>(base(pacienteId));
export const crearConsentimiento = (pacienteId: number, procedimientoId: number, pieza?: number) =>
  apiFetch<Consentimiento>(base(pacienteId), { method: 'POST', body: JSON.stringify({ procedimiento_id: procedimientoId, pieza }) });
export const enviarCodigo = (pacienteId: number, id: number) =>
  apiFetch<{ enviado_a: string; canal: string; expira_en: string }>(`${base(pacienteId)}/${id}/otp`, { method: 'POST' });
export const firmarConConsentimientoOtp = (pacienteId: number, id: number, codigo: string) =>
  apiFetch<Consentimiento>(`${base(pacienteId)}/${id}/firmar`, {
    method: 'POST', body: JSON.stringify({ modalidad: 'otp', codigo, dispositivo: navigator.userAgent }),
  });
export const firmarManuscrito = (pacienteId: number, id: number, firma: FirmaManuscrita) =>
  apiFetch<Consentimiento>(`${base(pacienteId)}/${id}/firmar`, {
    method: 'POST',
    body: JSON.stringify({ modalidad: 'manuscrita', firma, dispositivo: `${navigator.userAgent} · pantalla ${window.screen.width}x${window.screen.height}` }),
  });
export const revocarConsentimiento = (pacienteId: number, id: number, motivo: string) =>
  apiFetch<Consentimiento>(`${base(pacienteId)}/${id}/revocar`, { method: 'POST', body: JSON.stringify({ motivo }) });
export const listarCodigosEnviados = (pacienteId: number) => apiFetch<NotificacionCodigo[]>(`/api/pacientes/${pacienteId}/notificaciones`);

export const listarPlantillas = () => apiFetch<Plantilla[]>('/api/consentimiento-plantillas');
export const guardarPlantilla = (id: number | null, datos: Omit<Plantilla, 'id' | 'procedimiento_nombre'>) =>
  apiFetch<Plantilla>(id ? `/api/consentimiento-plantillas/${id}` : '/api/consentimiento-plantillas', {
    method: id ? 'PUT' : 'POST', body: JSON.stringify(datos),
  });
