// src/services/productosService.ts
// Capa de servicios del catálogo.
// Consume la API del backend (Go + Fiber), que lee los productos desde PostgreSQL.
// Los componentes no cambian: siguen recibiendo Promise<Producto[]> y Promise<Producto | undefined>.

import type { Producto } from "../data/productos";
import { apiFetch, ApiError } from "./http";

// GET /api/productos (lista completa)
export const getProductos = async (): Promise<Producto[]> => {
  return apiFetch<Producto[]>("/api/productos");
};

// GET /api/productos/:id (detalle de un producto)
// Un 404 significa que el producto no existe: devolvemos undefined.
export const getProductoById = async (id: number): Promise<Producto | undefined> => {
  try {
    return await apiFetch<Producto>(`/api/productos/${id}`);
  } catch (error) {
    if (error instanceof ApiError && error.status === 404) {
      return undefined;
    }
    throw error;
  }
};
