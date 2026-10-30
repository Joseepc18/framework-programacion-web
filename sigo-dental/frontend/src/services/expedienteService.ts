// Acceso a la API del expediente: signos vitales, archivos del paciente y consulta de la auditoría (F2-1).
import { apiBlob, apiFetch } from './http';

export interface SignosVitales {
  id: number;
  paciente_id: number;
  cita_id: number | null;
  sistolica: number | null;
  diastolica: number | null;
  frecuencia_cardiaca: number | null;
  frecuencia_respiratoria: number | null;
  temperatura: number | null;
  saturacion: number | null;
  peso_kg: number | null;
  talla_cm: number | null;
  observaciones: string;
  registrado_por: string;
  registrado_en: string;
}

export type DatosSignos = Partial<Omit<SignosVitales, 'id' | 'paciente_id' | 'registrado_por' | 'registrado_en'>>;

export const listarSignos = (pacienteId: number) => apiFetch<SignosVitales[]>(`/api/pacientes/${pacienteId}/signos-vitales`);

export const registrarSignos = (pacienteId: number, datos: DatosSignos) =>
  apiFetch<SignosVitales>(`/api/pacientes/${pacienteId}/signos-vitales`, { method: 'POST', body: JSON.stringify(datos) });

export type CategoriaArchivo = 'radiografia' | 'fotografia' | 'documento' | 'consentimiento' | 'receta' | 'presupuesto';

export const ETIQUETA_CATEGORIA: Record<CategoriaArchivo, string> = {
  radiografia: 'Radiografía',
  fotografia: 'Fotografía',
  documento: 'Documento',
  consentimiento: 'Consentimiento',
  receta: 'Receta',
  presupuesto: 'Presupuesto',
};

export interface Archivo {
  id: number;
  paciente_id: number;
  categoria: CategoriaArchivo;
  descripcion: string;
  nombre_original: string;
  tipo_mime: string;
  tamano_bytes: number;
  sha256: string;
  subido_por: string;
  subido_en: string;
}

export const listarArchivos = (pacienteId: number) => apiFetch<Archivo[]>(`/api/pacientes/${pacienteId}/archivos`);

export function subirArchivo(pacienteId: number, categoria: string, descripcion: string, archivo: File) {
  const cuerpo = new FormData();
  cuerpo.set('categoria', categoria);
  cuerpo.set('descripcion', descripcion);
  cuerpo.set('archivo', archivo);
  return apiFetch<Archivo>(`/api/pacientes/${pacienteId}/archivos`, { method: 'POST', body: cuerpo });
}

export const descargarArchivo = (id: number) => apiBlob(`/api/archivos/${id}/descargar`);

export type AccionAuditoria = 'consulta' | 'creacion' | 'modificacion' | 'exportacion' | 'firma' | 'anulacion';

export const ETIQUETA_ACCION: Record<AccionAuditoria, string> = {
  consulta: 'Consulta',
  creacion: 'Creación',
  modificacion: 'Modificación',
  exportacion: 'Exportación',
  firma: 'Firma',
  anulacion: 'Anulación',
};

export interface RegistroAuditoria {
  id: number;
  ocurrido_en: string;
  usuario_email: string;
  rol: string;
  accion: AccionAuditoria;
  entidad: string;
  entidad_id: number | null;
  paciente_id: number | null;
  paciente_nombre: string;
  ip: string;
  metodo: string;
  ruta: string;
}

export interface PaginaAuditoria {
  total: number;
  pagina: number;
  limite: number;
  registros: RegistroAuditoria[];
}

export interface FiltroAuditoria {
  usuario: string;
  paciente: string;
  accion: string;
  desde: string; // AAAA-MM-DD en la hora local, o vacío
  hasta: string;
  pagina: number;
}

export const LIMITE_AUDITORIA = 25;

// Las fechas del filtro son días locales: "hasta" incluye el día completo (se envía el inicio del día siguiente).
export function consultarAuditoria(filtro: FiltroAuditoria) {
  const params = new URLSearchParams({ limite: String(LIMITE_AUDITORIA), pagina: String(filtro.pagina) });
  if (filtro.usuario.trim()) params.set('usuario', filtro.usuario.trim());
  if (filtro.paciente.trim()) params.set('paciente', filtro.paciente.trim());
  if (filtro.accion) params.set('accion', filtro.accion);
  if (filtro.desde) params.set('desde', new Date(`${filtro.desde}T00:00:00`).toISOString());
  if (filtro.hasta) {
    const fin = new Date(`${filtro.hasta}T00:00:00`);
    fin.setDate(fin.getDate() + 1);
    params.set('hasta', fin.toISOString());
  }
  return apiFetch<PaginaAuditoria>(`/api/auditoria?${params}`);
}
