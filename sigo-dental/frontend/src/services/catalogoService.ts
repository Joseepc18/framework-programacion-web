// Acceso a la API del catálogo médico. Los listados reciben `todos` para incluir lo desactivado
// (el servidor solo lo respeta para el administrador TI).
import { apiFetch } from './http';

export type RecursoSimple = 'consultorios' | 'convenios' | 'condiciones-criticas';

export interface ItemCatalogo {
  id: number;
  nombre: string;
  activo: boolean;
}

export interface Sillon {
  id: number;
  consultorio_id: number;
  consultorio_nombre: string;
  nombre: string;
  activo: boolean;
}

export interface Profesional {
  id: number;
  usuario_id: number;
  nombre: string;
  email: string;
  especialidad: string;
  registro_profesional: string;
  activo: boolean;
}

export interface UsuarioResumen {
  id: number;
  nombre: string;
  email: string;
}

export interface Procedimiento {
  id: number;
  codigo: string;
  nombre: string;
  especialidad: string;
  precio_centavos: number;
  duracion_min: number;
  requiere_consentimiento: boolean;
  activo: boolean;
}

export type DatosProcedimiento = Omit<Procedimiento, 'id'>;

export interface Parametro {
  clave: string;
  valor: string;
  descripcion: string;
}

const conTodos = (todos: boolean) => (todos ? '?todos=true' : '');
const enviar = (metodo: 'POST' | 'PUT', cuerpo: unknown): RequestInit => ({ method: metodo, body: JSON.stringify(cuerpo) });

export const listarSimples = (recurso: RecursoSimple, todos = false) =>
  apiFetch<ItemCatalogo[]>(`/api/catalogo/${recurso}${conTodos(todos)}`);
export const crearSimple = (recurso: RecursoSimple, nombre: string) =>
  apiFetch<ItemCatalogo>(`/api/catalogo/${recurso}`, enviar('POST', { nombre }));
export const actualizarSimple = (recurso: RecursoSimple, id: number, nombre: string, activo: boolean) =>
  apiFetch<ItemCatalogo>(`/api/catalogo/${recurso}/${id}`, enviar('PUT', { nombre, activo }));

export const listarSillones = (todos = false, consultorioId = 0) => {
  const params = new URLSearchParams();
  if (todos) params.set('todos', 'true');
  if (consultorioId) params.set('consultorio_id', String(consultorioId));
  const consulta = params.toString();
  return apiFetch<Sillon[]>(`/api/sillones${consulta ? `?${consulta}` : ''}`);
};
export const crearSillon = (consultorioId: number, nombre: string) =>
  apiFetch<Sillon>('/api/sillones', enviar('POST', { consultorio_id: consultorioId, nombre }));
export const actualizarSillon = (id: number, nombre: string, activo: boolean) =>
  apiFetch<Sillon>(`/api/sillones/${id}`, enviar('PUT', { nombre, activo }));

export const listarProfesionales = (todos = false) =>
  apiFetch<Profesional[]>(`/api/profesionales${conTodos(todos)}`);
export const listarCandidatos = () => apiFetch<UsuarioResumen[]>('/api/profesionales/candidatos');
export const crearProfesional = (usuarioId: number, especialidad: string, registro: string) =>
  apiFetch<Profesional>('/api/profesionales', enviar('POST', { usuario_id: usuarioId, especialidad, registro_profesional: registro }));
export const actualizarProfesional = (id: number, especialidad: string, registro: string, activo: boolean) =>
  apiFetch<Profesional>(`/api/profesionales/${id}`, enviar('PUT', { especialidad, registro_profesional: registro, activo }));

export const listarProcedimientos = (todos = false) =>
  apiFetch<Procedimiento[]>(`/api/procedimientos${conTodos(todos)}`);
export const crearProcedimiento = (datos: DatosProcedimiento) =>
  apiFetch<Procedimiento>('/api/procedimientos', enviar('POST', datos));
export const actualizarProcedimiento = (id: number, datos: DatosProcedimiento) =>
  apiFetch<Procedimiento>(`/api/procedimientos/${id}`, enviar('PUT', datos));

export const listarParametros = () => apiFetch<Parametro[]>('/api/parametros');
export const actualizarParametro = (clave: string, valor: string) =>
  apiFetch<Parametro>(`/api/parametros/${clave}`, enviar('PUT', { valor }));
