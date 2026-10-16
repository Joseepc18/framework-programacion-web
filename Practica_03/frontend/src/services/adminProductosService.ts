// src/services/adminProductosService.ts
// Capa de servicio del catálogo administrativo (Fase 2 del back-office).
// Todas las operaciones viajan con el token JWT (apiFetch) y el backend exige rol admin.
import type { Producto } from "../data/productos";
import { apiFetch } from "./http";

export interface PaginaProductos {
  data: Producto[];
  page: number;
  limit: number;
  total: number;
}

// GET /api/admin/productos — paginado, incluye inactivos y acepta ?q= de búsqueda server-side
export const getProductosAdmin = async (
  q = "",
  page = 1,
  limit = 10
): Promise<PaginaProductos> => {
  const params = new URLSearchParams({ page: String(page), limit: String(limit) });
  if (q.trim() !== "") params.set("q", q.trim());
  return apiFetch<PaginaProductos>(`/api/admin/productos?${params.toString()}`);
};

// POST /api/productos — alta (solo admin)
export const crearProducto = async (
  producto: Omit<Producto, "id">
): Promise<Producto> => {
  return apiFetch<Producto>("/api/productos", {
    method: "POST",
    body: JSON.stringify(producto),
  });
};

// PUT /api/productos/:id — edición (solo admin)
export const actualizarProducto = async (
  id: number,
  producto: Omit<Producto, "id">
): Promise<Producto> => {
  return apiFetch<Producto>(`/api/productos/${id}`, {
    method: "PUT",
    body: JSON.stringify(producto),
  });
};

// DELETE /api/productos/:id — borrado lógico; ?activar=true lo reactiva
export const cambiarEstadoProducto = async (
  id: number,
  activo: boolean
): Promise<{ id: number; activo: boolean; message: string }> => {
  const sufijo = activo ? "?activar=true" : "";
  return apiFetch<{ id: number; activo: boolean; message: string }>(
    `/api/productos/${id}${sufijo}`,
    { method: "DELETE" }
  );
};
