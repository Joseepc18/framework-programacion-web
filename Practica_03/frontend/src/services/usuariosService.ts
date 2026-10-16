// src/services/usuariosService.ts
// Capa de servicio de administración de cuentas (Fase 7, solo admin).
import { apiFetch } from "./http";

export interface UsuarioAdmin {
  id: number;
  email: string;
  nombre: string;
  rol: "admin" | "cliente";
  activo: boolean;
  referido_id?: number;
  creado_en: string;
}

export interface PaginaUsuarios {
  data: UsuarioAdmin[];
  page: number;
  limit: number;
  total: number;
}

// GET /api/usuarios — listado paginado con búsqueda por nombre o correo
export const getUsuarios = async (
  q = "",
  page = 1,
  limit = 10
): Promise<PaginaUsuarios> => {
  const params = new URLSearchParams({ page: String(page), limit: String(limit) });
  if (q.trim() !== "") params.set("q", q.trim());
  return apiFetch<PaginaUsuarios>(`/api/usuarios?${params.toString()}`);
};

// PATCH /api/usuarios/:id — cambia rol y/o estado activo
export const actualizarUsuario = async (
  id: number,
  cambios: { rol?: "admin" | "cliente"; activo?: boolean }
): Promise<UsuarioAdmin> => {
  return apiFetch<UsuarioAdmin>(`/api/usuarios/${id}`, {
    method: "PATCH",
    body: JSON.stringify(cambios),
  });
};
