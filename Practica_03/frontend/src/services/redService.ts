// src/services/redService.ts
// Capa de servicios de la red multinivel.
// Consume GET /api/red del backend (Go + Fiber), que reconstruye el árbol desde PostgreSQL
// y lo sirve desde la caché de Redis.
import type { Referido } from "../data/red";

const API_URL = "http://localhost:3000/api";

// GET /api/red (el email identifica la clave de caché red:tree:{email} del usuario)
export const getRed = async (email?: string): Promise<Referido> => {
  const url = email ? `${API_URL}/red?email=${encodeURIComponent(email)}` : `${API_URL}/red`;
  const response = await fetch(url);
  if (!response.ok) throw new Error("No se pudo cargar la red multinivel");
  return response.json();
};
