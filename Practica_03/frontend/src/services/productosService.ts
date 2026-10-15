// src/services/productosService.ts
// Capa de servicios del catálogo.
// Consume la API del backend (Go + Fiber), que lee los productos desde PostgreSQL.
// Los componentes no cambian: siguen recibiendo Promise<Producto[]> y Promise<Producto | undefined>.

import type { Producto } from "../data/productos";

const API_URL = "http://localhost:3000/api";

// GET /api/productos (lista completa)
export const getProductos = async (): Promise<Producto[]> => {
  const response = await fetch(`${API_URL}/productos`);
  if (!response.ok) throw new Error("No se pudo cargar el catálogo");
  return response.json();
};

// GET /api/productos/:id (detalle de un producto)
// Un 404 significa que el producto no existe: devolvemos undefined, como hacía el mock.
export const getProductoById = async (id: number): Promise<Producto | undefined> => {
  const response = await fetch(`${API_URL}/productos/${id}`);
  if (response.status === 404) return undefined;
  if (!response.ok) throw new Error("No se pudo cargar el producto");
  return response.json();
};
